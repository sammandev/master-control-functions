package api_test

import (
	"strings"
	"testing"
)

func TestDeviceDeleteCascadesAndPreservesSharedData(t *testing.T) {
	for _, registered := range []bool{false, true} {
		t.Run(map[bool]string{false: "derived", true: "registered"}[registered], func(t *testing.T) {
			app, database := newDictionaryTestApp(t)
			if registered {
				if resp, body := registerDevice(t, app, "WIFI_TEST"); resp.StatusCode != 201 {
					t.Fatalf("register: %d %s", resp.StatusCode, body)
				}
			}
			seedDeviceData(t, func(path, body string) {
				if resp, result := postJSON(t, app, path, body); resp.StatusCode != 201 {
					t.Fatalf("seed %s: %d %s", path, resp.StatusCode, result)
				}
			})
			if resp, body := postJSON(t, app, "/api/functions", `{"source":"DLL","interface":"OTHER","name":"TX_MULTI_VERIFICATION"}`); resp.StatusCode != 201 {
				t.Fatalf("seed shared function: %d %s", resp.StatusCode, body)
			}
			for _, table := range []string{"parameter_links", "criteria_links", "examples"} {
				values := ` (function_source, function_name, device_name, name, created_at, updated_at)
					VALUES ('DLL', $1, $2, 'GLOBAL_P', 'old', 'old')`
				if table == "examples" {
					values = ` (function_source, function_name, device_name, script, created_at, updated_at)
						VALUES ('DLL', $1, $2, 'unchanged script', 'old', 'old')`
				}
				for _, function := range []string{"TX_MULTI_VERIFICATION", "RX_SENSITIVITY"} {
					for _, device := range []string{"OTHER", ""} {
						if _, err := database.Exec(`INSERT INTO `+table+values, function, device); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			for _, statement := range []string{
				`INSERT INTO function_grammars
					(function_source, function_interface, function_name, encoding, scope, audited_by, audited_at, created_at, updated_at)
					VALUES ('DLL', 'WIFI_TEST', 'DELETE_ONLY', 'named', 'line', 'test', 'old', 'old', 'old'),
					       ('DLL', 'OTHER', 'TX_MULTI_VERIFICATION', 'named', 'line', 'test', 'old', 'old', 'old')`,
				`INSERT INTO script_key_audits
					(function_source, function_interface, function_name, kind, name, scope, audited_by, audited_at, created_at, updated_at)
					VALUES ('DLL', 'WIFI_TEST', 'DELETE_ONLY', 'parameter', 'ANT', 'line', 'test', 'old', 'old', 'old'),
					       ('DLL', 'OTHER', 'TX_MULTI_VERIFICATION', 'parameter', 'ANT', 'line', 'test', 'old', 'old', 'old')`,
				`INSERT INTO sequences (name, created_at, updated_at)
					VALUES ('mixed', 'old', 'old'), ('untouched', 'old', 'old'), ('device-only', 'old', 'old')`,
				`INSERT INTO sequence_steps
					(sequence_id, position, function_source, function_interface, function_name, device_name)
					VALUES (1, 1, 'DLL', 'WIFI_TEST', 'TX_MULTI_VERIFICATION', 'OTHER'),
					       (1, 2, 'DLL', 'OTHER', 'TX_MULTI_VERIFICATION', 'OTHER'),
					       (1, 3, 'DLL', 'OTHER', 'TX_MULTI_VERIFICATION', 'WIFI_TEST'),
					       (1, 4, 'DLL', 'OTHER', 'TX_MULTI_VERIFICATION', ''),
					       (2, 1, 'DLL', 'OTHER', 'TX_MULTI_VERIFICATION', 'OTHER'),
					       (3, 1, 'DLL', 'WIFI_TEST', 'RX_SENSITIVITY', '')`,
			} {
				if _, err := database.Exec(statement); err != nil {
					t.Fatal(err)
				}
			}

			// A mismatched confirmation must not change any populated content.
			if resp, body := doJSON(t, app, "DELETE", "/api/devices/WIFI_TEST", `{"confirmation":"DELETE OTHER"}`); resp.StatusCode != 422 {
				t.Fatalf("confirmation: %d %s", resp.StatusCode, body)
			}
			var count int
			if err := database.QueryRow(`SELECT COUNT(*) FROM functions WHERE interface = 'WIFI_TEST'`).Scan(&count); err != nil || count != 2 {
				t.Fatalf("invalid confirmation changed data: count=%d err=%v", count, err)
			}
			if resp, body := doJSON(t, app, "DELETE", "/api/devices/WIFI_TEST", `{"confirmation":"DELETE WIFI_TEST"}`); resp.StatusCode != 204 {
				t.Fatalf("delete: %d %s", resp.StatusCode, body)
			}
			for _, table := range []string{"devices", "functions", "parameter_definitions", "criteria_definitions", "parameter_links", "criteria_links", "examples", "function_grammars", "script_key_audits", "sequence_steps"} {
				where := `device_name = 'WIFI_TEST'`
				switch table {
				case "devices":
					where = `name = 'WIFI_TEST'`
				case "functions":
					where = `interface = 'WIFI_TEST'`
				case "function_grammars", "script_key_audits":
					where = `function_interface = 'WIFI_TEST'`
				case "sequence_steps":
					where = `device_name = 'WIFI_TEST' OR function_interface = 'WIFI_TEST'`
				}
				if err := database.QueryRow(`SELECT COUNT(*) FROM ` + table + ` WHERE ` + where).Scan(&count); err != nil || count != 0 {
					t.Fatalf("%s still has device data: count=%d err=%v", table, count, err)
				}
			}
			for query, want := range map[string]int{
				`SELECT COUNT(*) FROM functions WHERE interface = 'OTHER'`:                              1,
				`SELECT COUNT(*) FROM parameter_definitions WHERE device_name = ''`:                     1,
				`SELECT COUNT(*) FROM parameter_definitions WHERE device_name = 'OSP'`:                  1,
				`SELECT COUNT(*) FROM criteria_definitions WHERE device_name = ''`:                      1,
				`SELECT COUNT(*) FROM parameter_links WHERE function_name = 'TX_MULTI_VERIFICATION'`:    2,
				`SELECT COUNT(*) FROM criteria_links WHERE function_name = 'TX_MULTI_VERIFICATION'`:     2,
				`SELECT COUNT(*) FROM examples WHERE function_name = 'TX_MULTI_VERIFICATION'`:           2,
				`SELECT COUNT(*) FROM parameter_links WHERE function_name = 'RX_SENSITIVITY'`:           0,
				`SELECT COUNT(*) FROM criteria_links WHERE function_name = 'RX_SENSITIVITY'`:            0,
				`SELECT COUNT(*) FROM examples WHERE function_name = 'RX_SENSITIVITY'`:                  0,
				`SELECT COUNT(*) FROM function_grammars WHERE function_interface = 'OTHER'`:             1,
				`SELECT COUNT(*) FROM script_key_audits WHERE function_interface = 'OTHER'`:             1,
				`SELECT COUNT(*) FROM sequences`:                                                        3,
				`SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = 3`:                             0,
				`SELECT COUNT(*) FROM sequences WHERE name = 'untouched' AND updated_at = 'old'`:        1,
				`SELECT COUNT(*) FROM sequences WHERE name <> 'untouched' AND updated_at <> 'old'`:      2,
				`SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = 1 AND id = 2 AND position = 1`: 1,
				`SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = 1 AND id = 4 AND position = 2`: 1,
			} {
				if err := database.QueryRow(query).Scan(&count); err != nil || count != want {
					t.Fatalf("%s: count=%d want=%d err=%v", query, count, want, err)
				}
			}
			if _, body := doRequest(t, app, "GET", "/api/devices"); strings.Contains(body, `"device_name":"WIFI_TEST"`) {
				t.Fatalf("deleted device still listed: %s", body)
			}
		})
	}
}

func TestDeviceDeleteRollsBackOnFailure(t *testing.T) {
	app, database := newDictionaryTestApp(t)
	if resp, body := registerDevice(t, app, "WIFI_TEST"); resp.StatusCode != 201 {
		t.Fatalf("register: %d %s", resp.StatusCode, body)
	}
	seedDeviceData(t, func(path, body string) {
		if resp, result := postJSON(t, app, path, body); resp.StatusCode != 201 {
			t.Fatalf("seed %s: %d %s", path, resp.StatusCode, result)
		}
	})
	for _, statement := range []string{
		`CREATE FUNCTION reject_device_delete() RETURNS trigger LANGUAGE plpgsql AS $$
			BEGIN RAISE EXCEPTION 'forced delete failure'; END $$`,
		`CREATE TRIGGER reject_device_delete BEFORE DELETE ON devices
			FOR EACH ROW EXECUTE FUNCTION reject_device_delete()`,
	} {
		if _, err := database.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if resp, body := doJSON(t, app, "DELETE", "/api/devices/WIFI_TEST", `{"confirmation":"DELETE WIFI_TEST"}`); resp.StatusCode != 500 {
		t.Fatalf("failed transaction: %d %s", resp.StatusCode, body)
	}
	for query, want := range map[string]int{
		`SELECT COUNT(*) FROM devices WHERE name = 'WIFI_TEST'`:                      1,
		`SELECT COUNT(*) FROM functions WHERE interface = 'WIFI_TEST'`:               2,
		`SELECT COUNT(*) FROM parameter_definitions WHERE device_name = 'WIFI_TEST'`: 2,
		`SELECT COUNT(*) FROM criteria_definitions WHERE device_name = 'WIFI_TEST'`:  1,
		`SELECT COUNT(*) FROM parameter_links WHERE device_name = 'WIFI_TEST'`:       3,
		`SELECT COUNT(*) FROM criteria_links WHERE device_name = 'WIFI_TEST'`:        2,
		`SELECT COUNT(*) FROM examples WHERE device_name = 'WIFI_TEST'`:              1,
	} {
		var count int
		if err := database.QueryRow(query).Scan(&count); err != nil || count != want {
			t.Fatalf("rollback %s: count=%d want=%d err=%v", query, count, want, err)
		}
	}
}
