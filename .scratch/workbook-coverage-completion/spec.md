# Spec: Script-definition coverage completion (trunk-usage axis)

**Labels:** done

## Problem Statement

The team maintains a script-definition knowledge workbook (`Master_Control_V2_Script_Definition_Knowledge_merged.xlsx` in `.dev-reference/`) that documents every test-script function the V2 test programs can call, with parameters, criteria, and usage examples. The workbook is the repeatable upsert source for the web app's SQLite database (per ADR-0004), and orphan link rows are kept and flagged (per ADR-0003).

A trunk-wide scan of the SVN sources (`MasterControl` scripts + `General_DUT` STB specs) shows that until recently, whole families of script functions that production test flows call every day — SSH sessions, the DUT console, GPIB instruments, DAQ audio measurement, video measurement — had no rows at all. An engineer reading a script line like `DUT|SEND_AND_WAIT_RESP|ver\x0D|*I:Version-Diag#|10000|BL|FAIL|BSFR01|Check diag version` could not look up what it does, which parameters it takes, or what the criteria segment means. Most of this gap has been closed by six committed batches, but the tail of the work (low-usage stragglers, the undefined IOT interface, and the documentation of the conventions learned along the way) is still open.

## Solution

Complete the trunk-usage coverage axis the same way the previous batches did: scan the SVN trunk for pipe-delimited script calls, diff them against the workbook's function registry, extract evidence (usage lines + implementation source where it exists), and append rows to the canonical workbook through the established backup → pre-flight → append → verify → fixture-sync → test-cycle → commit procedure. Additionally, write down the conventions the batches discovered (convention 1 vs 2 layout, source taxonomy, spec-style parameter patterns) so future maintainers do not have to reverse-engineer them from the data. Close with a quality pass over examples and criteria coverage.

## User Stories

1. As a workbook maintainer, I want every function call that appears in production trunk scripts to exist in the workbook, so that no script line is undocumented.
2. As a workbook maintainer, I want each documented function to carry at least one real example line harvested from trunk scripts, so that I can see the exact argument layout in context.
3. As a workbook maintainer, I want parameter definitions reused across functions when their semantics match, so that the parameter dictionary stays lean and consistent.
4. As a workbook maintainer, I want new parameter definitions to carry a `value_pattern` derived from real usage, so that pattern documentation does not drift from reality.
5. As a workbook maintainer, I want pre-flight assertions before every append (no duplicate natural keys, reused definitions exist, no pre-existing rows for target functions), so that an append can never silently corrupt the workbook.
6. As a workbook maintainer, I want a dated backup of the workbook before each batch, so that any bad append can be rolled back.
7. As a workbook maintainer, I want the append to be verified as additions-only (+N −0) against the backup, so that no one's existing rows are lost.
8. As a workbook maintainer, I want the 18 pre-existing `DLL|IOT` orphan links either resolved with real IOT function definitions or explicitly accepted as a known gap, so that the orphan report is trustworthy.
9. As a test engineer, I want the DUT console family (OPEN_PORT, CLOSE_PORT, SENDSTR, SEND_AND_WAIT_RESP, WAITSTR, RECEIVE_RSP) documented from the actual console-control implementation, so that my descriptions of send/wait/receive semantics are grounded in code rather than guesswork.
10. As a test engineer, I want instrument families (GPIB_DEVICE, DAQ, VIDEOMASTER, SPECTRUM, CISCO, RS_SFE100, NI6008) documented as DLL-sourced device drivers, so that the source taxonomy reflects how the functions are really implemented.
11. As a test engineer, I want `EXPECTED_RESPONSE` criteria reused for every function whose second segment is a response expectation, so that response-verification semantics are recorded uniformly.
12. As a backend developer, I want the importer test fixture re-synced from the canonical workbook after every batch, so that the repo's tests exercise the real current state.
13. As a backend developer, I want the pinned import counts in the four expectation files updated with each fixture sync, so that the suite fails loudly if the workbook and repo drift apart.
14. As a backend developer, I want the fixture copy hash-verified against the canonical workbook before running the suite, so that an external revert of the fixture cannot silently produce a false test result.
15. As a backend developer, I want the batch commits to touch only the fixture and the four test files, so that unrelated working-tree changes (README, frontend, later ADRs) stay out of the workbook history.
16. As a frontend developer, I want the imported database to contain consistent conventions (convention 1 = 5-segment data operations, convention 2 = 9-field flow-controlled operations), so that UI grouping by convention is meaningful.
17. As a frontend developer, I want examples to reference the exact device scope they came from, so that example browsing by device works.
18. As a QA engineer, I want the import gate run through the real app (import/reset + export) after every batch, so that counts in the UI reflect the workbook exactly.
19. As a QA engineer, I want the orphan report to distinguish pre-existing known orphans from batch-introduced ones, so that regressions are attributable.
20. As a QA engineer, I want a final coverage audit listing every trunk function name that still has no workbook row, so that completion is provable, not asserted.
21. As a new maintainer, I want the conventions (source taxonomy, convention layouts, spec-pattern parameter style, evidence hierarchy: usage line > implementation source > inference) written into project docs, so that batch 8 work follows the same discipline without re-deriving it.
22. As a new maintainer, I want the append-script pattern (temp-file Python, per-column style copying, bool-type matching) recorded, so that future batches do not mangle formatting through inline shell commands.
23. As a maintainer, I want sparse name-only parameter rows left alone unless evidence clarifies them, so that speculation does not replace honest gaps.
24. As an agent picking this up, I want the remaining work split into small independent batches (stragglers, IOT, quality pass, docs), so that each can be implemented, verified, and committed in one sitting.

## Implementation Decisions

- The canonical workbook is the merged master in `.dev-reference/`; it is gitignored, so only the importer test-fixture copy and expectation files are committed to the repo.
- Six batches are already committed (function families: SSH session, TEST PROGRAM data ops, TEST PROGRAM stragglers + NI6008, DUT console, GPIB_DEVICE + DAQ, VIDEOMASTER + SPECTRUM + CISCO + RS_SFE100), bringing the registry to 289 functions / 199 parameter definitions / 53 criteria / 542 parameter links / 184 criteria links / 174 examples.
- Source taxonomy: comm/console/TEST PROGRAM families are `BUILT-IN`; device drivers and instruments (NI6008, GPIB_DEVICE, DAQ, VIDEOMASTER, SPECTRUM, CISCO, RS_SFE100) are `DLL`.
- Convention 2 (9-field) is the canonical layout for all flow-controlled families; convention 1 (5-segment) only for plain MAIN_FLOW data operations.
- Parameter specs that pack multiple sub-fields into one segment use a single definition with a loose `<a>,<b>,...` pattern and MULTIPLE cardinality instead of inventing grammars without evidence (e.g. DAQ_MEASURE_SPEC, SPECTRUM_SEARCH_SPEC, RS_SFE100_SETTING).
- Cross-family reuse is preferred: EXPECTED_RESPONSE, VARIABLE, GPIB_COMMAND, EXPECTED_VALUE, GET_ABS_MAX_MIN_SOURCES/TARGETS were reused instead of duplicated.
- Examples are harvested verbatim from trunk `.ini` files (deduplicated by shape); the "meaning" column paraphrases the flow targets and beeps from the same line.
- The sync procedure after every batch: copy merged workbook → importer fixture, probe counts through a temporary test, update the pinned counts in the four expectation files, run the full backend suite, commit only those files.
- The remaining batches: (7) low-usage stragglers from the trunk diff; (8) either document the IOT interface functions with evidence or record the 18 orphans as an accepted gap; (9) a coverage audit that re-runs the trunk scan and produces the final missing list; (10) conventions write-up in the project docs.
- Batch 6 produced one correction worth recording: example rows for DLL families must carry `DLL` as function_source — the interface name goes in device_name only; the batch-6 in-place repair is the reference handling.

## Testing Decisions

- Tests assert external behavior only: the pinned workbook counts and the import gate response, never internal parsing details.
- The testing seam is the existing importer fixture + expectation files plus the API import/export endpoints; no new seams are introduced — six batches have proven this seam sufficient.
- Good tests fail when the workbook changes and stay green when nothing changed; the count-pinning style in the importer and API test files is the prior art to follow.
- Every batch cycle runs `go vet` and the full backend test suite once at the end; single-package probe tests may run during development and are deleted before commit.
- The workbook itself is verified outside Go: additions-only diff vs backup, duplicate natural-key scan, join-integrity check, and orphan-count comparison against the pre-batch baseline.

## Out of Scope

- The user's unrelated working-tree changes (README, frontend packaging, ADR-0007/0008) — not part of this work, must not be committed.
- Defining IOT interface functions without trunk evidence — either find the evidence or record the gap, do not invent.
- Over-specifying opaque single-string parameters with more than one honest generic definition.
- Any frontend or backend feature work beyond the workbook ecosystem.
- Modifying the SQLite database directly; the workbook is the only write path (ADR-0004).

## Further Notes

- The companion tracker `.scratch/workbook-ecosystem-expansion/` (tickets 01–05, all done) covers the Script.cs switch-case axis; this spec covers the trunk-usage axis. Both feed the same canonical workbook.
- The workbook canonical file lives on the user's Desktop project folder, not in the repo; the fixture under the importer testdata directory is the only tracked copy.
- Trunk scan candidates use the match set {('BUILT-IN', src, name), ('DLL', src, name), ('DLL', src without 'DLL' prefix, name)}; the STB_spec scripts in General_DUT bin folders are the richest evidence source for DUT and instrument families.
