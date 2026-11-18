package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"webprogfndb/internal/api"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/db"
	"webprogfndb/tests/support/testdb"
)

func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	d, err := testdb.Open(t)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	return api.New(d)
}

func doRequest(t *testing.T, app *fiber.App, method, target string) (*http.Response, string) {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, string(body)
}

func TestHealthReturnsOK(t *testing.T) {
	app := newTestApp(t)

	resp, body := doRequest(t, app, "GET", "/health")

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", resp.StatusCode, body)
	}
}

func TestHealthReportsDatabaseState(t *testing.T) {
	app := newTestApp(t)

	_, body := doRequest(t, app, "GET", "/health")

	if !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("body %s does not contain %q", body, `"status":"ok"`)
	}
	if !strings.Contains(body, `"db":"ok"`) {
		t.Fatalf("body %s does not contain %q", body, `"db":"ok"`)
	}
}

func TestUnknownAPIRouteReturns404JSON(t *testing.T) {
	app := newTestApp(t)

	resp, body := doRequest(t, app, "GET", "/api/nope")

	if resp.StatusCode != 404 {
		t.Fatalf("status = %d, want 404 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"error"`) {
		t.Fatalf("body %s does not contain %q", body, `"error"`)
	}
}

// withFakeBundle creates a directory that looks like a built SPA bundle
// and registers it, mirroring production route wiring.
func withFakeBundle(t *testing.T, app *fiber.App) {
	t.Helper()
	// os.MkdirTemp, not t.TempDir: fasthttp's static handler keeps file
	// handles open past test completion on Windows, which would make
	// t.TempDir's RemoveAll cleanup fail. The OS temp cleaner reaps these.
	dir, err := os.MkdirTemp("", "wpfdb-bundle-")
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	index := filepath.Join(dir, "index.html")
	if err := os.WriteFile(index, []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if !api.RegisterStatic(app, dir) {
		t.Fatalf("RegisterStatic reported no bundle")
	}
}

func TestRegisterStaticServesRootAndFallsBack(t *testing.T) {
	app := newTestApp(t)
	withFakeBundle(t, app)

	resp, body := doRequest(t, app, "GET", "/")
	if resp.StatusCode != 200 || !strings.Contains(body, "spa") {
		t.Fatalf("GET /: status = %d, body = %q", resp.StatusCode, body)
	}

	resp, body = doRequest(t, app, "GET", "/functions")
	if resp.StatusCode != 200 || !strings.Contains(body, "spa") {
		t.Fatalf("deep link fallback: status = %d, body = %q", resp.StatusCode, body)
	}

	if resp, _ := doRequest(t, app, "GET", "/health"); resp.StatusCode != 200 {
		t.Fatalf("health must still work after static registration, got %d", resp.StatusCode)
	}
}

func TestUnknownAPIRouteStays404JSONWithSPAPresent(t *testing.T) {
	app := newTestApp(t)
	withFakeBundle(t, app)

	resp, body := doRequest(t, app, "GET", "/api/nope")

	if resp.StatusCode != 404 {
		t.Fatalf("status = %d, want 404 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"error"`) {
		t.Fatalf("body %s does not contain %q", body, `"error"`)
	}
	if strings.Contains(body, "spa") {
		t.Fatalf("unknown API route must not return the SPA shell, got: %s", body)
	}
}

func TestRegisterStaticWithoutBundleReturnsFalse(t *testing.T) {
	app := newTestApp(t)

	if api.RegisterStatic(app, filepath.Join(t.TempDir(), "missing")) {
		t.Fatalf("RegisterStatic should report false when no bundle exists")
	}
}
