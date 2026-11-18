package api

import (
	"database/sql"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// Orphan report (ticket 05, ADR-0003): every link or Example row whose
// target Function / Parameter definition / Criteria definition does not
// exist, grouped by kind. Nothing is dropped or quarantined; the report is
// read-only — resolution happens through the regular create endpoints with
// the row's values prefilled, which clears the flags in place.

// orphanLinkItem is the JSON shape of one orphaned link row. The link's
// name IS the definition's name; function_exists/definition_exists say
// which side is missing. function_id identifies the (possibly
// multi-interface) Function for jump-to-row and is null when it is missing.
type orphanLinkItem struct {
	ID               int64  `json:"id"`
	FunctionSource   string `json:"function_source"`
	FunctionName     string `json:"function_name"`
	FunctionExists   bool   `json:"function_exists"`
	FunctionID       *int64 `json:"function_id"`
	Name             string `json:"name"`
	DefinitionExists bool   `json:"definition_exists"`
	DeviceName       string `json:"device_name"`
	Position         *int64 `json:"position"`
	Required         bool   `json:"required"`
}

// nullablePosition converts the generated nullable column into a JSON
// number-or-null.
func nullablePosition(p sql.NullInt64) *int64 {
	if !p.Valid {
		return nil
	}
	v := p.Int64
	return &v
}

// nullableAny converts a scalar-subquery column the generator could not
// type (interface{} in the Row struct) into a nullable int64.
func nullableAny(v any) *int64 {
	if n, ok := v.(int64); ok {
		return &n
	}
	return nil
}

// orphanExampleItem is the JSON shape of one orphaned Example row.
type orphanExampleItem struct {
	ID             int64  `json:"id"`
	FunctionSource string `json:"function_source"`
	FunctionName   string `json:"function_name"`
	DeviceName     string `json:"device_name"`
	Script         string `json:"script"`
	Meaning        string `json:"meaning"`
}

type orphanGroup struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
	Items any    `json:"items"`
}

// registerOrphanRoutes wires GET /api/orphans.
func registerOrphanRoutes(app *fiber.App, q *store.Queries) {
	app.Get("/api/orphans", orphanReport(q))
}

func orphanReport(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		parameterItems, err := orphanParameterItems(c, q)
		if err != nil {
			return serverError(c, "list orphan parameter links", err)
		}
		criteriaItems, err := orphanCriteriaItems(c, q)
		if err != nil {
			return serverError(c, "list orphan criteria links", err)
		}
		exampleItems, err := orphanExampleItems(c, q)
		if err != nil {
			return serverError(c, "list orphan examples", err)
		}

		groups := []orphanGroup{
			{Kind: "parameter-links", Count: len(parameterItems), Items: parameterItems},
			{Kind: "criteria-links", Count: len(criteriaItems), Items: criteriaItems},
			{Kind: "examples", Count: len(exampleItems), Items: exampleItems},
		}
		total := 0
		for _, g := range groups {
			total += g.Count
		}
		return c.JSON(fiber.Map{"total": total, "groups": groups})
	}
}

func orphanParameterItems(c fiber.Ctx, q *store.Queries) ([]orphanLinkItem, error) {
	rows, err := q.ListOrphanParameterLinks(c.Context())
	if err != nil {
		return nil, err
	}
	items := make([]orphanLinkItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, orphanLinkItem{
			ID:               r.ID,
			FunctionSource:   r.FunctionSource,
			FunctionName:     r.FunctionName,
			FunctionExists:   r.FunctionExists,
			FunctionID:       nullableAny(r.FunctionID),
			Name:             r.Name,
			DefinitionExists: r.DefinitionExists,
			DeviceName:       r.DeviceName,
			Position:         nullablePosition(r.Position),
			Required:         r.Required != 0,
		})
	}
	return items, nil
}

func orphanCriteriaItems(c fiber.Ctx, q *store.Queries) ([]orphanLinkItem, error) {
	rows, err := q.ListOrphanCriteriaLinks(c.Context())
	if err != nil {
		return nil, err
	}
	items := make([]orphanLinkItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, orphanLinkItem{
			ID:               r.ID,
			FunctionSource:   r.FunctionSource,
			FunctionName:     r.FunctionName,
			FunctionExists:   r.FunctionExists,
			FunctionID:       nullableAny(r.FunctionID),
			Name:             r.Name,
			DefinitionExists: r.DefinitionExists,
			DeviceName:       r.DeviceName,
			Position:         nullablePosition(r.Position),
			Required:         r.Required != 0,
		})
	}
	return items, nil
}

func orphanExampleItems(c fiber.Ctx, q *store.Queries) ([]orphanExampleItem, error) {
	rows, err := q.ListOrphanExamples(c.Context())
	if err != nil {
		return nil, err
	}
	items := make([]orphanExampleItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, orphanExampleItem{
			ID:             r.ID,
			FunctionSource: r.FunctionSource,
			FunctionName:   r.FunctionName,
			DeviceName:     r.DeviceName,
			Script:         r.Script,
			Meaning:        r.Meaning,
		})
	}
	return items, nil
}
