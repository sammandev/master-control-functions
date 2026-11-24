-- name: GetFunction :one
SELECT * FROM functions WHERE id = $1;

-- name: ListFunctions :many
SELECT * FROM functions
WHERE (sqlc.narg('search')::text IS NULL
       OR strpos(lower(name || ' ' || description), lower(sqlc.narg('search')::text)) > 0)
  AND (sqlc.narg('source')::text IS NULL OR source = sqlc.narg('source')::text)
  AND (sqlc.narg('interface')::text IS NULL OR interface = sqlc.narg('interface')::text)
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category')::text)
  AND (sqlc.narg('execution_role')::text IS NULL OR execution_role = sqlc.narg('execution_role')::text)
  AND (sqlc.narg('execution_phase')::text IS NULL OR execution_phase = sqlc.narg('execution_phase')::text)
ORDER BY id;

-- name: ListFunctionsByNamePage :many
SELECT * FROM functions
WHERE (sqlc.narg('search')::text IS NULL
       OR strpos(lower(name || ' ' || description), lower(sqlc.narg('search')::text)) > 0)
  AND (sqlc.narg('source')::text IS NULL OR source = sqlc.narg('source')::text)
  AND (sqlc.narg('interface')::text IS NULL OR interface = sqlc.narg('interface')::text)
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category')::text)
  AND (sqlc.narg('execution_role')::text IS NULL OR execution_role = sqlc.narg('execution_role')::text)
  AND (sqlc.narg('execution_phase')::text IS NULL OR execution_phase = sqlc.narg('execution_phase')::text)
-- Match the Go comparator bytewise, keeping the stable ID order even in descending sort.
ORDER BY
    (CASE WHEN NOT sqlc.arg('sort_desc')::boolean THEN name END) COLLATE "C",
    (CASE WHEN NOT sqlc.arg('sort_desc')::boolean THEN source || interface || name END) COLLATE "C",
    (CASE WHEN sqlc.arg('sort_desc')::boolean THEN name END) COLLATE "C" DESC,
    (CASE WHEN sqlc.arg('sort_desc')::boolean THEN source || interface || name END) COLLATE "C" DESC,
    id
LIMIT sqlc.arg('page_size')::bigint OFFSET sqlc.arg('page_offset')::bigint;

-- name: ListFunctionOptions :many
-- Match the catalog's ascending name sort, including its bytewise tiebreak.
SELECT id, name, source, interface FROM functions
ORDER BY name COLLATE "C", (source || interface || name) COLLATE "C", id;

-- name: CountFunctions :one
SELECT COUNT(*) FROM functions
WHERE (sqlc.narg('search')::text IS NULL
       OR strpos(lower(name || ' ' || description), lower(sqlc.narg('search')::text)) > 0)
  AND (sqlc.narg('source')::text IS NULL OR source = sqlc.narg('source')::text)
  AND (sqlc.narg('interface')::text IS NULL OR interface = sqlc.narg('interface')::text)
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category')::text)
  AND (sqlc.narg('execution_role')::text IS NULL OR execution_role = sqlc.narg('execution_role')::text)
  AND (sqlc.narg('execution_phase')::text IS NULL OR execution_phase = sqlc.narg('execution_phase')::text);

-- name: InsertFunction :one
INSERT INTO functions (
    source, interface, name, description, convention,
    category, execution_role, execution_phase, created_at, updated_at
) VALUES (
    @source, @interface, @name, @description, @convention,
    @category, @execution_role, @execution_phase, @created_at, @updated_at
) RETURNING *;

-- name: UpdateFunction :one
UPDATE functions SET
    description = @description,
    convention = @convention,
    category = @category,
    execution_role = @execution_role,
    execution_phase = @execution_phase,
    updated_at = @updated_at
WHERE id = @id
RETURNING *;

-- name: UpdateFunctionKey :one
UPDATE functions SET
    source = @source,
    interface = @interface,
    name = @name,
    updated_at = @updated_at
WHERE id = @id
RETURNING *;

-- name: DeleteFunction :execrows
DELETE FROM functions WHERE id = $1;
