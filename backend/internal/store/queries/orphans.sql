-- Orphan report queries (ticket 05, ADR-0003). An orphan is a link or
-- Example row whose target Function, Parameter definition, or Criteria
-- definition does not exist. Nothing is ever dropped or filtered out here.
--
-- Target existence is tested with EXISTS subqueries, not LEFT JOINs:
-- functions are unique on (source, interface, name), so a join on
-- (source, name) alone would multiply rows for functions that exist under
-- several interfaces.
--
-- Definition existence uses exact-scope-then-global resolution
-- (ADR-0009): a link resolves when (name, device_name) matches OR
-- (name, device_name='') matches. A link to a WIFI_TEST-only definition
-- is an orphan when its device_name is something else.

-- name: ListOrphanParameterLinks :many
SELECT l.id, l.function_source, l.function_name, l.device_name, l.name,
       l.position, l.required, l.created_at, l.updated_at, l.source_file,
       EXISTS (SELECT 1 FROM functions f
               WHERE f.source = l.function_source AND f.name = l.function_name) AS function_exists,
       EXISTS (SELECT 1 FROM parameter_definitions d
               WHERE d.name = l.name
                 AND (d.device_name = l.device_name OR d.device_name = '')) AS definition_exists,
       (SELECT MAX(f.id) FROM functions f
        WHERE f.source = l.function_source AND f.name = l.function_name) AS function_id
FROM parameter_links l
WHERE NOT EXISTS (SELECT 1 FROM functions f
                  WHERE f.source = l.function_source AND f.name = l.function_name)
   OR NOT EXISTS (SELECT 1 FROM parameter_definitions d
                  WHERE d.name = l.name
                    AND (d.device_name = l.device_name OR d.device_name = ''));

-- name: ListOrphanCriteriaLinks :many
SELECT l.id, l.function_source, l.function_name, l.device_name, l.name,
       l.position, l.required, l.created_at, l.updated_at, l.source_file,
       EXISTS (SELECT 1 FROM functions f
               WHERE f.source = l.function_source AND f.name = l.function_name) AS function_exists,
       EXISTS (SELECT 1 FROM criteria_definitions d
               WHERE d.name = l.name
                 AND (d.device_name = l.device_name OR d.device_name = '')) AS definition_exists,
       (SELECT MAX(f.id) FROM functions f
        WHERE f.source = l.function_source AND f.name = l.function_name) AS function_id
FROM criteria_links l
WHERE NOT EXISTS (SELECT 1 FROM functions f
                  WHERE f.source = l.function_source AND f.name = l.function_name)
   OR NOT EXISTS (SELECT 1 FROM criteria_definitions d
                  WHERE d.name = l.name
                    AND (d.device_name = l.device_name OR d.device_name = ''));

-- name: ListOrphanExamples :many
SELECT e.id, e.function_source, e.function_name, e.device_name, e.script,
       e.meaning, e.created_at, e.updated_at
FROM examples e
WHERE NOT EXISTS (SELECT 1 FROM functions f
                  WHERE f.source = e.function_source AND f.name = e.function_name);

-- Orphan usage for the dictionary pages: links referencing a definition by
-- name whose Function does not exist. Distinct from Count*Usage (total).
-- name: CountParameterOrphanUsage :one
SELECT COUNT(*)
FROM parameter_links l
WHERE l.name = $1
  AND NOT EXISTS (SELECT 1 FROM functions f
                  WHERE f.source = l.function_source AND f.name = l.function_name);

-- name: CountCriteriaOrphanUsage :one
SELECT COUNT(*)
FROM criteria_links l
WHERE l.name = $1
  AND NOT EXISTS (SELECT 1 FROM functions f
                  WHERE f.source = l.function_source AND f.name = l.function_name);
