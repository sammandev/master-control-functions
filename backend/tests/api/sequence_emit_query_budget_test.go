package api_test

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"webprogfndb/internal/api"
)

func TestSequenceEmitQueryBudget(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedSequenceFunctions(t, app)
	if resp, body := postJSON(t, app, "/api/functions",
		`{"source":"DLL","interface":"GPIB","name":"FN_A"}`); resp.StatusCode != 201 {
		t.Fatalf("seed FN_A on GPIB: %d (body: %s)", resp.StatusCode, body)
	}
	stepOnGPIB := strings.Replace(seqStepA, `"TELNET"`, `"GPIB"`, 1)
	stepWithOtherValue := strings.Replace(seqStepA, `"FREQ":"10"`, `"FREQ":"20"`, 1)
	if resp, body := postJSON(t, app, "/api/sequences", `{"name":"repeat","steps":[`+
		strings.Join([]string{seqStepA, stepOnGPIB, stepWithOtherValue, stepOnGPIB}, ",")+`]}`); resp.StatusCode != 201 {
		t.Fatalf("create sequence: %d (body: %s)", resp.StatusCode, body)
	}

	conn, err := d.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var config *pgx.ConnConfig
	err = conn.Raw(func(driverConn any) error {
		config = driverConn.(*stdlib.Conn).Conn().Config()
		return nil
	})
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	counter := &queryCounter{}
	config.Tracer = counter
	traced := stdlib.OpenDB(*config)
	t.Cleanup(func() { traced.Close() })
	emitApp := api.New(traced)

	const wantText = `"text":"DLLTELNET|FN_A|FREQ=10|NULL|1000|NEXT|FAIL\nDLLGPIB|FN_A|FREQ=10|NULL|1000|NEXT|FAIL\nDLLTELNET|FN_A|FREQ=20|NULL|1000|NEXT|FAIL\nDLLGPIB|FN_A|FREQ=10|NULL|1000|NEXT|FAIL"`
	for request := 1; request <= 2; request++ {
		expectedText := wantText
		if request == 2 {
			if resp, body := doJSON(t, app, "PUT", "/api/parameter-definitions/1",
				`{"name":"FREQ","type":"int","allowed_values":"{20,30}","cardinality":"SINGLE"}`); resp.StatusCode != 200 {
				t.Fatalf("update FREQ definition: %d (body: %s)", resp.StatusCode, body)
			}
			expectedText = `"text":"DLLTELNET|FN_A|FREQ=20|NULL|1000|NEXT|FAIL"`
		}
		before := counter.calls.Load()
		resp, body := doRequest(t, emitApp, "POST", "/api/sequences/1/emit")
		calls := counter.calls.Load() - before
		if resp.StatusCode != 200 || !strings.Contains(body, expectedText) {
			t.Fatalf("emit %d: %d (body: %s)", request, resp.StatusCode, body)
		}
		if request == 2 && strings.Count(body, `"FREQ":"must be one of: {20,30}"`) != 3 {
			t.Fatalf("emit %d: want validation errors on all three steps with FREQ=10 (body: %s)", request, body)
		}
		const wantCalls = 1 + 2*3 // step list plus three reads per distinct Function
		if calls != wantCalls {
			t.Errorf("emit %d: got %d SQL queries, want %d", request, calls, wantCalls)
		} else {
			t.Logf("emit %d: got %d SQL queries", request, calls)
		}
	}
}
