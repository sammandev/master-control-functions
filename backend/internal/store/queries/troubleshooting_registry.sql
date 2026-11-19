-- Managed troubleshooting taxonomy registry queries.

-- name: ListTroubleshootingCategories :many
SELECT key, label, description, sort_order, built_in
FROM troubleshooting_categories
ORDER BY sort_order, key;

-- name: GetTroubleshootingCategory :one
SELECT key, label, description, sort_order, built_in
FROM troubleshooting_categories
WHERE key = $1;

-- name: InsertTroubleshootingCategory :one
INSERT INTO troubleshooting_categories (key, label, description, sort_order)
VALUES (sqlc.arg(key), sqlc.arg(label), sqlc.arg(description), sqlc.arg(sort_order))
RETURNING key, label, description, sort_order, built_in;

-- name: UpdateTroubleshootingCategory :one
UPDATE troubleshooting_categories
SET label = sqlc.arg(label),
    description = sqlc.arg(description),
    sort_order = sqlc.arg(sort_order)
WHERE key = sqlc.arg(key)
RETURNING key, label, description, sort_order, built_in;

-- name: UpdateTroubleshootingCategorySortOrder :one
UPDATE troubleshooting_categories
SET sort_order = sqlc.arg(sort_order)
WHERE key = sqlc.arg(key)
RETURNING key, label, description, sort_order, built_in;

-- name: ListTroubleshootingCategoryGuideCounts :many
SELECT category AS key, COUNT(*)::BIGINT AS guide_count
FROM troubleshooting_guides
GROUP BY category;

-- name: CountTroubleshootingCategoryGuides :one
SELECT COUNT(*)::BIGINT
FROM troubleshooting_guides
WHERE category = $1;

-- name: DeleteTroubleshootingCategory :execrows
DELETE FROM troubleshooting_categories
WHERE key = $1
  AND built_in = FALSE;

-- name: ReassignTroubleshootingGuideCategories :execrows
UPDATE troubleshooting_guides
SET category = sqlc.arg(new_key),
    updated_at = sqlc.arg(updated_at)
WHERE category = sqlc.arg(old_key);

-- name: ListTroubleshootingSeverities :many
SELECT key, label, description, sort_order, built_in, tone
FROM troubleshooting_severities
ORDER BY sort_order, key;

-- name: GetTroubleshootingSeverity :one
SELECT key, label, description, sort_order, built_in, tone
FROM troubleshooting_severities
WHERE key = $1;

-- name: InsertTroubleshootingSeverity :one
INSERT INTO troubleshooting_severities (key, label, description, sort_order, tone)
VALUES (sqlc.arg(key), sqlc.arg(label), sqlc.arg(description), sqlc.arg(sort_order), sqlc.arg(tone))
RETURNING key, label, description, sort_order, built_in, tone;

-- name: UpdateTroubleshootingSeverity :one
UPDATE troubleshooting_severities
SET label = sqlc.arg(label),
    description = sqlc.arg(description),
    sort_order = sqlc.arg(sort_order),
    tone = sqlc.arg(tone)
WHERE key = sqlc.arg(key)
RETURNING key, label, description, sort_order, built_in, tone;

-- name: UpdateTroubleshootingSeveritySortOrder :one
UPDATE troubleshooting_severities
SET sort_order = sqlc.arg(sort_order)
WHERE key = sqlc.arg(key)
RETURNING key, label, description, sort_order, built_in, tone;

-- name: ListTroubleshootingSeverityGuideCounts :many
SELECT severity AS key, COUNT(*)::BIGINT AS guide_count
FROM troubleshooting_guides
GROUP BY severity;

-- name: CountTroubleshootingSeverityGuides :one
SELECT COUNT(*)::BIGINT
FROM troubleshooting_guides
WHERE severity = $1;

-- name: DeleteTroubleshootingSeverity :execrows
DELETE FROM troubleshooting_severities
WHERE key = $1
  AND built_in = FALSE;

-- name: ReassignTroubleshootingGuideSeverities :execrows
UPDATE troubleshooting_guides
SET severity = sqlc.arg(new_key),
    updated_at = sqlc.arg(updated_at)
WHERE severity = sqlc.arg(old_key);
