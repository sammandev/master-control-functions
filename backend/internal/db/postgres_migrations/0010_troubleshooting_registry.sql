-- 0010: Managed troubleshooting category and severity registries.

CREATE TABLE troubleshooting_categories (
    key TEXT PRIMARY KEY CHECK (btrim(key) <> '' AND length(key) <= 64),
    label TEXT NOT NULL CHECK (btrim(label) <> '' AND length(label) <= 100),
    description TEXT NOT NULL CHECK (btrim(description) <> '' AND length(description) <= 1000),
    sort_order BIGINT NOT NULL CHECK (sort_order > 0),
    built_in BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE UNIQUE INDEX idx_troubleshooting_categories_label_ci
    ON troubleshooting_categories (lower(label));

CREATE TABLE troubleshooting_severities (
    key TEXT PRIMARY KEY CHECK (btrim(key) <> '' AND length(key) <= 64),
    label TEXT NOT NULL CHECK (btrim(label) <> '' AND length(label) <= 100),
    description TEXT NOT NULL CHECK (btrim(description) <> '' AND length(description) <= 1000),
    sort_order BIGINT NOT NULL CHECK (sort_order > 0),
    built_in BOOLEAN NOT NULL DEFAULT FALSE,
    tone TEXT NOT NULL CHECK (tone IN ('destructive', 'warn', 'default', 'outline'))
);
CREATE UNIQUE INDEX idx_troubleshooting_severities_label_ci
    ON troubleshooting_severities (lower(label));

INSERT INTO troubleshooting_categories (key, label, description, sort_order, built_in) VALUES
    ('connection', 'DUT Connection & Setup', 'Issues establishing and maintaining DUT communication.', 1, TRUE),
    ('execution', 'Script & Flow Execution', 'Issues in script parsing, flow control, and runtime execution.', 2, TRUE),
    ('measurement', 'Measurement & Criteria', 'Issues with measurements, limits, criteria, and test results.', 3, TRUE),
    ('production', 'Production Run & Handoff', 'Issues with production runs, cleanup, and handoff.', 4, TRUE),
    ('fixture-hardware', 'Fixture & Hardware', 'Issues with fixtures, cabling, connectors, and station hardware.', 5, TRUE),
    ('instrument-calibration', 'Instrument & Calibration', 'Issues with instruments, calibration, and compensation.', 6, TRUE),
    ('dut-firmware-image', 'DUT Firmware & Image', 'Issues specific to DUT firmware and image behavior.', 7, TRUE),
    ('network-shopfloor-mes', 'Network / Shopfloor / MES', 'Issues with network infrastructure, shopfloor systems, and MES.', 8, TRUE),
    ('ini-spec-configuration', 'INI & Spec Configuration', 'Issues with INI files, specifications, and configuration values.', 9, TRUE),
    ('station-environment', 'Station Environment', 'Issues caused by station environment, power, timing, or external state.', 10, TRUE),
    ('logging-reporting', 'Logging & Reporting', 'Issues with logs, evidence collection, reports, and diagnostics.', 11, TRUE);

INSERT INTO troubleshooting_severities (key, label, description, sort_order, built_in, tone) VALUES
    ('critical', 'Critical (Blocks Flow)', 'Blocks the test flow and requires immediate intervention.', 1, TRUE, 'destructive'),
    ('high', 'High', 'Major impact requiring prompt investigation.', 2, TRUE, 'destructive'),
    ('warning', 'Warning (Intermittent / Margin)', 'Intermittent failure or insufficient operating margin.', 3, TRUE, 'warn'),
    ('low', 'Low', 'Limited impact that can be addressed during normal maintenance.', 4, TRUE, 'outline'),
    ('info', 'Info (Best Practice)', 'Informational guidance or recommended best practice.', 5, TRUE, 'default');

ALTER TABLE troubleshooting_guides
    DROP CONSTRAINT troubleshooting_guides_category_check,
    DROP CONSTRAINT troubleshooting_guides_severity_check;

ALTER TABLE troubleshooting_guides
    ADD CONSTRAINT troubleshooting_guides_category_fkey
        FOREIGN KEY (category) REFERENCES troubleshooting_categories (key),
    ADD CONSTRAINT troubleshooting_guides_severity_fkey
        FOREIGN KEY (severity) REFERENCES troubleshooting_severities (key);
