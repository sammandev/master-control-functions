package api

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"math"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

const troubleshootingExportName = "troubleshooting-guides"

type troubleshootingPreviewRow struct {
	Index   int      `json:"index"`
	ID      *int64   `json:"id,omitempty"`
	Title   string   `json:"title"`
	Action  string   `json:"action"`
	Reasons []string `json:"reasons,omitempty"`
}

type troubleshootingPreview struct {
	Valid  bool                        `json:"valid"`
	Rows   []troubleshootingPreviewRow `json:"rows"`
	Errors []string                    `json:"errors,omitempty"`
}

type troubleshootingTransferLookups interface {
	GetTroubleshootingCategory(context.Context, string) (store.TroubleshootingCategory, error)
	GetTroubleshootingSeverity(context.Context, string) (store.TroubleshootingSeverity, error)
	GetTroubleshootingGuide(context.Context, int64) (store.TroubleshootingGuide, error)
}

func registerTroubleshootingTransferRoutes(app *fiber.App, q *store.Queries, d *sql.DB) {
	app.Get("/api/troubleshooting/templates/:format", troubleshootingTemplate(q))
	app.Post("/api/troubleshooting/import/single", singleTroubleshootingTransfer)
	app.Post("/api/troubleshooting/import/preview", troubleshootingImport(q, d, false))
	app.Post("/api/troubleshooting/import/commit", troubleshootingImport(q, d, true))
	app.Post("/api/troubleshooting/export", troubleshootingExport(q, d))
}

func readTroubleshootingUpload(c fiber.Ctx) (troubleshootingTransfer, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return troubleshootingTransfer{}, errors.New("multipart file field 'file' is required")
	}
	f, err := file.Open()
	if err != nil {
		return troubleshootingTransfer{}, errors.New("unable to open uploaded file")
	}
	defer f.Close()
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Filename)), ".")
	return parseTroubleshootingTransfer(f, format)
}

func singleTroubleshootingTransfer(c fiber.Ctx) error {
	doc, err := readTroubleshootingUpload(c)
	if err != nil {
		return troubleshootingTransferError(c, fiber.StatusUnprocessableEntity, err.Error())
	}
	if len(doc.Guides) > 1 {
		return troubleshootingTransferError(c, fiber.StatusUnprocessableEntity, "This file contains multiple guides. Use bulk import on the Troubleshooting Guide page.")
	}
	if len(doc.Guides) == 0 {
		return troubleshootingTransferError(c, fiber.StatusUnprocessableEntity, "The file does not contain a guide.")
	}
	if problems := append(doc.issues, doc.Guides[0].issues...); len(problems) > 0 {
		return troubleshootingTransferError(c, fiber.StatusUnprocessableEntity, strings.Join(problems, "; "))
	}
	return c.JSON(doc.Guides[0].troubleshootingBody)
}

func previewTroubleshootingTransfer(ctx context.Context, q troubleshootingTransferLookups, doc *troubleshootingTransfer) (troubleshootingPreview, error) {
	out := troubleshootingPreview{
		Rows: make([]troubleshootingPreviewRow, 0, len(doc.Guides)), Errors: doc.issues,
	}
	if len(doc.Guides) == 0 {
		out.Errors = append(out.Errors, "The file does not contain any guides.")
	}
	out.Valid = len(out.Errors) == 0
	idCounts := make(map[int64]int)
	for _, g := range doc.Guides {
		if g.ID != nil {
			idCounts[*g.ID]++
		}
	}
	for i := range doc.Guides {
		g := &doc.Guides[i]
		fields := g.validateWithTextNormalization(false)
		row := troubleshootingPreviewRow{Index: i + 1, ID: g.ID, Title: g.Title, Action: "create", Reasons: g.issues}
		if g.Category != "" {
			if _, err := q.GetTroubleshootingCategory(ctx, g.Category); errors.Is(err, sql.ErrNoRows) {
				fields["category"] = fmt.Sprintf("unknown category %q", g.Category)
			} else if err != nil {
				return out, err
			}
		}
		if g.Severity != "" {
			if _, err := q.GetTroubleshootingSeverity(ctx, g.Severity); errors.Is(err, sql.ErrNoRows) {
				fields["severity"] = fmt.Sprintf("unknown severity %q", g.Severity)
			} else if err != nil {
				return out, err
			}
		}
		if g.ID != nil {
			row.Action = "update"
			if idCounts[*g.ID] > 1 {
				row.Reasons = append(row.Reasons, "duplicate guide id")
			}
			if _, err := q.GetTroubleshootingGuide(ctx, *g.ID); errors.Is(err, sql.ErrNoRows) {
				row.Reasons = append(row.Reasons, "id not found")
			} else if err != nil {
				return out, err
			}
		}
		for _, key := range slices.Sorted(maps.Keys(fields)) {
			message := fields[key]
			if !strings.HasPrefix(key, "steps.") {
				message = key + ": " + message
				if key != "steps" && fields[key] == "is required" {
					message = key + " is required"
				}
			}
			row.Reasons = append(row.Reasons, message)
		}
		if len(row.Reasons) > 0 {
			row.Action = "error"
			out.Valid = false
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}

func troubleshootingImport(q *store.Queries, d *sql.DB, commit bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		doc, err := readTroubleshootingUpload(c)
		if err != nil {
			return troubleshootingTransferError(c, fiber.StatusUnprocessableEntity, err.Error())
		}
		if !commit {
			out, err := previewTroubleshootingTransfer(c.Context(), q, &doc)
			if err != nil {
				return serverError(c, "preview troubleshooting import", err)
			}
			return c.JSON(out)
		}
		tx, err := d.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return serverError(c, "start troubleshooting import", err)
		}
		defer tx.Rollback()
		tq := q.WithTx(tx)
		preview, err := previewTroubleshootingTransfer(c.Context(), tq, &doc)
		if err != nil {
			return serverError(c, "validate troubleshooting import", err)
		}
		if !preview.Valid {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error":   "Import validation failed. Review the errors and upload the corrected file.",
				"preview": preview,
			})
		}
		created, updated := 0, 0
		now := time.Now().UTC().Format(time.RFC3339)
		for _, guide := range doc.Guides {
			if err := writeTroubleshootingTransferGuide(c, tq, guide, now); err != nil {
				return serverError(c, "write troubleshooting import", err)
			}
			if guide.ID == nil {
				created++
			} else {
				updated++
			}
		}
		if err := tx.Commit(); err != nil {
			return serverError(c, "commit troubleshooting import", err)
		}
		return c.JSON(fiber.Map{"created": created, "updated": updated})
	}
}

func writeTroubleshootingTransferGuide(c fiber.Ctx, q *store.Queries, guide troubleshootingTransferGuide, now string) error {
	b := guide.troubleshootingBody
	var id int64
	if guide.ID == nil {
		g, err := q.InsertTroubleshootingGuide(c.Context(), store.InsertTroubleshootingGuideParams{
			Title: b.Title, Category: b.Category, Severity: b.Severity, DeviceName: b.DeviceName,
			FunctionName: b.FunctionName, ErrorCode: b.ErrorCode, Symptom: b.Symptom, RootCause: b.RootCause,
			CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			return err
		}
		id = g.ID
	} else {
		id = *guide.ID
		if err := q.UpdateTroubleshootingGuideMeta(c.Context(), store.UpdateTroubleshootingGuideMetaParams{
			ID: id, Title: b.Title, Category: b.Category, Severity: b.Severity, DeviceName: b.DeviceName,
			FunctionName: b.FunctionName, ErrorCode: b.ErrorCode, Symptom: b.Symptom, RootCause: b.RootCause,
			UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	return writeTroubleshootingSteps(c, q, id, b.Steps)
}

func troubleshootingTransferError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func troubleshootingTemplate(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		format := c.Params("format")
		if format != "json" && format != "xlsx" {
			return troubleshootingTransferError(c, fiber.StatusBadRequest, "choose JSON or XLSX")
		}
		cats, err := q.ListTroubleshootingCategories(c.Context())
		if err != nil {
			return serverError(c, "read template categories", err)
		}
		sevs, err := q.ListTroubleshootingSeverities(c.Context())
		if err != nil {
			return serverError(c, "read template severities", err)
		}
		if len(cats) == 0 || len(sevs) == 0 {
			return troubleshootingTransferError(c, fiber.StatusUnprocessableEntity, "Configure at least one category and severity before downloading a template.")
		}
		doc := troubleshootingTransfer{Guides: []troubleshootingTransferGuide{{troubleshootingBody: troubleshootingBody{
			Title: "Example troubleshooting guide", Category: cats[0].Key, Severity: sevs[0].Key,
			Symptom: "- Describe the symptom\n- Include a log signature if useful",
			Steps:   []troubleshootingStepBody{{Title: "Example resolution step", Instruction: "Describe what to do."}},
		}}}}
		instructions := [][]string{
			{"Replace the example guide and step with your own data."},
			{"Use the same guide_ref on Guides and Steps. Order steps with step_no."},
			{"Leave id blank to create a guide. Use an existing id only to replace that guide."},
			{"Leave device_name blank for a Global guide. Use registry keys below."},
			{"Registry", "Valid key", "Label"},
		}
		for _, cat := range cats {
			instructions = append(instructions, []string{"Category", cat.Key, cat.Label})
		}
		for _, sev := range sevs {
			instructions = append(instructions, []string{"Severity", sev.Key, sev.Label})
		}
		raw, err := writeTroubleshootingTransfer(doc, format, instructions)
		if err != nil {
			return serverError(c, "generate troubleshooting template", err)
		}
		return sendTroubleshootingDownload(c, "troubleshooting-guide-template", format, raw)
	}
}

type troubleshootingExportRequest struct {
	Format string  `json:"format"`
	Scope  string  `json:"scope"`
	IDs    []int64 `json:"ids"`
}

func troubleshootingExport(q *store.Queries, d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		var in troubleshootingExportRequest
		if err := c.Bind().Body(&in); err != nil {
			return troubleshootingTransferError(c, fiber.StatusBadRequest, "invalid JSON body")
		}
		if in.Format != "json" && in.Format != "xlsx" && in.Format != "both" {
			return troubleshootingTransferError(c, fiber.StatusBadRequest, "choose JSON, XLSX or Both")
		}
		switch in.Scope {
		case "", "selected":
			if len(in.IDs) == 0 {
				return troubleshootingTransferError(c, fiber.StatusBadRequest, "select at least one guide")
			}
		case "matching", "all":
			if len(in.IDs) != 0 {
				return troubleshootingTransferError(c, fiber.StatusBadRequest, "choose selected guides or an export scope")
			}
		default:
			return troubleshootingTransferError(c, fiber.StatusBadRequest, "unknown export scope")
		}
		doc, err := loadTroubleshootingExport(c, q, d, in)
		if errors.Is(err, sql.ErrNoRows) {
			return troubleshootingTransferError(c, fiber.StatusNotFound, "a selected guide no longer exists")
		}
		if err != nil {
			return serverError(c, "read troubleshooting export", err)
		}
		if len(doc.Guides) == 0 {
			return troubleshootingTransferError(c, fiber.StatusUnprocessableEntity, "No guides match this export.")
		}
		format := in.Format
		var raw []byte
		if format == "both" {
			raw, err = zipTroubleshootingTransfer(doc)
			format = "zip"
		} else {
			raw, err = writeTroubleshootingTransfer(doc, format, nil)
		}
		if err != nil {
			return serverError(c, "generate troubleshooting export", err)
		}
		return sendTroubleshootingDownload(c, troubleshootingExportName, format, raw)
	}
}

func loadTroubleshootingExport(c fiber.Ctx, q *store.Queries, d *sql.DB, in troubleshootingExportRequest) (troubleshootingTransfer, error) {
	doc := troubleshootingTransfer{Guides: []troubleshootingTransferGuide{}}
	tx, err := d.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return doc, err
	}
	defer tx.Rollback()
	q = q.WithTx(tx)
	ids := in.IDs
	if in.Scope == "matching" || in.Scope == "all" {
		params := store.ListTroubleshootingGuidesParams{Limit: math.MaxInt64}
		if in.Scope == "matching" {
			params.Search, params.Category, params.DeviceName, params.ErrorCode = troubleshootingFilters(c)
		}
		rows, err := q.ListTroubleshootingGuides(c.Context(), params)
		if err != nil {
			return doc, err
		}
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
	}
	seen := make(map[int64]bool)
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		g, err := troubleshootingDetail(c, q, id)
		if err != nil {
			return doc, err
		}
		b := troubleshootingBody{
			Title: g.Title, Category: g.Category, Severity: g.Severity, DeviceName: g.DeviceName,
			FunctionName: g.FunctionName, ErrorCode: g.ErrorCode, Symptom: g.Symptom, RootCause: g.RootCause,
			Steps: make([]troubleshootingStepBody, 0, len(g.Steps)),
		}
		for _, step := range g.Steps {
			b.Steps = append(b.Steps, troubleshootingStepBody{
				Title: step.Title, Instruction: step.Instruction, ScriptSnippet: step.ScriptSnippet, Verification: step.Verification,
			})
		}
		doc.Guides = append(doc.Guides, troubleshootingTransferGuide{ID: &g.ID, troubleshootingBody: b})
	}
	return doc, tx.Commit()
}

func zipTroubleshootingTransfer(doc troubleshootingTransfer) ([]byte, error) {
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	for _, format := range []string{"json", "xlsx"} {
		raw, err := writeTroubleshootingTransfer(doc, format, nil)
		if err != nil {
			return nil, err
		}
		entry, err := z.Create(troubleshootingExportName + "." + format)
		if err != nil {
			return nil, err
		}
		if _, err := entry.Write(raw); err != nil {
			return nil, err
		}
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sendTroubleshootingDownload(c fiber.Ctx, name, format string, raw []byte) error {
	mime := map[string]string{
		"json": "application/json",
		"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"zip":  "application/zip",
	}
	c.Set(fiber.HeaderContentType, mime[format])
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", name+"."+format))
	return c.Send(raw)
}
