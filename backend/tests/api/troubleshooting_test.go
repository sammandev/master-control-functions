package api_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestTroubleshootingCRUDAndMultiSolutionPerErrorCode(t *testing.T) {
	app, _ := newLinksTestApp(t)

	// Seeded guides from 0009_troubleshooting.sql should be present, including 2 guides for ERRCODE:104.
	resp, body := doRequest(t, app, "GET", "/api/troubleshooting?error_code=ERRCODE:104")
	if resp.StatusCode != 200 {
		t.Fatalf("list by error_code: %d (%s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"total":2`) {
		t.Fatalf("expected 2 seeded guides for ERRCODE:104, got: %s", body)
	}

	// Create a 3rd guide for the same ERRCODE:104 (different manifestation/root cause).
	payload1 := `{
		"title": "TX_MULTI_VERIFICATION Fails on 6GHz Due to Regulatory Country Code",
		"category": "measurement",
		"severity": "critical",
		"device_name": "wifi_test",
		"function_name": "TX_MULTI_VERIFICATION",
		"error_code": "ERRCODE:104",
		"symptom": "6GHz channels return -40 dBm while 2.4GHz and 5GHz pass.",
		"root_cause": "DUT country code in nvram locks 6GHz band transmit chains.",
		"steps": [
			{
				"title": "Set regulatory country code before DUT_INITIALIZE",
				"instruction": "Configure country code in RD_SETTING.ini so 6GHz channels are unlocked.",
				"script_snippet": "BUILT-IN:SSH,SENDSTR_AND_WAITSTR,wl country US,*V:SSH-PROMPT#,,5000,NEXT,FAIL",
				"verification": "wl channels includes 6GHz UNII-5 frequencies."
			}
		]
	}`
	resp, createdBody := postJSON(t, app, "/api/troubleshooting", payload1)
	if resp.StatusCode != 201 {
		t.Fatalf("create guide: %d (%s)", resp.StatusCode, createdBody)
	}
	if !strings.Contains(createdBody, `"device_name":"WIFI_TEST"`) {
		t.Fatalf("expected normalized WIFI_TEST device_name: %s", createdBody)
	}

	// Now ERRCODE:104 should return 3 guides with related_count >= 2.
	resp, listBody := doRequest(t, app, "GET", "/api/troubleshooting?error_code=ERRCODE:104")
	if resp.StatusCode != 200 || !strings.Contains(listBody, `"total":3`) {
		t.Fatalf("expected 3 guides for ERRCODE:104, got %d (%s)", resp.StatusCode, listBody)
	}
}

func TestTroubleshootingRelatedEndpointReturnsMoreThanOneHundredSiblings(t *testing.T) {
	app, d := newLinksTestApp(t)

	for i := 0; i < 101; i++ {
		_, err := d.Exec(`INSERT INTO troubleshooting_guides
			(title, category, severity, device_name, function_name, error_code, symptom, root_cause, created_at, updated_at)
			VALUES ($1, 'measurement', 'warning', '', 'BULK_RELATED', 'ERRCODE:BULK', $2, '', '2026-11-19T00:00:00Z', '2026-11-19T00:00:00Z')`,
			fmt.Sprintf("Bulk sibling %03d", i), fmt.Sprintf("Symptom %03d", i))
		if err != nil {
			t.Fatalf("seed sibling %d: %v", i, err)
		}
	}

	var sourceID int64
	if err := d.QueryRow(`SELECT id FROM troubleshooting_guides WHERE error_code = 'ERRCODE:BULK' ORDER BY id LIMIT 1`).Scan(&sourceID); err != nil {
		t.Fatalf("find source guide: %v", err)
	}
	resp, body := doRequest(t, app, "GET", fmt.Sprintf("/api/troubleshooting/%d/related", sourceID))
	if resp.StatusCode != 200 {
		t.Fatalf("related endpoint: %d (%s)", resp.StatusCode, body)
	}
	var result struct {
		Items []json.RawMessage `json:"items"`
		Total int               `json:"total"`
	}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("decode related response: %v", err)
	}
	if result.Total != 100 || len(result.Items) != 100 {
		t.Fatalf("related result = total %d, items %d; want 100 excluding source", result.Total, len(result.Items))
	}
}
