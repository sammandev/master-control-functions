-- name: GetParameterDefinition :one
SELECT * FROM parameter_definitions WHERE id = $1;

-- name: ListParameterDefinitions :many
SELECT * FROM parameter_definitions
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('device_name')::text IS NULL OR device_name = sqlc.narg('device_name')::text)
ORDER BY id;

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
