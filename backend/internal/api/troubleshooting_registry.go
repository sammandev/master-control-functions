package api

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/text/unicode/norm"

	"webprogfndb/internal/store"
)

const (
	maxTroubleshootingRegistryLabelRunes       = 64
	maxTroubleshootingRegistryDescriptionRunes = 240
	maxTroubleshootingRegistryKeyBytes         = 64
)

var validTroubleshootingTones = map[string]bool{
	"destructive": true,
	"warn":        true,
	"default":     true,
	"outline":     true,
}

type troubleshootingRegistryBody struct {
	Label       *string `json:"label"`
	Description *string `json:"description"`
	SortOrder   *int64  `json:"sort_order"`
	Tone        *string `json:"tone"`
}

type troubleshootingRegistryResponse struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	SortOrder   int64  `json:"sort_order"`
	BuiltIn     bool   `json:"built_in"`
	GuideCount  int64  `json:"guide_count"`
	Tone        string `json:"tone,omitempty"`
}

func registerTroubleshootingRegistryRoutes(app *fiber.App, q *store.Queries, d *sql.DB) {
	app.Get("/api/troubleshooting/categories", listTroubleshootingCategories(q))
	app.Post("/api/troubleshooting/categories", createTroubleshootingCategory(q))
	app.Put("/api/troubleshooting/categories/:key", updateTroubleshootingCategory(q))
	app.Delete("/api/troubleshooting/categories/:key", deleteTroubleshootingCategory(q, d))
	app.Get("/api/troubleshooting/severities", listTroubleshootingSeverities(q))
	app.Post("/api/troubleshooting/severities", createTroubleshootingSeverity(q))
	app.Put("/api/troubleshooting/severities/:key", updateTroubleshootingSeverity(q))
	app.Delete("/api/troubleshooting/severities/:key", deleteTroubleshootingSeverity(q, d))
}

func slugTroubleshootingKey(label string) string {
	var out strings.Builder
	separator := false
	for _, r := range norm.NFKD.String(strings.ToLower(strings.TrimSpace(label))) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			if separator && out.Len() > 0 {
				out.WriteByte('-')
			}
			out.WriteRune(r)
			separator = false
		} else {
			separator = true
		}
	}
	if out.Len() == 0 {
		return "item"
	}
	return capTroubleshootingKey(out.String(), maxTroubleshootingRegistryKeyBytes)
}

func capTroubleshootingKey(key string, limit int) string {
	if len(key) > limit {
		key = key[:limit]
	}
	key = strings.TrimRight(key, "-")
	if key == "" {
		return "item"
	}
	return key
}

func nextTroubleshootingKey(label string, existing map[string]bool) string {
	base := slugTroubleshootingKey(label)
	if !existing[base] {
		return base
	}
	for suffix := 2; ; suffix++ {
		token := "-" + strconv.Itoa(suffix)
		candidateBase := capTroubleshootingKey(base, maxTroubleshootingRegistryKeyBytes-len(token))
		candidate := candidateBase + token
		if !existing[candidate] {
			return candidate
		}
	}
}

func bindTroubleshootingRegistryBody(c fiber.Ctx) (troubleshootingRegistryBody, error) {
	var body troubleshootingRegistryBody
	if err := c.Bind().Body(&body); err != nil {
		return body, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}
	return body, nil
}

func registryLabelProblem(label string) string {
	if label == "" {
		return "is required"
	}
	if utf8.RuneCountInString(label) > maxTroubleshootingRegistryLabelRunes {
		return fmt.Sprintf("is too long (max %d characters)", maxTroubleshootingRegistryLabelRunes)
	}
	return ""
}

func registryDescriptionProblem(description string) string {
	if utf8.RuneCountInString(description) > maxTroubleshootingRegistryDescriptionRunes {
		return fmt.Sprintf("is too long (max %d characters)", maxTroubleshootingRegistryDescriptionRunes)
	}
	return ""
}

func categoryValues(body troubleshootingRegistryBody, existing *store.TroubleshootingCategory) (string, string, int64, map[string]string) {
	fields := map[string]string{}
	label := ""
	if existing != nil {
		label = existing.Label
	}
	if body.Label != nil {
		label = strings.TrimSpace(*body.Label)
	}
	if problem := registryLabelProblem(label); problem != "" {
		fields["label"] = problem
	}
	description := ""
	if existing != nil {
		description = existing.Description
	}
	if body.Description != nil {
		description = strings.TrimSpace(*body.Description)
	}
	if problem := registryDescriptionProblem(description); problem != "" {
		fields["description"] = problem
	}
	sortOrder := int64(0)
	if existing != nil {
		sortOrder = existing.SortOrder
	}
	if body.SortOrder != nil {
		sortOrder = *body.SortOrder
		if sortOrder <= 0 {
			fields["sort_order"] = "must be greater than zero"
		}
	}
	return label, description, sortOrder, fields
}

func severityValues(body troubleshootingRegistryBody, existing *store.TroubleshootingSeverity) (string, string, int64, string, map[string]string) {
	fields := map[string]string{}
	label := ""
	if existing != nil {
		label = existing.Label
	}
	if body.Label != nil {
		label = strings.TrimSpace(*body.Label)
	}
	if problem := registryLabelProblem(label); problem != "" {
		fields["label"] = problem
	}
	description := ""
	if existing != nil {
		description = existing.Description
	}
	if body.Description != nil {
		description = strings.TrimSpace(*body.Description)
	}
	if problem := registryDescriptionProblem(description); problem != "" {
		fields["description"] = problem
	}
	sortOrder := int64(0)
	if existing != nil {
		sortOrder = existing.SortOrder
	}
	if body.SortOrder != nil {
		sortOrder = *body.SortOrder
		if sortOrder <= 0 {
			fields["sort_order"] = "must be greater than zero"
		}
	}
	tone := "default"
	if existing != nil {
		tone = existing.Tone
	}
	if body.Tone != nil {
		tone = strings.ToLower(strings.TrimSpace(*body.Tone))
	}
	if !validTroubleshootingTones[tone] {
		fields["tone"] = "must be one of: destructive, warn, default, outline"
	}
	return label, description, sortOrder, tone, fields
}

func categoryLabelExists(rows []store.TroubleshootingCategory, label, exceptKey string) bool {
	for _, row := range rows {
		if row.Key != exceptKey && strings.EqualFold(row.Label, label) {
			return true
		}
	}
	return false
}

func severityLabelExists(rows []store.TroubleshootingSeverity, label, exceptKey string) bool {
	for _, row := range rows {
		if row.Key != exceptKey && strings.EqualFold(row.Label, label) {
			return true
		}
	}
	return false
}

func nextCategoryOrder(rows []store.TroubleshootingCategory) int64 {
	var max int64
	for _, row := range rows {
		if row.SortOrder > max {
			max = row.SortOrder
		}
	}
	return max + 1
}

func nextSeverityOrder(rows []store.TroubleshootingSeverity) int64 {
	var max int64
	for _, row := range rows {
		if row.SortOrder > max {
			max = row.SortOrder
		}
	}
	return max + 1
}

func createTroubleshootingCategory(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		body, bindErr := bindTroubleshootingRegistryBody(c)
		if bindErr != nil {
			return bindErr
		}
		body.SortOrder = nil
		label, description, _, fields := categoryValues(body, nil)
		if len(fields) > 0 {
			return validationError(c, fields)
		}
		rows, err := q.ListTroubleshootingCategories(c.Context())
		if err != nil {
			return serverError(c, "list troubleshooting categories", err)
		}
		if categoryLabelExists(rows, label, "") {
			return validationError(c, map[string]string{"label": "already exists"})
		}
		keys := make(map[string]bool, len(rows))
		for _, row := range rows {
			keys[row.Key] = true
		}
		row, err := q.InsertTroubleshootingCategory(c.Context(), store.InsertTroubleshootingCategoryParams{
			Key:         nextTroubleshootingKey(label, keys),
			Label:       label,
			Description: description,
			SortOrder:   nextCategoryOrder(rows),
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				return validationError(c, map[string]string{"label": "already exists"})
			}
			return serverError(c, "create troubleshooting category", err)
		}
		return c.Status(fiber.StatusCreated).JSON(categoryRegistryResponse(row, 0))
	}
}

func createTroubleshootingSeverity(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		body, bindErr := bindTroubleshootingRegistryBody(c)
		if bindErr != nil {
			return bindErr
		}
		body.SortOrder = nil
		label, description, _, tone, fields := severityValues(body, nil)
		if len(fields) > 0 {
			return validationError(c, fields)
		}
		rows, err := q.ListTroubleshootingSeverities(c.Context())
		if err != nil {
			return serverError(c, "list troubleshooting severities", err)
		}
		if severityLabelExists(rows, label, "") {
			return validationError(c, map[string]string{"label": "already exists"})
		}
		keys := make(map[string]bool, len(rows))
		for _, row := range rows {
			keys[row.Key] = true
		}
		row, err := q.InsertTroubleshootingSeverity(c.Context(), store.InsertTroubleshootingSeverityParams{
			Key:         nextTroubleshootingKey(label, keys),
			Label:       label,
			Description: description,
			SortOrder:   nextSeverityOrder(rows),
			Tone:        tone,
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				return validationError(c, map[string]string{"label": "already exists"})
			}
			return serverError(c, "create troubleshooting severity", err)
		}
		return c.Status(fiber.StatusCreated).JSON(severityRegistryResponse(row, 0))
	}
}

func updateTroubleshootingCategory(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := c.Params("key")
		existing, err := q.GetTroubleshootingCategory(c.Context(), key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting category not found"})
			}
			return serverError(c, "read troubleshooting category", err)
		}
		body, bindErr := bindTroubleshootingRegistryBody(c)
		if bindErr != nil {
			return bindErr
		}
		label, description, sortOrder, fields := categoryValues(body, &existing)
		if len(fields) == 0 {
			rows, listErr := q.ListTroubleshootingCategories(c.Context())
			if listErr != nil {
				return serverError(c, "list troubleshooting categories", listErr)
			}
			if categoryLabelExists(rows, label, key) {
				fields["label"] = "already exists"
			}
		}
		if len(fields) > 0 {
			return validationError(c, fields)
		}
		row, err := q.UpdateTroubleshootingCategory(c.Context(), store.UpdateTroubleshootingCategoryParams{
			Label: label, Description: description, SortOrder: sortOrder, Key: key,
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				return validationError(c, map[string]string{"label": "already exists"})
			}
			return serverError(c, "update troubleshooting category", err)
		}
		count, err := registryCategoryCount(c, q, key)
		if err != nil {
			return serverError(c, "count troubleshooting category guides", err)
		}
		return c.JSON(categoryRegistryResponse(row, count))
	}
}

func updateTroubleshootingSeverity(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := c.Params("key")
		existing, err := q.GetTroubleshootingSeverity(c.Context(), key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting severity not found"})
			}
			return serverError(c, "read troubleshooting severity", err)
		}
		body, bindErr := bindTroubleshootingRegistryBody(c)
		if bindErr != nil {
			return bindErr
		}
		label, description, sortOrder, tone, fields := severityValues(body, &existing)
		if len(fields) == 0 {
			rows, listErr := q.ListTroubleshootingSeverities(c.Context())
			if listErr != nil {
				return serverError(c, "list troubleshooting severities", listErr)
			}
			if severityLabelExists(rows, label, key) {
				fields["label"] = "already exists"
			}
		}
		if len(fields) > 0 {
			return validationError(c, fields)
		}
		row, err := q.UpdateTroubleshootingSeverity(c.Context(), store.UpdateTroubleshootingSeverityParams{
			Label: label, Description: description, SortOrder: sortOrder, Tone: tone, Key: key,
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				return validationError(c, map[string]string{"label": "already exists"})
			}
			return serverError(c, "update troubleshooting severity", err)
		}
		count, err := registrySeverityCount(c, q, key)
		if err != nil {
			return serverError(c, "count troubleshooting severity guides", err)
		}
		return c.JSON(severityRegistryResponse(row, count))
	}
}

func troubleshootingGuideRegistryCounts(c fiber.Ctx, q *store.Queries) (map[string]int64, map[string]int64, error) {
	total, err := q.CountTroubleshootingGuides(c.Context(), store.CountTroubleshootingGuidesParams{})
	if err != nil {
		return nil, nil, err
	}
	rows, err := q.ListTroubleshootingGuides(c.Context(), store.ListTroubleshootingGuidesParams{Limit: total})
	if err != nil {
		return nil, nil, err
	}
	categories := map[string]int64{}
	severities := map[string]int64{}
	for _, row := range rows {
		categories[row.Category]++
		severities[row.Severity]++
	}
	return categories, severities, nil
}

func registryCategoryCount(c fiber.Ctx, q *store.Queries, key string) (int64, error) {
	categories, _, err := troubleshootingGuideRegistryCounts(c, q)
	return categories[key], err
}

func registrySeverityCount(c fiber.Ctx, q *store.Queries, key string) (int64, error) {
	_, severities, err := troubleshootingGuideRegistryCounts(c, q)
	return severities[key], err
}

func listTroubleshootingCategories(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		rows, err := q.ListTroubleshootingCategories(c.Context())
		if err != nil {
			return serverError(c, "list troubleshooting categories", err)
		}
		counts, _, err := troubleshootingGuideRegistryCounts(c, q)
		if err != nil {
			return serverError(c, "count troubleshooting categories", err)
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].SortOrder != rows[j].SortOrder {
				return rows[i].SortOrder < rows[j].SortOrder
			}
			return strings.ToLower(rows[i].Label) < strings.ToLower(rows[j].Label)
		})
		items := make([]troubleshootingRegistryResponse, 0, len(rows))
		for _, row := range rows {
			items = append(items, categoryRegistryResponse(row, counts[row.Key]))
		}
		return c.JSON(fiber.Map{"items": items, "total": len(items)})
	}
}

func listTroubleshootingSeverities(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		rows, err := q.ListTroubleshootingSeverities(c.Context())
		if err != nil {
			return serverError(c, "list troubleshooting severities", err)
		}
		_, counts, err := troubleshootingGuideRegistryCounts(c, q)
		if err != nil {
			return serverError(c, "count troubleshooting severities", err)
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].SortOrder != rows[j].SortOrder {
				return rows[i].SortOrder < rows[j].SortOrder
			}
			return strings.ToLower(rows[i].Label) < strings.ToLower(rows[j].Label)
		})
		items := make([]troubleshootingRegistryResponse, 0, len(rows))
		for _, row := range rows {
			items = append(items, severityRegistryResponse(row, counts[row.Key]))
		}
		return c.JSON(fiber.Map{"items": items, "total": len(items)})
	}
}

func categoryRegistryResponse(row store.TroubleshootingCategory, guideCount int64) troubleshootingRegistryResponse {
	return troubleshootingRegistryResponse{
		Key: row.Key, Label: row.Label, Description: row.Description,
		SortOrder: row.SortOrder, BuiltIn: row.BuiltIn, GuideCount: guideCount,
	}
}

func severityRegistryResponse(row store.TroubleshootingSeverity, guideCount int64) troubleshootingRegistryResponse {
	return troubleshootingRegistryResponse{
		Key: row.Key, Label: row.Label, Description: row.Description,
		SortOrder: row.SortOrder, BuiltIn: row.BuiltIn, Tone: row.Tone, GuideCount: guideCount,
	}
}

func deleteTroubleshootingCategory(q *store.Queries, d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		return deleteTroubleshootingRegistryItem(c, q, d, true)
	}
}

func deleteTroubleshootingSeverity(q *store.Queries, d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		return deleteTroubleshootingRegistryItem(c, q, d, false)
	}
}

func deleteTroubleshootingRegistryItem(c fiber.Ctx, q *store.Queries, d *sql.DB, category bool) error {
	key := c.Params("key")
	replacement := strings.ToLower(strings.TrimSpace(c.Query("reassign_to")))
	if replacement == "" {
		return validationError(c, map[string]string{"reassign_to": "is required"})
	}
	if replacement == key {
		return validationError(c, map[string]string{"reassign_to": "must be different from the deleted key"})
	}

	tx, err := d.BeginTx(c.Context(), nil)
	if err != nil {
		return serverError(c, "begin troubleshooting registry deletion", err)
	}
	defer tx.Rollback()
	tq := q.WithTx(tx)
	if category {
		source, err := tq.GetTroubleshootingCategory(c.Context(), key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting category not found"})
			}
			return serverError(c, "read troubleshooting category", err)
		}
		if source.BuiltIn {
			return validationError(c, map[string]string{"key": "built-in entries cannot be deleted"})
		}
		if _, err := tq.GetTroubleshootingCategory(c.Context(), replacement); err != nil {
			return troubleshootingReplacementError(c, tq, replacement, err, false)
		}
		if _, err := tq.ReassignTroubleshootingGuideCategories(c.Context(), store.ReassignTroubleshootingGuideCategoriesParams{
			NewKey: replacement, UpdatedAt: time.Now().UTC().Format(time.RFC3339), OldKey: key,
		}); err != nil {
			return serverError(c, "reassign troubleshooting guide categories", err)
		}
		if affected, err := tq.DeleteTroubleshootingCategory(c.Context(), key); err != nil {
			return serverError(c, "delete troubleshooting category", err)
		} else if affected != 1 {
			return validationError(c, map[string]string{"key": "cannot be deleted"})
		}
	} else {
		source, err := tq.GetTroubleshootingSeverity(c.Context(), key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "troubleshooting severity not found"})
			}
			return serverError(c, "read troubleshooting severity", err)
		}
		if source.BuiltIn {
			return validationError(c, map[string]string{"key": "built-in entries cannot be deleted"})
		}
		if _, err := tq.GetTroubleshootingSeverity(c.Context(), replacement); err != nil {
			return troubleshootingReplacementError(c, tq, replacement, err, true)
		}
		if _, err := tq.ReassignTroubleshootingGuideSeverities(c.Context(), store.ReassignTroubleshootingGuideSeveritiesParams{
			NewKey: replacement, UpdatedAt: time.Now().UTC().Format(time.RFC3339), OldKey: key,
		}); err != nil {
			return serverError(c, "reassign troubleshooting guide severities", err)
		}
		if affected, err := tq.DeleteTroubleshootingSeverity(c.Context(), key); err != nil {
			return serverError(c, "delete troubleshooting severity", err)
		} else if affected != 1 {
			return validationError(c, map[string]string{"key": "cannot be deleted"})
		}
	}
	if err := tx.Commit(); err != nil {
		return serverError(c, "commit troubleshooting registry deletion", err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func troubleshootingReplacementError(c fiber.Ctx, q *store.Queries, replacement string, err error, checkingSeverity bool) error {
	if !errors.Is(err, sql.ErrNoRows) {
		return serverError(c, "read troubleshooting replacement", err)
	}
	var oppositeErr error
	if checkingSeverity {
		_, oppositeErr = q.GetTroubleshootingCategory(c.Context(), replacement)
	} else {
		_, oppositeErr = q.GetTroubleshootingSeverity(c.Context(), replacement)
	}
	message, classifyErr := classifyTroubleshootingReplacementError(err, oppositeErr)
	if classifyErr != nil {
		return serverError(c, "read troubleshooting replacement", classifyErr)
	}
	return validationError(c, map[string]string{"reassign_to": message})
}

func classifyTroubleshootingReplacementError(replacementErr, oppositeErr error) (string, error) {
	if replacementErr != nil && !errors.Is(replacementErr, sql.ErrNoRows) {
		return "", replacementErr
	}
	if oppositeErr == nil {
		return "must reference the same registry type", nil
	}
	if !errors.Is(oppositeErr, sql.ErrNoRows) {
		return "", oppositeErr
	}
	return "does not exist", nil
}
