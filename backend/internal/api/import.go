package api

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/importer"
	"webprogfndb/internal/store"
)

// registerImportRoutes wires workbook import, reset, and the change-log
// views. All of ticket 03's maintainer surface lives here.
func registerImportRoutes(app *fiber.App, q *store.Queries, d *sql.DB) {
	app.Post("/api/import", importHandler(q, d, false))
	app.Post("/api/import/reset", importHandler(q, d, true))
	app.Get("/api/import/runs", listImportRuns(q))
	app.Get("/api/import/runs/:id", getImportRun(q))
}

// importHandler accepts a multipart xlsx upload and runs the importer.
// With reset=true the six entity tables are emptied first (same tx).
func importHandler(q *store.Queries, d *sql.DB, reset bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "multipart file field 'file' is required"})
		}
		f, err := file.Open()
		if err != nil {
			return serverError(c, "open upload", err)
		}
		defer f.Close()

		wb, err := importer.ParseWorkbook(f)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		}

		res, err := importer.Import(c.Context(), d, wb, file.Filename, reset)
		if err != nil {
			return serverError(c, "import", err)
		}
		return c.JSON(res)
	}
}

// listImportRuns returns the recent change-log runs.
func listImportRuns(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		runs, err := q.ListImportRuns(c.Context())
		if err != nil {
			return serverError(c, "list import runs", err)
		}
		out := make([]fiber.Map, 0, len(runs))
		for _, r := range runs {
			out = append(out, importRunResponse(r))
		}
		return c.JSON(fiber.Map{"items": out})
	}
}

func getImportRun(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid run id"})
		}
		run, err := q.GetImportRun(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "import run not found"})
		}
		norms, err := q.ListImportNormalizations(c.Context(), id)
		if err != nil {
			return serverError(c, "list normalizations", err)
		}
		// Map store models to the snake_case JSON shape the frontend
		// expects (sqlc models carry no JSON tags).
		out := make([]fiber.Map, 0, len(norms))
		for _, n := range norms {
			out = append(out, fiber.Map{
				"id":               n.ID,
				"run_id":           n.RunID,
				"table_name":       n.TableName,
				"row_key":          n.RowKey,
				"column_name":      n.ColumnName,
				"source_value":     n.SourceValue,
				"normalized_value": n.NormalizedValue,
				"rule":             n.Rule,
			})
		}
		return c.JSON(fiber.Map{
			"run":            importRunResponse(run),
			"normalizations": out,
		})
	}
}

func importRunResponse(r store.ImportRun) fiber.Map {
	return fiber.Map{
		"id":          r.ID,
		"file_name":   r.FileName,
		"reset":       r.Reset != 0,
		"started_at":  r.StartedAt,
		"finished_at": r.FinishedAt,
		"counts":      r.Counts,
		"skipped":     r.Skipped,
	}
}
