-- 0007: Troubleshooting Playbook & AI Knowledge Base (SQLite legacy parity).

CREATE TABLE troubleshooting_guides (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('connection', 'execution', 'measurement', 'production')),
    severity TEXT NOT NULL CHECK (severity IN ('critical', 'warning', 'info')),
    device_name TEXT NOT NULL DEFAULT '',
    function_name TEXT NOT NULL DEFAULT '',
    error_code TEXT NOT NULL DEFAULT '',
    symptom TEXT NOT NULL DEFAULT '',
    root_cause TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX idx_troubleshooting_guides_category ON troubleshooting_guides (category);
CREATE INDEX idx_troubleshooting_guides_device ON troubleshooting_guides (device_name);
CREATE INDEX idx_troubleshooting_guides_error_code ON troubleshooting_guides (error_code);
CREATE INDEX idx_troubleshooting_guides_function ON troubleshooting_guides (function_name);

CREATE TABLE troubleshooting_steps (
    id INTEGER PRIMARY KEY,
    guide_id INTEGER NOT NULL REFERENCES troubleshooting_guides (id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    title TEXT NOT NULL,
    instruction TEXT NOT NULL,
    script_snippet TEXT NOT NULL DEFAULT '',
    verification TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_troubleshooting_steps_guide ON troubleshooting_steps (guide_id, position);

CREATE VIEW mcp_v1_troubleshooting_guides AS
SELECT g.id, g.title, g.category, g.severity, g.device_name, g.function_name,
       g.error_code, g.symptom, g.root_cause, g.created_at, g.updated_at,
       (SELECT COUNT(*) FROM troubleshooting_steps s WHERE s.guide_id = g.id) AS step_count
FROM troubleshooting_guides g;

CREATE VIEW mcp_v1_troubleshooting_steps AS
SELECT s.id, s.guide_id, s.position, s.title, s.instruction, s.script_snippet, s.verification
FROM troubleshooting_steps s
ORDER BY s.guide_id, s.position, s.id;
