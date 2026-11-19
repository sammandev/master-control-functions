package api

import "testing"

func TestSlugTroubleshootingKey(t *testing.T) {
	tests := []struct {
		label string
		want  string
	}{
		{label: "Fixture Check", want: "fixture-check"},
		{label: "  Signal / Noise & Timing!  ", want: "signal-noise-timing"},
		{label: "Crème Brûlée", want: "creme-brulee"},
		{label: "日本語", want: "item"},
		{label: "!!!", want: "item"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			if got := slugTroubleshootingKey(tt.label); got != tt.want {
				t.Fatalf("slugTroubleshootingKey(%q) = %q, want %q", tt.label, got, tt.want)
			}
		})
	}
}

func TestNextTroubleshootingKeySuffixesCollisions(t *testing.T) {
	existing := map[string]bool{"fixture-check": true}
	if got := nextTroubleshootingKey("Fixture Check", existing); got != "fixture-check-2" {
		t.Fatalf("nextTroubleshootingKey() = %q, want fixture-check-2", got)
	}
	existing["fixture-check-2"] = true
	if got := nextTroubleshootingKey("Fixture Check", existing); got != "fixture-check-3" {
		t.Fatalf("nextTroubleshootingKey() = %q, want fixture-check-3", got)
	}
	if got := nextTroubleshootingKey("Fixture Check", map[string]bool{}); got != "fixture-check" {
		t.Fatalf("nextTroubleshootingKey() without collision = %q, want fixture-check", got)
	}
}
