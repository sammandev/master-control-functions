-- name: ListScopedParameterDefinitions :many
SELECT id, name, description, type, unit, default_value, allowed_values, cardinality, value_pattern, required, device_name, created_at, updated_at
FROM parameter_definitions
WHERE device_name = $1
ORDER BY name;

-- name: CountScopedParameterDefinitions :one
SELECT COUNT(*) FROM parameter_definitions WHERE device_name = $1;

-- Global definitions resolved by a device's links: a link in the device
-- references the name and no device-scoped sibling shadows it.

-- name: ListScopedCriteriaDefinitions :many
SELECT id, name, description, type, unit, default_value, allowed_values, cardinality, value_pattern, required, device_name, created_at, updated_at
FROM criteria_definitions
WHERE device_name = $1
ORDER BY name;

-- name: CountScopedCriteriaDefinitions :one
SELECT COUNT(*) FROM criteria_definitions WHERE device_name = $1;

