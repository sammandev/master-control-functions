# 03: Conventions write-up

**What to build:** Write the conventions the previous six batches discovered into the project docs, so the remaining batches (and future maintainers) follow the same discipline without re-deriving it. Content: convention 1 (5-segment data operations) vs convention 2 (9-field flow-controlled) layouts and when each applies; the source taxonomy (comm/console/TEST PROGRAM families BUILT-IN, device drivers and instruments DLL); spec-style parameter patterns (multi sub-field segments as one loose `MULTIPLE` definition, no invented grammars without evidence); the evidence hierarchy (usage line > implementation source > inference); the append-script pattern (temp-file script, per-column style copying, bool-type matching); and the batch-6 correction rule that DLL-family example rows carry `DLL` as function source with the interface name in device scope only.

**Blocked by:** None (can start immediately — deliberately early so later batches follow it).

**Status:** done

- [x] Conventions document exists in the project docs covering: layout conventions 1 vs 2, source taxonomy, spec-style parameter patterns, evidence hierarchy, append-script pattern, DLL example-source rule
- [x] Document reflects what the committed batches actually did, citing representative examples from the registry
- [x] No contradiction with the domain vocabulary in CONTEXT.md or the ADRs; cross-references where terms overlap
- [x] Later-batch discipline (backup → pre-flight → append → verify → fixture-sync → test-cycle → commit) described as the standard procedure

## Completion notes (2026-10-02, commit 616fdde)

`docs/workbook-conventions.md` written. Facts verified against the registry before commit: convention-1 population is 41 functions (session plain ops + SFIS writes + a few TEST PROGRAM/Common calls — the initial draft's HEX_TO_DEC example was a misquote and replaced with the real `SSH|OPEN` 5-segment call); the MULTIPLE-spec definitions' patterns (`DAQ_MEASURE_SPEC`, `SPECTRUM_SEARCH_SPEC`, `CADB_ACCESS_SPEC`) quoted verbatim; batch-6 counts (289/199/53/542/184/174), ADR-0009 split (290/261/70/562/192/178), and batch-7 counts (296/270/71/572/193/185) match the pinned test history. Includes the batch-8 learnings: app orphan join semantics and the IOT-is-a-scope correction, so the audit ticket (05) reads them from docs rather than re-deriving.
