package api

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// Examples CRUD (ticket 06). An Example belongs to a Function by natural
// key (source, name); the reference is validated at write time, so orphans
// only arise from imports or function deletions and are flagged with
// function_exists (ADR-0003). Editing replaces device scope, script and
// meaning only — the function reference is immutable.

// exampleBody is the create payload.
type exampleBody struct {
	FunctionSource string `json:"function_source"`
	FunctionName   string `json:"function_name"`
	DeviceName     string `json:"device_name"`
	Script         string `json:"script"`
	Meaning        string `json:"meaning"`
}

// exampleUpdateBody is the update payload; the function reference is not
// editable, so it does not appear here.
type exampleUpdateBody struct {
	DeviceName string `json:"device_name"`
	Script     string `json:"script"`
	Meaning    string `json:"meaning"`
}

// exampleResponse is the shared JSON shape for list and get.
type exampleResponse struct {
	ID             int64  `json:"id"`
	FunctionSource string `json:"function_source"`
	FunctionName   string `json:"function_name"`
	DeviceName     string `json:"device_name"`
	Script         string `json:"script"`
	Meaning        string `json:"meaning"`
	FunctionExists bool   `json:"function_exists"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	// ScriptIssues is returned on create/update only: structural problems the
	// MCP would report for this line. They warn; they never block a save.
	ScriptIssues []scriptIssue `json:"script_issues,omitempty"`
}

// exampleScriptIssues checks an example line against every Function sharing
// its (source, name) and reports the closest candidate's problems.
func exampleScriptIssues(c fiber.Ctx, q *store.Queries, source, name, script string) []scriptIssue {
	functions, err := q.ListFunctionsByNaturalName(c.Context(), store.ListFunctionsByNaturalNameParams{
		Source: source, Name: name,
	})
	if err != nil {
		return nil
	}
	targets := make([]exampleTarget, 0, len(functions))
	for _, f := range functions {
		targets = append(targets, exampleTarget{f.Source, f.Interface, f.Name, nullableConvention(f.Convention)})
	}
	return checkExampleScript(script, targets)
}

// exampleItem converts the generated list row into the JSON shape.
// exampleItem maps the shared columns positionally — sqlc emits distinct
// row types for the list and get queries with identical fields (same
// pattern as viewFromJoinedRow in links.go).
func exampleItem(
	id int64,
	functionSource, functionName string,
	deviceName string,
	script, meaning string,
	functionExists bool,
	createdAt, updatedAt string,
) exampleResponse {
	return exampleResponse{
		ID:             id,
		FunctionSource: functionSource,
		FunctionName:   functionName,
		DeviceName:     deviceName,
		Script:         script,
		Meaning:        meaning,
		FunctionExists: functionExists,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}

// listExampleItem adapts a generated list row to the mapper.
func listExampleItem(r store.ListExamplesRow) exampleResponse {
	return exampleItem(
		r.ID, r.FunctionSource, r.FunctionName, r.DeviceName,
		r.Script, r.Meaning, r.FunctionExists, r.CreatedAt, r.UpdatedAt,
	)
}

// registerExampleRoutes wires the /api/examples endpoints.
func registerExampleRoutes(app *fiber.App, q *store.Queries) {
	app.Get("/api/examples", listExamples(q))
	app.Post("/api/examples", createExample(q))
	app.Get("/api/examples/:id", getExample(q))
	app.Put("/api/examples/:id", updateExample(q))
	app.Delete("/api/examples/:id", deleteExample(q))
}

// exampleFilters reads the list filters: function_source, function_name,
// device and search (LIKE pattern over script + meaning).
func exampleFilters(c fiber.Ctx) store.ListExamplesParams {
	params := store.ListExamplesParams{}
	if v := strings.TrimSpace(c.Query("function_source")); v != "" {
		params.FunctionSource = sql.NullString{String: v, Valid: true}
	}
	if v := strings.TrimSpace(c.Query("function_name")); v != "" {
		params.FunctionName = sql.NullString{String: v, Valid: true}
	}
	// device distinguishes "param absent" (no filter) from "param present
	// but empty" (blank/global scope) — /api/devices/_global reuses this
	// endpoint for blank-scope examples. fiber's c.Query default fires on
	// empty values too, so presence is checked on the raw query args.
	if c.RequestCtx().QueryArgs().Has("device") {
		v := strings.TrimSpace(string(c.RequestCtx().QueryArgs().Peek("device")))
		params.Device = sql.NullString{String: v, Valid: true}
	}
	if s := strings.TrimSpace(c.Query("search")); s != "" {
		if p := likePattern(s); p != "" {
			params.Search = sql.NullString{String: p, Valid: true}
		}
	}
	return params
}

func listExamples(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		params := exampleFilters(c)
		total, err := q.CountExamples(c.Context(), store.CountExamplesParams{
			FunctionSource: params.FunctionSource,
			FunctionName:   params.FunctionName,
			Device:         params.Device,
			Search:         params.Search,
		})
		if err != nil {
			return serverError(c, "count examples", err)
		}
		rows, err := q.ListExamples(c.Context(), params)
		if err != nil {
			return serverError(c, "list examples", err)
		}
		// Go-side pagination over the SQL-sorted rows (device, id).
		lq := parseListQuery(c)
		start := (lq.page - 1) * lq.pageSize
		if start > int64(len(rows)) {
			start = int64(len(rows))
		}
		end := start + lq.pageSize
		if end > int64(len(rows)) {
			end = int64(len(rows))
		}
		items := make([]exampleResponse, 0, end-start)
		for _, r := range rows[start:end] {
			items = append(items, listExampleItem(r))
		}
		return c.JSON(fiber.Map{
			"items":     items,
			"total":     total,
			"page":      lq.page,
			"page_size": lq.pageSize,
		})
	}
}

func createExample(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		var b exampleBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		b.FunctionSource = strings.TrimSpace(b.FunctionSource)
		b.FunctionName = strings.TrimSpace(b.FunctionName)
		// ADR-0002: match the registry's normalization so casing twins
		// cannot diverge from a registered Device.
		b.DeviceName = normalizeDeviceName(b.DeviceName)
		b.Script = strings.TrimSpace(b.Script)

		fieldErrors := map[string]string{}
		if b.FunctionSource == "" {
			fieldErrors["function_source"] = "source is required"
		}
		if b.FunctionName == "" {
			fieldErrors["function_name"] = "name is required"
		}
		if b.Script == "" {
			fieldErrors["script"] = "script is required"
		}
		if len(fieldErrors) > 0 {
			return validationError(c, fieldErrors)
		}
		// Write-time reference validation: the example's function must
		// exist, so web-created examples are never orphans (ticket 06).
		exists, err := q.FunctionNaturalKeyExists(c.Context(), store.FunctionNaturalKeyExistsParams{
			Source: b.FunctionSource,
			Name:   b.FunctionName,
		})
		if err != nil {
			return serverError(c, "check function reference", err)
		}
		if !exists {
			return validationError(c, map[string]string{
				"function_name": "no function '" + b.FunctionName +
					"' with source '" + b.FunctionSource + "' exists",
			})
		}

		now := time.Now().UTC().Format(time.RFC3339)
		item, err := q.InsertExample(c.Context(), store.InsertExampleParams{
			FunctionSource: b.FunctionSource,
			FunctionName:   b.FunctionName,
			DeviceName:     b.DeviceName,
			Script:         b.Script,
			Meaning:        b.Meaning,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "this example already exists for the function and device",
				})
			}
			return serverError(c, "create example", err)
		}
		return c.Status(fiber.StatusCreated).JSON(exampleResponse{
			ID:             item.ID,
			FunctionSource: item.FunctionSource,
			FunctionName:   item.FunctionName,
			DeviceName:     item.DeviceName,
			Script:         item.Script,
			Meaning:        item.Meaning,
			FunctionExists: true,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			ScriptIssues:   exampleScriptIssues(c, q, item.FunctionSource, item.FunctionName, item.Script),
		})
	}
}

func getExample(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		row, err := q.GetExample(c.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "example not found"})
			}
			return serverError(c, "read example", err)
		}
		return c.JSON(exampleItem(
			row.ID, row.FunctionSource, row.FunctionName, row.DeviceName,
			row.Script, row.Meaning, row.FunctionExists, row.CreatedAt, row.UpdatedAt,
		))
	}
}

func updateExample(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		var b exampleUpdateBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		// ADR-0002: match the registry's normalization (see create).
		b.DeviceName = normalizeDeviceName(b.DeviceName)
		b.Script = strings.TrimSpace(b.Script)
		if b.Script == "" {
			return validationError(c, map[string]string{"script": "script is required"})
		}
		now := time.Now().UTC().Format(time.RFC3339)
		_, err = q.UpdateExample(c.Context(), store.UpdateExampleParams{
			DeviceName: b.DeviceName,
			Script:     b.Script,
			Meaning:    b.Meaning,
			UpdatedAt:  now,
			ID:         id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "example not found"})
			}
			// Editing device/script can collide with another example's
			// natural key; treat it like a create duplicate.
			if store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "this example already exists for the function and device",
				})
			}
			return serverError(c, "update example", err)
		}
		row, err := q.GetExample(c.Context(), id)
		if err != nil {
			return serverError(c, "read updated example", err)
		}
		updated := exampleItem(
			row.ID, row.FunctionSource, row.FunctionName, row.DeviceName,
			row.Script, row.Meaning, row.FunctionExists, row.CreatedAt, row.UpdatedAt,
		)
		updated.ScriptIssues = exampleScriptIssues(c, q, row.FunctionSource, row.FunctionName, row.Script)
		return c.JSON(updated)
	}
}

func deleteExample(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		affected, err := q.DeleteExample(c.Context(), id)
		if err != nil {
			return serverError(c, "delete example", err)
		}
		if affected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "example not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
