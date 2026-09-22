package mosaic

import (
	"fmt"
	"strings"
)

// Shared fragments of every warning that concerns a member's bill rate. The
// builders below compose their text from these and RateWarning matches on
// them, so the wording and the predicate cannot drift apart.
const (
	rateLookupWarningMarker = "could not look up project-specific rate"
	noBillRateWarningMarker = "no bill rate found"
)

// The three bill-rate warnings, as functions rather than inline formatting.
//
// Going through here is what keeps RateWarning honest: the tests call these
// and assert the predicate matches what comes out, so an edit that reworded a
// builder and dropped a marker fails the suite instead of quietly making
// RateWarning return false and putting every affected project back in the
// report. Re-inlining a message at the call site would still escape that, so
// new bill-rate warnings belong here rather than in resolveMemberRates.

// rateLookupFallbackWarning reports a project-specific rate that could not be
// read, where a global rate was available to stand in.
func rateLookupFallbackWarning(member string) string {
	return fmt.Sprintf("%s: %s, using fallback global rate instead", member, rateLookupWarningMarker)
}

// rateLookupNoFallbackWarning reports a project-specific rate that could not be
// read, with nothing to fall back on.
func rateLookupNoFallbackWarning(member string) string {
	return fmt.Sprintf("%s: %s and no fallback rate exists, so planned dollar amounts for this person will be $0",
		member, rateLookupWarningMarker)
}

// noBillRateWarning reports a member with no bill rate configured anywhere.
func noBillRateWarning(member string) string {
	return fmt.Sprintf("%s: %s, so planned dollar amounts for this person will be $0",
		member, noBillRateWarningMarker)
}

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
