-- 0003: import support. Three concerns:
--
-- 1. The workbook's link sheets carry a source_file column (which ini file
--    defines the entry); the link tables missed it in 0002. SQLite cannot
--    ALTER ADD CONSTRAINT, but ADD COLUMN is supported.
-- 2. Examples need a natural key for ADR-0004's idempotent upsert
--    (function + device scope + script text). SQLite cannot add a UNIQUE
--    constraint to an existing table, so examples is recreated with it.
-- 3. The change log: one row per import run plus one row per applied
--    Normalization (CONTEXT.md), auditable source-value → normalized-value.
ALTER TABLE parameter_links ADD COLUMN source_file TEXT NOT NULL DEFAULT '';
ALTER TABLE criteria_links ADD COLUMN source_file TEXT NOT NULL DEFAULT '';

CREATE TABLE examples_new (
    id INTEGER PRIMARY KEY,
    function_source TEXT NOT NULL,
    function_name TEXT NOT NULL,
    device_name TEXT NOT NULL DEFAULT '',
    script TEXT NOT NULL,
    meaning TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (function_source, function_name, device_name, script)
);

INSERT INTO examples_new
    (id, function_source, function_name, device_name, script, meaning, created_at, updated_at)
SELECT id, function_source, function_name, device_name, script, meaning, created_at, updated_at
FROM examples;

DROP TABLE examples;
ALTER TABLE examples_new RENAME TO examples;

CREATE INDEX idx_examples_function ON examples (function_source, function_name);

CREATE TABLE import_runs (
    id INTEGER PRIMARY KEY,
    file_name TEXT NOT NULL,
    reset INTEGER NOT NULL DEFAULT 0,
    started_at TEXT NOT NULL,
    finished_at TEXT NOT NULL,
    counts TEXT NOT NULL DEFAULT '{}',
    skipped TEXT NOT NULL DEFAULT '[]'
);

CREATE TABLE import_normalizations (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES import_runs (id),
    table_name TEXT NOT NULL,
    row_key TEXT NOT NULL,
    column_name TEXT NOT NULL,
    source_value TEXT NOT NULL,
    normalized_value TEXT NOT NULL,
    rule TEXT NOT NULL
);

CREATE INDEX idx_import_normalizations_run ON import_normalizations (run_id);
