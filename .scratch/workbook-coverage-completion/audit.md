# Coverage audit — trunk-usage axis (ticket 05, batch 9)

Date: 2026-10-02. Registry at audit time: **296 function rows** (254 distinct `(source, name)`), 270 parameter definitions, 71 criteria definitions, 572 parameter links, 193 criteria links, 219 examples.

## Method

- Full trunk walk (`F:\!SVN\WIFI_TEST\trunk`; `.ini`/`.cs`/`.txt`; `.vs`, `obj`, `packages`, `TempSpace` excluded): every quoted pipe-delimited call with ≥5 segments extracted (`KEY = "DEVICE|FUNC|PARAM|CRITERIA|…"`, including `#QUICK`/`#PROXY` mode suffixes).
- Resolution uses the **app's join semantics** (ADR-0002, `backend/internal/api/orphans.go`): function existence by `(function_source, function_name)` on any interface; a link's `device_name` is a device scope, not the interface. DLL device strings are checked with the prefix and mode suffix stripped (`DLLQSPR#QUICK` → `QSPR`).
- Prior-batch deferral decisions and scan-noise classes (from tickets 01/02/04 and `docs/workbook-conventions.md`) are applied before anything counts as undocumented.

## Result: no undocumented production trunk calls

Every `(device, function)` usage in the trunk resolves to a registry row, or falls into an explicitly recorded deferral/noise class. Breakdown of the 15 residual diff entries:

| Class | Count | Entries |
|---|---|---|
| Deferred — SDK sample DLL | 8 | `DLLDLLSAMPLE` / `DLLDLLSAMPLE#QUICK`: OPEN, WRITE, TIMES, TEST, RANDOM, PARSE_XML — `DLLSample` is the programming sample listed in `RD_SETTING.ini [Dll_Version]`; its call sites are sample flows (`bin\BK`, `bin\old`, `Scripts_`), not production content (ticket 01) |
| Deferred — macro label | 1 | `MACRO\|MARCOT7` — covered by the generic `MACRO\|RUN` decision (ecosystem-expansion ticket 05) |
| Deferred — variant spellings | 3 | `DLLQSPR`: `XML_CHECH_ITEM_PASS`, `HTML_CHECH_ITEM_PASS`, `PARSE_XML` — the canonical rows `XML_CHECK_ITEM_PASS`, `HTML_CHECK_ITEM_PASS`, `XML_PARSE` carry the variant spellings in their descriptions (ecosystem ticket 04) |
| Scan noise | 7 | `WL -I WL0/WL1 …` and `GET NAME\|QUALCOMM…` — C# source command arrays and device-enumeration strings in `DutControl\SSHcmd_*.cs` / `TELNETcmd_*.cs`, not script calls |

**Genuinely new (unclassified) missing names: 0.** The last audit candidate set was reclassified into the sample-DLL deferral bucket above.

## Example coverage (ticket 04 outcome)

154 of 254 distinct functions carry at least one verbatim example. The 100 functions without examples have no usage line in the trunk (full list in `.scratch/workbook-coverage-completion/issues/04-quality-pass.md`) — recorded as honest evidence-backed gaps per the evidence hierarchy, not silently dropped.

## Gate results (2026-10-02)

| Gate | Result |
|---|---|
| Import/reset counts through the API (`TestImportRealWorkbookAndIdempotency`, `TestImportResetClearsAndReimports`) | green — 296/270/71/572/193/219 |
| Import change log — only expected Normalization entries (`TestImportRealWorkbook…`, `normalizations` expectations in the suite) | green |
| Export round-trip cell-by-cell + re-import idempotency (`TestExportRoundTripCellByCell`, `TestExportReimportIdempotent`) | green |
| Orphan report — `0/0/0`, reconciles with ticket 02 (IOT = device scope, no gap) | green (`TestOrphanReportFromRealWorkbook`) |
| Backend suite (`go vet`, `go -C backend test ./...`) | green |
| Frontend typecheck + build (`pnpm check`, `pnpm build`) | green |
| E2E smoke suite against the real binary + real workbook fixture (`pnpm test:e2e`) | green |

Fixture copy hash-verified against the canonical workbook before the suite (`60CBBBDD518463751E3B8902EA9D92B84634AD00687E770D22834916A65AC2D3`).

## Verdict

**Trunk-usage coverage is complete.** Every script line an engineer reads in the trunk sources resolves to a documented Function in the workbook, except the entries explicitly deferred above with recorded reasons. Completion is asserted by this audit's diff, not by claim.
