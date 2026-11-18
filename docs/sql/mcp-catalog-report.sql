-- MCP catalog health report (ADR-0012). Read-only: run with psql against the
-- web catalog, e.g. psql "$DATABASE_URL" -f docs/sql/mcp-catalog-report.sql
-- Every query returns offending rows; an empty result means healthy.

BEGIN READ ONLY;

\echo '1. Ambiguous (source, name): links/examples cannot infer an interface'
SELECT source, name, string_agg(interface, ', ' ORDER BY interface) AS interfaces
FROM functions GROUP BY source, name HAVING COUNT(*) > 1 ORDER BY source, name;

\echo '2. Orphan parameter/criteria links and examples (no Function)'
SELECT 'parameter_link' AS kind, l.id, l.function_source, l.function_name, l.name
FROM parameter_links l
WHERE NOT EXISTS (SELECT 1 FROM functions f WHERE f.source = l.function_source AND f.name = l.function_name)
UNION ALL
SELECT 'criteria_link', l.id, l.function_source, l.function_name, l.name
FROM criteria_links l
WHERE NOT EXISTS (SELECT 1 FROM functions f WHERE f.source = l.function_source AND f.name = l.function_name)
UNION ALL
SELECT 'example', e.id, e.function_source, e.function_name, left(e.script, 60)
FROM examples e
WHERE NOT EXISTS (SELECT 1 FROM functions f WHERE f.source = e.function_source AND f.name = e.function_name)
ORDER BY 1, 2;

\echo '3. Links without a definition (device-scoped or global)'
SELECT 'parameter' AS kind, l.id, l.function_source, l.function_name, l.device_name, l.name
FROM parameter_links l
WHERE NOT EXISTS (SELECT 1 FROM parameter_definitions d
                  WHERE d.name = l.name AND d.device_name IN (l.device_name, ''))
UNION ALL
SELECT 'criterion', l.id, l.function_source, l.function_name, l.device_name, l.name
FROM criteria_links l
WHERE NOT EXISTS (SELECT 1 FROM criteria_definitions d
                  WHERE d.name = l.name AND d.device_name IN (l.device_name, ''))
ORDER BY 1, 2;

\echo '4. Definitions outside the vocabulary (type, cardinality, required)'
SELECT 'parameter' AS kind, id, name, device_name, type, cardinality, required
FROM parameter_definitions
WHERE type NOT IN ('string', 'int', 'double', 'list[int]')
   OR cardinality NOT IN ('SINGLE', 'MULTIPLE')
   OR required NOT IN (0, 1)
UNION ALL
SELECT 'criterion', id, name, device_name, type, cardinality, required
FROM criteria_definitions
WHERE type NOT IN ('string', 'int', 'double', 'list[int]')
   OR cardinality NOT IN ('SINGLE', 'MULTIPLE')
   OR required NOT IN (0, 1)
ORDER BY 1, 3, 4;

\echo '5. Grammar and script-key audit rows with no matching Function'
SELECT 'grammar' AS kind, g.function_source, g.function_interface, g.function_name, '' AS key
FROM function_grammars g
WHERE NOT EXISTS (SELECT 1 FROM functions f WHERE f.source = g.function_source
                  AND f.interface = g.function_interface AND f.name = g.function_name)
UNION ALL
SELECT 'script_key', k.function_source, k.function_interface, k.function_name, k.kind || ':' || k.name
FROM script_key_audits k
WHERE NOT EXISTS (SELECT 1 FROM functions f WHERE f.source = k.function_source
                  AND f.interface = k.function_interface AND f.name = k.function_name)
ORDER BY 1, 2, 3, 4;

\echo '6. Grammar encoding/convention mismatches (raw needs 1, named needs 2)'
SELECT f.source, f.interface, f.name, f.convention, g.encoding
FROM functions f
JOIN function_grammars g ON g.function_source = f.source
 AND g.function_interface = f.interface AND g.function_name = f.name
WHERE (g.encoding = 'raw' AND f.convention IS DISTINCT FROM 1)
   OR (g.encoding = 'named' AND f.convention IS DISTINCT FROM 2)
ORDER BY 1, 2, 3;

\echo '7. Definitions linked to audited Functions but missing value_pattern'
SELECT DISTINCT 'parameter' AS kind, d.name, d.device_name
FROM parameter_definitions d
JOIN parameter_links l ON l.name = d.name AND d.device_name IN (l.device_name, '')
JOIN function_grammars g ON g.function_source = l.function_source AND g.function_name = l.function_name
 AND l.device_name IN ('', g.function_interface)
WHERE d.value_pattern = ''
UNION
SELECT DISTINCT 'criterion', d.name, d.device_name
FROM criteria_definitions d
JOIN criteria_links l ON l.name = d.name AND d.device_name IN (l.device_name, '')
JOIN function_grammars g ON g.function_source = l.function_source AND g.function_name = l.function_name
 AND l.device_name IN ('', g.function_interface)
WHERE d.value_pattern = ''
ORDER BY 1, 2, 3;

ROLLBACK;
