package mosaic

import (
	"strings"
	"testing"
)

// Every bill-rate warning must be recognizable to a caller deciding what is
// worth reporting. All three affect planned dollar amounts only and say
// nothing about the project's budget figures.
//
// The messages come from the builders the producers call, not from literals
// copied here. An edit that drops a marker fails this test rather than quietly
// making RateWarning return false and putting every affected project back in
// the report.
func TestRateWarningMatchesEveryBuilder(t *testing.T) {
	for name, w := range map[string]string{
		"lookup failed, fallback used": rateLookupFallbackWarning("Allison Zimmerman"),
		"lookup failed, no fallback":   rateLookupNoFallbackWarning("Cori McGovern"),
		"no rate configured":           noBillRateWarning("Allison Zimmerman"),
	} {
		t.Run(name, func(t *testing.T) {
			if !RateWarning(w) {
				t.Errorf("RateWarning(%q) = false, want true", w)
			}
		})
	}
}

// The member's name has to survive into the message, or a reader cannot tell
// whose rate is missing.
func TestRateWarningBuildersNameTheMember(t *testing.T) {
	const member = "Allison Zimmerman"
	for _, w := range []string{
		rateLookupFallbackWarning(member),
		rateLookupNoFallbackWarning(member),
		noBillRateWarning(member),
	} {
		if !strings.HasPrefix(w, member+": ") {
			t.Errorf("warning %q does not lead with the member name", w)
		}
	}
}

// A warning about the project's own budget data is not a rate warning and must
// stay reportable on its own.
func TestRateWarningIgnoresBudgetWarnings(t *testing.T) {
	for _, w := range []string{
		"phase 01 Design: sub-phase budgets sum to $1000.00 but parent budget is $1200.00",
		"",
	} {
		if RateWarning(w) {
			t.Errorf("RateWarning(%q) = true, want false", w)
		}
	}
}
