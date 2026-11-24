package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// Device browser + registry (ADR-0010). A "device" is a distinct
// device_name value or Function interface (blank = Global). A Function
// belongs to a device through its interface or its parameter/criteria
// links. The definition tabs split the device's own rows
// from the global rows its links resolve to (exact-scope-then-global). The
// path segment "_global" addresses the blank device: real device values are
// upper-case device strings, so the lower-case segment cannot collide.

const globalDeviceSegment = "_global"

func registerDevices(app *fiber.App, q *store.Queries, db *sql.DB) {
	app.Get("/api/devices", listDevices(q))
	app.Post("/api/devices", createDevice(q))
	app.Get("/api/devices/:device/functions", deviceFunctions(q))
	app.Get("/api/devices/:device/parameters", deviceDefinitions(q.ListScopedParameterDefinitions, q.ListGlobalParameterDefinitionsForScope, parameterItem))
	app.Get("/api/devices/:device/criteria", deviceDefinitions(q.ListScopedCriteriaDefinitions, q.ListGlobalCriteriaDefinitionsForScope, criteriaItem))
	app.Put("/api/devices/:device", renameDevice(db))
	app.Delete("/api/devices/:device", deleteDevice(db))
}

func renameDevice(db *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		oldName := deviceFromSegment(c)
		if oldName == "" {
			return validationError(c, map[string]string{"device_name": "Global cannot be renamed"})
		}
		var body deviceBody
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		name := normalizeDeviceName(body.DeviceName)
		if name == "" {
			return validationError(c, map[string]string{"device_name": "is required"})
		}
		if problem := validateDeviceName(name); problem != "" {
			return validationError(c, map[string]string{"device_name": problem})
		}
		device, err := store.RenameDevice(c.Context(), db, oldName, name, time.Now().UTC().Format(time.RFC3339))
		if errors.Is(err, store.ErrDeviceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, store.ErrDeviceExists) || store.IsUniqueViolation(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "a device named '" + name + "' already exists"})
		}
		if err != nil {
			return serverError(c, "rename device", err)
		}
		return c.JSON(registeredDeviceResponse(device))
	}
}

// deviceFromSegment resolves the path segment to a device_name value;
// "_global" maps to the blank device. Fiber delivers route params raw, so
// percent-encoded names ("TEST%20PROGRAM") are decoded here; a segment
// carrying a literal '%' that is not a valid escape passes through as-is.
func deviceFromSegment(c fiber.Ctx) string {
	segment := c.Params("device")
	if segment == globalDeviceSegment {
		return ""
	}
	if decoded, err := url.PathUnescape(segment); err == nil {
		return decoded
	}
	return segment
}

// deviceBody keeps source optional for older clients and imported registrations.
type deviceBody struct {
	DeviceName string  `json:"device_name"`
	Source     *string `json:"source"`
}

func (body deviceBody) deviceSource() (sql.NullString, string) {
	if body.Source == nil {
		return sql.NullString{}, ""
	}
	source := strings.TrimSpace(*body.Source)
	if !validSources[source] {
		return sql.NullString{}, "must be BUILT-IN or DLL"
	}
	return sql.NullString{String: source, Valid: true}, ""
}

func registeredDeviceResponse(device store.Device) fiber.Map {
	var source *string
	if device.Source.Valid {
		source = &device.Source.String
	}
	return fiber.Map{"device_name": device.Name, "source": source, "created_at": device.CreatedAt, "updated_at": device.UpdatedAt}
}

// maxDeviceNameRunes bounds a registered Device's name so list rows,
// URLs, and error messages stay sane (ticket 06).
const maxDeviceNameRunes = 256

// normalizeDeviceName applies ADR-0002's canonical device-value rule —
// trimmed and uppercased — everywhere a device name enters the system
// (registry create/rename, Function interfaces, dictionary, links, examples).
func normalizeDeviceName(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

// validateDeviceName checks the already-normalized name against the
// registry's input rules and returns a field error, or "" when valid.
// The reserved names collide with the Global device's reserved "_global"
// segment / "(Global)" label in user-visible ways, so they are rejected.
func validateDeviceName(name string) string {
	if name == "_GLOBAL" || name == "GLOBAL" {
		return "is reserved for the Global device"
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return "must not contain control characters"
		}
	}
	if utf8.RuneCountInString(name) > maxDeviceNameRunes {
		return fmt.Sprintf("is too long (max %d characters)", maxDeviceNameRunes)
	}
	return ""
}

// createDevice reserves a name in the devices registry (ADR-0010) before
// any content references it. The stored name follows ADR-0002's
// normalization rule: trimmed and uppercased. A name already present as a
// derived device (any Function, link, definition, or example row) or in the registry
// is rejected with a conflict.
func createDevice(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		var b deviceBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		name := normalizeDeviceName(b.DeviceName)
		if name == "" {
			return validationError(c, map[string]string{"device_name": "is required"})
		}
		if problem := validateDeviceName(name); problem != "" {
			return validationError(c, map[string]string{"device_name": problem})
		}
		source, problem := b.deviceSource()
		if problem != "" {
			return validationError(c, map[string]string{"source": problem})
		}

		// Derived devices are reserved by content: a name carried by any
		// Function, link, definition, or example row already exists.
		uses, err := q.CountDeviceNameUsages(c.Context(), name)
		if err != nil {
			return serverError(c, "count device usages", err)
		}
		if uses > 0 {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "a device named '" + name + "' already exists",
			})
		}

		now := time.Now().UTC().Format(time.RFC3339)
		dev, err := q.InsertDevice(c.Context(), store.InsertDeviceParams{
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
			Source:    source,
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "a device named '" + name + "' already exists",
				})
			}
			return serverError(c, "create device", err)
		}
		return c.Status(fiber.StatusCreated).JSON(registeredDeviceResponse(dev))
	}
}

func deleteDevice(db *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		name := deviceFromSegment(c)
		if name == "" {
			return validationError(c, map[string]string{"device_name": "Global cannot be deleted"})
		}
		var body struct {
			Confirmation string `json:"confirmation"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if body.Confirmation != "DELETE "+name {
			return validationError(c, map[string]string{"confirmation": "must exactly match DELETE " + name})
		}
		err := store.DeleteDevice(c.Context(), db, name)
		if errors.Is(err, store.ErrDeviceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "device not found"})
		}
		var inUse *store.DeviceInUseError
		if errors.As(err, &inUse) {
			refusal := fiber.Map{"error": inUse.Error()}
			if len(inUse.Sequences) > 0 {
				refusal["sequences"] = inUse.Sequences
			}
			if len(inUse.Guides) > 0 {
				refusal["guides"] = inUse.Guides
			}
			return c.Status(fiber.StatusConflict).JSON(refusal)
		}
		if err != nil {
			return serverError(c, "delete device", err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// definitionItem is the lean read-only shape for the device detail tabs —
// the dictionary pages own the full CRUD shape (with usage counts).
type definitionItem struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	Unit          string `json:"unit"`
	DefaultValue  string `json:"default_value"`
	AllowedValues string `json:"allowed_values"`
	Cardinality   string `json:"cardinality"`
	ValuePattern  string `json:"value_pattern"`
	Required      bool   `json:"required"`
	DeviceName    string `json:"device_name"`
}

func parameterItem(d store.ParameterDefinition) definitionItem {
	return definitionItem{
		ID:            d.ID,
		Name:          d.Name,
		Description:   d.Description,
		Type:          d.Type,
		Unit:          d.Unit,
		DefaultValue:  nullStringValue(d.DefaultValue),
		AllowedValues: d.AllowedValues,
		Cardinality:   d.Cardinality,
		ValuePattern:  d.ValuePattern,
		Required:      d.Required == 1,
		DeviceName:    d.DeviceName,
	}
}

func criteriaItem(d store.CriteriaDefinition) definitionItem {
	return definitionItem{
		ID:            d.ID,
		Name:          d.Name,
		Description:   d.Description,
		Type:          d.Type,
		Unit:          d.Unit,
		DefaultValue:  nullStringValue(d.DefaultValue),
		AllowedValues: d.AllowedValues,
		Cardinality:   d.Cardinality,
		ValuePattern:  d.ValuePattern,
		Required:      d.Required == 1,
		DeviceName:    d.DeviceName,
	}
}

func nullStringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func listDevices(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		rows, err := q.ListDeviceScopes(c.Context())
		if err != nil {
			return serverError(c, "list devices", err)
		}
		sourceRows, err := q.ListDeviceSources(c.Context())
		if err != nil {
			return serverError(c, "list device types", err)
		}
		sourcesByName := map[string][]string{}
		for _, row := range sourceRows {
			if row.Source.Valid {
				sourcesByName[row.Name] = append(sourcesByName[row.Name], row.Source.String)
			}
		}
		items := make([]fiber.Map, 0, len(rows))
		for _, r := range rows {
			sources := sourcesByName[r.DeviceName]
			if sources == nil {
				sources = []string{}
			}
			var source *string
			if len(sources) == 1 {
				source = &sources[0]
			}
			items = append(items, fiber.Map{
				"device_name": r.DeviceName,
				"source":      source,
				"sources":     sources,
				"functions":   r.FunctionCount,
				"parameters":  r.ParameterCount,
				"criteria":    r.CriteriaCount,
				"examples":    r.ExampleCount,
			})
		}
		return c.JSON(fiber.Map{"items": items, "total": len(items)})
	}
}

func deviceFunctions(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		device := deviceFromSegment(c)
		total, err := q.CountFunctionsByDeviceScope(c.Context(), device)
		if err != nil {
			return serverError(c, "count device functions", err)
		}
		rows, err := q.ListFunctionsByDeviceScope(c.Context(), device)
		if err != nil {
			return serverError(c, "list device functions", err)
		}
		lq := parseListQuery(c)
		start, end := lq.sliceBounds(len(rows))
		items := make([]fiber.Map, 0, end-start)
		for _, f := range rows[start:end] {
			items = append(items, functionResponse(f))
		}
		return c.JSON(fiber.Map{
			"items":     items,
			"total":     total,
			"page":      lq.page,
			"page_size": lq.pageSize,
		})
	}
}

// deviceDefinitions serves both definition tabs: the device's own rows
// (paginated) plus the global rows its links resolve to (returned in
// full — bounded by the distinct definition names in the device).
func deviceDefinitions[T any](
	listScoped, listGlobal func(context.Context, string) ([]T, error),
	toItem func(T) definitionItem,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		device := deviceFromSegment(c)
		rows, err := listScoped(c.Context(), device)
		if err != nil {
			return serverError(c, "list scoped device definitions", err)
		}
		globalRows, err := listGlobal(c.Context(), device)
		if err != nil {
			return serverError(c, "list global device definitions", err)
		}

		lq := parseListQuery(c)
		term := strings.ToLower(strings.TrimSpace(lq.search))
		scoped := matchingDeviceDefinitions(rows, toItem, term)
		global := matchingDeviceDefinitions(globalRows, toItem, term)
		items := scoped[:0]
		if lq.page-1 <= int64(len(scoped))/lq.pageSize {
			start := (lq.page - 1) * lq.pageSize
			items = scoped[start:min(start+lq.pageSize, int64(len(scoped)))]
		}
		return c.JSON(fiber.Map{
			"items":        items,
			"total":        len(scoped),
			"page":         lq.page,
			"page_size":    lq.pageSize,
			"global_items": global,
			"global_total": len(global),
		})
	}
}

func matchingDeviceDefinitions[T any](rows []T, toItem func(T) definitionItem, term string) []definitionItem {
	items := make([]definitionItem, 0, len(rows))
	for _, row := range rows {
		item := toItem(row)
		if term == "" || strings.Contains(strings.ToLower(item.Name+" "+item.Type+" "+item.AllowedValues+" "+item.DefaultValue+" "+item.Description), term) {
			items = append(items, item)
		}
	}
	return items
}
