-- name: GetParameterDefinition :one
SELECT * FROM parameter_definitions WHERE id = $1;

-- name: ListParameterDefinitions :many
SELECT * FROM parameter_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text)
ORDER BY id;

-- name: ListParameterDefinitionsPage :many
SELECT * FROM parameter_definitions
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

-- name: LocateParameterDefinition :one
SELECT position FROM (
    SELECT id, row_number() OVER (ORDER BY
        (CASE WHEN NOT sqlc.arg('sort_desc')::boolean THEN CASE sqlc.arg('sort_col')::text WHEN 'name' THEN name WHEN 'type' THEN type WHEN 'updated_at' THEN updated_at END END) COLLATE "C",
        (CASE WHEN sqlc.arg('sort_desc')::boolean THEN CASE sqlc.arg('sort_col')::text WHEN 'name' THEN name WHEN 'type' THEN type WHEN 'updated_at' THEN updated_at END END) COLLATE "C" DESC,
        (CASE WHEN sqlc.arg('sort_col')::text IN ('type', 'updated_at') THEN name END) COLLATE "C",
        id
    ) AS position
    FROM parameter_definitions
    WHERE (sqlc.narg('search')::text IS NULL
           OR lower(name) LIKE lower(sqlc.narg('search')::text)
           OR lower(description) LIKE lower(sqlc.narg('search')::text))
      AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text)
) AS ranked
WHERE id = sqlc.arg('definition_id');

-- name: CountParameterDefinitions :one
SELECT COUNT(*) FROM parameter_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text);

-- name: InsertParameterDefinition :one
INSERT INTO parameter_definitions (
    name, description, type, unit, default_value,
    allowed_values, cardinality, value_pattern, required,
    device_name, created_at, updated_at
) VALUES (
    @name, @description, @type, @unit, @default_value,
    @allowed_values, @cardinality, @value_pattern, @required,
    @device_name, @created_at, @updated_at
) RETURNING *;

-- name: UpdateParameterDefinition :one
UPDATE parameter_definitions SET
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

-- name: DeleteParameterDefinition :execrows
DELETE FROM parameter_definitions WHERE id = $1;

-- name: CountParameterPageUsage :many
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
FROM parameter_definitions d
LEFT JOIN parameter_links l ON l.name = d.name
LEFT JOIN parameter_definitions scoped ON scoped.name = l.name
    AND scoped.device_name = l.device_name AND scoped.device_name <> ''
WHERE d.id = ANY(string_to_array(sqlc.arg('ids')::text, ',')::bigint[])
GROUP BY d.id;

-- name: CountParameterUsageScoped :one
-- Scoped definition (ADR-0009): count links of the same device scope.
SELECT COUNT(*) FROM parameter_links WHERE name = $1 AND device_name = $2;

-- name: CountParameterUsageGlobal :one
-- Global definition (ADR-0009): count links whose device scope has no
-- scoped same-name sibling, the links that actually resolve to this
-- global definition under exact-scope-then-global.
SELECT COUNT(*) FROM parameter_links l
WHERE l.name = $1
  AND NOT EXISTS (
    SELECT 1 FROM parameter_definitions pd
    WHERE pd.name = l.name
      AND pd.device_name <> ''
      AND pd.device_name = l.device_name
  );
