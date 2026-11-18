-- 0006: devices registry (ADR-0010). A registered Device is app-created
-- (reserved before it has any content) and survives workbook import/reset
-- like sequences (ADR-0005 precedent). Registry membership is by name —
-- derived devices remain distinct device_name values on links, definitions,
-- and examples, with no FKs anywhere (ADR-0002 rule). The table never
-- enters the six-sheet workbook (ADR-0004 export contract stays frozen).
-- The blank device (Global) is a derived concept only and can never be
-- registered, hence the CHECK.
CREATE TABLE devices (
    name TEXT NOT NULL UNIQUE CHECK (name <> ''),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
