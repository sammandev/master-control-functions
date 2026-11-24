package api

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/importer"
)

// Export endpoint (ticket 08): regenerates the six-sheet knowledge
// workbook from the database (ADR-0004) as an xlsx download. Values are
// the database's normalized state; orphans are exported faithfully.

func registerExportRoutes(app *fiber.App, d *sql.DB) {
	app.Get("/api/export", exportWorkbook(d))
}

func exportWorkbook(d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		var buf bytes.Buffer
		if err := importer.Export(c.Context(), d, &buf); err != nil {
			var limit *importer.ExportCellLimitError
			if errors.As(err, &limit) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": limit.Error()})
			}
			return serverError(c, "export workbook", err)
		}
		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", `attachment; filename="knowledge_workbook.xlsx"`)
		return c.Send(buf.Bytes())
	}
}
