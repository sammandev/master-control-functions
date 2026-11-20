package api

import (
	"testing"
)

func TestTroubleshootingBodyValidation(t *testing.T) {
	valid := troubleshootingBody{
		Title:        "  TX Power Low Due to Cable Loss  ",
		Category:     "MEASUREMENT",
		Severity:     "WARNING",
		DeviceName:   " wifi_test ",
		FunctionName: " TX_MULTI_VERIFICATION ",
		ErrorCode:    " errcode:104 ",
		Symptom:      "  Measured TX power is 4 dB low across all channels. ",
		RootCause:    "  Path loss table not calibrated. ",
		Steps: []troubleshootingStepBody{
			{
				Title:         " Check golden unit ",
				Instruction:   " Run TX_MULTI_VERIFICATION with golden unit. ",
				ScriptSnippet: "DLL:WIFI_TEST,TX_MULTI_VERIFICATION,...",
				Verification:  " Within 0.5 dB ",
			},
		},
	}
	if fields := valid.validate(); len(fields) != 0 {
		t.Fatalf("unexpected validation errors: %v", fields)
	}
	if valid.Title != "TX Power Low Due to Cable Loss" {
		t.Fatalf("title = %q", valid.Title)
	}
	if valid.Category != "measurement" {
		t.Fatalf("category = %q, want measurement", valid.Category)
	}
	if valid.Severity != "warning" {
		t.Fatalf("severity = %q, want warning", valid.Severity)
	}
	if valid.DeviceName != "WIFI_TEST" {
		t.Fatalf("device_name = %q, want WIFI_TEST", valid.DeviceName)
	}
	if valid.ErrorCode != "ERRCODE:104" {
		t.Fatalf("error_code = %q, want ERRCODE:104", valid.ErrorCode)
	}

	registryManaged := troubleshootingBody{
		Title:      "Registry Managed",
		Category:   "custom-category",
		Severity:   "custom-severity",
		DeviceName: "WIFI_TEST",
		Symptom:    "Custom registry keys are checked against storage later.",
		Steps: []troubleshootingStepBody{
			{Title: "Check", Instruction: "Check it"},
		},
	}
	if fields := registryManaged.validate(); len(fields) != 0 {
		t.Fatalf("shape validation rejected registry-managed keys: %v", fields)
	}

	empty := troubleshootingBody{}
	fields := empty.validate()
	for _, key := range []string{"title", "category", "severity", "symptom", "steps"} {
		if fields[key] == "" {
			t.Fatalf("expected error on %q, got %v", key, fields)
		}
	}

	badStep := troubleshootingBody{
		Title:      "Test Guide",
		Category:   "connection",
		Severity:   "critical",
		DeviceName: "GLOBAL", // reserved word
		Symptom:    "Timeout on SSH login",
		Steps: []troubleshootingStepBody{
			{Title: "", Instruction: ""},
		},
	}
	stepFields := badStep.validate()
	for _, key := range []string{"device_name", "steps.title", "steps.instruction"} {
		if stepFields[key] == "" {
			t.Fatalf("expected error on %q, got %v", key, stepFields)
		}
	}
}
