package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestListEndpointsReturnEmptyPageForOversizedPage(t *testing.T) {
	app := newTestApp(t)
	post := func(path, body string) {
		t.Helper()
		if resp, body := postJSON(t, app, path, body); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed %s: %d (%s)", path, resp.StatusCode, body)
		}
	}
	seedDeviceData(t, post)
	post("/api/sequences", `{"name":"Paging fixture","steps":[{"function_source":"DLL","function_interface":"WIFI_TEST","function_name":"RX_SENSITIVITY","timeout_ms":"1000","flow":"NEXT","on_fail":"FAIL"}]}`)

	endpoints := []struct {
		path        string
		querySuffix string
		total       int64
		globalName  string
	}{
		{path: "/api/functions", total: 2},
		{path: "/api/parameter-definitions", total: 4},
		{path: "/api/criteria-definitions", total: 2},
		{path: "/api/devices/WIFI_TEST/functions", total: 2},
		{path: "/api/troubleshooting", total: 7},
		{path: "/api/examples", total: 1},
		{path: "/api/sequences", total: 1},
		{path: "/api/sequences", querySuffix: "&sort=name", total: 1},
		{path: "/api/devices/WIFI_TEST/parameters", total: 2, globalName: "GLOBAL_P"},
		{path: "/api/devices/WIFI_TEST/criteria", total: 1, globalName: "GLOBAL_C"},
	}
	pages := []struct {
		name     string
		query    string
		page     int64
		pageSize int64
	}{
		{"negative offset wrap", "?page=9223372036854775807", 9223372036854775807, 20},
		{"zero offset wrap", "?page=4611686018427387905&page_size=4", 4611686018427387905, 4},
		{"page beyond int64", "?page=9223372036854775808", 9223372036854775807, 20},
	}
	for _, endpoint := range endpoints {
		t.Run(endpoint.path+endpoint.querySuffix, func(t *testing.T) {
			for _, tc := range pages {
				t.Run(tc.name, func(t *testing.T) {
					resp, body := doRequest(t, app, http.MethodGet, endpoint.path+tc.query+endpoint.querySuffix)
					if resp.StatusCode != http.StatusOK {
						t.Fatalf("status = %d, want 200 (body: %s)", resp.StatusCode, body)
					}
					var got struct {
						Items       []json.RawMessage `json:"items"`
						Page        int64             `json:"page"`
						PageSize    int64             `json:"page_size"`
						Total       int64             `json:"total"`
						GlobalItems []struct {
							Name string `json:"name"`
						} `json:"global_items"`
						GlobalTotal int `json:"global_total"`
					}
					if err := json.Unmarshal([]byte(body), &got); err != nil {
						t.Fatalf("decode list: %v (%s)", err, body)
					}
					if got.Items == nil || len(got.Items) != 0 || got.Page != tc.page || got.PageSize != tc.pageSize || got.Total != endpoint.total {
						t.Fatalf("want empty items, page=%d, page_size=%d, total=%d; got %s", tc.page, tc.pageSize, endpoint.total, body)
					}
					if endpoint.globalName != "" && (got.GlobalTotal != 1 || len(got.GlobalItems) != 1 || got.GlobalItems[0].Name != endpoint.globalName) {
						t.Fatalf("want global_total=1 and global_items=[%s]; got %s", endpoint.globalName, body)
					}
				})
			}
		})
	}
}
