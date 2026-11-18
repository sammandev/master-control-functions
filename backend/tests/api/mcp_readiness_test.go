package api_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type readinessBody struct {
	Status  string `json:"status"`
	Grammar *struct {
		Encoding string `json:"encoding"`
		Scope    string `json:"scope"`
	} `json:"grammar"`
	Line struct {
		Ready  bool `json:"ready"`
		Issues []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"issues"`
	} `json:"line"`
	Bundle struct {
		Ready bool `json:"ready"`
	} `json:"bundle"`
	Warnings []struct {
		Field string `json:"field"`
		Code  string `json:"code"`
	} `json:"warnings"`
}

func createFunctionID(t *testing.T, app *fiber.App, body string) string {
	t.Helper()
	resp, out := postJSON(t, app, "/api/functions", body)
	if resp.StatusCode != 201 {
		t.Fatalf("create function: %d %s", resp.StatusCode, out)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("decode function: %s", out)
	}
	return strconv.FormatInt(created.ID, 10)
}

func readiness(t *testing.T, app *fiber.App, id string) readinessBody {
	t.Helper()
	resp, out := doRequest(t, app, "GET", "/api/functions/"+id+"/mcp-readiness")
	if resp.StatusCode != 200 {
		t.Fatalf("readiness: %d %s", resp.StatusCode, out)
	}
	var r readinessBody
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("decode readiness: %s", out)
	}
	return r
}

func TestMCPReadinessUsesSeededGrammar(t *testing.T) {
	app, d := newDictionaryTestApp(t)

	var grammars, audits int
	if err := d.QueryRow(`SELECT COUNT(*) FROM function_grammars`).Scan(&grammars); err != nil || grammars != 12 {
		t.Fatalf("seeded grammars = %d (%v), want 12", grammars, err)
	}
	if err := d.QueryRow(`SELECT COUNT(*) FROM script_key_audits WHERE function_name = 'TX_MULTI_VERIFICATION'`).Scan(&audits); err != nil || audits != 17 {
		t.Fatalf("TX_MULTI_VERIFICATION audits = %d (%v), want 17", audits, err)
	}

	id := createFunctionID(t, app, `{"source":"DLL","interface":"WIFI_TEST","name":"TX_MULTI_VERIFICATION","convention":2}`)
	r := readiness(t, app, id)
	if r.Grammar == nil || r.Grammar.Encoding != "named" || r.Grammar.Scope != "line" {
		t.Fatalf("seeded grammar not applied: %+v", r.Grammar)
	}
	if r.Status != "ready" {
		t.Fatalf("function without fields should be ready, got %+v", r)
	}

	// An unaudited key blocks composition until its mapping is audited.
	link := `{"definition":{"name":"MYSTERY","type":"int","value_pattern":"<v>"}}`
	if resp, out := postJSON(t, app, "/api/functions/"+id+"/parameter-links", link); resp.StatusCode != 201 {
		t.Fatalf("attach: %d %s", resp.StatusCode, out)
	}
	r = readiness(t, app, id)
	if r.Status != "not_ready" || len(r.Line.Issues) != 1 || r.Line.Issues[0].Field != "params.MYSTERY" {
		t.Fatalf("unaudited key not reported: %+v", r)
	}
	key := `{"kind":"parameter","name":"MYSTERY","scope":"bundle","audited_by":"tester"}`
	if resp, out := doJSON(t, app, "PUT", "/api/functions/"+id+"/script-keys", key); resp.StatusCode != 200 {
		t.Fatalf("audit key: %d %s", resp.StatusCode, out)
	}
	if r = readiness(t, app, id); r.Status != "bundle_ready" {
		t.Fatalf("bundle-scoped key should make the function bundle_ready: %+v", r)
	}
	key = strings.Replace(key, `"bundle"`, `"line"`, 1)
	doJSON(t, app, "PUT", "/api/functions/"+id+"/script-keys", key)
	if r = readiness(t, app, id); r.Status != "ready" {
		t.Fatalf("line-scoped key should make the function ready: %+v", r)
	}

	// Renaming the function carries its grammar and key audits.
	rename := `{"source":"DLL","interface":"WIFI_TEST","name":"TX_RENAMED","convention":2}`
	if resp, out := doJSON(t, app, "PUT", "/api/functions/"+id, rename); resp.StatusCode != 200 {
		t.Fatalf("rename: %d %s", resp.StatusCode, out)
	}
	var moved int
	d.QueryRow(`SELECT COUNT(*) FROM script_key_audits WHERE function_name = 'TX_RENAMED'`).Scan(&moved)
	if moved != 18 {
		t.Fatalf("renamed audits = %d, want 18", moved)
	}
	resp, out := doRequest(t, app, "GET", "/api/functions/"+id+"/grammar")
	if resp.StatusCode != 200 || !strings.Contains(out, `"encoding":"named"`) {
		t.Fatalf("grammar did not follow rename: %d %s", resp.StatusCode, out)
	}

	if resp, out := doJSON(t, app, "DELETE", "/api/functions/"+id+"/grammar", ""); resp.StatusCode != 204 {
		t.Fatalf("delete grammar: %d %s", resp.StatusCode, out)
	}
	if r = readiness(t, app, id); r.Status != "not_ready" || r.Grammar != nil {
		t.Fatalf("function without grammar must be not_ready: %+v", r)
	}
	body := `{"encoding":"named","scope":"sideways","audited_by":""}`
	if resp, out := doJSON(t, app, "PUT", "/api/functions/"+id+"/grammar", body); resp.StatusCode != 422 ||
		!strings.Contains(out, `"scope"`) || !strings.Contains(out, `"audited_by"`) {
		t.Fatalf("invalid grammar accepted: %d %s", resp.StatusCode, out)
	}
}

func TestDeviceRenameCarriesGrammarAndKeyAudits(t *testing.T) {
	app, d := newDictionaryTestApp(t)
	id := createFunctionID(t, app, `{"source":"DLL","interface":"WIFI_TEST","name":"TX_MULTI_VERIFICATION","convention":2}`)

	if resp, out := doJSON(t, app, "PUT", "/api/devices/WIFI_TEST", `{"device_name":"WIFI_BENCH"}`); resp.StatusCode != 200 {
		t.Fatalf("rename device: %d %s", resp.StatusCode, out)
	}
	var stale, moved int
	d.QueryRow(`SELECT COUNT(*) FROM function_grammars WHERE function_interface = 'WIFI_TEST'`).Scan(&stale)
	d.QueryRow(`SELECT COUNT(*) FROM script_key_audits WHERE function_interface = 'WIFI_BENCH' AND function_name = 'TX_MULTI_VERIFICATION'`).Scan(&moved)
	if stale != 0 || moved != 17 {
		t.Fatalf("grammar rows left on old device = %d, audits moved = %d (want 0, 17)", stale, moved)
	}
	if r := readiness(t, app, id); r.Grammar == nil || r.Status != "ready" {
		t.Fatalf("grammar did not follow device rename: %+v", r)
	}
}

func TestMCPReadinessWarnsAboutExamplesAndAmbiguousKeys(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	id := createFunctionID(t, app, `{"source":"BUILT-IN","interface":"SSH","name":"OPEN","convention":1}`)
	example := `{"function_source":"BUILT-IN","function_name":"OPEN","script":"SSH|OPEN|host"}`
	resp, out := postJSON(t, app, "/api/examples", example)
	if resp.StatusCode != 201 || !strings.Contains(out, `"code":"field_count"`) {
		t.Fatalf("example save should succeed with a field_count warning: %d %s", resp.StatusCode, out)
	}
	good := `{"function_source":"BUILT-IN","function_name":"OPEN","script":"SSH|OPEN|host|NULL|5"}`
	if resp, out := postJSON(t, app, "/api/examples", good); resp.StatusCode != 201 || strings.Contains(out, "script_issues") {
		t.Fatalf("valid example flagged: %d %s", resp.StatusCode, out)
	}

	createFunctionID(t, app, `{"source":"BUILT-IN","interface":"TELNET","name":"OPEN","convention":1}`)
	r := readiness(t, app, id)
	codes := map[string]bool{}
	for _, w := range r.Warnings {
		codes[w.Code] = true
	}
	if !codes["field_count"] || !codes["ambiguous_function_key"] {
		t.Fatalf("expected example and ambiguity warnings, got %+v", r.Warnings)
	}

	resp, out = doRequest(t, app, "GET", "/api/mcp-readiness")
	if resp.StatusCode != 200 || !strings.Contains(out, `"total":2`) {
		t.Fatalf("summary: %d %s", resp.StatusCode, out)
	}
}

func TestMCPViewsInferInterfaceOnlyWhenUnambiguous(t *testing.T) {
	app, d := newDictionaryTestApp(t)

	a := createFunctionID(t, app, `{"source":"DLL","interface":"WIFI_TEST","name":"ONLY","convention":2}`)
	postJSON(t, app, "/api/functions/"+a+"/parameter-links", `{"definition":{"name":"ANT","type":"int","value_pattern":"<a>"}}`)
	b := createFunctionID(t, app, `{"source":"DLL","interface":"WIFI_TEST","name":"TWIN","convention":2}`)
	createFunctionID(t, app, `{"source":"DLL","interface":"IOT","name":"TWIN","convention":2}`)
	postJSON(t, app, "/api/functions/"+b+"/parameter-links", `{"name":"ANT"}`)

	rows, err := d.Query(`SELECT function_name, function_interface FROM mcp_v1_parameter_links ORDER BY function_name`)
	if err != nil {
		t.Fatalf("query view: %v", err)
	}
	defer rows.Close()
	got := map[string]string{}
	for rows.Next() {
		var name, iface string
		rows.Scan(&name, &iface)
		got[name] = iface
	}
	if got["ONLY"] != "WIFI_TEST" || got["TWIN"] != "" {
		t.Fatalf("inferred interfaces wrong: %v", got)
	}
	var encoding string
	if err := d.QueryRow(`SELECT grammar_encoding FROM mcp_v1_functions WHERE name = 'ONLY'`).Scan(&encoding); err != nil || encoding != "" {
		t.Fatalf("unaudited function grammar_encoding = %q (%v)", encoding, err)
	}
}

func TestDefinitionValuePatternAndConstraintRules(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	for _, tc := range []struct {
		body   string
		status int
		field  string
	}{
		{`{"name":"A","type":"int","value_pattern":"[<a>]"}`, 422, "value_pattern"},
		{`{"name":"B","type":"int","cardinality":"SINGLE","value_pattern":"<a>,<b>,..."}`, 422, "value_pattern"},
		{`{"name":"C","type":"double","allowed_values":"[-10..30]","default_value":"40"}`, 422, "default_value"},
		{`{"name":"D","type":"string","allowed_values":"[5..1]"}`, 422, "allowed_values"},
		{`{"name":"E","type":"double","allowed_values":"[-10..30]","default_value":"5","value_pattern":"<v>"}`, 201, ""},
		{`{"name":"F","type":"string","allowed_values":"[A-Z0-9]","default_value":"AB1"}`, 201, ""},
		{`{"name":"G","type":"string","allowed_values":"{}"}`, 201, ""},
	} {
		resp, out := postJSON(t, app, "/api/parameter-definitions", tc.body)
		if resp.StatusCode != tc.status || (tc.field != "" && !strings.Contains(out, `"`+tc.field+`"`)) {
			t.Fatalf("%s: %d %s", tc.body, resp.StatusCode, out)
		}
	}
}
