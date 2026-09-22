package mosaic

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

// budgetSumTolerance is how far sub-phase budgets may fall from their parent
// before it is worth reporting. Half a cent: the values are dollar amounts, so
// anything smaller is float representation rather than a real difference.
const budgetSumTolerance = 0.005

// PhaseBudgetReport contains per-phase budget metrics for a project.
type PhaseBudgetReport struct {
	// ProjectID is the project this report is for.
	ProjectID int `json:"project_id"`

	// Phases contains per-phase budget summaries.
	Phases []PhaseBudgetSummary `json:"phases"`

	// Totals contains the aggregated metrics across all phases.
	Totals BudgetMetrics `json:"totals"`

	// Warnings contains human-readable warnings generated during report
	// computation (e.g. missing rate data for members).
	Warnings []string `json:"warnings,omitempty"`
}

// PhaseBudgetSummary contains budget metrics for a single phase.
type PhaseBudgetSummary struct {
	// PhaseID is the Mosaic ID of the phase.
	PhaseID int `json:"phase_id"`

	// PhaseName is the display name of the phase.
	PhaseName string `json:"phase_name"`

	// PhaseNumber is the phase number (e.g. "#01"), if set.
	PhaseNumber *string `json:"phase_number,omitempty"`

	// BudgetStatus is the budget status (e.g. "active").
	BudgetStatus string `json:"budget_status,omitempty"`

	// Metrics contains the budget/spent/planned/remaining breakdown.
	Metrics BudgetMetrics `json:"metrics"`

	// Tasks contains per-sub-phase (task) budget summaries, if this phase
	// has sub-phases. Nil when there are no sub-phases.
	Tasks []TaskBudgetSummary `json:"tasks,omitempty"`
}

// TaskBudgetSummary contains budget metrics for a single sub-phase (task)
// within a parent phase.
type TaskBudgetSummary struct {
	TaskID     int           `json:"task_id"`
	TaskName   string        `json:"task_name"`
	TaskNumber string        `json:"task_number"`
	Metrics    BudgetMetrics `json:"metrics"`
}

// BudgetMetrics contains dollar and hour breakdowns for budget tracking.
type BudgetMetrics struct {
	// BudgetDollars is the budgeted dollar amount.
	BudgetDollars float64 `json:"budget_dollars"`

	// BudgetHours is the budgeted hours.
	BudgetHours float64 `json:"budget_hours"`

	// SpentDollars is the total dollars spent (hours * rate from time entries).
	SpentDollars float64 `json:"spent_dollars"`

	// SpentHours is the total hours logged in time entries.
	SpentHours float64 `json:"spent_hours"`

	// PlannedDollars is the total dollars from future work plans (hours * rate).
	PlannedDollars float64 `json:"planned_dollars"`

	// PlannedHours is the total hours from future work plans.
	PlannedHours float64 `json:"planned_hours"`

	// RemainingDollars is budget minus spent minus planned dollars.
	RemainingDollars float64 `json:"remaining_dollars"`

	// RemainingHours is budget minus spent minus planned hours.
	RemainingHours float64 `json:"remaining_hours"`

	// PercentSpent is (spent / budget) * 100, or 0 if budget is zero.
	PercentSpent float64 `json:"percent_spent"`
}

// ComputePhaseBudget computes per-phase budget/spent/planned/remaining metrics
// from pre-fetched data. The memberRates map provides effective bill rates
// keyed by member ID, used to compute planned dollar amounts from work plans.
// The today parameter determines which work plans count as "future" (planned).
//
// Phases are organized hierarchically: sub-phases (those with a ParentID) have
// their spent and planned values rolled up into their parent phase. Only
// top-level phases appear in the report output. Budget values come only from
// top-level phases to avoid double-counting.
func ComputePhaseBudget(projectID int, phases []Phase, entries []TimeEntry, plans []WorkPlan, memberRates map[int]float64, today time.Time) *PhaseBudgetReport {
	todayStr := today.Format("2006-01-02")

	// Build child-to-parent mapping, parent-to-children mapping, and phase lookup.
	childToParent := make(map[int]int)      // childID -> parentID
	parentToChildren := make(map[int][]int) // parentID -> []childID
	phaseByID := make(map[int]*Phase, len(phases))
	for i := range phases {
		p := &phases[i]
		phaseByID[p.MosaicID] = p
		if p.ParentID != nil {
			childToParent[p.MosaicID] = *p.ParentID
			parentToChildren[*p.ParentID] = append(parentToChildren[*p.ParentID], p.MosaicID)
		}
	}

	// resolveParent walks up the hierarchy to find the top-level phase ID.
	resolveParent := func(phaseID int) int {
		for {
			parent, ok := childToParent[phaseID]
			if !ok {
				return phaseID
			}
			phaseID = parent
		}
	}

	// Index top-level phases by MosaicID, preserving order.
	type accumulator struct {
		summary PhaseBudgetSummary
	}
	phaseMap := make(map[int]*accumulator, len(phases))
	var phaseOrder []int

	for _, p := range phases {
		if p.ParentID != nil {
			continue // skip sub-phases
		}
		phaseMap[p.MosaicID] = &accumulator{
			summary: PhaseBudgetSummary{
				PhaseID:      p.MosaicID,
				PhaseName:    p.Name,
				PhaseNumber:  p.PhaseNumber,
				BudgetStatus: p.BudgetStatus,
				Metrics: BudgetMetrics{
					BudgetDollars: parseOptionalFloat(p.Total),
					BudgetHours:   parseOptionalFloat(p.EstimatedHours),
				},
			},
		}
		phaseOrder = append(phaseOrder, p.MosaicID)
	}

	// ensurePhase creates a top-level entry for phases not in the original list
	// (e.g. time entries referencing unknown phases).
	ensurePhase := func(phaseID int, name string) *accumulator {
		acc, ok := phaseMap[phaseID]
		if !ok {
			acc = &accumulator{
				summary: PhaseBudgetSummary{
					PhaseID:   phaseID,
					PhaseName: name,
				},
			}
			phaseMap[phaseID] = acc
			phaseOrder = append(phaseOrder, phaseID)
		}
		return acc
	}

	// Sub-phase (task) accumulators: subPhaseID -> metrics.
	type subAccumulator struct {
		spentDollars   float64
		spentHours     float64
		plannedDollars float64
		plannedHours   float64
	}
	subMap := make(map[int]*subAccumulator)

	// ensureSub returns or creates the sub-phase accumulator.
	ensureSub := func(subPhaseID int) *subAccumulator {
		sa, ok := subMap[subPhaseID]
		if !ok {
			sa = &subAccumulator{}
			subMap[subPhaseID] = sa
		}
		return sa
	}

	// Accumulate spent from time entries, rolling up to parent phases.
	// Also track per-sub-phase spent when the entry is on a sub-phase.
	for _, e := range entries {
		parentID := resolveParent(e.PhaseID)
		acc := ensurePhase(parentID, e.PhaseName)
		hours := parseStringFloat(e.Hours)
		rate := parseOptionalFloat(e.Rate)
		acc.summary.Metrics.SpentHours += hours
		acc.summary.Metrics.SpentDollars += hours * rate

		// Track sub-phase level if this entry is on a child phase.
		if _, isChild := childToParent[e.PhaseID]; isChild {
			sa := ensureSub(e.PhaseID)
			sa.spentHours += hours
			sa.spentDollars += hours * rate
		}
	}

	// Accumulate planned from future work plans, rolling up to parent phases.
	// Also track per-sub-phase planned when the plan is on a sub-phase.
	for _, wp := range plans {
		if wp.StartDate < todayStr {
			continue
		}
		parentID := resolveParent(wp.PhaseID)
		acc := ensurePhase(parentID, wp.PhaseName)
		hours := parseStringFloat(wp.TotalHours)
		rate := memberRates[wp.MemberID]
		acc.summary.Metrics.PlannedHours += hours
		acc.summary.Metrics.PlannedDollars += hours * rate

		// Track sub-phase level if this plan is on a child phase.
		if _, isChild := childToParent[wp.PhaseID]; isChild {
			sa := ensureSub(wp.PhaseID)
			sa.plannedHours += hours
			sa.plannedDollars += hours * rate
		}
	}

	var warnings []string

	// Compute remaining and percentages, build result.
	report := &PhaseBudgetReport{
		ProjectID: projectID,
		Phases:    make([]PhaseBudgetSummary, 0, len(phaseOrder)),
	}

	for _, pid := range phaseOrder {
		acc := phaseMap[pid]
		m := &acc.summary.Metrics
		m.RemainingDollars = m.BudgetDollars - m.SpentDollars - m.PlannedDollars
		m.RemainingHours = m.BudgetHours - m.SpentHours - m.PlannedHours
		if m.BudgetDollars > 0 {
			m.PercentSpent = m.SpentDollars / m.BudgetDollars * 100
		}

		// Build Tasks slice for phases that have children.
		if children, ok := parentToChildren[pid]; ok && len(children) > 0 {
			var taskBudgetSum float64
			tasks := make([]TaskBudgetSummary, 0, len(children))
			for _, childID := range children {
				child := phaseByID[childID]
				taskNum := ""
				if child.PhaseNumber != nil {
					taskNum = strings.TrimPrefix(*child.PhaseNumber, "#")
				}
				budget := parseOptionalFloat(child.Total)
				taskBudgetSum += budget

				tm := BudgetMetrics{
					BudgetDollars: budget,
					BudgetHours:   parseOptionalFloat(child.EstimatedHours),
				}
				if sa, ok := subMap[childID]; ok {
					tm.SpentDollars = sa.spentDollars
					tm.SpentHours = sa.spentHours
					tm.PlannedDollars = sa.plannedDollars
					tm.PlannedHours = sa.plannedHours
				}
				tm.RemainingDollars = tm.BudgetDollars - tm.SpentDollars - tm.PlannedDollars
				tm.RemainingHours = tm.BudgetHours - tm.SpentHours - tm.PlannedHours
				if tm.BudgetDollars > 0 {
					tm.PercentSpent = tm.SpentDollars / tm.BudgetDollars * 100
				}

				tasks = append(tasks, TaskBudgetSummary{
					TaskID:     childID,
					TaskName:   child.Name,
					TaskNumber: taskNum,
					Metrics:    tm,
				})
			}
			acc.summary.Tasks = tasks

			// Warn if sub-phase budgets don't sum to parent budget.
			//
			// Compared with a tolerance, not exactly: taskBudgetSum is an
			// accumulation of floats parsed from strings, so two amounts that
			// are equal to the cent routinely differ in the last bits. An
			// exact comparison reported "sum to $99096.14 but parent budget is
			// $99096.14", which tells a reader nothing and buries the real
			// mismatches.
			//
			// Sub-phases that carry no budget at all are a different shape, not
			// a discrepancy: the fee was entered on the parent and the work was
			// broken out beneath it without splitting the money. Warning there
			// would say "sum to $0.00 but parent budget is $100000.00" on every
			// such project, with nothing behind it to act on.
			if m.BudgetDollars > 0 && taskBudgetSum > 0 && math.Abs(taskBudgetSum-m.BudgetDollars) > budgetSumTolerance {
				phaseName := acc.summary.PhaseName
				if acc.summary.PhaseNumber != nil {
					phaseName = *acc.summary.PhaseNumber + " " + phaseName
				}
				warnings = append(warnings, fmt.Sprintf(
					"phase %s: sub-phase budgets sum to $%.2f but parent budget is $%.2f",
					phaseName, taskBudgetSum, m.BudgetDollars,
				))
			}
		}

		report.Phases = append(report.Phases, acc.summary)

		report.Totals.BudgetDollars += m.BudgetDollars
		report.Totals.BudgetHours += m.BudgetHours
		report.Totals.SpentDollars += m.SpentDollars
		report.Totals.SpentHours += m.SpentHours
		report.Totals.PlannedDollars += m.PlannedDollars
		report.Totals.PlannedHours += m.PlannedHours
	}

	t := &report.Totals
	t.RemainingDollars = t.BudgetDollars - t.SpentDollars - t.PlannedDollars
	t.RemainingHours = t.BudgetHours - t.SpentHours - t.PlannedHours
	if t.BudgetDollars > 0 {
		t.PercentSpent = t.SpentDollars / t.BudgetDollars * 100
	}

	report.Warnings = append(report.Warnings, warnings...)

	return report
}

// GetPhaseBudgetReport fetches all required data from the Mosaic API and
// computes per-phase budget/spent/planned/remaining metrics for a project.
func (c *Client) GetPhaseBudgetReport(ctx context.Context, projectID int) (*PhaseBudgetReport, error) {
	phases, err := c.ListPhasesByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("fetching phases: %w", err)
	}
	slog.Info("fetched phases", "count", len(phases))

	allTrue := true
	entries, err := c.ListTimeEntries(ctx, TimeEntryFilter{
		ListParams: ListParams{All: &allTrue},
		ProjectIDs: []int{projectID},
	})
	if err != nil {
		return nil, fmt.Errorf("fetching time entries: %w", err)
	}
	slog.Info("fetched time entries", "count", len(entries))

	plans, err := c.ListWorkPlans(ctx, WorkPlanFilter{
		ProjectIDs: []int{projectID},
		All:        true,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching work plans: %w", err)
	}
	slog.Info("fetched work plans", "count", len(plans))

	// Collect unique member IDs and names from future work plans for rate resolution.
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	memberNames := make(map[int]string)
	for _, wp := range plans {
		if wp.StartDate >= todayStr {
			if _, ok := memberNames[wp.MemberID]; !ok {
				memberNames[wp.MemberID] = wp.MemberName
			}
		}
	}
	memberIDs := make([]int, 0, len(memberNames))
	for id := range memberNames {
		memberIDs = append(memberIDs, id)
	}

	memberRates, rateWarnings, err := c.resolveMemberRates(ctx, projectID, memberIDs, memberNames, today)
	if err != nil {
		return nil, fmt.Errorf("resolving member rates: %w", err)
	}

	report := ComputePhaseBudget(projectID, phases, entries, plans, memberRates, today)
	// Append: ComputePhaseBudget has already recorded its own warnings, such as
	// sub-phase budgets that do not sum to their parent. Assigning here dropped
	// those, so callers only ever saw rate warnings.
	report.Warnings = append(report.Warnings, rateWarnings...)
	return report, nil
}

// resolveMemberRates builds a map of memberID -> effective bill rate for the
// given project. It fetches the rate table and member-project-rates, preferring
// project-specific rates over global bill rates. The rate amount comes from
// the Rate table via rate_id, not from the MemberProjectRate directly.
func (c *Client) resolveMemberRates(ctx context.Context, projectID int, memberIDs []int, memberNames map[int]string, today time.Time) (map[int]float64, []string, error) {
	if len(memberIDs) == 0 {
		return make(map[int]float64), nil, nil
	}

	var warnings []string

	// Fetch all rates to build rateID -> amount lookup.
	rates, err := c.ListRates(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching rates: %w", err)
	}
	rateAmounts := make(map[int]float64, len(rates))
	for _, r := range rates {
		amount, _ := strconv.ParseFloat(r.Rate, 64)
		rateAmounts[r.MosaicID] = amount
	}
	slog.Info("fetched rates for lookup", "count", len(rates))

	// Fetch bill rates as fallback.
	billRates, err := c.ListBillRates(ctx, BillRateFilter{})
	if err != nil {
		return nil, nil, fmt.Errorf("fetching bill rates: %w", err)
	}
	todayStr := today.Format("2006-01-02")
	fallbackRates := make(map[int]float64)
	for _, br := range billRates {
		if br.StartDate != "" && br.StartDate > todayStr {
			continue
		}
		if br.EndDate != "" && br.EndDate < todayStr {
			continue
		}
		fallbackRates[br.MemberID] = br.RateAmount
	}
	slog.Info("fetched bill rates for fallback", "count", len(billRates))

	result := make(map[int]float64, len(memberIDs))
	for _, mid := range memberIDs {
		mpRates, err := c.ListMemberProjectRates(ctx, MemberProjectRateFilter{
			MemberID:  mid,
			ProjectID: projectID,
		})
		if err != nil {
			slog.Warn("failed to fetch member project rates, using fallback", "member_id", mid, "error", err)
			if fallback, ok := fallbackRates[mid]; ok {
				warnings = append(warnings, rateLookupFallbackWarning(memberDisplayName(memberNames, mid)))
				result[mid] = fallback
			} else {
				warnings = append(warnings, rateLookupNoFallbackWarning(memberDisplayName(memberNames, mid)))
			}
			continue
		}

		// Find the best matching non-cost-rate for today.
		// Prefer project-level (PhaseID == 0) rate; the phase-specific rate
		// would need per-work-plan resolution which we handle below.
		var bestRateID int
		var bestStart string
		for _, mpr := range mpRates {
			if mpr.IsCostRate != nil && *mpr.IsCostRate {
				continue
			}
			if mpr.PhaseID != 0 {
				continue // skip phase-specific for now
			}
			if mpr.StartDate != "" && mpr.StartDate > todayStr {
				continue
			}
			if mpr.EndDate != "" && mpr.EndDate < todayStr {
				continue
			}
			if mpr.StartDate >= bestStart {
				bestRateID = mpr.RateID
				bestStart = mpr.StartDate
			}
		}

		if bestRateID > 0 {
			result[mid] = rateAmounts[bestRateID]
		} else if fallback, ok := fallbackRates[mid]; ok {
			result[mid] = fallback
		} else {
			slog.Warn("no rate found for member", "member_id", mid)
			warnings = append(warnings, noBillRateWarning(memberDisplayName(memberNames, mid)))
		}
	}

	return result, warnings, nil
}

// memberDisplayName returns a human-readable label for a member, preferring
// the name from memberNames and falling back to the numeric ID.
func memberDisplayName(memberNames map[int]string, memberID int) string {
	if name, ok := memberNames[memberID]; ok && name != "" {
		return name
	}
	return fmt.Sprintf("Member %d", memberID)
}

// parseOptionalFloat parses a *string to float64, returning 0 if nil or unparseable.
func parseOptionalFloat(s *string) float64 {
	if s == nil {
		return 0
	}
	v, _ := strconv.ParseFloat(*s, 64)
	return v
}

// parseStringFloat parses a string to float64, returning 0 if empty or unparseable.
func parseStringFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
