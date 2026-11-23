package api_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// Device browser (read-only): /api/devices lists the distinct device_name
// values (ADR-0009/ADR-0010) with per-device counts; the per-device
// endpoints list the functions linked in the device and the definitions
// scoped to it, split from the global definitions its links resolve to
// (exact-scope-then-global). The reserved segment "_global" addresses the
// blank device.
func seedDeviceData(t *testing.T, post func(path, body string)) {
	t.Helper()
	post("/api/functions", `{"source":"DLL","interface":"WIFI_TEST","name":"TX_MULTI_VERIFICATION"}`)
	post("/api/functions", `{"source":"DLL","interface":"WIFI_TEST","name":"RX_SENSITIVITY"}`)
	post("/api/parameter-definitions", `{"name":"ANT","type":"string","cardinality":"SINGLE","device_name":"WIFI_TEST"}`)
	post("/api/parameter-definitions", `{"name":"BW","type":"double","cardinality":"SINGLE","device_name":"WIFI_TEST"}`)
	post("/api/parameter-definitions", `{"name":"GLOBAL_P","type":"string","cardinality":"SINGLE"}`)
	post("/api/parameter-definitions", `{"name":"OSP_P","type":"string","cardinality":"SINGLE","device_name":"OSP"}`)
	post("/api/criteria-definitions", `{"name":"CRIT_A","type":"string","cardinality":"SINGLE","device_name":"WIFI_TEST"}`)
	post("/api/criteria-definitions", `{"name":"GLOBAL_C","type":"string","cardinality":"SINGLE"}`)
	post("/api/functions/1/parameter-links", `{"name":"ANT","device_name":"WIFI_TEST","position":1}`)
	post("/api/functions/1/parameter-links", `{"name":"GLOBAL_P","device_name":"WIFI_TEST","position":2}`)
	post("/api/functions/2/parameter-links", `{"name":"BW","device_name":"WIFI_TEST","position":1}`)
	post("/api/functions/1/criteria-links", `{"name":"CRIT_A","device_name":"WIFI_TEST"}`)
	post("/api/functions/2/criteria-links", `{"name":"GLOBAL_C","device_name":"WIFI_TEST"}`)
	post("/api/examples", `{"function_source":"DLL","function_name":"TX_MULTI_VERIFICATION","device_name":"WIFI_TEST","script":"DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=15|10000|NEXT|FAIL","meaning":"device test"}`)
}

type deviceSummary struct {
	DeviceName string `json:"device_name"`
	Functions  int    `json:"functions"`
	Parameters int    `json:"parameters"`
	Criteria   int    `json:"criteria"`
	Examples   int    `json:"examples"`
}

func TestDeviceSourcesPersistAndInfer(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	for _, source := range []string{"BUILT-IN", "DLL"} {
		name := "TYPE_" + source
		resp, body := postJSON(t, app, "/api/devices", `{"device_name":"`+name+`","source":"`+source+`"}`)
		if resp.StatusCode != 201 || !strings.Contains(body, `"source":"`+source+`"`) {
			t.Fatalf("create typed device: %d %s", resp.StatusCode, body)
		}
		resp, body = doJSON(t, app, "PUT", "/api/devices/"+name, `{"device_name":"RENAMED_`+source+`"}`)
		if resp.StatusCode != 200 || !strings.Contains(body, `"source":"`+source+`"`) {
			t.Fatalf("rename typed device: %d %s", resp.StatusCode, body)
		}
	}
	for _, input := range []string{
		`{"source":"BUILT-IN","interface":"INFERRED","name":"OPEN"}`,
		`{"source":"BUILT-IN","interface":"MIXED","name":"OPEN"}`,
		`{"source":"DLL","interface":"MIXED","name":"OPEN"}`,
	} {
		if resp, body := postJSON(t, app, "/api/functions", input); resp.StatusCode != 201 {
			t.Fatalf("seed sources: %d %s", resp.StatusCode, body)
		}
	}
	if resp, body := registerDevice(t, app, "UNTYPED"); resp.StatusCode != 201 {
		t.Fatalf("legacy create: %d %s", resp.StatusCode, body)
	}
	resp, body := doRequest(t, app, "GET", "/api/devices")
	var index struct {
		Items []struct {
			DeviceName string   `json:"device_name"`
			Source     *string  `json:"source"`
			Sources    []string `json:"sources"`
		} `json:"items"`
	}
	if resp.StatusCode != 200 || json.Unmarshal([]byte(body), &index) != nil {
		t.Fatalf("index: %d %s", resp.StatusCode, body)
	}
	want := map[string]string{"RENAMED_BUILT-IN": "BUILT-IN", "RENAMED_DLL": "DLL", "INFERRED": "BUILT-IN", "MIXED": "BUILT-IN,DLL", "UNTYPED": ""}
	for _, item := range index.Items {
		expected, found := want[item.DeviceName]
		if !found {
			t.Fatalf("unexpected device: %+v", item)
		}
		if strings.Join(item.Sources, ",") != expected {
			t.Fatalf("source inference: %+v, want %q", item, expected)
		}
		if len(item.Sources) == 1 && (item.Source == nil || *item.Source != expected) {
			t.Fatalf("single source: %+v", item)
		}
		if len(item.Sources) != 1 && item.Source != nil {
			t.Fatalf("ambiguous source: %+v", item)
		}
		delete(want, item.DeviceName)
	}
	if len(want) != 0 {
		t.Fatalf("missing devices: %v", want)
	}
}

func TestDeviceFunctionOnlyAndGlobalDiscovery(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	if resp, body := registerDevice(t, app, "FUNCTION_ONLY"); resp.StatusCode != 201 {
		t.Fatalf("register: %d %s", resp.StatusCode, body)
	}
	for _, input := range []string{
		`{"source":"DLL","interface":" function_only ","name":"SCOPED_FN"}`,
		`{"source":"DLL","interface":"","name":"GLOBAL_FN"}`,
	} {
		if resp, body := postJSON(t, app, "/api/functions", input); resp.StatusCode != 201 {
			t.Fatalf("create function: %d %s", resp.StatusCode, body)
		}
	}
	for device, name := range map[string]string{"FUNCTION_ONLY": "SCOPED_FN", "_global": "GLOBAL_FN"} {
		resp, body := doRequest(t, app, "GET", "/api/devices/"+device+"/functions")
		if resp.StatusCode != 200 || !strings.Contains(body, `"total":1`) || !strings.Contains(body, `"name":"`+name+`"`) {
			t.Fatalf("%s membership: %d %s", device, resp.StatusCode, body)
		}
	}
	resp, body := doRequest(t, app, "GET", "/api/devices")
	var index struct {
		Items []deviceSummary `json:"items"`
	}
	if resp.StatusCode != 200 || json.Unmarshal([]byte(body), &index) != nil || len(index.Items) != 2 {
		t.Fatalf("device index: %d %s", resp.StatusCode, body)
	}
	for _, device := range index.Items {
		if device.Functions != 1 {
			t.Fatalf("function count: %+v", device)
		}
	}
	if resp, body := doJSON(t, app, "DELETE", "/api/devices/FUNCTION_ONLY", `{"confirmation":"DELETE FUNCTION_ONLY"}`); resp.StatusCode != 204 {
		t.Fatalf("delete populated device: %d %s", resp.StatusCode, body)
	}
}

func TestDeviceRenamePreservesStructuredReferences(t *testing.T) {
	for _, registered := range []bool{false, true} {
		t.Run(strconv.FormatBool(registered), func(t *testing.T) {
			app, _ := newDictionaryTestApp(t)
			var original registryDevice
			if registered {
				resp, body := registerDevice(t, app, "WIFI_TEST")
				if resp.StatusCode != 201 || json.Unmarshal([]byte(body), &original) != nil {
					t.Fatalf("register: %d %s", resp.StatusCode, body)
				}
			}
			seedDeviceData(t, func(path, body string) {
				if resp, result := postJSON(t, app, path, body); resp.StatusCode != 201 {
					t.Fatalf("seed %s: %d %s", path, resp.StatusCode, result)
				}
			})
			if resp, body := postJSON(t, app, "/api/sequences", `{"name":"rename sequence","steps":[{"function_source":"DLL","function_interface":"WIFI_TEST","function_name":"TX_MULTI_VERIFICATION","device_name":"WIFI_TEST","parameters":{},"criteria":{},"timeout_ms":"1000","flow":"NEXT","on_fail":"FAIL"}]}`); resp.StatusCode != 201 {
				t.Fatalf("seed sequence: %d %s", resp.StatusCode, body)
			}
			resp, body := doJSON(t, app, "PUT", "/api/devices/WIFI_TEST", `{"device_name":" renamed_device "}`)
			var renamed registryDevice
			if resp.StatusCode != 200 || json.Unmarshal([]byte(body), &renamed) != nil || renamed.DeviceName != "RENAMED_DEVICE" {
				t.Fatalf("rename: %d %s", resp.StatusCode, body)
			}
			if registered && renamed.CreatedAt != original.CreatedAt {
				t.Fatalf("created_at changed: %s -> %s", original.CreatedAt, renamed.CreatedAt)
			}
			for _, path := range []string{"/api/functions/1", "/api/functions/1/parameter-links", "/api/functions/1/criteria-links", "/api/parameter-definitions?search=ANT", "/api/criteria-definitions?search=CRIT_A", "/api/examples", "/api/sequences/1"} {
				resp, result := doRequest(t, app, "GET", path)
				if resp.StatusCode != 200 || strings.Contains(result, `"device_name":"WIFI_TEST"`) || strings.Contains(result, `"interface":"WIFI_TEST"`) || strings.Contains(result, `"function_interface":"WIFI_TEST"`) || !strings.Contains(result, "RENAMED_DEVICE") {
					t.Fatalf("reference %s: %d %s", path, resp.StatusCode, result)
				}
				if path == "/api/examples" && !strings.Contains(result, "DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=15|10000|NEXT|FAIL") {
					t.Fatalf("saved script changed: %s", result)
				}
				if path == "/api/sequences/1" && !strings.Contains(result, `"function_exists":true`) {
					t.Fatalf("sequence reference orphaned: %s", result)
				}
			}
			_, body = doRequest(t, app, "GET", "/api/devices")
			var index struct {
				Items []deviceSummary `json:"items"`
			}
			if err := json.Unmarshal([]byte(body), &index); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, device := range index.Items {
				if device.DeviceName == "WIFI_TEST" {
					t.Fatalf("old device retained: %s", body)
				}
				if device.DeviceName == "RENAMED_DEVICE" {
					found = true
					if device != (deviceSummary{DeviceName: "RENAMED_DEVICE", Functions: 2, Parameters: 2, Criteria: 1, Examples: 1}) {
						t.Fatalf("counts changed: %+v", device)
					}
				}
			}
			if !found {
				t.Fatalf("renamed device missing: %s", body)
			}
		})
	}
}

func TestDeviceRenameRollsBackAllReferencesOnFailure(t *testing.T) {
	app, database := newDictionaryTestApp(t)
	if resp, body := registerDevice(t, app, "WIFI_TEST"); resp.StatusCode != 201 {
		t.Fatalf("register: %d %s", resp.StatusCode, body)
	}
	seedDeviceData(t, func(path, body string) {
		if resp, result := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed: %d %s", resp.StatusCode, result)
		}
	})
	before := map[string]string{}
	for _, path := range []string{"/api/devices", "/api/functions/1", "/api/functions/1/parameter-links", "/api/functions/1/criteria-links", "/api/parameter-definitions", "/api/criteria-definitions", "/api/examples"} {
		resp, body := doRequest(t, app, "GET", path)
		if resp.StatusCode != 200 {
			t.Fatalf("snapshot %s: %d %s", path, resp.StatusCode, body)
		}
		before[path] = body
	}
	if _, err := database.Exec(`ALTER TABLE examples ADD CONSTRAINT reject_test_device_rename CHECK (device_name <> 'RENAMED_DEVICE')`); err != nil {
		t.Fatal(err)
	}
	if resp, body := doJSON(t, app, "PUT", "/api/devices/WIFI_TEST", `{"device_name":"RENAMED_DEVICE"}`); resp.StatusCode != 500 {
		t.Fatalf("forced failure: %d %s", resp.StatusCode, body)
	}
	for path, original := range before {
		resp, body := doRequest(t, app, "GET", path)
		if resp.StatusCode != 200 || body != original {
			t.Fatalf("rollback failed for %s: %d %s; before %s", path, resp.StatusCode, body, original)
		}
	}
}

func TestDeviceRenameConflictsPreserveOriginal(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	seedDeviceData(t, func(path, body string) {
		if resp, result := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed: %d %s", resp.StatusCode, result)
		}
	})
	if resp, body := registerDevice(t, app, "REGISTERED_TARGET"); resp.StatusCode != 201 {
		t.Fatalf("register: %d %s", resp.StatusCode, body)
	}
	for _, target := range []string{"OSP", "REGISTERED_TARGET"} {
		resp, body := doJSON(t, app, "PUT", "/api/devices/WIFI_TEST", `{"device_name":"`+target+`"}`)
		if resp.StatusCode != 409 {
			t.Fatalf("collision %s: %d %s", target, resp.StatusCode, body)
		}
		_, body = doRequest(t, app, "GET", "/api/functions/1")
		if !strings.Contains(body, `"interface":"WIFI_TEST"`) {
			t.Fatalf("original changed: %s", body)
		}
	}
	if resp, body := doJSON(t, app, "PUT", "/api/devices/MISSING", `{"device_name":"UNUSED"}`); resp.StatusCode != 404 {
		t.Fatalf("missing device: %d %s", resp.StatusCode, body)
	}
}

func TestDeviceIndexCounts(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	seedDeviceData(t, func(path, body string) {
		if resp, b := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed %s failed: %s", path, b)
		}
	})

	resp, body := doRequest(t, app, "GET", "/api/devices")
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d (body: %s)", resp.StatusCode, body)
	}
	var got struct {
		Items []deviceSummary `json:"items"`
		Total int             `json:"total"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	if got.Total != 3 {
		t.Fatalf("total = %d, want 3 (body: %s)", got.Total, body)
	}
	want := []deviceSummary{
		{DeviceName: "", Functions: 0, Parameters: 1, Criteria: 1, Examples: 0},
		{DeviceName: "OSP", Functions: 0, Parameters: 1, Criteria: 0, Examples: 0},
		{DeviceName: "WIFI_TEST", Functions: 2, Parameters: 2, Criteria: 1, Examples: 1},
	}
	if len(got.Items) != len(want) {
		t.Fatalf("items = %d, want %d (body: %s)", len(got.Items), len(want), body)
	}
	for i, w := range want {
		if got.Items[i] != w {
			t.Fatalf("items[%d] = %+v, want %+v", i, got.Items[i], w)
		}
	}
}

type deviceFunctionRow struct {
	Name string `json:"name"`
}

func TestDeviceDetailFunctions(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	seedDeviceData(t, func(path, body string) {
		if resp, b := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed %s failed: %s", path, b)
		}
	})

	resp, body := doRequest(t, app, "GET", "/api/devices/WIFI_TEST/functions")
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d (body: %s)", resp.StatusCode, body)
	}
	var got struct {
		Items []deviceFunctionRow `json:"items"`
		Total int                 `json:"total"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	if got.Total != 2 || len(got.Items) != 2 {
		t.Fatalf("total/items = %d/%d, want 2/2 (body: %s)", got.Total, len(got.Items), body)
	}
	names := map[string]bool{}
	for _, it := range got.Items {
		names[it.Name] = true
	}
	if !names["TX_MULTI_VERIFICATION"] || !names["RX_SENSITIVITY"] {
		t.Fatalf("functions missing from device: %v (body: %s)", names, body)
	}

	// A device with no links lists no functions; the reserved "_global"
	// segment addresses the blank device (no blank links seeded → 0).
	if resp, body := doRequest(t, app, "GET", "/api/devices/OSP/functions"); resp.StatusCode != 200 {
		t.Fatalf("OSP status = %d", resp.StatusCode)
	} else if !strings.Contains(body, `"total":0`) {
		t.Fatalf("OSP functions total not 0: %s", body)
	}
	if resp, body := doRequest(t, app, "GET", "/api/devices/_global/functions"); resp.StatusCode != 200 {
		t.Fatalf("_global status = %d", resp.StatusCode)
	} else if !strings.Contains(body, `"total":0`) {
		t.Fatalf("_global functions total not 0: %s", body)
	}
}

func TestDeviceDetailDefinitions(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	seedDeviceData(t, func(path, body string) {
		if resp, b := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed %s failed: %s", path, b)
		}
	})

	resp, body := doRequest(t, app, "GET", "/api/devices/WIFI_TEST/parameters")
	if resp.StatusCode != 200 {
		t.Fatalf("parameters status = %d (body: %s)", resp.StatusCode, body)
	}
	var got struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
		Total       int `json:"total"`
		GlobalItems []struct {
			Name string `json:"name"`
		} `json:"global_items"`
		GlobalTotal int `json:"global_total"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	if got.Total != 2 || len(got.Items) != 2 {
		t.Fatalf("scoped parameters = %d/%d, want 2/2 (body: %s)", got.Total, len(got.Items), body)
	}
	if got.Items[0].Name != "ANT" || got.Items[1].Name != "BW" {
		t.Fatalf("scoped parameters = %v, want [ANT BW]", got.Items)
	}
	// GLOBAL_P has a WIFI_TEST link and no scoped sibling → resolved from
	// Global; GLOBAL_C is a criteria definition, never mixed in here.
	if got.GlobalTotal != 1 || len(got.GlobalItems) != 1 || got.GlobalItems[0].Name != "GLOBAL_P" {
		t.Fatalf("global parameters = %d/%v, want 1/[GLOBAL_P] (body: %s)", got.GlobalTotal, got.GlobalItems, body)
	}

	resp, body = doRequest(t, app, "GET", "/api/devices/WIFI_TEST/criteria")
	if resp.StatusCode != 200 {
		t.Fatalf("criteria status = %d (body: %s)", resp.StatusCode, body)
	}
	var gotCrit struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
		GlobalItems []struct {
			Name string `json:"name"`
		} `json:"global_items"`
	}
	if err := json.Unmarshal([]byte(body), &gotCrit); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	if len(gotCrit.Items) != 1 || gotCrit.Items[0].Name != "CRIT_A" {
		t.Fatalf("scoped criteria = %v, want [CRIT_A] (body: %s)", gotCrit.Items, body)
	}
	if len(gotCrit.GlobalItems) != 1 || gotCrit.GlobalItems[0].Name != "GLOBAL_C" {
		t.Fatalf("global criteria = %v, want [GLOBAL_C] (body: %s)", gotCrit.GlobalItems, body)
	}
}

// ---- device registry (ADR-0010) ------------------------------------------

// registryDevice is the create/GET shape of a registered Device: name plus
// web-native timestamps, nothing else (spec).
type registryDevice struct {
	DeviceName string `json:"device_name"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func registerDevice(t *testing.T, app *fiber.App, name string) (*http.Response, string) {
	t.Helper()
	return postJSON(t, app, "/api/devices", `{"device_name":"`+name+`"}`)
}

func TestDeviceRegistryCreateNormalizesAndLists(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	resp, body := registerDevice(t, app, " wifi_test ")
	if resp.StatusCode != 201 {
		t.Fatalf("create status = %d (body: %s)", resp.StatusCode, body)
	}
	var got registryDevice
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	// ADR-0002 rule: the stored device_name is uppercase.
	if got.DeviceName != "WIFI_TEST" {
		t.Fatalf("device_name = %q, want WIFI_TEST (body: %s)", got.DeviceName, body)
	}
	if got.CreatedAt == "" || got.UpdatedAt == "" {
		t.Fatalf("timestamps missing: %s", body)
	}

	// The registered device appears in the index with zero counts even
	// though no link/definition/example references it (reserved).
	resp, body = doRequest(t, app, "GET", "/api/devices")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	if !strings.Contains(body, `"device_name":"WIFI_TEST"`) || !strings.Contains(body, `"total":1`) {
		t.Fatalf("registry device missing from index: %s", body)
	}
	// The blank device is never registrable and never listed.
	if strings.Contains(body, `"device_name":""`) {
		t.Fatalf("blank device listed: %s", body)
	}
}

func TestDeviceRegistryCreateRejectsDuplicates(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	seedDeviceData(t, func(path, body string) {
		if resp, b := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed %s failed: %s", path, b)
		}
	})

	// WIFI_TEST already exists as a derived device (seeded content).
	resp, body := registerDevice(t, app, "wifi_test")
	if resp.StatusCode != 409 {
		t.Fatalf("derived duplicate status = %d (body: %s)", resp.StatusCode, body)
	}

	// Registering a fresh name twice: second attempt hits the registry.
	if resp, body = registerDevice(t, app, "OSP2"); resp.StatusCode != 201 {
		t.Fatalf("first create status = %d (body: %s)", resp.StatusCode, body)
	}
	resp, body = registerDevice(t, app, "osp2")
	if resp.StatusCode != 409 {
		t.Fatalf("registry duplicate status = %d (body: %s)", resp.StatusCode, body)
	}
}

func TestDeviceRegistryCreateValidation(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	resp, body := postJSON(t, app, "/api/devices", `{"device_name":"   "}`)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("blank name status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, "device_name") {
		t.Fatalf("blank name error not field-keyed: %s", body)
	}
}

func TestDeviceRegistryCreateRejectsInvalidNames(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	cases := []struct {
		why  string
		body string
		want string
	}{
		{"reserved global segment", `{"device_name":"_GLOBAL"}`, "reserved"},
		{"reserved global name", `{"device_name":"GLOBAL"}`, "reserved"},
		{"interior control character", "{\"device_name\":\"A\\u0007B\"}", "device_name"},
		{"too long", `{"device_name":"` + strings.Repeat("A", 257) + `"}`, "device_name"},
	}
	for _, tc := range cases {
		resp, body := postJSON(t, app, "/api/devices", tc.body)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("%s: status = %d (body: %s)", tc.why, resp.StatusCode, body)
		}
		if !strings.Contains(body, tc.want) {
			t.Fatalf("%s: error %q does not mention %q", tc.why, body, tc.want)
		}
	}
}

// Device names must mean the same thing on every path that creates them:
// the content paths (dictionary, examples) apply the same ADR-0002
// normalization as the registry, so a casing twin can never sneak in.
func TestDeviceNameNormalizedAcrossContentPaths(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	// Dictionary create with a lowercase, padded device name.
	resp, body := postJSON(t, app, "/api/parameter-definitions",
		`{"name":"TWIN_P","type":"string","cardinality":"SINGLE","device_name":"  wifi_test  "}`)
	if resp.StatusCode != 201 {
		t.Fatalf("definition create status = %d (body: %s)", resp.StatusCode, body)
	}
	if _, body = doRequest(t, app, "GET", "/api/parameter-definitions?search=TWIN_P"); !strings.Contains(body, `"device_name":"WIFI_TEST"`) {
		t.Fatalf("dictionary device_name not normalized: %s", body)
	}

	// The registry now sees WIFI_TEST as a derived duplicate.
	if resp, body = registerDevice(t, app, "wifi_test"); resp.StatusCode != 409 {
		t.Fatalf("registry duplicate status = %d (body: %s)", resp.StatusCode, body)
	}

	// Dictionary update normalizes too (full-replace contract).
	resp, body = doRequest(t, app, "GET", "/api/parameter-definitions?search=TWIN_P")
	var listed struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(body), &listed); err != nil || len(listed.Items) != 1 {
		t.Fatalf("list decode: %v (body: %s)", err, body)
	}
	defID := listed.Items[0].ID
	resp, body = doJSON(t, app, "PUT", "/api/parameter-definitions/"+strconv.FormatInt(defID, 10),
		`{"name":"TWIN_P","type":"string","cardinality":"SINGLE","device_name":" osp "}`)
	if resp.StatusCode != 200 {
		t.Fatalf("definition update status = %d (body: %s)", resp.StatusCode, body)
	}
	if _, body = doRequest(t, app, "GET", "/api/parameter-definitions?search=TWIN_P"); !strings.Contains(body, `"device_name":"OSP"`) {
		t.Fatalf("dictionary update device_name not normalized: %s", body)
	}

	// Example create normalizes its device field as well.
	if resp, body = postJSON(t, app, "/api/functions", `{"source":"DLL","interface":"WIFI_TEST","name":"TX_MULTI_VERIFICATION"}`); resp.StatusCode != 201 {
		t.Fatalf("function seed status = %d (body: %s)", resp.StatusCode, body)
	}
	resp, body = postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"TX_MULTI_VERIFICATION","device_name":" iot ","script":"DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=1|1000|NEXT|FAIL","meaning":"x"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("example create status = %d (body: %s)", resp.StatusCode, body)
	}
	if _, body = doRequest(t, app, "GET", "/api/devices"); !strings.Contains(body, `"device_name":"IOT"`) {
		t.Fatalf("example device_name not normalized in index: %s", body)
	}
}

// Percent-encoded device segments resolve to the stored name: the client
// encodes devicePath with encodeURIComponent, so "TEST PROGRAM" travels as
// TEST%20PROGRAM on detail GETs and on DELETE.
func TestDevicePathSegmentIsDecoded(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	seedDeviceData(t, func(path, body string) {
		if resp, b := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed %s failed: %s", path, b)
		}
	})

	// A definition scoped to a spaced name gives the device content, so a
	// wrong decode shows up as a zero total.
	if resp, body := postJSON(t, app, "/api/parameter-definitions",
		`{"name":"SPACE_P","type":"string","cardinality":"SINGLE","device_name":"test program"}`); resp.StatusCode != 201 {
		t.Fatalf("seed definition status = %d (body: %s)", resp.StatusCode, body)
	}
	if _, body := doRequest(t, app, "GET", "/api/devices/TEST%20PROGRAM/parameters"); !strings.Contains(body, `"total":1`) {
		t.Fatalf("encoded segment not decoded on GET: %s", body)
	}

	// Registering the same name makes the DELETE path resolvable too.
	if resp, body := registerDevice(t, app, "test program"); resp.StatusCode != 409 {
		t.Fatalf("derived duplicate status = %d (body: %s)", resp.StatusCode, body)
	}

	// A registered spaced name must be deletable through its encoded
	// segment (204, then gone from the list).
	if resp, body := registerDevice(t, app, "space test"); resp.StatusCode != 201 {
		t.Fatalf("register spaced name: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, _ := doJSON(t, app, "DELETE", "/api/devices/SPACE%20TEST", `{"confirmation":"DELETE SPACE TEST"}`); resp.StatusCode != 204 {
		t.Fatalf("encoded DELETE status = %d", resp.StatusCode)
	}
	if _, body := doRequest(t, app, "GET", "/api/devices"); strings.Contains(body, "SPACE TEST") {
		t.Fatalf("deleted device still listed: %s", body)
	}
}

func TestDeviceRegistryDeleteRegisteredAndDerived(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	if resp, body := registerDevice(t, app, "LONELY"); resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}

	// Unregistered / unknown names are 404, not 204.
	if resp, _ := doJSON(t, app, "DELETE", "/api/devices/NEVER_EXISTED", `{"confirmation":"DELETE NEVER_EXISTED"}`); resp.StatusCode != 404 {
		t.Fatalf("unknown delete status = %d", resp.StatusCode)
	}

	if resp, body := doJSON(t, app, "DELETE", "/api/devices/LONELY", `{"confirmation":"DELETE LONELY"}`); resp.StatusCode != 204 {
		t.Fatalf("delete empty registered device: %d (body: %s)", resp.StatusCode, body)
	}
	if _, body := doRequest(t, app, "GET", "/api/devices"); strings.Contains(body, "LONELY") {
		t.Fatalf("deleted device still listed: %s", body)
	}

	// Populated registrations and content-derived devices are both deletable.
	if resp, body := registerDevice(t, app, "BUSY"); resp.StatusCode != 201 {
		t.Fatalf("create BUSY: %d (body: %s)", resp.StatusCode, body)
	}
	resp, body := postJSON(t, app, "/api/parameter-definitions",
		`{"name":"BUSY_P","type":"string","cardinality":"SINGLE","device_name":"BUSY"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("seed BUSY definition: %d (body: %s)", resp.StatusCode, body)
	}
	resp, body = doJSON(t, app, "DELETE", "/api/devices/BUSY", `{"confirmation":"DELETE BUSY"}`)
	if resp.StatusCode != 204 {
		t.Fatalf("delete referenced device status = %d (body: %s)", resp.StatusCode, body)
	}
	if _, body = doRequest(t, app, "GET", "/api/devices"); strings.Contains(body, `"device_name":"BUSY"`) {
		t.Fatalf("deleted device still listed: %s", body)
	}

	if resp, body = postJSON(t, app, "/api/parameter-definitions",
		`{"name":"DERIVED_P","type":"string","cardinality":"SINGLE","device_name":"DERIVED_ONLY"}`); resp.StatusCode != 201 {
		t.Fatalf("seed DERIVED_ONLY definition: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body = doJSON(t, app, "DELETE", "/api/devices/DERIVED_ONLY", `{"confirmation":"DELETE DERIVED_ONLY"}`); resp.StatusCode != 204 {
		t.Fatalf("delete derived-only device status = %d (body: %s)", resp.StatusCode, body)
	}
	if _, body = doRequest(t, app, "GET", "/api/devices"); strings.Contains(body, `"device_name":"DERIVED_ONLY"`) {
		t.Fatalf("derived device still listed after deletion: %s", body)
	}
}

func TestDeviceIndexIncludesRegistered(t *testing.T) {
	app, _ := newDictionaryTestApp(t)
	seedDeviceData(t, func(path, body string) {
		if resp, b := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed %s failed: %s", path, b)
		}
	})
	if resp, body := registerDevice(t, app, "ZZZ_RESERVED"); resp.StatusCode != 201 {
		t.Fatalf("register: %d (body: %s)", resp.StatusCode, body)
	}

	resp, body := doRequest(t, app, "GET", "/api/devices")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	var got struct {
		Items []deviceSummary `json:"items"`
		Total int             `json:"total"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	// Union of derived (seeded: blank, OSP, WIFI_TEST) and registered
	// (ZZZ_RESERVED), name-ordered, Global first.
	want := []deviceSummary{
		{DeviceName: "", Functions: 0, Parameters: 1, Criteria: 1, Examples: 0},
		{DeviceName: "OSP", Functions: 0, Parameters: 1, Criteria: 0, Examples: 0},
		{DeviceName: "WIFI_TEST", Functions: 2, Parameters: 2, Criteria: 1, Examples: 1},
		{DeviceName: "ZZZ_RESERVED", Functions: 0, Parameters: 0, Criteria: 0, Examples: 0},
	}
	if got.Total != 4 || len(got.Items) != 4 {
		t.Fatalf("total/items = %d/%d, want 4/4 (body: %s)", got.Total, len(got.Items), body)
	}
	for i, w := range want {
		if got.Items[i] != w {
			t.Fatalf("items[%d] = %+v, want %+v (body: %s)", i, got.Items[i], w, body)
		}
	}
}

// TestResetLeavesDevicesRegistered: the workbook reset truncates the six
// entity tables only (ADR-0005 precedent); registered devices survive.
func TestResetLeavesDevicesRegistered(t *testing.T) {
	app, _ := newLinksTestApp(t)

	if resp, body := registerDevice(t, app, "RESERVED"); resp.StatusCode != 201 {
		t.Fatalf("register: %d (body: %s)", resp.StatusCode, body)
	}

	ct, upload := multipartBody(t, "file", "wb.xlsx", fixtureBytes(t))
	resp, body := postFile(t, app, "/api/import/reset", ct, upload)
	if resp.StatusCode != 200 {
		t.Fatalf("reset: %d (body: %s)", resp.StatusCode, body)
	}
	resp, body = doRequest(t, app, "GET", "/api/devices")
	if resp.StatusCode != 200 || !strings.Contains(body, `"device_name":"RESERVED"`) {
		t.Fatalf("registered device lost after reset: %d (body: %s)", resp.StatusCode, body)
	}
}

// registryScopeCounts reads the registry size and the distinct non-blank
// device-scope count the content tables currently carry. Equality of the
// two is the ADR-0013 invariant after any import.
func registryScopeCounts(t *testing.T, d *sql.DB) (registered, scopes int) {
	t.Helper()
	err := d.QueryRow(`SELECT
			(SELECT COUNT(*) FROM devices),
			(SELECT COUNT(*) FROM (
				SELECT device_name FROM parameter_links
				UNION SELECT device_name FROM criteria_links
				UNION SELECT device_name FROM parameter_definitions
				UNION SELECT device_name FROM criteria_definitions
				UNION SELECT device_name FROM examples
				UNION SELECT interface FROM functions
			) v WHERE v.device_name <> '')`).Scan(&registered, &scopes)
	if err != nil {
		t.Fatalf("registry/scope counts: %v", err)
	}
	return registered, scopes
}

// TestImportRegistersDeviceScopes: importing the workbook auto-populates
// the devices registry with every distinct non-blank device scope found in
// the content (ADR-0013). The registry mirrors the content after a reset
// import on a fresh schema, a repeat import does not duplicate or disturb
// rows, and the run report reports how many names were newly registered.
func TestImportRegistersDeviceScopes(t *testing.T) {
	app, d := newImportTestAppWithDB(t)

	ct, upload := multipartBody(t, "file", "wb.xlsx", fixtureBytes(t))
	resp, body := postFile(t, app, "/api/import/reset", ct, upload)
	if resp.StatusCode != 200 {
		t.Fatalf("reset import: %d (body: %s)", resp.StatusCode, body)
	}
	var result struct {
		Counts map[string]int `json:"counts"`
	}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("decode import result: %v (body: %s)", err, body)
	}

	registered, scopes := registryScopeCounts(t, d)
	if registered != scopes {
		t.Fatalf("registry (%d) does not mirror content scopes (%d)", registered, scopes)
	}
	if registered == 0 {
		t.Fatal("fixture workbook has no device scopes; fixture drift — update the oracle fixture")
	}
	if got := result.Counts["devices"]; got != registered {
		t.Fatalf("run report devices count = %d, want %d (body: %s)", got, registered, body)
	}

	// A second import without reset re-registers nothing: counts stay put
	// and no device row gains a new created_at.
	var firstTimestamps []string
	rows, err := d.Query(`SELECT name, created_at, updated_at FROM devices ORDER BY name`)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	for rows.Next() {
		var name, created, updated string
		if err := rows.Scan(&name, &created, &updated); err != nil {
			t.Fatalf("scan registry: %v", err)
		}
		firstTimestamps = append(firstTimestamps, name+"\x00"+created+"\x00"+updated)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate registry: %v", err)
	}

	ct, upload = multipartBody(t, "file", "wb.xlsx", fixtureBytes(t))
	if resp, body = postFile(t, app, "/api/import", ct, upload); resp.StatusCode != 200 {
		t.Fatalf("second import: %d (body: %s)", resp.StatusCode, body)
	}
	if registered2, scopes2 := registryScopeCounts(t, d); registered2 != registered || scopes2 != scopes {
		t.Fatalf("second import changed counts: registry %d->%d, scopes %d->%d",
			registered, registered2, scopes, scopes2)
	}

	rows, err = d.Query(`SELECT name, created_at, updated_at FROM devices ORDER BY name`)
	if err != nil {
		t.Fatalf("reread registry: %v", err)
	}
	defer rows.Close()
	var secondTimestamps []string
	for rows.Next() {
		var name, created, updated string
		if err := rows.Scan(&name, &created, &updated); err != nil {
			t.Fatalf("rescan registry: %v", err)
		}
		secondTimestamps = append(secondTimestamps, name+"\x00"+created+"\x00"+updated)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reiterate registry: %v", err)
	}
	if len(firstTimestamps) != len(secondTimestamps) {
		t.Fatalf("second import changed registry size: %d -> %d",
			len(firstTimestamps), len(secondTimestamps))
	}
	for i := range firstTimestamps {
		if firstTimestamps[i] != secondTimestamps[i] {
			t.Fatalf("second import disturbed row: %q -> %q", firstTimestamps[i], secondTimestamps[i])
		}
	}

	// An imported scope is now registered: creating it again conflicts.
	var scopeName string
	if err := d.QueryRow(`SELECT device_name FROM parameter_definitions
			WHERE device_name <> '' LIMIT 1`).Scan(&scopeName); err != nil {
		if err := d.QueryRow(`SELECT device_name FROM parameter_links
				WHERE device_name <> '' LIMIT 1`).Scan(&scopeName); err != nil {
			t.Fatalf("find an imported scope: %v", err)
		}
	}
	if resp, body = registerDevice(t, app, scopeName); resp.StatusCode != 409 {
		t.Fatalf("register imported scope %q: got %d, want 409 (body: %s)", scopeName, resp.StatusCode, body)
	}
}
