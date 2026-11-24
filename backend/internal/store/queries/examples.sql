-- Examples CRUD queries (ticket 06). Examples are keyed naturally by
-- (function_source, function_name, device_name, script); the Function
-- reference is validated at write time in the API, so orphans only arise
-- from imports or function deletions and are flagged, never dropped
-- (ADR-0003).

-- name: ListExamples :many
WITH matching_examples AS MATERIALIZED (
    SELECT e.id, e.device_name,
           (CASE WHEN NOT sqlc.arg('descending')::boolean THEN sort_value END)
               COLLATE example_sort_natural AS asc_key,
           (CASE WHEN sqlc.arg('descending')::boolean THEN sort_value END)
               COLLATE example_sort_natural AS desc_key
    FROM examples e
    CROSS JOIN LATERAL (
        SELECT CASE sqlc.arg('sort_column')::text
            WHEN 'function' THEN e.function_source || '/' || e.function_name
            WHEN 'script' THEN e.script
            WHEN 'meaning' THEN e.meaning
        END AS sort_value
    ) sort
    WHERE (sqlc.narg('function_source')::text IS NULL OR e.function_source = sqlc.narg('function_source')::text)
      AND (sqlc.narg('function_name')::text IS NULL OR e.function_name = sqlc.narg('function_name')::text)
      AND (sqlc.narg('device')::text IS NULL OR e.device_name = sqlc.narg('device')::text)
      AND (sqlc.narg('search')::text IS NULL
         OR lower(e.function_source) LIKE lower(sqlc.narg('search')::text)
         OR lower(e.function_name) LIKE lower(sqlc.narg('search')::text)
         OR lower(e.function_source || ' ' || e.function_name) LIKE lower(sqlc.narg('search')::text)
         OR lower(COALESCE(NULLIF(e.device_name, ''), 'Global')) LIKE lower(sqlc.narg('search')::text)
         OR lower(e.script) LIKE lower(sqlc.narg('search')::text)
         OR lower(e.meaning) LIKE lower(sqlc.narg('search')::text))
), page_start AS (
    SELECT COALESCE(
        -- Round the matched example's zero-based position down to its page start.
        (SELECT (SELECT COUNT(*) FROM matching_examples preceding
                 WHERE preceding.asc_key < target.asc_key
                    OR preceding.desc_key > target.desc_key
                    OR (preceding.asc_key IS NOT DISTINCT FROM target.asc_key
                        AND preceding.desc_key IS NOT DISTINCT FROM target.desc_key
                        AND (preceding.device_name, preceding.id) < (target.device_name, target.id)))
                / sqlc.arg('page_size')::bigint * sqlc.arg('page_size')::bigint
         FROM matching_examples target
         WHERE target.id = sqlc.narg('example_id')::bigint),
        sqlc.arg('requested_offset')::bigint
    )::bigint AS page_offset
), page AS (
    SELECT id, device_name, asc_key, desc_key
    FROM matching_examples
    ORDER BY asc_key, desc_key DESC, device_name, id
    LIMIT sqlc.arg('page_size')::bigint
    OFFSET (SELECT page_offset FROM page_start)
)
SELECT e.*,
       EXISTS (SELECT 1 FROM functions f
               WHERE f.source = e.function_source AND f.name = e.function_name) AS function_exists,
       (SELECT page_offset FROM page_start) AS page_offset
FROM page
JOIN examples e ON e.id = page.id
ORDER BY page.asc_key, page.desc_key DESC, page.device_name, page.id;

-- name: ListExampleFunctionNames :many
SELECT function_name
FROM (
    SELECT DISTINCT ON (function_name) function_name, device_name, id
    FROM examples
    ORDER BY function_name, device_name, id
) first_examples
ORDER BY device_name, id;

-- name: CountExamples :one
SELECT COUNT(*)
FROM examples e
WHERE (sqlc.narg('function_source')::text IS NULL OR e.function_source = sqlc.narg('function_source')::text)
  AND (sqlc.narg('function_name')::text IS NULL OR e.function_name = sqlc.narg('function_name')::text)
  AND (sqlc.narg('device')::text IS NULL OR e.device_name = sqlc.narg('device')::text)
  AND (sqlc.narg('search')::text IS NULL
     OR lower(e.function_source) LIKE lower(sqlc.narg('search')::text)
     OR lower(e.function_name) LIKE lower(sqlc.narg('search')::text)
     OR lower(e.function_source || ' ' || e.function_name) LIKE lower(sqlc.narg('search')::text)
     OR lower(COALESCE(NULLIF(e.device_name, ''), 'Global')) LIKE lower(sqlc.narg('search')::text)
       OR lower(e.script) LIKE lower(sqlc.narg('search')::text)
       OR lower(e.meaning) LIKE lower(sqlc.narg('search')::text));

-- name: GetExample :one
SELECT e.*,
       EXISTS (SELECT 1 FROM functions f
               WHERE f.source = e.function_source AND f.name = e.function_name) AS function_exists
FROM examples e
WHERE e.id = $1;

-- name: InsertExample :one
INSERT INTO examples (
    function_source, function_name, device_name, script, meaning,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateExample :one
UPDATE examples SET
    device_name = $1, script = $2, meaning = $3, updated_at = $4
WHERE id = $5
RETURNING *;

-- name: DeleteExample :execrows
DELETE FROM examples WHERE id = $1;

-- Write-time reference validation for creates (ticket 06): EXISTS instead
-- of a join, because functions are unique on (source, interface, name) and
-- (source, name) alone can match several rows.
-- name: FunctionNaturalKeyExists :one
SELECT EXISTS (SELECT 1 FROM functions f
               WHERE f.source = $1 AND f.name = $2);
