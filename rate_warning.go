package mosaic

import "strings"

// Shared fragments of every warning that concerns a member's bill rate. The
// producers build their text from these and RateWarning matches on them, so
// the wording and the predicate cannot drift apart.
const (
	rateLookupWarningMarker = "could not look up project-specific rate"
	noBillRateWarningMarker = "no bill rate found"
)

// RateWarning reports whether w concerns a member's bill rate rather than the
// project's budget data.
//
// Three warnings qualify: the project-specific rate lookup failing with a
// global fallback available, the same lookup failing with no fallback, and a
// member who has no bill rate configured at all. All three affect planned
// dollar amounts only, and none of them says anything about whether the
// project's budget figures are right.
//
// A caller deciding whether a project is worth reporting on can use this to
// tell those apart from a warning about the data itself, such as sub-phase
// budgets that do not sum to their parent.
func RateWarning(w string) bool {
	return strings.Contains(w, rateLookupWarningMarker) ||
		strings.Contains(w, noBillRateWarningMarker)
}
