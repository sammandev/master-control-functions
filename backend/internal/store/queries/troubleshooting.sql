-- Troubleshooting Playbook & AI Knowledge Base queries.
-- Multiple guides can share the same error_code or function_name to represent different
-- manifestations/solutions for the same problem.

-- name: InsertTroubleshootingGuide :one
INSERT INTO troubleshooting_guides (
    title, category, severity, device_name, function_name, error_code,
    symptom, root_cause, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, title, category, severity, device_name, function_name, error_code,
          symptom, root_cause, created_at, updated_at;

-- name: UpdateTroubleshootingGuideMeta :exec
UPDATE troubleshooting_guides
SET title = $1,
    category = $2,
    severity = $3,
    device_name = $4,
    function_name = $5,
    error_code = $6,
    symptom = $7,
    root_cause = $8,
    updated_at = $9
WHERE id = $10;

-- name: DeleteTroubleshootingGuide :execrows
DELETE FROM troubleshooting_guides WHERE id = $1;

-- name: GetTroubleshootingGuide :one
SELECT id, title, category, severity, device_name, function_name, error_code,
       symptom, root_cause, created_at, updated_at
FROM troubleshooting_guides
WHERE id = $1;

-- name: CountTroubleshootingGuides :one
SELECT COUNT(*)
FROM troubleshooting_guides g
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(g.title) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.symptom) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.root_cause) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.error_code) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.function_name) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('category')::text IS NULL OR g.category = sqlc.narg('category')::text)
  AND (sqlc.narg('device_name')::text IS NULL OR g.device_name = sqlc.narg('device_name')::text)
  AND (sqlc.narg('error_code')::text IS NULL OR lower(g.error_code) = lower(sqlc.narg('error_code')::text));

-- name: ListTroubleshootingGuides :many
SELECT g.id, g.title, g.category, g.severity, g.device_name, g.function_name,
       g.error_code, g.symptom, g.root_cause, g.created_at, g.updated_at,
       (SELECT COUNT(*) FROM troubleshooting_steps s WHERE s.guide_id = g.id) AS step_count,
       (SELECT COUNT(*) FROM troubleshooting_guides g2
        WHERE g2.id <> g.id
          AND ((g.error_code <> '' AND lower(g2.error_code) = lower(g.error_code))
               OR (g.function_name <> '' AND lower(g2.function_name) = lower(g.function_name)))) AS related_count
FROM troubleshooting_guides g
JOIN troubleshooting_severities taxonomy_severity ON taxonomy_severity.key = g.severity
JOIN troubleshooting_categories taxonomy_category ON taxonomy_category.key = g.category
WHERE (sqlc.narg('search')::text IS NULL
       OR lower(g.title) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.symptom) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.root_cause) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.error_code) LIKE lower(sqlc.narg('search')::text)
       OR lower(g.function_name) LIKE lower(sqlc.narg('search')::text))
  AND (sqlc.narg('category')::text IS NULL OR g.category = sqlc.narg('category')::text)
  AND (sqlc.narg('device_name')::text IS NULL OR g.device_name = sqlc.narg('device_name')::text)
  AND (sqlc.narg('error_code')::text IS NULL OR lower(g.error_code) = lower(sqlc.narg('error_code')::text))
ORDER BY taxonomy_severity.sort_order,
         taxonomy_category.sort_order,
         lower(g.title),
         g.id
LIMIT sqlc.arg('limit')::bigint OFFSET sqlc.arg('offset')::bigint;

-- name: ListTroubleshootingExportGuides :many
SELECT g.id, g.title, g.category, g.severity, g.device_name, g.function_name,
       g.error_code, g.symptom, g.root_cause
FROM troubleshooting_guides g
JOIN troubleshooting_severities taxonomy_severity ON taxonomy_severity.key = g.severity
JOIN troubleshooting_categories taxonomy_category ON taxonomy_category.key = g.category
ORDER BY taxonomy_severity.sort_order,
         taxonomy_category.sort_order,
         lower(g.title),
         g.id;

-- name: GetTroubleshootingRelatedGuides :many
SELECT related.id, related.title, related.category, related.severity,
       related.device_name, related.function_name, related.error_code,
       related.symptom, related.root_cause, related.created_at, related.updated_at,
       (SELECT count(*) FROM troubleshooting_steps steps WHERE steps.guide_id = related.id) AS step_count
FROM troubleshooting_guides source
JOIN troubleshooting_guides related
  ON related.id <> source.id
 AND ((source.error_code <> '' AND lower(related.error_code) = lower(source.error_code))
      OR (source.function_name <> '' AND lower(related.function_name) = lower(source.function_name)))
WHERE source.id = $1
ORDER BY CASE
             WHEN source.error_code <> ''
              AND lower(related.error_code) = lower(source.error_code) THEN 0
             ELSE 1
         END,
         lower(related.title),
         related.id;

-- name: InsertTroubleshootingStep :one
INSERT INTO troubleshooting_steps (
    guide_id, position, title, instruction, script_snippet, verification
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: DeleteTroubleshootingSteps :exec
DELETE FROM troubleshooting_steps WHERE guide_id = $1;

-- name: ListTroubleshootingSteps :many
SELECT id, guide_id, position, title, instruction, script_snippet, verification
FROM troubleshooting_steps
WHERE guide_id = $1
ORDER BY position, id;

-- name: ListTroubleshootingExportSteps :many
SELECT guide_id, title, instruction, script_snippet, verification
FROM troubleshooting_steps
ORDER BY guide_id, position, id;
