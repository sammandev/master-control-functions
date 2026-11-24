package api_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"
)

func TestDeviceFunctionsReadsOnlyRequestedPage(t *testing.T) {
	app, _, counter := newTracedApp(t)
	ids := map[string]int{}
	for _, fn := range []struct{ iface, name string }{
		{"WIFI_TEST", "C_FN"}, {"WIFI_TEST", "A_FN"}, {"WIFI_TEST", "E_FN"},
		{"OSP", "B_FN"}, {"OSP", "D_FN"}, {"OSP", "UNRELATED"},
	} {
		resp, body := postJSON(t, app, "/api/functions",
			fmt.Sprintf(`{"source":"DLL","interface":%q,"name":%q}`, fn.iface, fn.name))
		var created struct {
			ID int `json:"id"`
		}
		if resp.StatusCode != 201 || json.Unmarshal([]byte(body), &created) != nil {
			t.Fatalf("seed %s: %d %s", fn.name, resp.StatusCode, body)
		}
		ids[fn.name] = created.ID
	}
	for _, seed := range []struct{ path, body string }{
		{"/api/parameter-definitions", `{"name":"GLOBAL_P","type":"string","cardinality":"SINGLE"}`},
		{"/api/criteria-definitions", `{"name":"GLOBAL_C","type":"string","cardinality":"SINGLE"}`},
		{fmt.Sprintf("/api/functions/%d/parameter-links", ids["B_FN"]), `{"name":"GLOBAL_P","device_name":"WIFI_TEST","position":1}`},
		{fmt.Sprintf("/api/functions/%d/criteria-links", ids["D_FN"]), `{"name":"GLOBAL_C","device_name":"WIFI_TEST"}`},
	} {
		if resp, body := postJSON(t, app, seed.path, seed.body); resp.StatusCode != 201 {
			t.Fatalf("seed %s: %d %s", seed.path, resp.StatusCode, body)
		}
	}

	for page, want := range map[string][]string{
		"1":                     {"A_FN", "B_FN"},
		"2":                     {"C_FN", "D_FN"},
		"3":                     {"E_FN"},
		"4":                     {},
		"999999999999999999999": {},
	} {
		before := counter.rows.Load()
		resp, body := doRequest(t, app, "GET", "/api/devices/WIFI_TEST/functions?page_size=2&page="+page)
		rows := counter.rows.Load() - before
		if resp.StatusCode != 200 {
			t.Fatalf("page %s: status = %d: %s", page, resp.StatusCode, body)
		}
		var got struct {
			Items []deviceFunctionRow `json:"items"`
			Total int                 `json:"total"`
		}
		if err := json.Unmarshal([]byte(body), &got); err != nil {
			t.Fatal(err)
		}
		names := []string{}
		for _, item := range got.Items {
			names = append(names, item.Name)
		}
		if got.Total != 5 || !slices.Equal(names, want) {
			t.Errorf("page %s: total %d, names %v; want 5, %v", page, got.Total, names, want)
		}
		// One count row plus the page's own rows.
		if wantRows := int64(len(want) + 1); rows != wantRows {
			t.Errorf("page %s: read %d SQL rows, want %d", page, rows, wantRows)
		} else {
			t.Logf("page %s: read %d SQL rows", page, rows)
		}
	}
}
