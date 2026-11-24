package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestListEndpointsReturnEmptyPageForOversizedPage(t *testing.T) {
	app := newTestApp(t)
	seedDeviceData(t, func(path, body string) {
		t.Helper()
		if resp, body := postJSON(t, app, path, body); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed %s: %d (%s)", path, resp.StatusCode, body)
		}
	})

	endpoints := []struct {
		path  string
		total int64
	}{
		{"/api/functions", 2},
		{"/api/parameter-definitions", 4},
		{"/api/criteria-definitions", 2},
		{"/api/devices/WIFI_TEST/functions", 2},
		{"/api/troubleshooting", 7},
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
		t.Run(endpoint.path, func(t *testing.T) {
			for _, tc := range pages {
				t.Run(tc.name, func(t *testing.T) {
					resp, body := doRequest(t, app, http.MethodGet, endpoint.path+tc.query)
					if resp.StatusCode != http.StatusOK {
						t.Fatalf("status = %d, want 200 (body: %s)", resp.StatusCode, body)
					}
					var got struct {
						Items    []json.RawMessage `json:"items"`
						Page     int64             `json:"page"`
						PageSize int64             `json:"page_size"`
						Total    int64             `json:"total"`
					}
					if err := json.Unmarshal([]byte(body), &got); err != nil {
						t.Fatalf("decode list: %v (%s)", err, body)
					}
					if got.Items == nil || len(got.Items) != 0 || got.Page != tc.page || got.PageSize != tc.pageSize || got.Total != endpoint.total {
						t.Fatalf("want empty items, page=%d, page_size=%d, total=%d; got %s", tc.page, tc.pageSize, endpoint.total, body)
					}
				})
			}
		})
	}
}
