-- name: GetCriteriaDefinition :one
SELECT * FROM criteria_definitions WHERE id = $1;

-- name: ListCriteriaDefinitions :many
SELECT * FROM criteria_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
ORDER BY id;

-- name: CountCriteriaDefinitions :one
SELECT COUNT(*) FROM criteria_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text));

-- name: InsertCriteriaDefinition :one
INSERT INTO criteria_definitions (
    name, description, type, unit, default_value,
    allowed_values, cardinality, value_pattern, required,
    device_name, created_at, updated_at
) VALUES (
    @name, @description, @type, @unit, @default_value,
    @allowed_values, @cardinality, @value_pattern, @required,
    @device_name, @created_at, @updated_at
) RETURNING *;

-- name: UpdateCriteriaDefinition :one
UPDATE criteria_definitions SET
    name = @name,
    description = @description,
    type = @type,
    unit = @unit,
    default_value = @default_value,
    allowed_values = @allowed_values,
    cardinality = @cardinality,
    value_pattern = @value_pattern,
    required = @required,
    device_name = @device_name,
    updated_at = @updated_at
WHERE id = @id
RETURNING *;

-- name: DeleteCriteriaDefinition :execrows
DELETE FROM criteria_definitions WHERE id = $1;

-- name: CountCriteriaUsageScoped :one
-- Scoped definition (ADR-0009): count links of the same device scope.
SELECT COUNT(*) FROM criteria_links WHERE name = $1 AND device_name = $2;

-- name: CountCriteriaUsageGlobal :one
-- Global definition (ADR-0009): count links whose device scope has no
-- scoped same-name sibling.
SELECT COUNT(*) FROM criteria_links l
WHERE l.name = $1
  AND NOT EXISTS (
    SELECT 1 FROM criteria_definitions pd
    WHERE pd.name = l.name
      AND pd.device_name <> ''
      AND pd.device_name = l.device_name
  );
