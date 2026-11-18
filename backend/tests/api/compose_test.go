package api_test

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// seedComposeFunction creates function id 1 (DLL/WIFI_TEST/COMPOSE_FN)
// with five parameter definitions and links in position order plus one
// criteria definition and link. All rows go through the API except the
// FREQ definition: the workbook imports range-form allowed_values
// ("[0..7500]") which the definitions API deliberately refuses to create,
// so that one is seeded raw to mirror importer output. Function id 2
// (DLL/TELNET/COMPOSE_PLAIN) has no links at all.
func seedComposeFunction(t *testing.T, app *fiber.App, d *sql.DB) {
	post := func(target, body string, want int) {
		if resp, b := postJSON(t, app, target, body); resp.StatusCode != want {
			t.Fatalf("%s: %d (body: %s)", target, resp.StatusCode, b)
		}
	}
	post("/api/functions", `{"source":"DLL","interface":"WIFI_TEST","name":"COMPOSE_FN"}`, 201)
	post("/api/parameter-definitions", `{"name":"ANT","type":"int","allowed_values":"{1,2,3,4,7,8,15}","cardinality":"MULTIPLE","default_value":"1"}`, 201)
	post("/api/parameter-definitions", `{"name":"BW","type":"int","allowed_values":"{20,40,80}","cardinality":"SINGLE","default_value":"20"}`, 201)
	post("/api/parameter-definitions", `{"name":"STANDARD","type":"string","cardinality":"SINGLE","default_value":"11AC"}`, 201)
	post("/api/parameter-definitions", `{"name":"TX_TARGET_POWER","type":"int","unit":"dBm","allowed_values":"{0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24}","cardinality":"SINGLE"}`, 201)
	if _, err := d.Exec(`INSERT INTO parameter_definitions
		(name, type, unit, default_value, allowed_values, cardinality, value_pattern, required, created_at, updated_at)
		VALUES ('FREQ', 'int', 'MHz', '2412', '[0..7500]', 'MULTIPLE', '<freq1>,<freq2>…', 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("seed FREQ definition: %v", err)
	}
	post("/api/functions/1/parameter-links", `{"name":"ANT","position":1,"required":true,"device_name":"WIFI_TEST"}`, 201)
	post("/api/functions/1/parameter-links", `{"name":"STANDARD","position":2,"required":true}`, 201)
	post("/api/functions/1/parameter-links", `{"name":"FREQ","position":3,"required":true}`, 201)
	post("/api/functions/1/parameter-links", `{"name":"BW","position":4,"required":true}`, 201)
	post("/api/functions/1/parameter-links", `{"name":"TX_TARGET_POWER","position":5,"required":false}`, 201)
	post("/api/criteria-definitions", `{"name":"POWER","type":"string","cardinality":"SINGLE"}`, 201)
	post("/api/functions/1/criteria-links", `{"name":"POWER","required":false}`, 201)
	post("/api/functions", `{"source":"DLL","interface":"TELNET","name":"COMPOSE_PLAIN"}`, 201)
}

// compose1 posts a compose request and returns status + body.
func compose1(t *testing.T, app *fiber.App, body string) (int, string) {
	resp, body := postJSON(t, app, "/api/functions/1/compose", body)
	return resp.StatusCode, body
}

// decodeScript extracts the emitted script from a 200 compose response.
func decodeScript(t *testing.T, body string) string {
	t.Helper()
	var out struct {
		Script string `json:"script"`
	}
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("decode script: %v (body: %s)", err, body)
	}
	return out.Script
}

// TestComposeKnownFunctionScript: valid input emits the exact 7-segment
// script string per the workbook grammar.
func TestComposeKnownFunctionScript(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedComposeFunction(t, app, d)

	status, body := compose1(t, app, `{
		"parameters":{"ANT":"15","STANDARD":"11AX","FREQ":"6165","BW":"20"},
		"criteria":{"POWER":"5dB"},
		"timeout_ms":"20000","flow":"6175_11AX_BW20_MCS9_CAL","on_fail":"FAIL"}`)
	if status != 200 {
		t.Fatalf("compose: %d (body: %s)", status, body)
	}
	want := "DLLWIFI_TEST|COMPOSE_FN|ANT=15&STANDARD=11AX&FREQ=6165&BW=20|POWER=5dB|20000|6175_11AX_BW20_MCS9_CAL|FAIL"
	if got := decodeScript(t, body); got != want {
		t.Fatalf("script wrong:\n got: %s\nwant: %s", got, want)
	}

	// Unit suffix on an int value is accepted and emitted as typed; a
	// MULTIPLE parameter takes comma-joined values.
	status, body = compose1(t, app, `{
		"parameters":{"ANT":"15,2","STANDARD":"11AX","FREQ":"6165","BW":"20","TX_TARGET_POWER":"15dBm"},
		"criteria":{},"timeout_ms":"10000","flow":"NEXT","on_fail":"FAIL"}`)
	if status != 200 {
		t.Fatalf("compose 2: %d (body: %s)", status, body)
	}
	want = "DLLWIFI_TEST|COMPOSE_FN|ANT=15,2&STANDARD=11AX&FREQ=6165&BW=20&TX_TARGET_POWER=15dBm|NULL|10000|NEXT|FAIL"
	if got := decodeScript(t, body); got != want {
		t.Fatalf("script 2 wrong:\n got: %s\nwant: %s", got, want)
	}
}

// TestComposeValidationRejections: every constraint surfaces as a
// field-level message and NO script key is emitted.
func TestComposeValidationRejections(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedComposeFunction(t, app, d)

	cases := []struct {
		name   string
		body   string
		fields []string // substrings expected in the fields JSON
	}{
		{"missing required", `{"parameters":{"ANT":"15","STANDARD":"X","FREQ":"1"},"criteria":{},"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL"}`, []string{`"BW":"is required"`}},
		{"not in set", `{"parameters":{"ANT":"15","STANDARD":"X","FREQ":"1","BW":"30"},"criteria":{},"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL"}`, []string{`"BW":"must be one of: {20,40,80}"`}},
		{"out of range", `{"parameters":{"ANT":"15","STANDARD":"X","FREQ":"9999","BW":"20"},"criteria":{},"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL"}`, []string{`"FREQ":"must be between 0 and 7500"`}},
		{"bad int", `{"parameters":{"ANT":"abc","STANDARD":"X","FREQ":"1","BW":"20"},"criteria":{},"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL"}`, []string{`"ANT":"must be an integer`}},
		{"forbidden char", `{"parameters":{"ANT":"a|b","STANDARD":"X","FREQ":"1","BW":"20"},"criteria":{},"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL"}`, []string{`"ANT":"must not contain`}},
		{"bad timeout", `{"parameters":{"ANT":"15","STANDARD":"X","FREQ":"1","BW":"20"},"criteria":{},"timeout_ms":"abc","flow":"NEXT","on_fail":"FAIL"}`, []string{`"timeout_ms":"timeout-ms must be a positive integer`}},
		{"missing flow", `{"parameters":{"ANT":"15","STANDARD":"X","FREQ":"1","BW":"20"},"criteria":{},"timeout_ms":"1","on_fail":"FAIL"}`, []string{`"flow":"flow segment is required`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := compose1(t, app, tc.body)
			if status != 422 {
				t.Fatalf("status: %d (body: %s)", status, body)
			}
			if strings.Contains(body, `"script"`) {
				t.Fatalf("script emitted despite invalid input: %s", body)
			}
			if !strings.Contains(body, `"error":"validation failed"`) {
				t.Fatalf("error envelope missing: %s", body)
			}
			for _, f := range tc.fields {
				if !strings.Contains(body, f) {
					t.Fatalf("field message %q missing: %s", f, body)
				}
			}
		})
	}

	// Required criteria missing → criteria.<name> field key.
	if resp, b := postJSON(t, app, "/api/functions/1/criteria-links",
		`{"definition":{"name":"LIMIT","type":"string","cardinality":"SINGLE"},"required":true}`); resp.StatusCode != 201 {
		t.Fatalf("seed required criteria: %d (body: %s)", resp.StatusCode, b)
	}
	status, body := compose1(t, app, `{"parameters":{"ANT":"15","STANDARD":"X","FREQ":"1","BW":"20"},"criteria":{},"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL"}`)
	if status != 422 || !strings.Contains(body, `"criteria.LIMIT":"is required"`) {
		t.Fatalf("required criteria not enforced: %d (body: %s)", status, body)
	}
}

// TestComposeNoLinksFunction: a function without links composes with an
// empty parameter segment and the literal NULL criteria segment.
func TestComposeNoLinksFunction(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedComposeFunction(t, app, d)

	resp, body := postJSON(t, app, "/api/functions/2/compose",
		`{"parameters":{},"criteria":{},"timeout_ms":"10000","flow":"NEXT","on_fail":"FAIL"}`)
	if resp.StatusCode != 200 {
		t.Fatalf("compose: %d (body: %s)", resp.StatusCode, body)
	}
	if got := decodeScript(t, body); got != "DLLTELNET|COMPOSE_PLAIN||NULL|10000|NEXT|FAIL" {
		t.Fatalf("script wrong: %s", got)
	}
}

// TestComposeFieldsEndpoint: field metadata in position order with
// definition constraints attached.
func TestComposeFieldsEndpoint(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedComposeFunction(t, app, d)

	resp, body := doRequest(t, app, "GET", "/api/functions/1/compose/fields")
	if resp.StatusCode != 200 {
		t.Fatalf("fields: %d (body: %s)", resp.StatusCode, body)
	}
	var got struct {
		Function struct {
			Source    string `json:"source"`
			Interface string `json:"interface"`
			Name      string `json:"name"`
		} `json:"function"`
		Parameters []struct {
			Name          string `json:"name"`
			Position      *int64 `json:"position"`
			Required      bool   `json:"required"`
			Type          string `json:"type"`
			AllowedValues string `json:"allowed_values"`
			Cardinality   string `json:"cardinality"`
			Unit          string `json:"unit"`
			Resolved      bool   `json:"resolved"`
		} `json:"parameters"`
		Criteria []struct {
			Name     string `json:"name"`
			Resolved bool   `json:"resolved"`
		} `json:"criteria"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	if got.Function.Source != "DLL" || got.Function.Interface != "WIFI_TEST" || got.Function.Name != "COMPOSE_FN" {
		t.Fatalf("function wrong: %+v", got.Function)
	}
	wantOrder := []string{"ANT", "STANDARD", "FREQ", "BW", "TX_TARGET_POWER"}
	if len(got.Parameters) != len(wantOrder) {
		t.Fatalf("parameter count: %d (body: %s)", len(got.Parameters), body)
	}
	for i, want := range wantOrder {
		if got.Parameters[i].Name != want {
			t.Fatalf("position order wrong at %d: got %s want %s", i, got.Parameters[i].Name, want)
		}
	}
	if p := got.Parameters[0]; !p.Required || p.Cardinality != "MULTIPLE" || p.AllowedValues != "{1,2,3,4,7,8,15}" || !p.Resolved {
		t.Fatalf("ANT metadata wrong: %+v", p)
	}
	if p := got.Parameters[4]; p.Type != "int" || p.Unit != "dBm" {
		t.Fatalf("TX_TARGET_POWER metadata wrong: %+v", p)
	}
	if len(got.Criteria) != 1 || got.Criteria[0].Name != "POWER" || !got.Criteria[0].Resolved {
		t.Fatalf("criteria wrong: %+v", got.Criteria)
	}

	// Unknown function id → 404 on both endpoints, with an error body
	// only: no script, no empty field list may leak behind the 404.
	resp, fbody := doRequest(t, app, "GET", "/api/functions/99/compose/fields")
	if resp.StatusCode != 404 || !strings.Contains(fbody, `"error":"function not found"`) {
		t.Fatalf("fields 404: %d (body: %s)", resp.StatusCode, fbody)
	}
	if strings.Contains(fbody, `"parameters"`) || strings.Contains(fbody, `"script"`) {
		t.Fatalf("404 body leaked field/script content: %s", fbody)
	}
	resp, cbody := postJSON(t, app, "/api/functions/99/compose", `{"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL"}`)
	if resp.StatusCode != 404 || !strings.Contains(cbody, `"error":"function not found"`) {
		t.Fatalf("compose 404: %d (body: %s)", resp.StatusCode, cbody)
	}
	if strings.Contains(cbody, `"script"`) {
		t.Fatalf("404 body leaked a script: %s", cbody)
	}
}

// TestComposeSaveAsExample: the emitted script is a valid Example payload;
// saving through POST /api/examples and reading it back must match the
// composed string exactly.
func TestComposeSaveAsExample(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedComposeFunction(t, app, d)

	status, body := compose1(t, app, `{
		"parameters":{"ANT":"15","STANDARD":"11AX","FREQ":"6165","BW":"20"},
		"criteria":{"POWER":"5dB"},
		"timeout_ms":"20000","flow":"NEXT","on_fail":"FAIL"}`)
	if status != 200 {
		t.Fatalf("compose: %d (body: %s)", status, body)
	}
	script := decodeScript(t, body)
	want := "DLLWIFI_TEST|COMPOSE_FN|ANT=15&STANDARD=11AX&FREQ=6165&BW=20|POWER=5dB|20000|NEXT|FAIL"
	if script != want {
		t.Fatalf("compose: got %q want %q", script, want)
	}

	// The save-as-example flow posts the composed script (editable in the
	// UI before save) to the existing examples endpoint.
	saveBody, err := json.Marshal(map[string]string{
		"function_source": "DLL",
		"function_name":   "COMPOSE_FN",
		"device_name":     "WIFI_TEST",
		"script":          script,
		"meaning":         "composed from the form",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, sbody := postJSON(t, app, "/api/examples", string(saveBody))
	if resp.StatusCode != 201 {
		t.Fatalf("save as example: %d (body: %s)", resp.StatusCode, sbody)
	}
	resp, sbody = doRequest(t, app, "GET", "/api/examples/1")
	if resp.StatusCode != 200 {
		t.Fatalf("read example: %d (body: %s)", resp.StatusCode, sbody)
	}
	var saved struct {
		Script string `json:"script"`
	}
	if err := json.Unmarshal([]byte(sbody), &saved); err != nil {
		t.Fatalf("decode saved: %v (body: %s)", err, sbody)
	}
	if saved.Script != script {
		t.Fatalf("saved example script mismatch:\n got: %s\nwant: %s", saved.Script, script)
	}
}
