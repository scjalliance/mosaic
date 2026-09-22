package mosaic

import "testing"

// Every bill-rate warning must be recognizable to a caller deciding what is
// worth reporting. All three affect planned dollar amounts only and say
// nothing about the project's budget figures.
func TestRateWarningMatchesProducedText(t *testing.T) {
	for _, w := range []string{
		"Allison Zimmerman: could not look up project-specific rate, using fallback global rate instead",
		"Cori McGovern: could not look up project-specific rate and no fallback rate exists, so planned dollar amounts for this person will be $0",
		"Allison Zimmerman: no bill rate found, so planned dollar amounts for this person will be $0",
	} {
		if !RateWarning(w) {
			t.Errorf("RateWarning(%q) = false, want true", w)
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
