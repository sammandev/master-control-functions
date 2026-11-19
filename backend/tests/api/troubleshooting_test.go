package api_test

import (
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
