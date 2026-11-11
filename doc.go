// Package mosaic provides a Go client for the Mosaic time management platform API.
//
// The Mosaic API uses bearer token authentication with multi-tenant headers.
// All API requests require an API key, team ID, and appropriate tenant/realm
// configuration.
//
// # Authentication
//
// Requests are authenticated via a bearer token in the Authorization header.
// Additional required headers include Origin, x-tenant, and x-realm-id for
// multi-tenant routing.
//
// # Usage
//
//	client := mosaic.NewClient("your-api-key", "your-team-id")
//
//	// List projects
//	projects, err := client.ListProjects(ctx, mosaic.ListParams{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get a specific member
//	member, err := client.GetMember(ctx, 123)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Rate Limiting
//
// The Mosaic API allows 500 requests per 5 minutes per API key. A Client
// starts its requests at least [DefaultRequestInterval] apart to stay under
// it; [WithRequestInterval] changes or disables that. On a 429 the Client
// holds back all of its requests for the Retry-After time (2 minutes when
// the header is missing or unreadable) and retries up to 3 times, then
// returns a [RateLimitError].
//
// # Supported Resources
//
// The client supports the following API resources:
//
//   - Projects, Phases, Activity Phases
//   - Time Entries, Time Tracking
//   - Members (Employees), Member Roles, Member Project Rates, Member Project Roles
//   - Portfolios, Clients, Invoices
//   - Tasks, Task Lists
//   - Budget Estimates, Work Plans
//   - Departments, Offices, Roles
//   - Standard Work Categories, Scopes, Calendar Events
//   - Rates, Rate Groups, Bill Rates, Cost Rates, Entity Rates
//   - Holidays, PTO
//   - Team Currencies, Currency Exchange Rates
//   - Request Logs
//
// # API Path Pattern
//
// All API endpoints follow the pattern /api/{team_id}/{data_type}.
package mosaic
