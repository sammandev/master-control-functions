-- 0002: remaining entity tables of the knowledge base. Definitions are the
-- two dictionaries (same shape, separate tables per CONTEXT.md); link rows
-- carry device scope as an attribute (ADR-0002) and intentionally have no
-- foreign keys so Orphan rows are storable (ADR-0003). ANSI SQL only.
CREATE TABLE parameter_definitions (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'string',
    unit TEXT NOT NULL DEFAULT '',
    default_value TEXT,
    allowed_values TEXT NOT NULL DEFAULT '',
    cardinality TEXT NOT NULL DEFAULT 'SINGLE',
    value_pattern TEXT NOT NULL DEFAULT '',
    required INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE criteria_definitions (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'string',
    unit TEXT NOT NULL DEFAULT '',
    default_value TEXT,
    allowed_values TEXT NOT NULL DEFAULT '',
    cardinality TEXT NOT NULL DEFAULT 'SINGLE',
    value_pattern TEXT NOT NULL DEFAULT '',
    required INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- A parameter link binds a Parameter definition to a Function at a
-- position with a required flag and a Device scope.
CREATE TABLE parameter_links (
    id INTEGER PRIMARY KEY,
    function_source TEXT NOT NULL,
    function_name TEXT NOT NULL,
    device_name TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    position INTEGER,
    required INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (function_source, function_name, device_name, name)
);

CREATE TABLE criteria_links (
    id INTEGER PRIMARY KEY,
    function_source TEXT NOT NULL,
    function_name TEXT NOT NULL,
    device_name TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    position INTEGER,
    required INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (function_source, function_name, device_name, name)
);

CREATE TABLE examples (
    id INTEGER PRIMARY KEY,
    function_source TEXT NOT NULL,
    function_name TEXT NOT NULL,
    device_name TEXT NOT NULL DEFAULT '',
    script TEXT NOT NULL,
    meaning TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_parameter_links_name ON parameter_links (name);
CREATE INDEX idx_criteria_links_name ON criteria_links (name);
CREATE INDEX idx_parameter_links_function
    ON parameter_links (function_source, function_name);
CREATE INDEX idx_criteria_links_function
    ON criteria_links (function_source, function_name);
CREATE INDEX idx_examples_function ON examples (function_source, function_name);
