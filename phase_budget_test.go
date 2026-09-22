package mosaic

import (
	"strings"
	"testing"
	"time"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func TestComputePhaseBudget_WithSubPhases(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Design", PhaseNumber: strPtr("#08"), Total: strPtr("150000")},
		{MosaicID: 201, Name: "Sub A", PhaseNumber: strPtr("01"), ParentID: intPtr(100), Total: strPtr("90000")},
		{MosaicID: 202, Name: "Sub B", PhaseNumber: strPtr("02"), ParentID: intPtr(100), Total: strPtr("60000")},
	}
	entries := []TimeEntry{
		{PhaseID: 201, Hours: "10", Rate: strPtr("100"), PhaseName: "Sub A"},
		{PhaseID: 202, Hours: "5", Rate: strPtr("120"), PhaseName: "Sub B"},
	}
	var plans []WorkPlan
	memberRates := map[int]float64{}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, entries, plans, memberRates, today)

	if len(report.Phases) != 1 {
		t.Fatalf("expected 1 top-level phase, got %d", len(report.Phases))
	}
	phase := report.Phases[0]

	// Parent budget and rollup spent
	if phase.Metrics.BudgetDollars != 150000 {
		t.Errorf("parent budget: want 150000, got %f", phase.Metrics.BudgetDollars)
	}
	if phase.Metrics.SpentDollars != 1600 {
		t.Errorf("parent spent$: want 1600, got %f", phase.Metrics.SpentDollars)
	}

	// Tasks
	if len(phase.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(phase.Tasks))
	}

	task01 := phase.Tasks[0]
	if task01.TaskID != 201 {
		t.Errorf("task01 TaskID: want 201, got %d", task01.TaskID)
	}
	if task01.TaskNumber != "01" {
		t.Errorf("task01 TaskNumber: want '01', got %q", task01.TaskNumber)
	}
	if task01.Metrics.BudgetDollars != 90000 {
		t.Errorf("task01 budget: want 90000, got %f", task01.Metrics.BudgetDollars)
	}
	if task01.Metrics.SpentDollars != 1000 {
		t.Errorf("task01 spent$: want 1000, got %f", task01.Metrics.SpentDollars)
	}
	if task01.Metrics.SpentHours != 10 {
		t.Errorf("task01 spent hours: want 10, got %f", task01.Metrics.SpentHours)
	}

	task02 := phase.Tasks[1]
	if task02.TaskID != 202 {
		t.Errorf("task02 TaskID: want 202, got %d", task02.TaskID)
	}
	if task02.Metrics.BudgetDollars != 60000 {
		t.Errorf("task02 budget: want 60000, got %f", task02.Metrics.BudgetDollars)
	}
	if task02.Metrics.SpentDollars != 600 {
		t.Errorf("task02 spent$: want 600, got %f", task02.Metrics.SpentDollars)
	}
	if task02.Metrics.SpentHours != 5 {
		t.Errorf("task02 spent hours: want 5, got %f", task02.Metrics.SpentHours)
	}
}

func TestComputePhaseBudget_NoSubPhases(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Design", PhaseNumber: strPtr("#01"), Total: strPtr("50000")},
	}
	entries := []TimeEntry{
		{PhaseID: 100, Hours: "10", Rate: strPtr("100"), PhaseName: "Design"},
	}
	var plans []WorkPlan
	memberRates := map[int]float64{}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, entries, plans, memberRates, today)

	if len(report.Phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(report.Phases))
	}
	phase := report.Phases[0]

	if phase.Tasks != nil {
		t.Errorf("expected Tasks to be nil for phase without sub-phases, got %v", phase.Tasks)
	}
	if phase.Metrics.SpentDollars != 1000 {
		t.Errorf("spent$: want 1000, got %f", phase.Metrics.SpentDollars)
	}
}

func TestComputePhaseBudget_SubPhaseBudgetWarning(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Design", PhaseNumber: strPtr("#01"), Total: strPtr("100000")},
		{MosaicID: 201, Name: "Sub A", PhaseNumber: strPtr("01"), ParentID: intPtr(100), Total: strPtr("60000")},
	}
	var entries []TimeEntry
	var plans []WorkPlan
	memberRates := map[int]float64{}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, entries, plans, memberRates, today)

	if len(report.Warnings) == 0 {
		t.Error("expected warnings about budget sum mismatch, got none")
	}
}

// Sub-phase budgets that add up to the parent must not warn. taskBudgetSum
// accumulates floats parsed from strings, so amounts equal to the cent
// routinely differ in the last bits; an exact comparison reported "sum to
// $99096.14 but parent budget is $99096.14" on real projects.
func TestComputePhaseBudget_NoWarningWhenSubPhasesSum(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Project Management", PhaseNumber: strPtr("#01"), Total: strPtr("99096.14")},
		{MosaicID: 201, Name: "Sub A", PhaseNumber: strPtr("01"), ParentID: intPtr(100), Total: strPtr("33032.05")},
		{MosaicID: 202, Name: "Sub B", PhaseNumber: strPtr("02"), ParentID: intPtr(100), Total: strPtr("33032.05")},
		{MosaicID: 203, Name: "Sub C", PhaseNumber: strPtr("03"), ParentID: intPtr(100), Total: strPtr("33032.04")},
	}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, nil, nil, map[int]float64{}, today)

	for _, w := range report.Warnings {
		if strings.Contains(w, "sub-phase budgets sum to") {
			t.Errorf("unexpected sum warning: %s", w)
		}
	}
}

// Sub-phases with no budget of their own are a shape, not a discrepancy: the
// fee was entered on the parent and the work broken out beneath it. Warning
// there would say "sum to $0.00 but parent budget is $100000.00" on every such
// project, with nothing behind it to act on.
func TestComputePhaseBudget_NoWarningWhenSubPhasesHaveNoBudget(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Design", PhaseNumber: strPtr("#01"), Total: strPtr("100000")},
		{MosaicID: 201, Name: "Sub A", PhaseNumber: strPtr("01"), ParentID: intPtr(100)},
		{MosaicID: 202, Name: "Sub B", PhaseNumber: strPtr("02"), ParentID: intPtr(100)},
	}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, nil, nil, map[int]float64{}, today)

	for _, w := range report.Warnings {
		if strings.Contains(w, "sub-phase budgets sum to") {
			t.Errorf("unexpected sum warning: %s", w)
		}
	}
}

// A sub-phase budget of an explicit zero is still a budget. Reading the guard
// off the sum instead of off the children would treat this as "nobody was
// given a budget" and suppress a real mismatch.
func TestComputePhaseBudget_WarnsWhenSubPhaseBudgetsAreExplicitZero(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Design", PhaseNumber: strPtr("#01"), Total: strPtr("100000")},
		{MosaicID: 201, Name: "Sub A", PhaseNumber: strPtr("01"), ParentID: intPtr(100), Total: strPtr("0")},
		{MosaicID: 202, Name: "Sub B", PhaseNumber: strPtr("02"), ParentID: intPtr(100), Total: strPtr("0")},
	}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, nil, nil, map[int]float64{}, today)

	var found bool
	for _, w := range report.Warnings {
		if strings.Contains(w, "sub-phase budgets sum to") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a sum warning, got %v", report.Warnings)
	}
}

// One sub-phase carrying a budget means the money was meant to be split, so a
// shortfall is still worth reporting.
func TestComputePhaseBudget_WarnsWhenSomeSubPhasesHaveBudget(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Design", PhaseNumber: strPtr("#01"), Total: strPtr("100000")},
		{MosaicID: 201, Name: "Sub A", PhaseNumber: strPtr("01"), ParentID: intPtr(100), Total: strPtr("40000")},
		{MosaicID: 202, Name: "Sub B", PhaseNumber: strPtr("02"), ParentID: intPtr(100)},
	}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, nil, nil, map[int]float64{}, today)

	var found bool
	for _, w := range report.Warnings {
		if strings.Contains(w, "sub-phase budgets sum to") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a sum warning, got %v", report.Warnings)
	}
}

// A difference big enough to matter still warns.
func TestComputePhaseBudget_WarnsOnRealSumGap(t *testing.T) {
	phases := []Phase{
		{MosaicID: 100, Name: "Design", PhaseNumber: strPtr("#01"), Total: strPtr("100000")},
		{MosaicID: 201, Name: "Sub A", PhaseNumber: strPtr("01"), ParentID: intPtr(100), Total: strPtr("99999.97")},
	}
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	report := ComputePhaseBudget(1, phases, nil, nil, map[int]float64{}, today)

	var found bool
	for _, w := range report.Warnings {
		if strings.Contains(w, "sub-phase budgets sum to") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a sum warning for a 3 cent gap, got %v", report.Warnings)
	}
}
