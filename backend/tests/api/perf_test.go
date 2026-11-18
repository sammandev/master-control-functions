package api_test

import (
	"strings"
	"testing"
	"time"
)

// TestFullDatasetListLatency pins the performance sanity requirement of
// ticket 09: with the full real workbook imported, every list page
// endpoint answers locally in well under a second, and the function
// detail payload set (function + both link lists + filtered examples) —
// the hotspots where N+1 patterns would show — stays within the bound.
func TestFullDatasetListLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test")
	}
	app := newImportTestApp(t)
	importFixture(t, app)

	// TX_MULTI_VERIFICATION is the fixture's first function (id 1, DLL /
	// WIFI_TEST) and has positioned parameter links + examples.
	const fixtureFn = "TX_MULTI_VERIFICATION"

	cases := []struct {
		name   string
		target string
	}{
		{"functions list", "/api/functions?page=1&page_size=20"},
		{"functions filtered", "/api/functions?source=DLL&interface=WIFI_TEST&page=1&page_size=20"},
		{"function facets", "/api/functions/facets"},
		{"parameter definitions", "/api/parameter-definitions?page_size=100"},
		{"criteria definitions", "/api/criteria-definitions?page_size=100"},
		{"examples list", "/api/examples?page_size=50"},
		{"examples by function", "/api/examples?function_source=DLL&function_name=" + fixtureFn},
		{"orphans report", "/api/orphans"},
		{"function detail", "/api/functions/1"},
		{"parameter links", "/api/functions/1/parameter-links"},
		{"criteria links", "/api/functions/1/criteria-links"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			resp, body := doRequest(t, app, "GET", tc.target)
			elapsed := time.Since(start)
			if resp.StatusCode != 200 {
				t.Fatalf("%s: %d (body: %.200s)", tc.name, resp.StatusCode, body)
			}
			if elapsed > time.Second {
				t.Errorf("%s took %v (> 1s)", tc.name, elapsed)
			}
			t.Logf("%s: %v", tc.name, elapsed)
		})
	}

	// The filtered-examples target must actually match a function,
	// otherwise the timing exercises an empty filter.
	resp, body := doRequest(t, app, "GET", "/api/examples?function_source=DLL&function_name="+fixtureFn+"&page_size=1")
	if resp.StatusCode != 200 || !strings.Contains(body, fixtureFn) {
		t.Fatalf("fixture examples filter empty: %d (body: %.200s)", resp.StatusCode, body)
	}
}
