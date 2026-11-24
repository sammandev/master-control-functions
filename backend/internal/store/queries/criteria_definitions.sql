-- name: GetCriteriaDefinition :one
SELECT * FROM criteria_definitions WHERE id = $1;

-- name: ListCriteriaDefinitions :many
SELECT * FROM criteria_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text)
ORDER BY id;

-- name: ListCriteriaDefinitionsPage :many
SELECT * FROM criteria_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text)
-- Match Go's bytewise order; secondary name and stable ID stay ascending.
ORDER BY
    (CASE WHEN NOT sqlc.arg('sort_desc')::boolean THEN CASE sqlc.arg('sort_col')::text WHEN 'name' THEN name WHEN 'type' THEN type WHEN 'updated_at' THEN updated_at END END) COLLATE "C",
    (CASE WHEN sqlc.arg('sort_desc')::boolean THEN CASE sqlc.arg('sort_col')::text WHEN 'name' THEN name WHEN 'type' THEN type WHEN 'updated_at' THEN updated_at END END) COLLATE "C" DESC,
    (CASE WHEN sqlc.arg('sort_col')::text IN ('type', 'updated_at') THEN name END) COLLATE "C",
    id
LIMIT sqlc.arg('page_size')::bigint OFFSET sqlc.arg('page_offset')::bigint;

-- name: LocateCriteriaDefinition :one
WITH target AS (
    SELECT t.id, t.name,
           (CASE sqlc.arg('sort_col')::text WHEN 'name' THEN t.name WHEN 'type' THEN t.type WHEN 'updated_at' THEN t.updated_at END) COLLATE "C" AS sort_key
    FROM criteria_definitions AS t
    WHERE t.id = sqlc.arg('definition_id')
    AND (sqlc.narg('search')::text IS NULL
         OR lower(t.name) LIKE lower(sqlc.narg('search')::text)
         OR lower(t.description) LIKE lower(sqlc.narg('search')::text))
    AND (sqlc.narg('device_name')::text IS NULL OR t.device_name = sqlc.narg('device_name')::text)
)
SELECT (1 + (
    SELECT count(*)
    FROM criteria_definitions AS d
    CROSS JOIN LATERAL (
        SELECT (CASE sqlc.arg('sort_col')::text WHEN 'name' THEN d.name WHEN 'type' THEN d.type WHEN 'updated_at' THEN d.updated_at END) COLLATE "C" AS sort_key
    ) AS candidate
    WHERE (sqlc.narg('search')::text IS NULL
         OR lower(d.name) LIKE lower(sqlc.narg('search')::text)
         OR lower(d.description) LIKE lower(sqlc.narg('search')::text))
    AND (sqlc.narg('device_name')::text IS NULL OR d.device_name = sqlc.narg('device_name')::text)
    AND CASE
        WHEN target.sort_key IS NULL THEN d.id < target.id
        WHEN candidate.sort_key = target.sort_key THEN (d.name COLLATE "C", d.id) < (target.name COLLATE "C", target.id)
        WHEN sqlc.arg('sort_desc')::boolean THEN candidate.sort_key > target.sort_key
        ELSE candidate.sort_key < target.sort_key
    END
))::bigint AS position
FROM target;

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
