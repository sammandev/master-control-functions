-- Device index query: the union of every distinct device_name value with
-- per-device counts. Kept in its own file: sqlc v1.31.1's analyzer mangles
-- multi-query files that combine this parameterless CTE with parameterized
-- queries (pre-existing latent breakage, exposed by the ticket-02 regen).

-- name: ListDeviceScopes :many
WITH all_names (device_name) AS (
    SELECT device_name FROM parameter_links
    UNION SELECT device_name FROM criteria_links
    UNION SELECT device_name FROM parameter_definitions
    UNION SELECT device_name FROM criteria_definitions
    UNION SELECT device_name FROM examples
    UNION SELECT name FROM devices
    UNION SELECT interface FROM functions
)
SELECT
    v.device_name AS device_name,
    (SELECT COUNT(*) FROM functions f
      WHERE f.interface = v.device_name
        OR EXISTS (SELECT 1 FROM parameter_links pl
                   WHERE pl.function_source = f.source
                     AND pl.function_name = f.name
                     AND pl.device_name = v.device_name)
        OR EXISTS (SELECT 1 FROM criteria_links cl
                   WHERE cl.function_source = f.source
                     AND cl.function_name = f.name
                     AND cl.device_name = v.device_name)) AS function_count,
    (SELECT COUNT(*) FROM parameter_definitions d
     WHERE d.device_name = v.device_name) AS parameter_count,
    (SELECT COUNT(*) FROM criteria_definitions cd
     WHERE cd.device_name = v.device_name) AS criteria_count,
    (SELECT COUNT(*) FROM examples e
     WHERE e.device_name = v.device_name) AS example_count
FROM all_names v
ORDER BY v.device_name;
