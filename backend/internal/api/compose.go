package api

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// Script composer (ticket 07). Emits the workbook's 7-segment grammar:
//
//	<source><interface>|<FUNCTION>|<p1=v1&p2=v2...>|<criteria or NULL>|<timeout-ms>|<flow>|<on-fail>
//
// derived from the real example scripts (e.g.
// "DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=15&STANDARD=11AX&...|POWER=5dB|20000|6175_11AX_BW20_MCS9_CAL|FAIL").
// Parameter fields come from the function's parameter links in position
// order (NULL positions last, first link per name across device scopes);
// criteria fields likewise and emit as a name=value list or the literal
// NULL segment when none have values. Validation is field-level and runs
// before any script is emitted.

// composeField is one form field: a link enriched with its definition's
// constraints. Resolved=false means the linked definition is missing
// (orphan) — the field is still composed as free text (ADR-0003).
type composeField struct {
	Name          string `json:"name"`
	Position      *int64 `json:"position"`
	Required      bool   `json:"required"`
	DeviceName    string `json:"device_name"`
	Resolved      bool   `json:"resolved"`
	Type          string `json:"type"`
	Unit          string `json:"unit"`
	DefaultValue  string `json:"default_value"`
	AllowedValues string `json:"allowed_values"`
	Cardinality   string `json:"cardinality"`
}

func composeFieldFromRow(r store.ListComposerParameterFieldsRow) composeField {
	return fieldFromCommon(r.Name, r.Position, r.Required, r.DeviceName,
		r.Type, r.Unit, r.DefaultValue, r.AllowedValues, r.Cardinality)
}

func composeFieldFromCriteriaRow(r store.ListComposerCriteriaFieldsRow) composeField {
	return fieldFromCommon(r.Name, r.Position, r.Required, r.DeviceName,
		r.Type, r.Unit, r.DefaultValue, r.AllowedValues, r.Cardinality)
}

// fieldFromCommon builds the composeField from the raw column values the
// two generated row types share.
func fieldFromCommon(
	name string, position sql.NullInt64, required int64, deviceName string,
	typ, unit, defaultValue, allowedValues, cardinality sql.NullString,
) composeField {
	f := composeField{
		Name:       name,
		Required:   required != 0,
		DeviceName: deviceName,
	}
	if position.Valid {
		p := position.Int64
		f.Position = &p
	}
	if typ.Valid {
		f.Resolved = true
		f.Type = typ.String
	}
	if unit.Valid {
		f.Unit = unit.String
	}
	if defaultValue.Valid {
		f.DefaultValue = defaultValue.String
	}
	if allowedValues.Valid {
		f.AllowedValues = allowedValues.String
	}
	if cardinality.Valid {
		f.Cardinality = cardinality.String
	}
	return f
}

// dedupeFields keeps the first row per name — the same parameter can be
// linked under several device scopes, and one script has one value per
// field. Works for both row types: convert them to composeField first.
func dedupeParameterRows(rows []store.ListComposerParameterFieldsRow) []composeField {
	seen := map[string]bool{}
	fields := make([]composeField, 0, len(rows))
	for _, r := range rows {
		if seen[r.Name] {
			continue
		}
		seen[r.Name] = true
		fields = append(fields, composeFieldFromRow(r))
	}
	return fields
}

func dedupeCriteriaRows(rows []store.ListComposerCriteriaFieldsRow) []composeField {
	seen := map[string]bool{}
	fields := make([]composeField, 0, len(rows))
	for _, r := range rows {
		if seen[r.Name] {
			continue
		}
		seen[r.Name] = true
		fields = append(fields, composeFieldFromCriteriaRow(r))
	}
	return fields
}

// ---- grammar ------------------------------------------------------------

// composeRequest is the POST /api/functions/:id/compose body. All values
// are raw strings; the server parses per field type.
type composeRequest struct {
	Parameters map[string]string `json:"parameters"`
	Criteria   map[string]string `json:"criteria"`
	TimeoutMs  string            `json:"timeout_ms"`
	Flow       string            `json:"flow"`
	OnFail     string            `json:"on_fail"`
}

// composeSegments is the 7-segment breakdown of an emitted script.
type composeSegments struct {
	Prefix     string
	Function   string
	Parameters string
	Criteria   string
	TimeoutMs  string
	Flow       string
	OnFail     string
}

// validateAndEmit validates the request against the fields and, when
// there are no field errors, emits the script string. Validation messages
// are keyed "criteria.<name>" for criteria and by parameter name
// otherwise; reserved keys timeout_ms / flow / on_fail cover the fixed
// segments.
func validateAndEmit(fn store.Function, params, criteria []composeField, req composeRequest) (string, map[string]string) {
	fields := map[string]string{}
	segs := composeSegments{
		Prefix:   fn.Source + fn.Interface,
		Function: fn.Name,
	}

	// Fixed segments.
	timeout := strings.TrimSpace(req.TimeoutMs)
	if timeout == "" {
		fields["timeout_ms"] = "timeout-ms is required"
	} else if n, err := strconv.Atoi(timeout); err != nil || n <= 0 {
		fields["timeout_ms"] = "timeout-ms must be a positive integer (milliseconds)"
	}
	if strings.TrimSpace(req.Flow) == "" {
		fields["flow"] = "flow segment is required (e.g. NEXT or QUIT)"
	}
	if strings.TrimSpace(req.OnFail) == "" {
		fields["on_fail"] = "on-fail segment is required (e.g. FAIL)"
	}

	segs.Parameters = emitAssignments(params, req.Parameters, "", fields)
	segs.Criteria = emitAssignments(criteria, req.Criteria, "criteria.", fields)
	if segs.Criteria == "" {
		segs.Criteria = "NULL"
	}
	segs.TimeoutMs = timeout
	segs.Flow = strings.TrimSpace(req.Flow)
	segs.OnFail = strings.TrimSpace(req.OnFail)

	if len(fields) > 0 {
		return "", fields
	}
	return segs.Prefix + "|" + segs.Function + "|" + segs.Parameters + "|" +
		segs.Criteria + "|" + segs.TimeoutMs + "|" + segs.Flow + "|" + segs.OnFail, nil
}

// emitAssignments renders "name=value" segments in field order. Empty
// optional fields are omitted; empty required fields produce an error.
func emitAssignments(fs []composeField, values map[string]string, keyPrefix string, fields map[string]string) string {
	parts := make([]string, 0, len(fs))
	for _, f := range fs {
		raw := strings.TrimSpace(values[f.Name])
		if raw == "" {
			if f.Required {
				fields[keyPrefix+f.Name] = "is required"
			}
			continue
		}
		if msgs := validateFieldValue(f, raw); len(msgs) > 0 {
			fields[keyPrefix+f.Name] = strings.Join(msgs, "; ")
			continue
		}
		parts = append(parts, f.Name+"="+raw)
	}
	return strings.Join(parts, "&")
}

// validateFieldValue checks one raw value (already trimmed) against the
// field's type and allowed-values constraint, returning 0..n messages.
func validateFieldValue(f composeField, raw string) []string {
	var msgs []string

	multi := f.Cardinality == "MULTIPLE" || f.Type == "list[int]"
	vals := []string{raw}
	if multi {
		vals = splitValues(raw)
		if len(vals) == 0 {
			return []string{"must contain at least one value"}
		}
	}

	for _, v := range vals {
		if strings.ContainsAny(v, "|&=") {
			msgs = append(msgs, "must not contain '|', '&' or '='")
			continue
		}
		msgs = append(msgs, checkType(f, v)...)
		msgs = append(msgs, checkAllowed(f, v)...)
	}
	return msgs
}

// splitValues splits a comma-separated multi-value input, dropping empty
// pieces.
func splitValues(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// numericPart strips a trailing alphabetic unit ("15dBm" → "15") so
// values like the workbook's TX_TARGET_POWER=15dBm still pass int checks.
func numericPart(v string) (string, bool) {
	i := len(v)
	for i > 0 && isAlpha(v[i-1]) {
		i--
	}
	if i == 0 || i == len(v) {
		return v, false
	}
	return v[:i], true
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// checkType enforces the definition's type; unresolved definitions and
// plain strings impose no constraint.
func checkType(f composeField, v string) []string {
	num, hadUnit := numericPart(v)
	if hadUnit && num == "" {
		return nil // pure unit text: leave to allowed-values / pattern
	}
	switch f.Type {
	case "int":
		if _, err := strconv.Atoi(num); err != nil {
			return []string{"must be an integer" + unitHint(f)}
		}
	case "double":
		if _, err := strconv.ParseFloat(num, 64); err != nil {
			return []string{"must be a number" + unitHint(f)}
		}
	case "list[int]":
		for _, p := range strings.Split(num, ",") {
			if _, err := strconv.Atoi(strings.TrimSpace(p)); err != nil {
				return []string{"must be a comma-separated list of integers" + unitHint(f)}
			}
		}
	}
	return nil
}

func unitHint(f composeField) string {
	if f.Unit != "" {
		return " (unit: " + f.Unit + ")"
	}
	return ""
}

// checkAllowed enforces the two workbook constraint shapes: set
// "{20,40,80}" and numeric range "[0..7500]". Anything else is free-form
// (the workbook also stores loose constraints like "[A-Z0-9]" paired with
// pattern text, which the composer does not enforce).
func checkAllowed(f composeField, v string) []string {
	a := strings.TrimSpace(f.AllowedValues)
	if a == "" {
		return nil
	}
	if strings.HasPrefix(a, "{") && strings.HasSuffix(a, "}") {
		allowed := map[string]bool{}
		for _, p := range splitValues(a[1 : len(a)-1]) {
			allowed[p] = true
		}
		if allowed[v] {
			return nil
		}
		if n, hadUnit := numericPart(v); hadUnit && allowed[n] {
			return nil
		}
		return []string{"must be one of: " + a}
	}
	if strings.HasPrefix(a, "[") && strings.HasSuffix(a, "]") && strings.Contains(a, "..") {
		bounds := strings.SplitN(a[1:len(a)-1], "..", 2)
		lo, errLo := strconv.ParseFloat(strings.TrimSpace(bounds[0]), 64)
		hi, errHi := strconv.ParseFloat(strings.TrimSpace(bounds[1]), 64)
		if errLo != nil || errHi != nil {
			return nil
		}
		num, _ := numericPart(v)
		n, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return nil // non-numeric value: type check reports it
		}
		if n < lo || n > hi {
			return []string{"must be between " + strings.TrimSpace(bounds[0]) + " and " + strings.TrimSpace(bounds[1])}
		}
	}
	return nil
}

// ---- handlers -----------------------------------------------------------

// registerComposeRoutes wires the composer endpoints for one function.
func registerComposeRoutes(app *fiber.App, q *store.Queries) {
	app.Get("/api/functions/:id/compose/fields", getComposeFields(q))
	app.Post("/api/functions/:id/compose", composeScript(q))
}

// loadFunctionOr404 resolves the :id param; when it returns ok=false the
// error response (400/404/500) is already written and the caller MUST
// return immediately — Ctx.JSON returns a nil error, so a bool (not the
// write result) is the only reliable signal.
func loadFunctionOr404(c fiber.Ctx, q *store.Queries) (store.Function, bool) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid function id"})
		return store.Function{}, false
	}
	fn, err := q.GetFunction(c.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
			return store.Function{}, false
		}
		_ = serverError(c, "read function", err)
		return store.Function{}, false
	}
	return fn, true
}

// composeFieldsFor loads and dedupes the composer fields for one function.
func composeFieldsFor(c fiber.Ctx, q *store.Queries, fn store.Function) ([]composeField, []composeField, error) {
	pRows, err := q.ListComposerParameterFields(c.Context(), store.ListComposerParameterFieldsParams{
		FunctionSource: fn.Source,
		FunctionName:   fn.Name,
	})
	if err != nil {
		return nil, nil, err
	}
	cRows, err := q.ListComposerCriteriaFields(c.Context(), store.ListComposerCriteriaFieldsParams{
		FunctionSource: fn.Source,
		FunctionName:   fn.Name,
	})
	if err != nil {
		return nil, nil, err
	}
	return dedupeParameterRows(pRows), dedupeCriteriaRows(cRows), nil
}

func getComposeFields(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		fn, ok := loadFunctionOr404(c, q)
		if !ok {
			return nil
		}
		params, criteria, err := composeFieldsFor(c, q, fn)
		if err != nil {
			return serverError(c, "list composer fields", err)
		}
		return c.JSON(fiber.Map{
			"function": fiber.Map{
				"source":    fn.Source,
				"interface": fn.Interface,
				"name":      fn.Name,
			},
			"parameters": params,
			"criteria":   criteria,
		})
	}
}

func composeScript(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		fn, ok := loadFunctionOr404(c, q)
		if !ok {
			return nil
		}
		var req composeRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if req.Parameters == nil {
			req.Parameters = map[string]string{}
		}
		if req.Criteria == nil {
			req.Criteria = map[string]string{}
		}
		params, criteria, err := composeFieldsFor(c, q, fn)
		if err != nil {
			return serverError(c, "list composer fields", err)
		}
		script, fieldErrors := validateAndEmit(fn, params, criteria, req)
		if len(fieldErrors) > 0 {
			return validationError(c, fieldErrors)
		}
		return c.JSON(fiber.Map{"script": script})
	}
}
