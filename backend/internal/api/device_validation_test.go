package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestFunctionDeviceValidation(t *testing.T) {
	for _, testCase := range []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: "   ", want: ""},
		{input: " wifi_test ", want: "WIFI_TEST"},
	} {
		body := functionBody{Source: "DLL", Interface: testCase.input, Name: "TEST"}
		if fields := body.validate(); len(fields) != 0 {
			t.Fatalf("interface %q: %v", testCase.input, fields)
		}
		if body.Interface != testCase.want {
			t.Fatalf("interface = %q, want %q", body.Interface, testCase.want)
		}
	}
	for _, name := range []string{"GLOBAL", "_global", "BAD\x00NAME"} {
		body := functionBody{Source: "DLL", Interface: name, Name: "TEST"}
		if body.validate()["interface"] == "" {
			t.Fatalf("invalid device %q accepted", name)
		}
	}
}

func TestSequenceBlankDeviceValidation(t *testing.T) {
	body := sequenceBody{Name: "Global test", Steps: []sequenceStepBody{{FunctionSource: "DLL", FunctionName: "TEST"}}}
	if fields := body.validate(); len(fields) != 0 {
		t.Fatalf("blank interface rejected: %v", fields)
	}
	body.Steps[0].FunctionInterface = " wifi_test "
	body.Steps[0].DeviceName = " wifi_test "
	if fields := body.validate(); len(fields) != 0 || body.Steps[0].FunctionInterface != "WIFI_TEST" || body.Steps[0].DeviceName != "WIFI_TEST" {
		t.Fatalf("device normalization failed: %v %+v", fields, body.Steps[0])
	}
	body.Steps[0].DeviceName = "_global"
	if body.validate()["steps.device_name"] == "" {
		t.Fatal("reserved sequence device accepted")
	}
}

func TestCreateDeviceSourceValidation(t *testing.T) {
	for _, source := range []string{"BUILT-IN", "DLL"} {
		got, problem := (deviceBody{Source: &source}).deviceSource()
		if problem != "" || !got.Valid || got.String != source {
			t.Fatalf("source %q: %+v %q", source, got, problem)
		}
	}
	if got, problem := (deviceBody{}).deviceSource(); got.Valid || problem != "" {
		t.Fatalf("legacy payload: %+v %q", got, problem)
	}
	app := fiber.New()
	app.Post("/api/devices", createDevice(nil))
	for _, source := range []string{"", "dll", "OTHER"} {
		request := httptest.NewRequest("POST", "/api/devices", strings.NewReader(`{"device_name":"TEST","source":"`+source+`"}`))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 422 {
			t.Fatalf("source %q: status %d", source, response.StatusCode)
		}
	}
}

func TestRenameDeviceValidation(t *testing.T) {
	app := fiber.New()
	app.Put("/api/devices/:device", renameDevice(nil))
	for _, testCase := range []struct {
		path, body string
		status     int
	}{
		{path: "_global", body: `{"device_name":"NEW"}`, status: 422},
		{path: "OLD", body: `{"device_name":""}`, status: 422},
		{path: "OLD", body: `{"device_name":" global "}`, status: 422},
		{path: "OLD", body: `{"device_name":"_global"}`, status: 422},
		{path: "OLD", body: `{"device_name":"BAD\u0000NAME"}`, status: 422},
		{path: "OLD", body: `{"device_name":"` + strings.Repeat("A", 257) + `"}`, status: 422},
		{path: "OLD", body: `{`, status: 400},
	} {
		request := httptest.NewRequest("PUT", "/api/devices/"+testCase.path, strings.NewReader(testCase.body))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}

		response.Body.Close()
		if response.StatusCode != testCase.status {
			t.Fatalf("%s %s: status %d, want %d", testCase.path, testCase.body, response.StatusCode, testCase.status)
		}
	}
}

func TestDeleteDeviceConfirmationValidation(t *testing.T) {
	app := fiber.New()
	app.Delete("/api/devices/:device", deleteDevice(nil))
	for _, testCase := range []struct {
		path, body string
		status     int
	}{
		{path: "_global", body: `{"confirmation":"DELETE "}`, status: 422},
		{path: "TEST", body: `{}`, status: 422},
		{path: "TEST", body: `{"confirmation":"delete TEST"}`, status: 422},
		{path: "TEST", body: `{"confirmation":"DELETE test"}`, status: 422},
		{path: "TEST", body: `{"confirmation":"DELETE OTHER"}`, status: 422},
		{path: "TEST", body: `{"confirmation":" DELETE TEST"}`, status: 422},
		{path: "TEST", body: `{"confirmation":"DELETE TEST "}`, status: 422},
		{path: "TEST", body: `{`, status: 400},
		{path: "TEST", body: ``, status: 400},
	} {
		request := httptest.NewRequest("DELETE", "/api/devices/"+testCase.path, strings.NewReader(testCase.body))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != testCase.status {
			t.Fatalf("%s %s: status %d, want %d", testCase.path, testCase.body, response.StatusCode, testCase.status)
		}
	}
}
