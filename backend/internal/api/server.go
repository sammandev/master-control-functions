// Package api wires HTTP routes onto the application's dependencies.
//
// The HTTP surface is the primary test seam: every behavior is exercised
// through these handlers in integration tests against a temporary SQLite
// database.
package api

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"

	"webprogfndb/internal/store"
)

// Serve listens on addr until ctx is done, then stops accepting connections
// and waits up to drainTimeout for in-flight requests to finish.
func Serve(ctx context.Context, app *fiber.App, addr string, drainTimeout time.Duration) error {
	return app.Listen(addr, fiber.ListenConfig{
		GracefulContext: ctx,
		ShutdownTimeout: drainTimeout,
	})
}

// New builds the application's Fiber instance with all routes registered.
// Static file serving for the built SPA is added separately (RegisterStatic)
// so tests can run without a frontend build present.
func New(d *sql.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Web Program Functions Database",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New())

	q := store.New(d)
	registerFunctionRoutes(app, q, d)
	registerImportRoutes(app, q, d)
	registerExportRoutes(app, d)
	registerParameterDictionary(app, q)
	registerCriteriaDictionary(app, q)
	registerLinkRoutes(app, q, d, parameterLinkStore(), "parameter")
	registerLinkRoutes(app, q, d, criteriaLinkStore(), "criteria")
	registerOrphanRoutes(app, q)
	registerExampleRoutes(app, q)
	registerDevices(app, q, d)
	registerComposeRoutes(app, q)
	registerSequenceRoutes(app, q, d)
	registerMCPRoutes(app, q)
	registerTroubleshootingRegistryRoutes(app, q, d)
	registerTroubleshootingTransferRoutes(app, q, d)
	registerTroubleshootingRoutes(app, q, d)

	app.Get("/health", func(c fiber.Ctx) error {
		if err := d.Ping(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status": "error", "db": "error",
			})
		}
		return c.JSON(fiber.Map{"status": "ok", "db": "ok"})
	})

	return app
}

// RegisterStatic serves the built SPA bundle if the directory exists and
// wires the SPA fallback for client-side routes. Unknown /api/* paths get a
// JSON 404 (the API contract), not the SPA shell.
//
// Route registration order matters: specific API routes (registered in New
// and by later feature routes) take precedence, then static files, then the
// /api 404, then the SPA entry point.
func RegisterStatic(app *fiber.App, dir string) bool {
	index := filepath.Join(dir, "index.html")
	if _, err := os.Stat(index); err != nil {
		return false
	}

	app.Use("/", static.New(dir))
	app.All("/api/*", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	})
	// SPA fallback: any remaining GET serves the entry point so client-side
	// routing works on refresh/deep links.
	app.Get("/*", func(c fiber.Ctx) error {
		return c.SendFile(index)
	})
	return true
}
