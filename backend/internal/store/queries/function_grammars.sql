-- Script grammar and script-key audits (ADR-0012). Keyed by the Function
-- natural key so they survive workbook import/reset.

-- name: GetFunctionGrammar :one
SELECT * FROM function_grammars
WHERE function_source = $1 AND function_interface = $2 AND function_name = $3;

-- name: UpsertFunctionGrammar :one
INSERT INTO function_grammars (
    function_source, function_interface, function_name, encoding, scope,
    audited_by, audited_at, note, created_at, updated_at
) VALUES (
    @function_source, @function_interface, @function_name, @encoding, @scope,
    @audited_by, @audited_at, @note, @now, @now
)
ON CONFLICT (function_source, function_interface, function_name) DO UPDATE SET
    encoding = EXCLUDED.encoding,
    scope = EXCLUDED.scope,
    audited_by = EXCLUDED.audited_by,
    audited_at = EXCLUDED.audited_at,
    note = EXCLUDED.note,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteFunctionGrammar :execrows
DELETE FROM function_grammars
WHERE function_source = $1 AND function_interface = $2 AND function_name = $3;

-- name: ListScriptKeyAudits :many
SELECT * FROM script_key_audits
WHERE function_source = $1 AND function_interface = $2 AND function_name = $3
ORDER BY kind, name;

-- name: UpsertScriptKeyAudit :one
INSERT INTO script_key_audits (
    function_source, function_interface, function_name, kind, name, scope,
    audited_by, audited_at, note, created_at, updated_at
) VALUES (
    @function_source, @function_interface, @function_name, @kind, @name, @scope,
    @audited_by, @audited_at, @note, @now, @now
)
ON CONFLICT (function_source, function_interface, function_name, kind, name) DO UPDATE SET
    scope = EXCLUDED.scope,
    audited_by = EXCLUDED.audited_by,
    audited_at = EXCLUDED.audited_at,
    note = EXCLUDED.note,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteScriptKeyAudit :execrows
DELETE FROM script_key_audits
WHERE function_source = $1 AND function_interface = $2 AND function_name = $3
  AND kind = $4 AND name = $5;

-- name: RenameFunctionGrammar :exec
UPDATE function_grammars SET
    function_source = @new_source,
    function_interface = @new_interface,
    function_name = @new_name,
    updated_at = @updated_at
WHERE function_source = @old_source
  AND function_interface = @old_interface
  AND function_name = @old_name;

-- name: RenameScriptKeyAudits :exec
UPDATE script_key_audits SET
    function_source = @new_source,
    function_interface = @new_interface,
    function_name = @new_name,
    updated_at = @updated_at
WHERE function_source = @old_source
  AND function_interface = @old_interface
  AND function_name = @old_name;

-- name: CountFunctionsSharingNaturalName :one
-- Links and examples key a Function by (source, name) only; more than one
-- match means their interface is ambiguous (ADR-0012).
SELECT COUNT(*) FROM functions WHERE source = $1 AND name = $2;

-- name: ListAllFunctionGrammars :many
SELECT * FROM function_grammars ORDER BY id;

-- name: ListAllScriptKeyAudits :many
SELECT * FROM script_key_audits ORDER BY id;

-- name: ListAllFunctionsForReadiness :many
SELECT * FROM functions ORDER BY source, interface, name;

-- name: ListAllParameterDefinitionsForReadiness :many
SELECT * FROM parameter_definitions ORDER BY id;

-- name: ListAllCriteriaDefinitionsForReadiness :many
SELECT * FROM criteria_definitions ORDER BY id;

-- name: ListFunctionsByNaturalName :many
SELECT * FROM functions WHERE source = $1 AND name = $2 ORDER BY interface;
