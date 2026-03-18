// Command mosaic-cli is a read-only CLI for querying the Mosaic API.
// It serves as both a test harness for the mosaic Go package and a quick
// way to inspect production data.
//
// # Environment Variables
//
//   - MOSAIC_API_KEY  — Bearer token for authentication (required)
//   - MOSAIC_TEAM_ID  — Team ID for API path construction (required)
//   - MOSAIC_BASE_URL — Override the default API base URL (optional)
//   - MOSAIC_ORIGIN  — Origin header value (optional, may be required by API)
//   - MOSAIC_TENANT  — x-tenant header (optional, default: "prod")
//   - MOSAIC_REALM_ID — x-realm-id header (optional, default: "prod")
//
// Use -env to load variables from a .env file (parsed before commands run):
//
//	mosaic-cli -env .secrets/.env <command> [flags]
//
// # Usage
//
//	mosaic-cli [-env <file>] <command> [flags]
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/scjalliance/mosaic"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("mosaic-cli: ")

	// Parse -env flag before subcommand.
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "-env" {
		if err := loadEnvFile(args[1]); err != nil {
			log.Fatalf("failed to load env file %s: %v", args[1], err)
		}
		log.Printf("loaded env file %s", args[1])
		args = args[2:]
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	subcmd := args[0]

	// Handle help before requiring credentials.
	switch subcmd {
	case "help", "-h", "--help":
		printUsage()
		return
	}

	client, err := buildClient()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	switch subcmd {
	case "projects":
		err = runProjects(ctx, client, args[1:])
	case "members":
		err = runMembers(ctx, client, args[1:])
	case "clients":
		err = runClients(ctx, client)
	case "phases":
		err = runPhases(ctx, client, args[1:])
	case "portfolios":
		err = runPortfolios(ctx, client)
	case "work-categories":
		err = runWorkCategories(ctx, client)
	case "departments":
		err = runDepartments(ctx, client)
	case "roles":
		err = runRoles(ctx, client)
	case "holidays":
		err = runHolidays(ctx, client, args[1:])
	case "offices":
		err = runOffices(ctx, client)
	case "pto":
		err = runPTO(ctx, client, args[1:])
	case "rates":
		err = runRates(ctx, client)
	case "rate-groups":
		err = runRateGroups(ctx, client)
	case "bill-rates":
		err = runBillRates(ctx, client, args[1:])
	case "cost-rates":
		err = runCostRates(ctx, client)
	case "member-roles":
		err = runMemberRoles(ctx, client, args[1:])
	case "member-project-rates":
		err = runMemberProjectRates(ctx, client, args[1:])
	case "member-project-roles":
		err = runMemberProjectRoles(ctx, client, args[1:])
	case "entity-rates":
		err = runEntityRates(ctx, client, args[1:])
	case "calendar-events":
		err = runCalendarEvents(ctx, client, args[1:])
	case "exchange-rates":
		err = runExchangeRates(ctx, client, args[1:])
	case "team-currencies":
		err = runTeamCurrencies(ctx, client)
	case "time-entries":
		err = runTimeEntries(ctx, client, args[1:])
	case "activity-phases":
		err = runActivityPhases(ctx, client, args[1:])
	case "time-tracking":
		err = runTimeTracking(ctx, client, args[1:])
	case "work-plans":
		err = runWorkPlans(ctx, client, args[1:])
	case "scopes":
		err = runScopes(ctx, client, args[1:])
	case "request-logs":
		err = runRequestLogs(ctx, client, args[1:])
	case "phase-budget":
		err = runPhaseBudget(ctx, client, args[1:])
	case "raw":
		err = runRaw(ctx, client, args[1:])
	default:
		log.Fatalf("unknown command: %s\n\nRun 'mosaic-cli help' for usage.", subcmd)
	}

	if err != nil {
		log.Fatal(err)
	}
}

// buildClient creates a mosaic.Client from environment variables.
func buildClient() (*mosaic.Client, error) {
	apiKey := os.Getenv("MOSAIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MOSAIC_API_KEY environment variable is required")
	}

	teamID := os.Getenv("MOSAIC_TEAM_ID")
	if teamID == "" {
		return nil, fmt.Errorf("MOSAIC_TEAM_ID environment variable is required")
	}

	var opts []mosaic.Option
	if baseURL := os.Getenv("MOSAIC_BASE_URL"); baseURL != "" {
		opts = append(opts, mosaic.WithBaseURL(baseURL))
	}
	if origin := os.Getenv("MOSAIC_ORIGIN"); origin != "" {
		opts = append(opts, mosaic.WithOrigin(origin))
	}
	if tenant := os.Getenv("MOSAIC_TENANT"); tenant != "" {
		opts = append(opts, mosaic.WithTenant(tenant))
	}
	if realmID := os.Getenv("MOSAIC_REALM_ID"); realmID != "" {
		opts = append(opts, mosaic.WithRealmID(realmID))
	}

	return mosaic.NewClient(apiKey, teamID, opts...), nil
}

// printJSON marshals v as indented JSON and writes it to stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// runProjects lists projects with optional filters.
// The Mosaic API requires a portfolio_id filter. If --portfolio-id is not
// provided, projects are fetched across all portfolios.
func runProjects(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("projects", flag.ExitOnError)
	limit := fs.Int("limit", 0, "maximum number of results")
	offset := fs.Int("offset", 0, "number of results to skip")
	archived := fs.Bool("archived", false, "include archived projects")
	portfolioID := fs.Int("portfolio-id", 0, "filter by portfolio ID (omit to query all portfolios)")
	budgetStatus := fs.String("budget-status", "", "filter by budget status (e.g. active, archived)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.ProjectFilter{
		ListParams: mosaic.ListParams{
			Limit:  *limit,
			Offset: *offset,
		},
		BudgetStatus: *budgetStatus,
	}
	if *archived {
		v := true
		filter.IsArchived = &v
	}

	// If portfolio_id is specified, use it directly.
	if *portfolioID > 0 {
		filter.PortfolioID = *portfolioID
		projects, err := client.ListProjects(ctx, filter)
		if err != nil {
			return fmt.Errorf("listing projects: %w", err)
		}
		fmt.Fprintf(os.Stderr, "fetched %d projects\n", len(projects))
		return printJSON(projects)
	}

	// Otherwise, list all portfolios and aggregate projects across them.
	portfolios, err := client.ListAllPortfolios(ctx)
	if err != nil {
		return fmt.Errorf("listing portfolios: %w", err)
	}

	var allProjects []mosaic.Project
	for _, p := range portfolios {
		filter.PortfolioID = p.MosaicID
		projects, err := client.ListProjects(ctx, filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping portfolio %d (%s): %v\n", p.MosaicID, p.Name, err)
			continue
		}
		allProjects = append(allProjects, projects...)
	}

	fmt.Fprintf(os.Stderr, "fetched %d projects across %d portfolios\n", len(allProjects), len(portfolios))
	return printJSON(allProjects)
}

// runMembers lists all team members.
func runMembers(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("members", flag.ExitOnError)
	includeDiscarded := fs.Bool("include-discarded", false, "include discarded members")
	if err := fs.Parse(args); err != nil {
		return err
	}

	members, err := client.ListAllMembers(ctx, *includeDiscarded)
	if err != nil {
		return fmt.Errorf("listing members: %w", err)
	}

	fmt.Fprintf(os.Stderr, "fetched %d members\n", len(members))
	return printJSON(members)
}

// runClients lists all clients.
func runClients(ctx context.Context, client *mosaic.Client) error {
	clients, err := client.ListClients(ctx)
	if err != nil {
		return fmt.Errorf("listing clients: %w", err)
	}

	fmt.Fprintf(os.Stderr, "fetched %d clients\n", len(clients))
	return printJSON(clients)
}

// runPhases lists phases, optionally filtered by project ID, search text, or budget status.
func runPhases(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("phases", flag.ExitOnError)
	projectID := fs.Int("project-id", 0, "filter phases by project ID")
	search := fs.String("search", "", "search phases by name (uses search_text API param)")
	isBudget := fs.String("is-budget", "", "filter by budget phase status (true/false)")
	limit := fs.Int("limit", 0, "maximum number of results")
	offset := fs.Int("offset", 0, "number of results to skip")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *projectID > 0 {
		phases, err := client.ListPhasesByProject(ctx, *projectID)
		if err != nil {
			return fmt.Errorf("listing phases for project %d: %w", *projectID, err)
		}
		fmt.Fprintf(os.Stderr, "fetched %d phases for project %d\n", len(phases), *projectID)
		return printJSON(phases)
	}

	filter := mosaic.PhaseIndexFilter{
		ListParams: mosaic.ListParams{
			Limit:  *limit,
			Offset: *offset,
		},
		SearchText: *search,
	}
	if *isBudget != "" {
		v := *isBudget == "true"
		filter.IsBudget = &v
	}

	phases, err := client.ListAllPhases(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing all phases: %w", err)
	}

	fmt.Fprintf(os.Stderr, "fetched %d phases\n", len(phases))
	return printJSON(phases)
}

// runPortfolios lists all portfolios.
func runPortfolios(ctx context.Context, client *mosaic.Client) error {
	portfolios, err := client.ListAllPortfolios(ctx)
	if err != nil {
		return fmt.Errorf("listing portfolios: %w", err)
	}

	fmt.Fprintf(os.Stderr, "fetched %d portfolios\n", len(portfolios))
	return printJSON(portfolios)
}

// runWorkCategories lists all standard work categories.
func runWorkCategories(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListStandardWorkCategories(ctx)
	if err != nil {
		return fmt.Errorf("listing work categories: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d work categories\n", len(items))
	return printJSON(items)
}

// runDepartments lists all departments.
func runDepartments(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListDepartments(ctx)
	if err != nil {
		return fmt.Errorf("listing departments: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d departments\n", len(items))
	return printJSON(items)
}

// runRoles lists all roles.
func runRoles(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListRoles(ctx)
	if err != nil {
		return fmt.Errorf("listing roles: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d roles\n", len(items))
	return printJSON(items)
}

// runHolidays lists holidays with optional date range filters.
func runHolidays(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("holidays", flag.ExitOnError)
	startDate := fs.String("start-date", "", "filter holidays starting on or after this date")
	endDate := fs.String("end-date", "", "filter holidays ending on or before this date")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.HolidayFilter{
		StartDate: *startDate,
		EndDate:   *endDate,
	}
	items, err := client.ListHolidays(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing holidays: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d holidays\n", len(items))
	return printJSON(items)
}

// runOffices lists all offices.
func runOffices(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListOffices(ctx)
	if err != nil {
		return fmt.Errorf("listing offices: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d offices\n", len(items))
	return printJSON(items)
}

// runPTO lists PTO types with optional filters.
func runPTO(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("pto", flag.ExitOnError)
	includeArchived := fs.Bool("include-archived", false, "include archived PTO types")
	includeDefault := fs.Bool("include-default", false, "include default PTO types")
	isCustom := fs.String("is-custom", "", "filter by custom status (true/false)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var filter mosaic.PTOFilter
	if *includeArchived {
		v := true
		filter.IncludeArchived = &v
	}
	if *includeDefault {
		v := true
		filter.IncludeDefault = &v
	}
	if *isCustom != "" {
		v := *isCustom == "true"
		filter.IsCustom = &v
	}
	items, err := client.ListPTOs(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing PTOs: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d PTOs\n", len(items))
	return printJSON(items)
}

// runRates lists all rates.
func runRates(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListRates(ctx)
	if err != nil {
		return fmt.Errorf("listing rates: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d rates\n", len(items))
	return printJSON(items)
}

// runRateGroups lists all rate groups.
func runRateGroups(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListRateGroups(ctx)
	if err != nil {
		return fmt.Errorf("listing rate groups: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d rate groups\n", len(items))
	return printJSON(items)
}

// runBillRates lists bill rates with optional filters.
func runBillRates(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("bill-rates", flag.ExitOnError)
	includeArchived := fs.Bool("include-archived", false, "include archived bill rates")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var filter mosaic.BillRateFilter
	if *includeArchived {
		v := true
		filter.IncludeArchived = &v
	}
	items, err := client.ListBillRates(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing bill rates: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d bill rates\n", len(items))
	return printJSON(items)
}

// runCostRates lists all cost rates.
func runCostRates(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListCostRates(ctx)
	if err != nil {
		return fmt.Errorf("listing cost rates: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d cost rates\n", len(items))
	return printJSON(items)
}

// runMemberRoles lists member roles with optional filters.
func runMemberRoles(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("member-roles", flag.ExitOnError)
	memberID := fs.Int("member-id", 0, "filter by member ID")
	roleID := fs.Int("role-id", 0, "filter by role ID")
	includeArchived := fs.Bool("include-archived", false, "include archived member roles")
	isActive := fs.String("is-active", "", "filter by active status (true/false)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.MemberRoleFilter{
		MemberID: *memberID,
		RoleID:   *roleID,
	}
	if *includeArchived {
		v := true
		filter.IncludeArchived = &v
	}
	if *isActive != "" {
		v := *isActive == "true"
		filter.IsActive = &v
	}
	items, err := client.ListMemberRoles(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing member roles: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d member roles\n", len(items))
	return printJSON(items)
}

// runMemberProjectRates lists member project rates filtered by member and project.
func runMemberProjectRates(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("member-project-rates", flag.ExitOnError)
	memberID := fs.Int("member-id", 0, "member ID to filter by (required)")
	projectID := fs.Int("project-id", 0, "project ID to filter by (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memberID == 0 || *projectID == 0 {
		return fmt.Errorf("both -member-id and -project-id are required")
	}
	filter := mosaic.MemberProjectRateFilter{
		MemberID:  *memberID,
		ProjectID: *projectID,
	}
	items, err := client.ListMemberProjectRates(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing member project rates: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d member project rates\n", len(items))
	return printJSON(items)
}

// runMemberProjectRoles lists member project roles with optional filters.
func runMemberProjectRoles(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("member-project-roles", flag.ExitOnError)
	memberID := fs.Int("member-id", 0, "filter by member ID")
	projectID := fs.Int("project-id", 0, "filter by project ID")
	roleID := fs.Int("role-id", 0, "filter by role ID")
	phaseID := fs.Int("phase-id", 0, "filter by phase ID")
	includeArchived := fs.Bool("include-archived", false, "include archived roles")
	isActive := fs.String("is-active", "", "filter by active status (true/false)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.MemberProjectRoleFilter{
		MemberID:  *memberID,
		ProjectID: *projectID,
		RoleID:    *roleID,
		PhaseID:   *phaseID,
	}
	if *includeArchived {
		v := true
		filter.IncludeArchived = &v
	}
	if *isActive != "" {
		v := *isActive == "true"
		filter.IsActive = &v
	}
	items, err := client.ListMemberProjectRoles(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing member project roles: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d member project roles\n", len(items))
	return printJSON(items)
}

// runEntityRates lists entity rates with optional filters.
func runEntityRates(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("entity-rates", flag.ExitOnError)
	entityType := fs.String("entity-type", "", "filter by entity type")
	rateGroupID := fs.Int("rate-group-id", 0, "filter by rate group ID")
	startDate := fs.String("start-date", "", "filter by start date")
	endDate := fs.String("end-date", "", "filter by end date")
	includeArchived := fs.Bool("include-archived", false, "include archived entity rates")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.EntityRateFilter{
		EntityType: *entityType,
		StartDate:  *startDate,
		EndDate:    *endDate,
	}
	if *rateGroupID > 0 {
		filter.RateGroupIDs = []int{*rateGroupID}
	}
	if *includeArchived {
		v := true
		filter.IncludeArchived = &v
	}
	items, err := client.ListEntityRates(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing entity rates: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d entity rates\n", len(items))
	return printJSON(items)
}

// runCalendarEvents lists calendar events with optional filters.
func runCalendarEvents(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("calendar-events", flag.ExitOnError)
	startDate := fs.String("start-date", "", "filter events starting on or after this date")
	endDate := fs.String("end-date", "", "filter events ending on or before this date")
	title := fs.String("title", "", "filter events by title")
	projectID := fs.Int("project-id", 0, "filter by project ID")
	phaseID := fs.Int("phase-id", 0, "filter by phase ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.CalendarEventFilter{
		StartDate: *startDate,
		EndDate:   *endDate,
		Title:     *title,
	}
	if *projectID > 0 {
		filter.ProjectIDs = []int{*projectID}
	}
	if *phaseID > 0 {
		filter.PhaseIDs = []int{*phaseID}
	}
	items, err := client.ListCalendarEvents(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing calendar events: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d calendar events\n", len(items))
	return printJSON(items)
}

// runExchangeRates lists currency exchange rates within a date range.
func runExchangeRates(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("exchange-rates", flag.ExitOnError)
	startDate := fs.String("start-date", "", "start date (required)")
	endDate := fs.String("end-date", "", "end date (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *startDate == "" || *endDate == "" {
		return fmt.Errorf("--start-date and --end-date are required for exchange-rates")
	}

	filter := mosaic.CurrencyExchangeRateFilter{
		StartDate: *startDate,
		EndDate:   *endDate,
	}
	items, err := client.ListCurrencyExchangeRates(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing exchange rates: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d exchange rates\n", len(items))
	return printJSON(items)
}

// runTeamCurrencies lists all team currencies.
func runTeamCurrencies(ctx context.Context, client *mosaic.Client) error {
	items, err := client.ListTeamCurrencies(ctx)
	if err != nil {
		return fmt.Errorf("listing team currencies: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d team currencies\n", len(items))
	return printJSON(items)
}

// runActivityPhases lists activity phases with optional filters.
func runActivityPhases(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("activity-phases", flag.ExitOnError)
	projectID := fs.Int("project-id", 0, "filter by project ID")
	phaseID := fs.Int("phase-id", 0, "filter by phase ID")
	startDate := fs.String("start-date", "", "filter by start date")
	endDate := fs.String("end-date", "", "filter by end date")
	all := fs.Bool("all", false, "return all activity phases")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.ActivityPhaseFilter{
		StartDate: *startDate,
		EndDate:   *endDate,
		All:       *all,
	}
	if *projectID > 0 {
		filter.ProjectIDs = []int{*projectID}
	}
	if *phaseID > 0 {
		filter.PhaseIDs = []int{*phaseID}
	}
	items, err := client.ListActivityPhases(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing activity phases: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d activity phases\n", len(items))
	return printJSON(items)
}

// runTimeTracking lists time tracking entries for a given date.
func runTimeTracking(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("time-tracking", flag.ExitOnError)
	date := fs.String("date", "", "date to fetch entries for (required, MM/DD/YYYY)")
	memberIDs := fs.String("member-ids", "", "comma-separated member IDs")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *date == "" {
		return fmt.Errorf("--date is required for time-tracking")
	}

	filter := mosaic.TimeTrackingFilter{
		Date:      *date,
		MemberIDs: *memberIDs,
	}
	items, err := client.ListTimeTracking(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing time tracking: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d time tracking entries\n", len(items))
	return printJSON(items)
}

// runWorkPlans lists work plans with optional filters.
func runWorkPlans(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("work-plans", flag.ExitOnError)
	all := fs.Bool("all", false, "return all work plans")
	limit := fs.Int("limit", 0, "maximum number of results")
	offset := fs.Int("offset", 0, "number of results to skip")
	startDate := fs.String("start-date", "", "filter by start date")
	endDate := fs.String("end-date", "", "filter by end date")
	memberID := fs.Int("member-id", 0, "filter by member ID")
	projectID := fs.Int("project-id", 0, "filter by project ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.WorkPlanFilter{
		ListParams: mosaic.ListParams{
			Limit:  *limit,
			Offset: *offset,
		},
		All:       *all,
		StartDate: *startDate,
		EndDate:   *endDate,
	}
	if *memberID > 0 {
		filter.MemberIDs = []int{*memberID}
	}
	if *projectID > 0 {
		filter.ProjectIDs = []int{*projectID}
	}
	items, err := client.ListWorkPlans(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing work plans: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d work plans\n", len(items))
	return printJSON(items)
}

// runScopes lists scopes with optional filters.
func runScopes(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("scopes", flag.ExitOnError)
	all := fs.Bool("all", false, "return all scopes")
	limit := fs.Int("limit", 0, "maximum number of results")
	offset := fs.Int("offset", 0, "number of results to skip")
	projectID := fs.Int("project-id", 0, "filter by project ID")
	scheduleStart := fs.String("schedule-start", "", "filter by schedule start date")
	scheduleEnd := fs.String("schedule-end", "", "filter by schedule end date")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.ScopeFilter{
		ListParams: mosaic.ListParams{
			Limit:  *limit,
			Offset: *offset,
		},
		All:           *all,
		ScheduleStart: *scheduleStart,
		ScheduleEnd:   *scheduleEnd,
	}
	if *projectID > 0 {
		filter.ProjectIDs = []int{*projectID}
	}
	items, err := client.ListScopes(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing scopes: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d scopes\n", len(items))
	return printJSON(items)
}

// runRequestLogs fetches the status of an API request.
func runRequestLogs(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("request-logs", flag.ExitOnError)
	apiRequestID := fs.String("api-request-id", "", "API request ID to check")
	dataType := fs.String("data-type", "", "data type to check")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *apiRequestID == "" {
		return fmt.Errorf("--api-request-id is required for request-logs")
	}

	result, err := client.GetRequestLog(ctx, *apiRequestID, *dataType)
	if err != nil {
		return fmt.Errorf("getting request log: %w", err)
	}
	return printJSON(result)
}

// runTimeEntries lists time entries with optional filters.
func runTimeEntries(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("time-entries", flag.ExitOnError)
	startDate := fs.String("start-date", "", "filter entries on or after this date (MM/DD/YYYY)")
	endDate := fs.String("end-date", "", "filter entries on or before this date (MM/DD/YYYY)")
	memberIDs := fs.String("member-ids", "", "comma-separated member IDs")
	projectIDs := fs.String("project-ids", "", "comma-separated project IDs")
	limit := fs.Int("limit", 0, "maximum number of results")
	offset := fs.Int("offset", 0, "number of results to skip")
	billable := fs.String("billable", "", "filter by billable status (true/false)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	filter := mosaic.TimeEntryFilter{
		ListParams: mosaic.ListParams{
			Limit:  *limit,
			Offset: *offset,
		},
		MemberIDs:  parseIntList(*memberIDs),
		ProjectIDs: parseIntList(*projectIDs),
	}
	if *startDate != "" {
		d, err := mosaic.ParseDate(*startDate)
		if err != nil {
			return fmt.Errorf("parsing start-date: %w", err)
		}
		filter.StartDate = &d
	}
	if *endDate != "" {
		d, err := mosaic.ParseDate(*endDate)
		if err != nil {
			return fmt.Errorf("parsing end-date: %w", err)
		}
		filter.EndDate = &d
	}
	if *billable != "" {
		v := *billable == "true"
		filter.Billable = &v
	}
	items, err := client.ListTimeEntries(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing time entries: %w", err)
	}
	fmt.Fprintf(os.Stderr, "fetched %d time entries\n", len(items))
	return printJSON(items)
}

// parseIntList parses a comma-separated string of integers into a slice.
// Empty strings return nil.
func parseIntList(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var ids []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if id, err := strconv.Atoi(p); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// runPhaseBudget generates a per-phase budget/spent/planned/remaining report for a project.
func runPhaseBudget(ctx context.Context, client *mosaic.Client, args []string) error {
	fs := flag.NewFlagSet("phase-budget", flag.ExitOnError)
	projectID := fs.Int("project-id", 0, "project ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *projectID == 0 {
		return fmt.Errorf("--project-id is required")
	}

	report, err := client.GetPhaseBudgetReport(ctx, *projectID)
	if err != nil {
		return fmt.Errorf("generating phase budget report: %w", err)
	}
	return printJSON(report)
}

// runRaw performs a raw GET request and prints the decoded JSON response.
// The path argument is relative to /api/{team_id}/ (e.g., "portfolio/index").
func runRaw(ctx context.Context, client *mosaic.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mosaic-cli raw <path>\n  example: mosaic-cli raw portfolio/index")
	}
	path := "/api/" + os.Getenv("MOSAIC_TEAM_ID") + "/" + args[0]
	fmt.Fprintf(os.Stderr, "GET %s\n", path)

	var result any
	if err := client.RawGet(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("raw GET %s: %w", path, err)
	}
	return printJSON(result)
}

// loadEnvFile reads a .env file and sets each KEY=VALUE pair as an environment
// variable. Lines that are empty, whitespace-only, or start with # are skipped.
// Values may optionally be wrapped in single or double quotes, which are stripped.
// Existing environment variables are NOT overwritten, so real env takes precedence.
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening env file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// Strip matching surrounding quotes.
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		// Don't overwrite existing env vars — real environment takes precedence.
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// printUsage writes usage information to stderr.
func printUsage() {
	fmt.Fprint(os.Stderr, `Usage: mosaic-cli [-env <file>] <command> [flags]

Read-only CLI for the Mosaic API.

Environment:
  MOSAIC_API_KEY   Bearer token for authentication (required)
  MOSAIC_TEAM_ID   Team ID (required)
  MOSAIC_BASE_URL  Override API base URL (optional)
  MOSAIC_ORIGIN   Origin header (optional, may be required by API)
  MOSAIC_TENANT   x-tenant header (default: "prod")
  MOSAIC_REALM_ID  x-realm-id header (default: "prod")

Commands:
  projects              List projects [--portfolio-id N] [--budget-status S] [--limit N] [--offset N] [--archived]
  members               List members [--include-discarded]
  clients               List clients
  phases                List phases [--project-id N] [--search TEXT] [--is-budget true|false] [--limit N] [--offset N]
  portfolios            List portfolios
  work-categories       List standard work categories
  departments           List departments
  roles                 List roles
  holidays              List holidays [--start-date DATE] [--end-date DATE]
  offices               List offices
  pto                   List PTO types [--include-archived] [--include-default] [--is-custom true|false]
  rates                 List rates
  rate-groups           List rate groups
  bill-rates            List bill rates [--include-archived]
  cost-rates            List cost rates
  member-roles          List member roles [--member-id N] [--role-id N] [--include-archived] [--is-active true|false]
  member-project-rates  List member project rates --member-id N --project-id N
  member-project-roles  List member project roles [--member-id N] [--project-id N] [--role-id N] [--phase-id N]
                          [--include-archived] [--is-active true|false]
  entity-rates          List entity rates [--entity-type S] [--rate-group-id N] [--start-date DATE] [--end-date DATE]
                          [--include-archived]
  calendar-events       List calendar events [--start-date DATE] [--end-date DATE] [--title TEXT]
                          [--project-id N] [--phase-id N]
  exchange-rates        List exchange rates --start-date DATE --end-date DATE
  team-currencies       List team currencies
  time-entries          List time entries [--start-date MM/DD/YYYY] [--end-date MM/DD/YYYY] [--member-ids IDs]
                          [--project-ids IDs] [--limit N] [--offset N] [--billable true|false]
  activity-phases       List activity phases [--project-id N] [--phase-id N] [--start-date DATE] [--end-date DATE]
                          [--all]
  time-tracking         List time tracking entries --date MM/DD/YYYY [--member-ids IDs]
  work-plans            List work plans [--all] [--limit N] [--offset N] [--start-date DATE] [--end-date DATE]
                          [--member-id N] [--project-id N]
  scopes                List scopes [--all] [--limit N] [--offset N] [--project-id N]
                          [--schedule-start DATE] [--schedule-end DATE]
  phase-budget          Per-phase budget/spent/planned/remaining report --project-id N
  request-logs          Get request log --api-request-id ID [--data-type TYPE]
  raw <path>            Raw GET (path relative to /api/{team_id}/)
  help                  Show this help
`)
}
