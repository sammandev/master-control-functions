-- name: GetRegisteredDevice :one
SELECT name, created_at, updated_at, source FROM devices WHERE name = $1;

-- name: InsertDevice :one
INSERT INTO devices (name, created_at, updated_at, source)
VALUES ($1, $2, $3, $4)
RETURNING name, created_at, updated_at, source;

-- name: ListDeviceSources :many
WITH content_sources AS (
        SELECT interface AS name, source FROM functions
        UNION SELECT device_name, function_source FROM parameter_links
        UNION SELECT device_name, function_source FROM criteria_links
        UNION SELECT device_name, function_source FROM examples
)
SELECT name, source FROM devices WHERE source IS NOT NULL
UNION
SELECT content.name, content.source FROM content_sources content
WHERE content.name <> '' AND content.source IN ('BUILT-IN', 'DLL')
    AND NOT EXISTS (
            SELECT 1 FROM devices registered
            WHERE registered.name = content.name AND registered.source IS NOT NULL
    )
ORDER BY name, source;

-- name: DeleteRegisteredDevice :execrows
DELETE FROM devices WHERE name = $1;

-- name: CountDeviceNameUsages :one
SELECT (SELECT COUNT(*) FROM parameter_links pl WHERE pl.device_name = sqlc.arg(device))
     + (SELECT COUNT(*) FROM criteria_links cl WHERE cl.device_name = sqlc.arg(device))
     + (SELECT COUNT(*) FROM parameter_definitions pd WHERE pd.device_name = sqlc.arg(device))
     + (SELECT COUNT(*) FROM criteria_definitions cd WHERE cd.device_name = sqlc.arg(device))
    + (SELECT COUNT(*) FROM examples ex WHERE ex.device_name = sqlc.arg(device))
    + (SELECT COUNT(*) FROM functions f WHERE f.interface = sqlc.arg(device));

-- name: RegisterImportedDeviceScopes :execrows
-- Auto-populate the registry from imported content (ADR-0013): every
-- distinct non-blank device_name across the five scope tables becomes a
-- registered device. Idempotent (conflicts skipped); never deletes, so
-- manual registrations survive imports and resets.
INSERT INTO devices (name, created_at, updated_at)
SELECT v.device_name, $1, $1
FROM (
    SELECT device_name FROM parameter_links
    UNION SELECT device_name FROM criteria_links
    UNION SELECT device_name FROM parameter_definitions
    UNION SELECT device_name FROM criteria_definitions
    UNION SELECT device_name FROM examples
    UNION SELECT interface FROM functions
) v
WHERE v.device_name <> ''
ON CONFLICT (name) DO NOTHING;
