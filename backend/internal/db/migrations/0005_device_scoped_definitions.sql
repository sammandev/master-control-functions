-- 0005: device-scoped definitions (ADR-0009). parameter_definitions and
-- criteria_definitions gain a device_name column and key on
-- (name, device_name); blank device_name is the global/shared definition
-- (CONTEXT.md) and the fallback target for link-to-definition resolution.
--
-- SQLite cannot ALTER ADD CONSTRAINT or replace a UNIQUE on an existing
-- table, so both tables are rebuilt in place (CREATE new → INSERT SELECT
-- → DROP old → RENAME), which is plain DDL: no PRAGMA, no INSERT OR
-- REPLACE, no SQLite-only syntax. The runner wraps the whole file in one
-- transaction (db.Migrate), so the rebuild is atomic.

CREATE TABLE parameter_definitions_new (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'string',
    unit TEXT NOT NULL DEFAULT '',
    default_value TEXT,
    allowed_values TEXT NOT NULL DEFAULT '',
    cardinality TEXT NOT NULL DEFAULT 'SINGLE',
    value_pattern TEXT NOT NULL DEFAULT '',
    required INTEGER NOT NULL DEFAULT 0,
    device_name TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (name, device_name)
);

INSERT INTO parameter_definitions_new
    (id, name, description, type, unit, default_value,
     allowed_values, cardinality, value_pattern, required,
     device_name, created_at, updated_at)
SELECT id, name, description, type, unit, default_value,
       allowed_values, cardinality, value_pattern, required,
       '', created_at, updated_at
FROM parameter_definitions
ORDER BY id;

DROP TABLE parameter_definitions;
ALTER TABLE parameter_definitions_new RENAME TO parameter_definitions;

CREATE TABLE criteria_definitions_new (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'string',
    unit TEXT NOT NULL DEFAULT '',
    default_value TEXT,
    allowed_values TEXT NOT NULL DEFAULT '',
    cardinality TEXT NOT NULL DEFAULT 'SINGLE',
    value_pattern TEXT NOT NULL DEFAULT '',
    required INTEGER NOT NULL DEFAULT 0,
    device_name TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (name, device_name)
);

INSERT INTO criteria_definitions_new
    (id, name, description, type, unit, default_value,
     allowed_values, cardinality, value_pattern, required,
     device_name, created_at, updated_at)
SELECT id, name, description, type, unit, default_value,
       allowed_values, cardinality, value_pattern, required,
       '', created_at, updated_at
FROM criteria_definitions
ORDER BY id;

DROP TABLE criteria_definitions;
ALTER TABLE criteria_definitions_new RENAME TO criteria_definitions;
