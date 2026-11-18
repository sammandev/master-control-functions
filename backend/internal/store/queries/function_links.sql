-- Function editor link queries (ticket 04). The two link tables share the
-- same shape (CONTEXT.md), so every query exists once per dictionary.
-- List/get queries LEFT JOIN the matching dictionary entry using the
-- exact-scope-then-global resolution (ADR-0009): the definition row
-- whose (name, device_name) matches the link's, or the blank-scope one
-- when no exact match exists. A NULL id means the link is an Orphan
-- (ADR-0003) and must render with an unresolved warning, not be hidden
-- or dropped.

-- name: ResolveParameterDefinitionByLink :one
-- Returns the definition row a link resolves to under
-- exact-scope-then-global (ADR-0009), or ErrNoRows if neither exists.
-- (name, device_name) is UNIQUE so the WHERE matches at most two rows;
-- device_name DESC prefers the non-empty (exact-scope) one over the
-- blank-scope global fallback.
SELECT * FROM parameter_definitions
WHERE name = $1 AND (device_name = $2 OR device_name = '')
ORDER BY device_name DESC
LIMIT 1;

-- name: ListParameterLinksByFunction :many
SELECT l.*, d.id AS definition_id
FROM parameter_links l
LEFT JOIN parameter_definitions d
       ON d.name = l.name
      AND d.id = COALESCE(
              (SELECT d2.id FROM parameter_definitions d2
               WHERE d2.name = l.name AND d2.device_name = l.device_name),
              (SELECT d3.id FROM parameter_definitions d3
               WHERE d3.name = l.name AND d3.device_name = ''))
WHERE l.function_source = $1 AND l.function_name = $2;

-- name: GetParameterLink :one
SELECT l.*, d.id AS definition_id
FROM parameter_links l
LEFT JOIN parameter_definitions d
       ON d.name = l.name
      AND d.id = COALESCE(
              (SELECT d2.id FROM parameter_definitions d2
               WHERE d2.name = l.name AND d2.device_name = l.device_name),
              (SELECT d3.id FROM parameter_definitions d3
               WHERE d3.name = l.name AND d3.device_name = ''))
WHERE l.id = $1;

-- name: InsertParameterLink :one
INSERT INTO parameter_links (
    function_source, function_name, device_name, name, position, required,
    created_at, updated_at
) VALUES (
    @function_source, @function_name, @device_name, @name, @position, @required,
    @created_at, @updated_at
) RETURNING *;

-- name: UpdateParameterLink :one
UPDATE parameter_links SET
    device_name = @device_name,
    position = @position,
    required = @required,
    updated_at = @updated_at
WHERE id = @id
RETURNING *;

-- name: DeleteParameterLink :execrows
DELETE FROM parameter_links WHERE id = $1;

-- name: ResolveCriteriaDefinitionByLink :one
SELECT * FROM criteria_definitions
WHERE name = $1 AND (device_name = $2 OR device_name = '')
ORDER BY device_name DESC
LIMIT 1;

-- name: ListCriteriaLinksByFunction :many
SELECT l.*, d.id AS definition_id
FROM criteria_links l
LEFT JOIN criteria_definitions d
       ON d.name = l.name
      AND d.id = COALESCE(
              (SELECT d2.id FROM criteria_definitions d2
               WHERE d2.name = l.name AND d2.device_name = l.device_name),
              (SELECT d3.id FROM criteria_definitions d3
               WHERE d3.name = l.name AND d3.device_name = ''))
WHERE l.function_source = $1 AND l.function_name = $2;

-- name: GetCriteriaLink :one
SELECT l.*, d.id AS definition_id
FROM criteria_links l
LEFT JOIN criteria_definitions d
       ON d.name = l.name
      AND d.id = COALESCE(
              (SELECT d2.id FROM criteria_definitions d2
               WHERE d2.name = l.name AND d2.device_name = l.device_name),
              (SELECT d3.id FROM criteria_definitions d3
               WHERE d3.name = l.name AND d3.device_name = ''))
WHERE l.id = $1;

-- name: InsertCriteriaLink :one
INSERT INTO criteria_links (
    function_source, function_name, device_name, name, position, required,
    created_at, updated_at
) VALUES (
    @function_source, @function_name, @device_name, @name, @position, @required,
    @created_at, @updated_at
) RETURNING *;

-- name: UpdateCriteriaLink :one
UPDATE criteria_links SET
    device_name = @device_name,
    position = @position,
    required = @required,
    updated_at = @updated_at
WHERE id = @id
RETURNING *;

-- name: DeleteCriteriaLink :execrows
DELETE FROM criteria_links WHERE id = $1;
