-- name: DistinctFunctionSources :many
SELECT DISTINCT source FROM functions ORDER BY source;

-- name: DistinctFunctionInterfaces :many
SELECT DISTINCT interface FROM functions ORDER BY interface;

-- name: DistinctFunctionCategories :many
SELECT DISTINCT category FROM functions WHERE category <> '' ORDER BY category;

-- name: DistinctFunctionExecutionRoles :many
SELECT DISTINCT execution_role FROM functions WHERE execution_role <> '' ORDER BY execution_role;

-- name: DistinctFunctionExecutionPhases :many
SELECT DISTINCT execution_phase FROM functions WHERE execution_phase <> '' ORDER BY execution_phase;
