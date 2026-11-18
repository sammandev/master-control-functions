-- name: CountFunctionsByDeviceScope :one
SELECT COUNT(*) FROM functions f
WHERE f.interface = sqlc.arg(device)
  OR EXISTS (SELECT 1 FROM parameter_links pl
              WHERE pl.function_source = f.source AND pl.function_name = f.name
                AND pl.device_name = sqlc.arg(device))
   OR EXISTS (SELECT 1 FROM criteria_links cl
              WHERE cl.function_source = f.source AND cl.function_name = f.name
                AND cl.device_name = sqlc.arg(device));

-- name: ListFunctionsByDeviceScope :many
-- Explicit columns, not f.*: sqlc v1.31.1's star-expansion edit mis-splices
-- at this file's offsets ("edited query syntax is invalid"); the CTE index
-- query lives apart in devices_index.sql for the same reason.
SELECT f.id, f.source, f.interface, f.name, f.description, f.convention, f.category, f.execution_role, f.execution_phase, f.created_at, f.updated_at
FROM functions f
WHERE f.interface = sqlc.arg(device)
  OR EXISTS (SELECT 1 FROM parameter_links pl
              WHERE pl.function_source = f.source AND pl.function_name = f.name
                AND pl.device_name = sqlc.arg(device))
   OR EXISTS (SELECT 1 FROM criteria_links cl
              WHERE cl.function_source = f.source AND cl.function_name = f.name
                AND cl.device_name = sqlc.arg(device))
ORDER BY f.name, f.source, f.interface;
