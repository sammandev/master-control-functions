-- name: UpsertFunctionNatural :one
INSERT INTO functions (
    source, interface, name, description, convention,
    category, execution_role, execution_phase, created_at, updated_at
) VALUES (
    @source, @interface, @name, @description, @convention,
    @category, @execution_role, @execution_phase, @created_at, @updated_at
)
ON CONFLICT (source, interface, name) DO UPDATE SET
    description = excluded.description,
    convention = excluded.convention,
    category = excluded.category,
    execution_role = excluded.execution_role,
    execution_phase = excluded.execution_phase,
    updated_at = excluded.updated_at
RETURNING *;

-- name: UpsertParameterDefinition :one
INSERT INTO parameter_definitions (
    name, description, type, unit, default_value,
    allowed_values, cardinality, value_pattern, required,
    device_name, created_at, updated_at
) VALUES (
    @name, @description, @type, @unit, @default_value,
    @allowed_values, @cardinality, @value_pattern, @required,
    @device_name, @created_at, @updated_at
)
ON CONFLICT (name, device_name) DO UPDATE SET
    description = excluded.description,
    type = excluded.type,
    unit = excluded.unit,
    default_value = excluded.default_value,
    allowed_values = excluded.allowed_values,
    cardinality = excluded.cardinality,
    value_pattern = excluded.value_pattern,
    required = excluded.required,
    updated_at = excluded.updated_at
RETURNING *;

-- name: UpsertCriteriaDefinition :one
INSERT INTO criteria_definitions (
    name, description, type, unit, default_value,
    allowed_values, cardinality, value_pattern, required,
    device_name, created_at, updated_at
) VALUES (
    @name, @description, @type, @unit, @default_value,
    @allowed_values, @cardinality, @value_pattern, @required,
    @device_name, @created_at, @updated_at
)
ON CONFLICT (name, device_name) DO UPDATE SET
    description = excluded.description,
    type = excluded.type,
    unit = excluded.unit,
    default_value = excluded.default_value,
    allowed_values = excluded.allowed_values,
    cardinality = excluded.cardinality,
    value_pattern = excluded.value_pattern,
    required = excluded.required,
    updated_at = excluded.updated_at
RETURNING *;

-- name: UpsertParameterLink :one
INSERT INTO parameter_links (
    function_source, function_name, device_name, name, source_file,
    position, required, created_at, updated_at
) VALUES (
    @function_source, @function_name, @device_name, @name, @source_file,
    @position, @required, @created_at, @updated_at
)
ON CONFLICT (function_source, function_name, device_name, name) DO UPDATE SET
    source_file = excluded.source_file,
    position = excluded.position,
    required = excluded.required,
    updated_at = excluded.updated_at
RETURNING *;

-- name: UpsertCriteriaLink :one
INSERT INTO criteria_links (
    function_source, function_name, device_name, name, source_file,
    position, required, created_at, updated_at
) VALUES (
    @function_source, @function_name, @device_name, @name, @source_file,
    @position, @required, @created_at, @updated_at
)
ON CONFLICT (function_source, function_name, device_name, name) DO UPDATE SET
    source_file = excluded.source_file,
    position = excluded.position,
    required = excluded.required,
    updated_at = excluded.updated_at
RETURNING *;

-- name: UpsertExample :one
INSERT INTO examples (
    function_source, function_name, device_name, script, meaning, created_at, updated_at
) VALUES (
    @function_source, @function_name, @device_name, @script, @meaning, @created_at, @updated_at
)
ON CONFLICT (function_source, function_name, device_name, script) DO UPDATE SET
    meaning = excluded.meaning,
    updated_at = excluded.updated_at
RETURNING *;

-- name: DeleteAllParameterLinks :exec
DELETE FROM parameter_links;

-- name: DeleteAllCriteriaLinks :exec
DELETE FROM criteria_links;

-- name: DeleteAllExamples :exec
DELETE FROM examples;

-- name: DeleteAllParameterDefinitions :exec
DELETE FROM parameter_definitions;

-- name: DeleteAllCriteriaDefinitions :exec
DELETE FROM criteria_definitions;

-- name: DeleteAllFunctions :exec
DELETE FROM functions;

-- name: InsertImportRun :one
INSERT INTO import_runs (
    file_name, reset, started_at, finished_at, counts, skipped
) VALUES (
    @file_name, @reset, @started_at, @finished_at, @counts, @skipped
)
RETURNING *;

-- name: ListImportRuns :many
SELECT * FROM import_runs ORDER BY id DESC LIMIT 50;

-- name: GetImportRun :one
SELECT * FROM import_runs WHERE id = $1;

-- name: InsertImportNormalization :exec
INSERT INTO import_normalizations (
    run_id, table_name, row_key, column_name, source_value, normalized_value, rule
) VALUES (
    @run_id, @table_name, @row_key, @column_name, @source_value, @normalized_value, @rule
);

-- name: ListImportNormalizations :many
SELECT * FROM import_normalizations WHERE run_id = $1 ORDER BY id;

-- name: ListAllParameterLinks :many
SELECT * FROM parameter_links ORDER BY id;

-- name: ListAllCriteriaLinks :many
SELECT * FROM criteria_links ORDER BY id;

-- name: ListAllExamples :many
SELECT * FROM examples ORDER BY id;
