# Sequences are web-native entities, outside the workbook

The composer emits single-Function script strings (ticket 07); the multi-function sequence mode was deferred (story 37). We decided a Sequence is a first-class stored entity in the app (sequences + sequence_steps tables): ordered Steps referencing Functions by natural key (source, interface, name), emitted as newline-joined per-step script strings. Every step is re-validated against the current definitions at emission; unresolved or now-invalid steps are flagged per-step while clean steps still emit — the same keep-and-flag spirit as ADR-0003, applied to staleness instead of just absence. Sequences never enter the workbook: the six-sheet import/export contract (ADR-0004) stays frozen, and the workbook has no sequence representation to import.

## Considered options

- Generator-only (no persistence): rejected — sequences must accumulate and be re-editable, which is the point of story 37.
- Seventh workbook sheet: rejected — changes the export contract, the cell-level round-trip test, and every consumer of the file.

## Consequences

- Import/export and the round-trip determinism test are untouched.
- Sequences survive "reset database from workbook": steps reference Functions by natural key and re-import recreates the same keys; a step whose key no longer resolves becomes a flagged orphan, not lost data.
- Emitted sequence text is always generated from steps and never parsed back (the pipe grammar is lossy for orphan and multi-value fields).
- Steps carry no database-level foreign key to Functions (orphan-able, mirror link tables); deleting a Function leaves referencing steps in place, flagged as orphans — the same treatment link rows get.
