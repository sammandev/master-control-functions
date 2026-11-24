-- Global definitions resolved by a device's links: a link in the device
-- references the name and no device-scoped sibling shadows it.
-- name: ListGlobalParameterDefinitionsForScope :many
SELECT pd.id, pd.name, pd.description, pd.type, pd.unit, pd.default_value, pd.allowed_values, pd.cardinality, pd.value_pattern, pd.required, pd.device_name, pd.created_at, pd.updated_at
FROM parameter_definitions pd
WHERE pd.device_name = ''
  AND EXISTS (SELECT 1 FROM parameter_links pl
              WHERE pl.name = pd.name AND pl.device_name = sqlc.arg(device))
  AND NOT EXISTS (SELECT 1 FROM parameter_definitions s
                  WHERE s.name = pd.name AND s.device_name = sqlc.arg(device))
ORDER BY pd.name;

-- name: ListGlobalCriteriaDefinitionsForScope :many
SELECT cd.id, cd.name, cd.description, cd.type, cd.unit, cd.default_value, cd.allowed_values, cd.cardinality, cd.value_pattern, cd.required, cd.device_name, cd.created_at, cd.updated_at
FROM criteria_definitions cd
WHERE cd.device_name = ''
  AND EXISTS (SELECT 1 FROM criteria_links cl
              WHERE cl.name = cd.name AND cl.device_name = sqlc.arg(device))
  AND NOT EXISTS (SELECT 1 FROM criteria_definitions s
                  WHERE s.name = cd.name AND s.device_name = sqlc.arg(device))
ORDER BY cd.name;

-- Registry CRUD (ADR-0010). The delete guard counts every content row a
-- device name appears on; a registered device may only be deleted when the
-- count is zero (registered-but-empty).
