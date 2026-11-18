# Workbook is a repeatable upsert source; SQLite is live truth between imports

The source xlsx is not a one-time seed. Import is an idempotent upsert by natural keys (Function: source+interface+name; definitions: name; links: their full natural key; examples: their natural key), repeatable any time the workbook changes, with import-time Normalization logged. Between imports, SQLite is the live source of truth: edits made in the web UI live in SQLite and are NOT written back to the workbook automatically — only a manual Export regenerates the workbook.

## Considered Options

- One-time seed, SQLite-only truth — rejected: the workbook remains the team's master artifact.
- Automatic two-way sync — rejected: conflict handling is out of scope for v1.

## Consequences

- A full round-trip Export (six sheets, original column layout) exists precisely so the workbook can be regenerated after UI edits.
- Re-importing a workbook over UI edits upserts by natural key: matching rows are updated from the workbook, non-matching rows survive. (Deliberate; revisit if the team wants SQLite-wins semantics.)
