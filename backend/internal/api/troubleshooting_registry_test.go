package api

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"testing"

	"webprogfndb/internal/store"
)

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

func TestTroubleshootingKeysStayWithinDatabaseBoundary(t *testing.T) {
	label := strings.Repeat("ﬃ", 64)
	base := slugTroubleshootingKey(label)
	if len(base) != 64 {
		t.Fatalf("slug length = %d, want 64 (%q)", len(base), base)
	}

	existing := map[string]bool{base: true}
	got := nextTroubleshootingKey(label, existing)
	if len(got) != 64 || got != base[:62]+"-2" {
		t.Fatalf("colliding key = %q (len %d), want %q", got, len(got), base[:62]+"-2")
	}

	existing[got] = true
	for suffix := 3; suffix <= 10; suffix++ {
		got = nextTroubleshootingKey(label, existing)
		if len(got) > 64 {
			t.Fatalf("key at suffix %d has length %d: %q", suffix, len(got), got)
		}
		if !strings.HasSuffix(got, "-"+strconv.Itoa(suffix)) {
			t.Fatalf("key at suffix %d = %q", suffix, got)
		}
		existing[got] = true
	}
}

func TestNewTroubleshootingSeverityDefaultsToOutlineTone(t *testing.T) {
	label := "New Severity"
	_, _, _, tone, fields := severityValues(troubleshootingRegistryBody{Label: &label}, nil)
	if len(fields) != 0 {
		t.Fatalf("severity fields rejected: %v", fields)
	}
	if tone != "outline" {
		t.Fatalf("new severity tone = %q, want outline", tone)
	}
}

func TestTroubleshootingSortOrderMustBePositiveOnUpdate(t *testing.T) {
	zero := int64(0)
	_, _, _, categoryFields := categoryValues(troubleshootingRegistryBody{SortOrder: &zero}, &store.TroubleshootingCategory{SortOrder: 1})
	if categoryFields["sort_order"] == "" {
		t.Fatalf("category sort_order=0 accepted: %v", categoryFields)
	}
	_, _, _, _, severityFields := severityValues(troubleshootingRegistryBody{SortOrder: &zero}, &store.TroubleshootingSeverity{SortOrder: 1, Tone: "default"})
	if severityFields["sort_order"] == "" {
		t.Fatalf("severity sort_order=0 accepted: %v", severityFields)
	}
}

func TestTroubleshootingReplacementErrorPreservesDatabaseFailures(t *testing.T) {
	databaseErr := errors.New("opposite registry lookup failed")
	message, err := classifyTroubleshootingReplacementError(sql.ErrNoRows, databaseErr)
	if message != "" || !errors.Is(err, databaseErr) {
		t.Fatalf("opposite lookup error = (%q, %v), want preserved %v", message, err, databaseErr)
	}

	message, err = classifyTroubleshootingReplacementError(databaseErr, sql.ErrNoRows)
	if message != "" || !errors.Is(err, databaseErr) {
		t.Fatalf("primary lookup error = (%q, %v), want preserved %v", message, err, databaseErr)
	}
	if message, err = classifyTroubleshootingReplacementError(sql.ErrNoRows, nil); err != nil || message != "must reference the same registry type" {
		t.Fatalf("wrong type = (%q, %v)", message, err)
	}
	if message, err = classifyTroubleshootingReplacementError(sql.ErrNoRows, sql.ErrNoRows); err != nil || message != "does not exist" {
		t.Fatalf("missing replacement = (%q, %v)", message, err)
	}
}
