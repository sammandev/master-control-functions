-- name: GetCriteriaDefinition :one
SELECT * FROM criteria_definitions WHERE id = $1;

-- name: ListCriteriaDefinitions :many
SELECT * FROM criteria_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text)
ORDER BY id;

-- name: CountCriteriaDefinitions :one
SELECT COUNT(*) FROM criteria_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text);

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

-- name: CountCriteriaPageUsage :many
-- Usage follows exact-scope-then-global resolution; orphan usage is
-- name-wide, regardless of device scope.
SELECT d.id,
       COUNT(l.id) FILTER (
           WHERE l.device_name = d.device_name
              OR (d.device_name = '' AND scoped.id IS NULL)
       ) AS usage_count,
       COUNT(l.id) FILTER (
           WHERE NOT EXISTS (
               SELECT 1 FROM functions f
               WHERE f.source = l.function_source AND f.name = l.function_name
           )
       ) AS orphan_usage_count
FROM criteria_definitions d
LEFT JOIN criteria_links l ON l.name = d.name
LEFT JOIN criteria_definitions scoped ON scoped.name = l.name
    AND scoped.device_name = l.device_name AND scoped.device_name <> ''
WHERE d.id = ANY(string_to_array(sqlc.arg('ids')::text, ',')::bigint[])
GROUP BY d.id;

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
