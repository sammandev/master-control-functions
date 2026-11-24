-- Sequence builder queries (ticket 10, ADR-0005). A Sequence is a named
-- aggregate; steps reference Functions by natural key with no FK to
-- functions (orphan steps are storable and flagged, ADR-0003). parameter
-- and criteria values are stored as JSON object TEXT.

-- name: InsertSequence :one
INSERT INTO sequences (name, description, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, created_at, updated_at;

-- name: UpdateSequenceMeta :execrows
UPDATE sequences SET name = $1, description = $2, updated_at = $3 WHERE id = $4;

-- name: TouchSequence :execrows
UPDATE sequences SET updated_at = $1 WHERE id = $2;

-- name: DeleteSequence :execrows
DELETE FROM sequences WHERE id = $1;

-- name: GetSequence :one
SELECT id, name, description, created_at, updated_at FROM sequences WHERE id = $1;

-- name: CountSequences :one
SELECT COUNT(*)
FROM sequences
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(name) LIKE lower(sqlc.narg('search')::text)
       OR lower(description) LIKE lower(sqlc.narg('search')::text));

-- name: ListSequences :many
SELECT s.id, s.name, s.description, s.created_at, s.updated_at,
       (SELECT COUNT(*) FROM sequence_steps st WHERE st.sequence_id = s.id) AS step_count
FROM (
    SELECT id, name, description, created_at, updated_at
    FROM sequences
    WHERE (sqlc.narg('search')::text IS NULL
           OR lower(name) LIKE lower(sqlc.narg('search')::text)
           OR lower(description) LIKE lower(sqlc.narg('search')::text))
    ORDER BY name, id
    LIMIT sqlc.narg('page_limit')::bigint OFFSET sqlc.arg('page_offset')::bigint
) s
ORDER BY s.name, s.id;

-- name: InsertSequenceStep :one
INSERT INTO sequence_steps (
    sequence_id, position, function_source, function_interface, function_name,
    parameters, criteria, timeout_ms, flow, on_fail, device_name, note
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, sequence_id, position, function_source, function_interface,
          function_name, parameters, criteria, timeout_ms, flow, on_fail,
          device_name, note;

-- name: DeleteSequenceSteps :exec
DELETE FROM sequence_steps WHERE sequence_id = $1;

-- name: ListSequenceSteps :many
SELECT st.id, st.sequence_id, st.position, st.function_source, st.function_interface,
       st.function_name, st.parameters, st.criteria, st.timeout_ms, st.flow,
       st.on_fail, st.device_name, st.note,
       EXISTS (SELECT 1 FROM functions f
               WHERE f.source = st.function_source
                 AND f.interface = st.function_interface
                 AND f.name = st.function_name) AS function_exists
FROM sequence_steps st
WHERE st.sequence_id = $1
ORDER BY st.position, st.id;

-- name: GetFunctionByNaturalKey :one
SELECT id, source, interface, name, description, convention, category,
       execution_role, execution_phase, created_at, updated_at
FROM functions
WHERE source = $1 AND interface = $2 AND name = $3;
