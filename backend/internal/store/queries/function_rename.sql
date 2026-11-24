-- name: RekeyFunctionParameterLinks :exec
WITH duplicates AS (
    SELECT old.id FROM parameter_links AS old
    JOIN parameter_links AS target ON target.device_name = old.device_name AND target.name = old.name
    WHERE NOT sqlc.arg(keep_old)::boolean
      AND old.function_source = @old_source AND old.function_name = @old_name
      AND target.function_source = @new_source AND target.function_name = @new_name
), removed_duplicates AS (
    DELETE FROM parameter_links WHERE id IN (SELECT id FROM duplicates)
), moved AS (
    UPDATE parameter_links SET function_source = @new_source, function_name = @new_name, updated_at = @updated_at
    WHERE NOT sqlc.arg(keep_old)::boolean AND function_source = @old_source AND function_name = @old_name
      AND id NOT IN (SELECT id FROM duplicates)
)
INSERT INTO parameter_links (function_source, function_name, device_name, created_at, updated_at, name, position, required, source_file)
SELECT @new_source, @new_name, src.device_name, src.created_at, @updated_at, src.name, src.position, src.required, src.source_file
FROM parameter_links AS src
WHERE sqlc.arg(keep_old)::boolean AND src.function_source = @old_source AND src.function_name = @old_name
ON CONFLICT (function_source, function_name, device_name, name) DO NOTHING;

-- name: RekeyFunctionCriteriaLinks :exec
WITH duplicates AS (
    SELECT old.id FROM criteria_links AS old
    JOIN criteria_links AS target ON target.device_name = old.device_name AND target.name = old.name
    WHERE NOT sqlc.arg(keep_old)::boolean
      AND old.function_source = @old_source AND old.function_name = @old_name
      AND target.function_source = @new_source AND target.function_name = @new_name
), removed_duplicates AS (
    DELETE FROM criteria_links WHERE id IN (SELECT id FROM duplicates)
), moved AS (
    UPDATE criteria_links SET function_source = @new_source, function_name = @new_name, updated_at = @updated_at
    WHERE NOT sqlc.arg(keep_old)::boolean AND function_source = @old_source AND function_name = @old_name
      AND id NOT IN (SELECT id FROM duplicates)
)
INSERT INTO criteria_links (function_source, function_name, device_name, created_at, updated_at, name, position, required, source_file)
SELECT @new_source, @new_name, src.device_name, src.created_at, @updated_at, src.name, src.position, src.required, src.source_file
FROM criteria_links AS src
WHERE sqlc.arg(keep_old)::boolean AND src.function_source = @old_source AND src.function_name = @old_name
ON CONFLICT (function_source, function_name, device_name, name) DO NOTHING;

-- name: RekeyFunctionExamples :exec
WITH duplicates AS (
    SELECT old.id FROM examples AS old
    JOIN examples AS target ON target.device_name = old.device_name AND target.script = old.script
    WHERE NOT sqlc.arg(keep_old)::boolean
      AND old.function_source = @old_source AND old.function_name = @old_name
      AND target.function_source = @new_source AND target.function_name = @new_name
), removed_duplicates AS (
    DELETE FROM examples WHERE id IN (SELECT id FROM duplicates)
), moved AS (
    UPDATE examples SET function_source = @new_source, function_name = @new_name, updated_at = @updated_at
    WHERE NOT sqlc.arg(keep_old)::boolean AND function_source = @old_source AND function_name = @old_name
      AND id NOT IN (SELECT id FROM duplicates)
)
INSERT INTO examples (function_source, function_name, device_name, created_at, updated_at, script, meaning)
SELECT @new_source, @new_name, src.device_name, src.created_at, @updated_at, src.script, src.meaning
FROM examples AS src
WHERE sqlc.arg(keep_old)::boolean AND src.function_source = @old_source AND src.function_name = @old_name
ON CONFLICT (function_source, function_name, device_name, script) DO NOTHING;
