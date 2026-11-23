package api

import "testing"

func TestNaturalCompare(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"Run 2", "Run 10", -1},
		{"Run 10", "Run 2", 1},
		{"run 02", "Run 2", 0},
		{"alpha", "Bravo", -1},
		{"Bravo", "charlie", -1},
		{"Run", "Run 1", -1},
		{"step10b", "step10a", 1},
		{"", "", 0},
	} {
		if got := naturalCompare(tc.a, tc.b); got != tc.want {
			t.Errorf("naturalCompare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
