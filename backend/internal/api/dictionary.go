package api

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// validTypes is the Type vocabulary observed in the knowledge workbook:
// int, double, string, list[int].
var validTypes = map[string]bool{
	"int":       true,
	"double":    true,
	"string":    true,
	"list[int]": true,
}

// validCardinalities is the Cardinality vocabulary from CONTEXT.md.
var validCardinalities = map[string]bool{
	"SINGLE":   true,
	"MULTIPLE": true,
}

// definitionBody is the JSON shape of a Parameter/Criteria definition on
// create/update. The two dictionaries share one shape (CONTEXT.md), so
// validation and response mapping are shared; only the store calls differ.
type definitionBody struct {
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	Type          string  `json:"type"`
	Unit          *string `json:"unit"`
	DefaultValue  *string `json:"default_value"`
	AllowedValues *string `json:"allowed_values"`
	Cardinality   *string `json:"cardinality"`
	ValuePattern  *string `json:"value_pattern"`
	Required      *bool   `json:"required"`
	// Device scope (ADR-0009): blank = global definition; an explicit
	// scope overrides the global one for matching device-scoped links.
	DeviceName *string `json:"device_name"`
}

// validateDefinition enforces the definitions' own constraints:
// - name required
// - type one of the known vocabulary
// - cardinality SINGLE/MULTIPLE
// - allowed_values must be a `{v1,v2,...}` set whose members parse per type
// - default_value must parse per type and (when a set exists) be a member
func validateDefinition(b *definitionBody) map[string]string {
	fields := map[string]string{}
	b.Name = strings.TrimSpace(b.Name)
	b.Type = strings.TrimSpace(b.Type)
	if b.DeviceName != nil {
		// ADR-0002: device values are trimmed and uppercased, matching the
		// registry's rule so a casing twin can never diverge from a
		// registered Device.
		*b.DeviceName = normalizeDeviceName(*b.DeviceName)
		if *b.DeviceName == "" {
			b.DeviceName = nil
		}
	}

	if b.Name == "" {
		fields["name"] = "is required"
	}
	if b.Type == "" {
		b.Type = "string"
	}
	if !validTypes[b.Type] {
		fields["type"] = "must be int, double, string or list[int]"
	}

	cardinality := "SINGLE"
	if b.Cardinality != nil && strings.TrimSpace(*b.Cardinality) != "" {
		cardinality = strings.ToUpper(strings.TrimSpace(*b.Cardinality))
	}
	if !validCardinalities[cardinality] {
		fields["cardinality"] = "must be SINGLE or MULTIPLE"
	}
	b.Cardinality = &cardinality

	raw := ""
	if b.AllowedValues != nil {
		raw = strings.TrimSpace(*b.AllowedValues)
		b.AllowedValues = &raw
	}
	var rule allowedRule
	// "{}" predates the MCP grammar and still means "unconstrained" here;
	// MCP readiness reports it.
	if raw != "" && raw != "{}" {
		parsed, err := parseAllowed(raw)
		if err != nil {
			fields["allowed_values"] = err.Error()
		} else {
			rule = parsed
			for _, v := range rule.set {
				if !valueParsesAs(b.Type, v) {
					fields["allowed_values"] = "value '" + v + "' does not match type " + b.Type
					break
				}
			}
			if rule.kind == "range" && b.Type != "int" && b.Type != "double" && b.Type != "list[int]" {
				fields["allowed_values"] = "a numeric range requires type int, double or list[int]"
			}
		}
	}

	if b.DefaultValue != nil {
		dv := strings.TrimSpace(*b.DefaultValue)
		b.DefaultValue = &dv
		if dv != "" {
			if !valueParsesAs(b.Type, dv) {
				fields["default_value"] = "'" + dv + "' does not match type " + b.Type
			} else if !defaultAllowed(rule, b.Type, dv) {
				fields["default_value"] = "'" + dv + "' is not permitted by the allowed values"
			}
		}
	}

	if b.ValuePattern != nil {
		vp := strings.TrimSpace(*b.ValuePattern)
		b.ValuePattern = &vp
		if vp != "" && fields["type"] == "" && fields["cardinality"] == "" {
			if _, err := compileValuePattern(vp, b.Type, cardinality); err != nil {
				fields["value_pattern"] = err.Error()
			}
		}
	}
	return fields
}

// defaultAllowed checks every comma-separated member of a list default.
func defaultAllowed(rule allowedRule, defType, value string) bool {
	if defType != "list[int]" {
		return rule.accepts(value)
	}
	for _, part := range strings.Split(value, ",") {
		if !rule.accepts(strings.TrimSpace(part)) {
			return false
		}
	}
	return true
}

// valueParsesAs reports whether a single value literal parses as the given
// definition type.
func valueParsesAs(defType, v string) bool {
	switch defType {
	case "int":
		_, err := strconv.ParseInt(v, 10, 64)
		return err == nil
	case "double":
		_, err := strconv.ParseFloat(v, 64)
		return err == nil
	case "list[int]":
		for _, part := range strings.Split(v, ",") {
			if _, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err != nil {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// dictionaryRoutes bundles the shared pieces one dictionary's handlers need.
type dictionaryRoutes[T any] struct {
	get   func(c fiber.Ctx, id int64) (T, error)
	list  func(c fiber.Ctx, search *string) ([]T, error)
	count func(c fiber.Ctx, search *string) (int64, error)
	// usage returns (total, orphaned): links referencing the definition by
	// (name, device_name), and how many of those point at a missing
	// Function (ADR-0003). ADR-0009 makes the count scope-aware: a scoped
	// definition counts links of its own scope; a global definition counts
	// links whose device has no scoped sibling.
	usage    func(c context.Context, name, deviceName string) (int64, int64, error)
	insert   func(c fiber.Ctx, b definitionBody, now string) (T, error)
	update   func(c fiber.Ctx, id int64, b definitionBody, now string) (T, error)
	remove   func(c fiber.Ctx, id int64) (int64, error)
	toResp   func(T, int64, int64) fiber.Map
	nameOf   func(T) string
	scopeOf  func(T) string
	sortKeys map[string]func(T) string
}

func registerDictionary[T any](app *fiber.App, base string, r dictionaryRoutes[T]) {
	app.Post(base, r.create())
	app.Get(base, r.listHandler())
	app.Get(base+"/:id", r.getHandler())
	app.Put(base+"/:id", r.updateHandler())
	app.Delete(base+"/:id", r.deleteHandler())
}

// parseDefinitionBody parses and validates the request body.
func parseDefinitionBody(c fiber.Ctx) (*definitionBody, error) {
	var b definitionBody
	if err := c.Bind().Body(&b); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	if fields := validateDefinition(&b); len(fields) > 0 {
		return nil, &fieldValidationError{fields: fields}
	}
	return &b, nil
}

// fieldValidationError carries per-field messages through the error
// handler chain.
type fieldValidationError struct {
	fields map[string]string
}

func (e *fieldValidationError) Error() string { return "validation failed" }

// definitionError handles parse/validation failures uniformly.
func definitionWriteError(c fiber.Ctx, err error) error {
	if fe, ok := err.(*fieldValidationError); ok {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":  "validation failed",
			"fields": fe.fields,
		})
	}
	if fe, ok := err.(*fiber.Error); ok {
		return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
	}
	return serverError(c, "dictionary write", err)
}

func (r dictionaryRoutes[T]) create() fiber.Handler {
	return func(c fiber.Ctx) error {
		b, err := parseDefinitionBody(c)
		if err != nil {
			return definitionWriteError(c, err)
		}
		now := time.Now().UTC().Format(time.RFC3339)
		item, err := r.insert(c, *b, now)
		if err != nil {
			if store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "definition '" + b.Name + "' already exists",
				})
			}
			return serverError(c, "create definition", err)
		}
		return c.Status(fiber.StatusCreated).JSON(r.toResp(item, 0, 0))
	}
}

func (r dictionaryRoutes[T]) listHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		lq := parseListQuery(c)
		var search *string
		if pattern := likePattern(lq.search); pattern != "" {
			search = &pattern
		}
		total, err := r.count(c, search)
		if err != nil {
			return serverError(c, "count definitions", err)
		}
		items, err := r.list(c, search)
		if err != nil {
			return serverError(c, "list definitions", err)
		}

		// Whitelist against THIS dictionary's sort keys — functions.go's
		// sortColumns admits different keys ("interface", "source") that
		// would nil-panic here.
		if keyFn := r.sortKeys[lq.sortCol]; keyFn != nil {
			sort.SliceStable(items, func(i, j int) bool {
				a, b := keyFn(items[i]), keyFn(items[j])
				if a != b {
					if lq.sortDesc {
						return a > b
					}
					return a < b
				}
				// Deterministic tiebreak on name, mirroring the natural key.
				return r.nameOf(items[i]) < r.nameOf(items[j])
			})
		}
		start := (lq.page - 1) * lq.pageSize
		if start > int64(len(items)) {
			start = int64(len(items))
		}
		end := start + lq.pageSize
		if end > int64(len(items)) {
			end = int64(len(items))
		}
		page := items[start:end]

		out := make([]fiber.Map, 0, len(page))
		for _, item := range page {
			usage, orphanUsage, err := r.usage(c.Context(), r.nameOf(item), r.scopeOf(item))
			if err != nil {
				return serverError(c, "count usage", err)
			}
			out = append(out, r.toResp(item, usage, orphanUsage))
		}
		return c.JSON(fiber.Map{
			"items":     out,
			"total":     total,
			"page":      lq.page,
			"page_size": lq.pageSize,
		})
	}
}

func (r dictionaryRoutes[T]) getHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		item, err := r.get(c, id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "definition not found"})
		}
		usage, orphanUsage, err := r.usage(c.Context(), r.nameOf(item), r.scopeOf(item))
		if err != nil {
			return serverError(c, "count usage", err)
		}
		return c.JSON(r.toResp(item, usage, orphanUsage))
	}
}

func (r dictionaryRoutes[T]) updateHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		b, err := parseDefinitionBody(c)
		if err != nil {
			return definitionWriteError(c, err)
		}
		if _, err := r.get(c, id); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "definition not found"})
		}
		now := time.Now().UTC().Format(time.RFC3339)
		item, err := r.update(c, id, *b, now)
		if err != nil {
			if store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "definition '" + b.Name + "' already exists",
				})
			}
			return serverError(c, "update definition", err)
		}
		usage, orphanUsage, err := r.usage(c.Context(), r.nameOf(item), r.scopeOf(item))
		if err != nil {
			return serverError(c, "count usage", err)
		}
		return c.JSON(r.toResp(item, usage, orphanUsage))
	}
}

func (r dictionaryRoutes[T]) deleteHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		item, err := r.get(c, id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "definition not found"})
		}
		usage, _, err := r.usage(c.Context(), r.nameOf(item), r.scopeOf(item))
		if err != nil {
			return serverError(c, "count usage", err)
		}
		// A referenced definition is not silently deletable: the caller
		// must force it, and surviving links become Orphans (ADR-0003).
		if usage > 0 && c.Query("force") != "true" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":       "definition is referenced by " + strconv.FormatInt(usage, 10) + " link(s)",
				"usage_count": usage,
			})
		}
		if _, err := r.remove(c, id); err != nil {
			return serverError(c, "delete definition", err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// registerParameterDictionary wires the parameter-definitions endpoints.
func registerParameterDictionary(app *fiber.App, q *store.Queries) {
	r := dictionaryRoutes[store.ParameterDefinition]{
		get: func(c fiber.Ctx, id int64) (store.ParameterDefinition, error) {
			return q.GetParameterDefinition(c.Context(), id)
		},
		list: func(c fiber.Ctx, search *string) ([]store.ParameterDefinition, error) {
			return q.ListParameterDefinitions(c.Context(), nullStrPtr(search))
		},
		count: func(c fiber.Ctx, search *string) (int64, error) {
			return q.CountParameterDefinitions(c.Context(), nullStrPtr(search))
		},
		usage: func(c context.Context, name, deviceName string) (int64, int64, error) {
			// ADR-0009 scope-aware count: a scoped definition (device_name !=
			// '') counts links of its own scope; a global definition counts
			// links whose device has no scoped sibling — the links that
			// actually resolve to it under exact-scope-then-global.
			var total int64
			var err error
			if deviceName == "" {
				total, err = q.CountParameterUsageGlobal(c, name)
			} else {
				total, err = q.CountParameterUsageScoped(c, store.CountParameterUsageScopedParams{
					Name:       name,
					DeviceName: deviceName,
				})
			}
			if err != nil {
				return 0, 0, err
			}
			orphaned, err := q.CountParameterOrphanUsage(c, name)
			return total, orphaned, err
		},
		insert: func(c fiber.Ctx, b definitionBody, now string) (store.ParameterDefinition, error) {
			return q.InsertParameterDefinition(c.Context(), store.InsertParameterDefinitionParams{
				Name:          b.Name,
				Description:   deref(b.Description),
				Type:          b.Type,
				Unit:          deref(b.Unit),
				DefaultValue:  nullStrPtr(b.DefaultValue),
				AllowedValues: deref(b.AllowedValues),
				Cardinality:   deref(b.Cardinality),
				ValuePattern:  deref(b.ValuePattern),
				Required:      boolToInt64(b.Required),
				DeviceName:    deref(b.DeviceName),
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		},
		update: func(c fiber.Ctx, id int64, b definitionBody, now string) (store.ParameterDefinition, error) {
			return q.UpdateParameterDefinition(c.Context(), store.UpdateParameterDefinitionParams{
				ID:            id,
				Name:          b.Name,
				Description:   deref(b.Description),
				Type:          b.Type,
				Unit:          deref(b.Unit),
				DefaultValue:  nullStrPtr(b.DefaultValue),
				AllowedValues: deref(b.AllowedValues),
				Cardinality:   deref(b.Cardinality),
				ValuePattern:  deref(b.ValuePattern),
				Required:      boolToInt64(b.Required),
				DeviceName:    deref(b.DeviceName),
				UpdatedAt:     now,
			})
		},
		remove: func(c fiber.Ctx, id int64) (int64, error) {
			return q.DeleteParameterDefinition(c.Context(), id)
		},
		toResp:  parameterDefinitionResponse,
		nameOf:  func(d store.ParameterDefinition) string { return d.Name },
		scopeOf: func(d store.ParameterDefinition) string { return d.DeviceName },
		sortKeys: map[string]func(store.ParameterDefinition) string{
			"name":       func(d store.ParameterDefinition) string { return d.Name },
			"type":       func(d store.ParameterDefinition) string { return d.Type },
			"updated_at": func(d store.ParameterDefinition) string { return d.UpdatedAt },
		},
	}
	registerDictionary(app, "/api/parameter-definitions", r)
}

// registerCriteriaDictionary wires the criteria-definitions endpoints.
func registerCriteriaDictionary(app *fiber.App, q *store.Queries) {
	r := dictionaryRoutes[store.CriteriaDefinition]{
		get: func(c fiber.Ctx, id int64) (store.CriteriaDefinition, error) {
			return q.GetCriteriaDefinition(c.Context(), id)
		},
		list: func(c fiber.Ctx, search *string) ([]store.CriteriaDefinition, error) {
			return q.ListCriteriaDefinitions(c.Context(), nullStrPtr(search))
		},
		count: func(c fiber.Ctx, search *string) (int64, error) {
			return q.CountCriteriaDefinitions(c.Context(), nullStrPtr(search))
		},
		usage: func(c context.Context, name, deviceName string) (int64, int64, error) {
			// ADR-0009 scope-aware count: see the parameter dictionary's
			// usage closure for the exact-scope-then-global rule.
			var total int64
			var err error
			if deviceName == "" {
				total, err = q.CountCriteriaUsageGlobal(c, name)
			} else {
				total, err = q.CountCriteriaUsageScoped(c, store.CountCriteriaUsageScopedParams{
					Name:       name,
					DeviceName: deviceName,
				})
			}
			if err != nil {
				return 0, 0, err
			}
			orphaned, err := q.CountCriteriaOrphanUsage(c, name)
			return total, orphaned, err
		},
		insert: func(c fiber.Ctx, b definitionBody, now string) (store.CriteriaDefinition, error) {
			return q.InsertCriteriaDefinition(c.Context(), store.InsertCriteriaDefinitionParams{
				Name:          b.Name,
				Description:   deref(b.Description),
				Type:          b.Type,
				Unit:          deref(b.Unit),
				DefaultValue:  nullStrPtr(b.DefaultValue),
				AllowedValues: deref(b.AllowedValues),
				Cardinality:   deref(b.Cardinality),
				ValuePattern:  deref(b.ValuePattern),
				Required:      boolToInt64(b.Required),
				DeviceName:    deref(b.DeviceName),
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		},
		update: func(c fiber.Ctx, id int64, b definitionBody, now string) (store.CriteriaDefinition, error) {
			return q.UpdateCriteriaDefinition(c.Context(), store.UpdateCriteriaDefinitionParams{
				ID:            id,
				Name:          b.Name,
				Description:   deref(b.Description),
				Type:          b.Type,
				Unit:          deref(b.Unit),
				DefaultValue:  nullStrPtr(b.DefaultValue),
				AllowedValues: deref(b.AllowedValues),
				Cardinality:   deref(b.Cardinality),
				ValuePattern:  deref(b.ValuePattern),
				Required:      boolToInt64(b.Required),
				DeviceName:    deref(b.DeviceName),
				UpdatedAt:     now,
			})
		},
		remove: func(c fiber.Ctx, id int64) (int64, error) {
			return q.DeleteCriteriaDefinition(c.Context(), id)
		},
		toResp:  criteriaDefinitionResponse,
		nameOf:  func(d store.CriteriaDefinition) string { return d.Name },
		scopeOf: func(d store.CriteriaDefinition) string { return d.DeviceName },
		sortKeys: map[string]func(store.CriteriaDefinition) string{
			"name":       func(d store.CriteriaDefinition) string { return d.Name },
			"type":       func(d store.CriteriaDefinition) string { return d.Type },
			"updated_at": func(d store.CriteriaDefinition) string { return d.UpdatedAt },
		},
	}
	registerDictionary(app, "/api/criteria-definitions", r)
}

func parameterDefinitionResponse(d store.ParameterDefinition, usage, orphanUsage int64) fiber.Map {
	return definitionResponseMap(
		d.ID, d.Name, d.Description, d.Type, d.Unit, d.DefaultValue,
		d.AllowedValues, d.Cardinality, d.ValuePattern, d.Required != 0,
		d.DeviceName, d.CreatedAt, d.UpdatedAt, usage, orphanUsage)
}

func criteriaDefinitionResponse(d store.CriteriaDefinition, usage, orphanUsage int64) fiber.Map {
	return definitionResponseMap(
		d.ID, d.Name, d.Description, d.Type, d.Unit, d.DefaultValue,
		d.AllowedValues, d.Cardinality, d.ValuePattern, d.Required != 0,
		d.DeviceName, d.CreatedAt, d.UpdatedAt, usage, orphanUsage)
}

// definitionResponseMap is the shared JSON shape for both dictionaries.
func definitionResponseMap(
	id int64, name, description, defType, unit string,
	defaultValue sql.NullString, allowedValues, cardinality, valuePattern string,
	required bool, deviceName, createdAt, updatedAt string, usage, orphanUsage int64,
) fiber.Map {
	var dv *string
	if defaultValue.Valid {
		v := defaultValue.String
		dv = &v
	}
	return fiber.Map{
		"id":                 id,
		"name":               name,
		"description":        description,
		"type":               defType,
		"unit":               unit,
		"default_value":      dv,
		"allowed_values":     allowedValues,
		"cardinality":        cardinality,
		"value_pattern":      valuePattern,
		"required":           required,
		"device_name":        deviceName,
		"created_at":         createdAt,
		"updated_at":         updatedAt,
		"usage_count":        usage,
		"orphan_usage_count": orphanUsage,
	}
}

// nullStrPtr converts an optional JSON string into a sql.NullString for
// the generated nullable-column parameter.
func nullStrPtr(p *string) sql.NullString {
	if p == nil || strings.TrimSpace(*p) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: strings.TrimSpace(*p), Valid: true}
}

func boolToInt64(b *bool) int64 {
	if b == nil || !*b {
		return 0
	}
	return 1
}
