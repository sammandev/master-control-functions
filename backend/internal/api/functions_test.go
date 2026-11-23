package api

import "testing"

func TestLikePattern(t *testing.T) {
	for _, tt := range []struct{ term, want string }{
		{` C:\Windows\System32 `, `%C:\\Windows\\System32%`},
		{`C:\`, `%C:\\%`},
		{" TX%_MULTI ", "%TX_MULTI%"},
		{" % ", ""},
	} {
		if got := likePattern(tt.term); got != tt.want {
			t.Errorf("likePattern(%q) = %q, want %q", tt.term, got, tt.want)
		}
	}
}
