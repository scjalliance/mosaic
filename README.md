# mosaic

Go client library for the [Mosaic](https://www.mosaicapp.com/) time management platform API.

Provides typed access to projects, phases, members, time entries, budgets, rates, and more.

## Installation

```bash
go get github.com/scjalliance/mosaic
```

## Authentication

The Mosaic API uses bearer token authentication with multi-tenant headers.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/scjalliance/mosaic"
)

func main() {
	client := mosaic.NewClient("your-api-key", "your-team-id",
		mosaic.WithOrigin("https://app.mosaicapp.com"),
	)

	ctx := context.Background()

	// List projects in a portfolio
	projects, err := client.ListProjects(ctx, mosaic.ProjectFilter{
		PortfolioID: 123,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range projects {
		fmt.Printf("%d  %s\n", p.ID, p.Title)
	}
}
```

## Rate Limiting

The Mosaic API enforces a rate limit of 300 requests per 5 minutes. When exceeded, methods return a `RateLimitError` containing the `Retry-After` duration from the response headers.

## Resources

The client covers the following Mosaic API resources:

- **Projects** — list, get, with portfolio filtering
- **Phases** — project phases, phase index, budget phases
- **Activity Phases** — phase activity tracking
- **Phase Budget** — per-phase budget/spent/planned/remaining reports
- **Time Entries** — time entry records with date/member/project filtering
- **Time Tracking** — daily time tracking entries
- **Members** — team members (employees)
- **Member Roles** — role assignments per member
- **Member Project Rates** — per-member per-project rate overrides
- **Member Project Roles** — role assignments per member per project/phase
- **Portfolios** — project portfolios
- **Clients** — client organizations
- **Invoices** — invoice records
- **Tasks** — task management
- **Budget Estimates** — project budget estimates
- **Work Plans** — resource work plans
- **Departments** — organizational departments
- **Offices** — office locations
- **Roles** — role definitions
- **Standard Work Categories** — work category taxonomy
- **Scopes** — project scopes
- **Calendar Events** — team calendar events
- **Rates** — base rate definitions
- **Rate Groups** — rate groupings
- **Bill Rates** — billing rates
- **Cost Rates** — cost rates
- **Entity Rates** — entity-level rate assignments
- **Holidays** — holiday schedules
- **PTO** — paid time off types
- **Team Currencies** — supported currencies
- **Currency Exchange Rates** — exchange rate history
- **Request Logs** — API request status tracking

## CLI Tool

The `cmd/mosaic-cli` directory contains a read-only CLI for querying the Mosaic API.

### Environment Variables

| Variable | Required | Description |
|---|---|---|
| `MOSAIC_API_KEY` | Yes | Bearer token for authentication |
| `MOSAIC_TEAM_ID` | Yes | Team ID for API path construction |
| `MOSAIC_BASE_URL` | No | Override the default API base URL |
| `MOSAIC_ORIGIN` | No | Origin header value (may be required by API) |
| `MOSAIC_TENANT` | No | x-tenant header (default: `prod`) |
| `MOSAIC_REALM_ID` | No | x-realm-id header (default: `prod`) |

Use `-env` to load variables from a file:

```bash
mosaic-cli -env .secrets/.env <command> [flags]
```

### Commands

```
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
entity-rates          List entity rates [--entity-type S] [--rate-group-id N] [--start-date DATE] [--end-date DATE]
calendar-events       List calendar events [--start-date DATE] [--end-date DATE] [--title TEXT]
exchange-rates        List exchange rates --start-date DATE --end-date DATE
team-currencies       List team currencies
time-entries          List time entries [--start-date MM/DD/YYYY] [--end-date MM/DD/YYYY] [--member-ids IDs]
activity-phases       List activity phases [--project-id N] [--phase-id N] [--start-date DATE] [--end-date DATE]
time-tracking         List time tracking entries --date MM/DD/YYYY [--member-ids IDs]
work-plans            List work plans [--all] [--limit N] [--offset N] [--start-date DATE] [--end-date DATE]
scopes                List scopes [--all] [--limit N] [--offset N] [--project-id N]
phase-budget          Per-phase budget report --project-id N
request-logs          Get request log --api-request-id ID [--data-type TYPE]
raw <path>            Raw GET (path relative to /api/{team_id}/)
```

## License

MIT — see [LICENSE](LICENSE).
