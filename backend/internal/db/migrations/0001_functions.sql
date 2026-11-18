-- 0001: functions table. One Function = a test-script command, keyed by
-- (source, interface, name) per CONTEXT.md. ANSI-compatible SQL only
-- (ADR-0001): INTEGER PRIMARY KEY, TEXT timestamps written from Go.
CREATE TABLE functions (
    id INTEGER PRIMARY KEY,
    source TEXT NOT NULL,
    interface TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    convention INTEGER,
    category TEXT NOT NULL DEFAULT '',
    execution_role TEXT NOT NULL DEFAULT '',
    execution_phase TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (source, interface, name)
);

CREATE INDEX idx_functions_name ON functions (name);
CREATE INDEX idx_functions_interface ON functions (interface);
