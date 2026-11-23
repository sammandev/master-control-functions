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

// Troubleshooting Playbook & AI Knowledge Base CRUD.
// Multiple guides can share the same error_code, function_name, or device_name
// because a problem or error code can manifest in several ways with different
// root causes and step-by-step solutions.

type troubleshootingStepBody struct {
	Title         string `json:"title"`
	Instruction   string `json:"instruction"`
	ScriptSnippet string `json:"script_snippet"`
	Verification  string `json:"verification"`
}

type troubleshootingBody struct {
	Title        string                    `json:"title"`
	Category     string                    `json:"category"`
	Severity     string                    `json:"severity"`
	DeviceName   string                    `json:"device_name"`
	FunctionName string                    `json:"function_name"`
	ErrorCode    string                    `json:"error_code"`
	Symptom      string                    `json:"symptom"`
	RootCause    string                    `json:"root_cause"`
	Steps        []troubleshootingStepBody `json:"steps"`
}

func (b *troubleshootingBody) validate() map[string]string {
	return b.validateWithTextNormalization(true)
}

func (b *troubleshootingBody) validateWithTextNormalization(trimText bool) map[string]string {
	fields := map[string]string{}
	b.Title = strings.TrimSpace(b.Title)
	b.Category = strings.ToLower(strings.TrimSpace(b.Category))
	b.Severity = strings.ToLower(strings.TrimSpace(b.Severity))
	b.DeviceName = normalizeDeviceName(b.DeviceName)
	b.FunctionName = strings.TrimSpace(b.FunctionName)
	b.ErrorCode = strings.ToUpper(strings.TrimSpace(b.ErrorCode))
	if trimText {
		b.Symptom = strings.TrimSpace(b.Symptom)
		b.RootCause = strings.TrimSpace(b.RootCause)
	}

	if b.Title == "" {
		fields["title"] = "is required"
	}
	if b.Category == "" {
		fields["category"] = "is required"
	}
	if b.Severity == "" {
		fields["severity"] = "is required"
	}
	if problem := validateDeviceName(b.DeviceName); problem != "" {
		fields["device_name"] = problem
	}
	if strings.TrimSpace(b.Symptom) == "" {
		fields["symptom"] = "is required"
	}
	if len(b.Steps) == 0 {
		fields["steps"] = "at least one resolution step is required"
	}
	for i := range b.Steps {
		s := &b.Steps[i]
		s.Title = strings.TrimSpace(s.Title)
		if trimText {
			s.Instruction = strings.TrimSpace(s.Instruction)
			s.Verification = strings.TrimSpace(s.Verification)
		}
		if s.Title == "" {
			fields["steps."+strconv.Itoa(i)+".title"] = "step " + strconv.Itoa(i+1) + ": title is required"
		}
		if strings.TrimSpace(s.Instruction) == "" {
			fields["steps."+strconv.Itoa(i)+".instruction"] = "step " + strconv.Itoa(i+1) + ": instruction is required"
		}
	}
	return fields
}

func validateTroubleshootingRegistryKeys(c fiber.Ctx, q *store.Queries, category, severity string) (map[string]string, error) {
	fields := map[string]string{}
	if _, err := q.GetTroubleshootingCategory(c.Context(), category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fields["category"] = "does not exist"
		} else {
			return nil, err
		}
	}
	if _, err := q.GetTroubleshootingSeverity(c.Context(), severity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fields["severity"] = "does not exist"
		} else {
			return nil, err
		}
	}
	return fields, nil
}

func registryFieldError(c fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error":  "validation failed",
		"fields": fields,
	})
}

type troubleshootingStepResponse struct {
	ID            int64  `json:"id"`
	Position      int64  `json:"position"`
	Title         string `json:"title"`
	Instruction   string `json:"instruction"`
	ScriptSnippet string `json:"script_snippet"`
	Verification  string `json:"verification"`
}

type troubleshootingGuideResponse struct {
	ID           int64                         `json:"id"`
	Title        string                        `json:"title"`
	Category     string                        `json:"category"`
	Severity     string                        `json:"severity"`
	DeviceName   string                        `json:"device_name"`
	FunctionName string                        `json:"function_name"`
	ErrorCode    string                        `json:"error_code"`
	Symptom      string                        `json:"symptom"`
	RootCause    string                        `json:"root_cause"`
	Steps        []troubleshootingStepResponse `json:"steps"`
	CreatedAt    string                        `json:"created_at"`
	UpdatedAt    string                        `json:"updated_at"`
}

type troubleshootingListItem struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	Severity     string `json:"severity"`
	DeviceName   string `json:"device_name"`
	FunctionName string `json:"function_name"`
	ErrorCode    string `json:"error_code"`
	Symptom      string `json:"symptom"`
	RootCause    string `json:"root_cause"`
	StepCount    int64  `json:"step_count"`
	RelatedCount int64  `json:"related_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func registerTroubleshootingRoutes(app *fiber.App, q *store.Queries, d *sql.DB) {
	app.Get("/api/troubleshooting", listTroubleshootingGuides(q))
	app.Post("/api/troubleshooting", createTroubleshootingGuide(q, d))
	app.Get("/api/troubleshooting/:id/related", listRelatedTroubleshootingGuides(q))
	app.Get("/api/troubleshooting/:id", getTroubleshootingGuide(q))
	app.Put("/api/troubleshooting/:id", updateTroubleshootingGuide(q, d))
	app.Delete("/api/troubleshooting/:id", deleteTroubleshootingGuide(q))
}

func troubleshootingDetail(c fiber.Ctx, q *store.Queries, id int64) (troubleshootingGuideResponse, error) {
	g, err := q.GetTroubleshootingGuide(c.Context(), id)
	if err != nil {
		return troubleshootingGuideResponse{}, err
	}
	rows, err := q.ListTroubleshootingSteps(c.Context(), id)
	if err != nil {
		return troubleshootingGuideResponse{}, err
	}
	steps := make([]troubleshootingStepResponse, 0, len(rows))
	for _, r := range rows {
		steps = append(steps, troubleshootingStepResponse{
			ID:            r.ID,
			Position:      r.Position,
			Title:         r.Title,
			Instruction:   r.Instruction,
			ScriptSnippet: r.ScriptSnippet,
			Verification:  r.Verification,
		})
	}
	return troubleshootingGuideResponse{
		ID:           g.ID,
		Title:        g.Title,
		Category:     g.Category,
		Severity:     g.Severity,
		DeviceName:   g.DeviceName,
		FunctionName: g.FunctionName,
		ErrorCode:    g.ErrorCode,
		Symptom:      g.Symptom,
		RootCause:    g.RootCause,
		Steps:        steps,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}, nil
}

func writeTroubleshootingSteps(c fiber.Ctx, q *store.Queries, guideID int64, steps []troubleshootingStepBody) error {
	if err := q.DeleteTroubleshootingSteps(c.Context(), guideID); err != nil {
		return err
	}
	for i, s := range steps {
		_, err := q.InsertTroubleshootingStep(c.Context(), store.InsertTroubleshootingStepParams{
			GuideID:       guideID,
			Position:      int64(i + 1),
			Title:         s.Title,
			Instruction:   s.Instruction,
			ScriptSnippet: s.ScriptSnippet,
			Verification:  s.Verification,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func troubleshootingFilters(c fiber.Ctx) (search, category, deviceName, errorCode sql.NullString) {
	if s := strings.TrimSpace(c.Query("search")); s != "" {
		if p := likePattern(s); p != "" {
			search = sql.NullString{String: p, Valid: true}
		}
	}
	if cat := strings.ToLower(strings.TrimSpace(c.Query("category"))); cat != "" && cat != "_all" {
		category = sql.NullString{String: cat, Valid: true}
	}
	if c.RequestCtx().QueryArgs().Has("device") {
		dev := strings.TrimSpace(string(c.RequestCtx().QueryArgs().Peek("device")))
		if dev != "_all" {
			if dev == "_global" {
				dev = ""
			} else {
				dev = normalizeDeviceName(dev)
			}
			deviceName = sql.NullString{String: dev, Valid: true}
		}
	}
	if ec := strings.TrimSpace(c.Query("error_code")); ec != "" {
		errorCode = sql.NullString{String: ec, Valid: true}
	}
	return
}

func listTroubleshootingGuides(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		search, category, deviceName, errorCode := troubleshootingFilters(c)
		lq := parseListQuery(c)
		total, err := q.CountTroubleshootingGuides(c.Context(), store.CountTroubleshootingGuidesParams{
			Search:     search,
			Category:   category,
			DeviceName: deviceName,
			ErrorCode:  errorCode,
		})
		if err != nil {
			return serverError(c, "count troubleshooting guides", err)
		}
		rows, err := q.ListTroubleshootingGuides(c.Context(), store.ListTroubleshootingGuidesParams{
			Search:     search,
			Category:   category,
			DeviceName: deviceName,
			ErrorCode:  errorCode,
			Limit:      lq.pageSize,
			Offset:     (lq.page - 1) * lq.pageSize,
		})
		if err != nil {
			return serverError(c, "list troubleshooting guides", err)
		}
		items := make([]troubleshootingListItem, 0, len(rows))
		for _, r := range rows {
			items = append(items, troubleshootingListItem{
				ID:           r.ID,
				Title:        r.Title,
				Category:     r.Category,
				Severity:     r.Severity,
				DeviceName:   r.DeviceName,
				FunctionName: r.FunctionName,
				ErrorCode:    r.ErrorCode,
				Symptom:      r.Symptom,
				RootCause:    r.RootCause,
				StepCount:    r.StepCount,
				RelatedCount: r.RelatedCount,
				CreatedAt:    r.CreatedAt,
				UpdatedAt:    r.UpdatedAt,
			})
		}
		return c.JSON(fiber.Map{
			"items":     items,
			"total":     total,
			"page":      lq.page,
			"page_size": lq.pageSize,
		})
	}
}

func createTroubleshootingGuide(q *store.Queries, d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		var b troubleshootingBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if fields := b.validate(); len(fields) > 0 {
			return validationError(c, fields)
		}
		fields, err := validateTroubleshootingRegistryKeys(c, q, b.Category, b.Severity)
		if err != nil {
			return serverError(c, "validate troubleshooting registry keys", err)
		}
		if len(fields) > 0 {
			return registryFieldError(c, fields)
		}
		now := time.Now().UTC().Format(time.RFC3339)

		var id int64
		err = func() error {
			tx, err := d.BeginTx(c.Context(), nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			tq := q.WithTx(tx)
			g, err := tq.InsertTroubleshootingGuide(c.Context(), store.InsertTroubleshootingGuideParams{
				Title:        b.Title,
				Category:     b.Category,
				Severity:     b.Severity,
				DeviceName:   b.DeviceName,
				FunctionName: b.FunctionName,
				ErrorCode:    b.ErrorCode,
				Symptom:      b.Symptom,
				RootCause:    b.RootCause,
				CreatedAt:    now,
				UpdatedAt:    now,
			})
			if err != nil {
				return err
			}
			id = g.ID
			if err := writeTroubleshootingSteps(c, tq, id, b.Steps); err != nil {
				return err
			}
			return tx.Commit()
		}()
		if err != nil {
			return serverError(c, "create troubleshooting guide", err)
		}
		out, err := troubleshootingDetail(c, q, id)
		if err != nil {
			return serverError(c, "read created troubleshooting guide", err)
		}
		return c.Status(fiber.StatusCreated).JSON(out)
	}
}

func listRelatedTroubleshootingGuides(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		if _, err := q.GetTroubleshootingGuide(c.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting guide not found"})
			}
			return serverError(c, "read troubleshooting guide for related guides", err)
		}
		rows, err := q.GetTroubleshootingRelatedGuides(c.Context(), id)
		if err != nil {
			return serverError(c, "list related troubleshooting guides", err)
		}
		items := make([]troubleshootingListItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, troubleshootingListItem{
				ID: row.ID, Title: row.Title, Category: row.Category, Severity: row.Severity,
				DeviceName: row.DeviceName, FunctionName: row.FunctionName, ErrorCode: row.ErrorCode,
				Symptom: row.Symptom, RootCause: row.RootCause, StepCount: row.StepCount,
				CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			})
		}
		return c.JSON(fiber.Map{"items": items, "total": len(items)})
	}
}

func getTroubleshootingGuide(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		out, err := troubleshootingDetail(c, q, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting guide not found"})
			}
			return serverError(c, "read troubleshooting guide", err)
		}
		return c.JSON(out)
	}
}

func updateTroubleshootingGuide(q *store.Queries, d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		var b troubleshootingBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if fields := b.validate(); len(fields) > 0 {
			return validationError(c, fields)
		}
		fields, err := validateTroubleshootingRegistryKeys(c, q, b.Category, b.Severity)
		if err != nil {
			return serverError(c, "validate troubleshooting registry keys", err)
		}
		if len(fields) > 0 {
			return registryFieldError(c, fields)
		}
		if _, err := q.GetTroubleshootingGuide(c.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting guide not found"})
			}
			return serverError(c, "read troubleshooting guide", err)
		}
		now := time.Now().UTC().Format(time.RFC3339)

		err = func() error {
			tx, err := d.BeginTx(c.Context(), nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			tq := q.WithTx(tx)
			if err := tq.UpdateTroubleshootingGuideMeta(c.Context(), store.UpdateTroubleshootingGuideMetaParams{
				Title:        b.Title,
				Category:     b.Category,
				Severity:     b.Severity,
				DeviceName:   b.DeviceName,
				FunctionName: b.FunctionName,
				ErrorCode:    b.ErrorCode,
				Symptom:      b.Symptom,
				RootCause:    b.RootCause,
				UpdatedAt:    now,
				ID:           id,
			}); err != nil {
				return err
			}
			if err := writeTroubleshootingSteps(c, tq, id, b.Steps); err != nil {
				return err
			}
			return tx.Commit()
		}()
		if err != nil {
			return serverError(c, "update troubleshooting guide", err)
		}
		out, err := troubleshootingDetail(c, q, id)
		if err != nil {
			return serverError(c, "read updated troubleshooting guide", err)
		}
		return c.JSON(out)
	}
}

func deleteTroubleshootingGuide(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		affected, err := q.DeleteTroubleshootingGuide(c.Context(), id)
		if err != nil {
			return serverError(c, "delete troubleshooting guide", err)
		}
		if affected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting guide not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
