-- Script composer queries (ticket 07). Each composer field is a link row
-- enriched with its definition's constraints (type, unit, allowed values,
-- cardinality, default). Position NULLs sort last (same order the links
-- tab shows); duplicate names across device scopes are deduped Go-side,
-- first row wins.
--
-- Definition resolution is exact-scope-then-global (ADR-0009): the
-- definition row whose (name, device_name) matches the link's, or the
-- blank-scope row when no exact match exists. When both exist, the
-- exact-scope one wins because it overrides the global for that scope.

-- name: ListComposerParameterFields :many
SELECT l.name,
       l.position,
       l.required,
       l.device_name,
       d.type,
       d.unit,
       d.default_value,
       d.allowed_values,
       d.cardinality
FROM parameter_links l
LEFT JOIN parameter_definitions d
       ON d.id = COALESCE(
              (SELECT d2.id FROM parameter_definitions d2
               WHERE d2.name = l.name AND d2.device_name = l.device_name),
              (SELECT d3.id FROM parameter_definitions d3
               WHERE d3.name = l.name AND d3.device_name = ''))
WHERE l.function_source = $1 AND l.function_name = $2
ORDER BY CASE WHEN l.position IS NULL THEN 1 ELSE 0 END,
          l.position, l.name, l.id;

-- name: ListComposerCriteriaFields :many
SELECT l.name,
       l.position,
       l.required,
       l.device_name,
       d.type,
       d.unit,
       d.default_value,
       d.allowed_values,
       d.cardinality
FROM criteria_links l
LEFT JOIN criteria_definitions d
       ON d.id = COALESCE(
              (SELECT d2.id FROM criteria_definitions d2
               WHERE d2.name = l.name AND d2.device_name = l.device_name),
              (SELECT d3.id FROM criteria_definitions d3
               WHERE d3.name = l.name AND d3.device_name = ''))
WHERE l.function_source = $1 AND l.function_name = $2
ORDER BY CASE WHEN l.position IS NULL THEN 1 ELSE 0 END,
          l.position, l.name, l.id;
