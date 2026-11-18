-- 0004: sequences + sequence_steps (ticket 10, ADR-0005). A Sequence is a
-- web-native, named, ordered set of Steps; it never enters the six-sheet
-- workbook (ADR-0004's export contract stays frozen). Steps reference
-- Functions by natural key and carry no FK to functions — orphan steps are
-- storable and flagged, never dropped (ADR-0003). The steps→sequences FK
-- is aggregate-internal (same pattern as import_runs in 0003): deleting a
-- sequence deletes its steps.
CREATE TABLE sequences (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE sequence_steps (
    id INTEGER PRIMARY KEY,
    sequence_id INTEGER NOT NULL REFERENCES sequences (id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    function_source TEXT NOT NULL,
    function_interface TEXT NOT NULL,
    function_name TEXT NOT NULL,
    parameters TEXT NOT NULL DEFAULT '{}',
    criteria TEXT NOT NULL DEFAULT '{}',
    timeout_ms TEXT NOT NULL DEFAULT '',
    flow TEXT NOT NULL DEFAULT '',
    on_fail TEXT NOT NULL DEFAULT '',
    device_name TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_sequence_steps_sequence ON sequence_steps (sequence_id, position);
