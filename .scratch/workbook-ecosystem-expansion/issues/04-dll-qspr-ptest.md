# 04: QSPR + PTest DLL functions

**Status:** done

Register the DLL device functions reachable through the `DLL*` prefix (`General_DUT\Script.cs:2283-2285` â†’ `DLLControlCmd`, which strips `DLL` and forwards `Func|Param|Criteria|WaitTime` verbatim):

- **`DLLQSPR` / `DLLQSPR#QUICK`**: functions seen in MAIN_FLOW.ini aliases â€” XML_PARSE, XML_TO_CSV, HTML_TO_CSV, HTML_PARSE, XML_CHECK_ITEM_PASS, HTML_CHECK_ITEM_PASS.
- **`DLLPTest` / `DLLPTest#PROXY`**: proxy-mode functions (e.g. Test1 seen in aliases); the `#PROXY` suffix routes through the named-pipe proxy (`MasterControl\Proxy\ProxyControl.cs` LOCK/UNLOCK/CMD), `#QUICK` loads the DLL in-process.

**Must do first:** locate the QSPR and PTest DLL source (candidates: `trunk\DLLStruct`, `trunk\DLLProxy`, `trunk\PredictiveEngine`, or a separate repo â€” the RD_SETTING.ini `[Dll_Version]` section lists QSPR/PTest as versioned components). If the source is not in this repo, register the functions with device-scoped pass-through semantics (params forwarded verbatim) and record that in each description.

**Blocked by:** none

**Acceptance:** every QSPR/PTest function used in MAIN_FLOW.ini has a function_main row + at least one Example derived from a real alias line; zero new orphans; `#PROXY`/`#QUICK` routing documented in the function descriptions.

## Comments

- Done 2026-10-02. Source hunt result: QSPR.dll and PTest.dll are BINARY ONLY in the whole SVN repo (trunk + NTGR/NS300/Predictive_PA_Trimming branches; no .csproj/.cs/.pdb anywhere). RD_SETTING.ini [Dll_Version]: QSPR=1.0.0.0, PTest=1.0.0.0, PTestFAIL=1.0.0.0. Per the ticket rule, rows use pass-through semantics documented from call sites only.
- Evidence base: DLLStruct\DLLAbstract.cs defines the DLL contract (Run(pipeIndex, "Func|Param|Criteria|WaitTime"), &-joined key=value params/criteria); every function registered cites MAIN_FLOW.ini call sites (Release + BK/bkM7 legacy flows).
- Added 10 function_main rows: QSPR XML_PARSE, HTML_PARSE, XML_TO_CSV, HTML_TO_CSV, XML_CHECK_ITEM_PASS, HTML_CHECK_ITEM_PASS, CHECH_CRITERIA, CHECH_ITEM_PASS, TO_CSV + PTest Test1. All descriptions note the pass-through semantics, the missing source, and the #PROXY/#QUICK routing.
- +9 parameter definitions from real call-site keys (QSPR_ROOT/REPORT_DIR/BAND/TEST_ITEM/ERR_ITEM/CSV_ITEM/FILTER/TO_CSV/DETAIL_CSV), +18 param links, +5 criteria links (CRITERIA_ARGUMENTS), +10 examples (all derived from real MAIN_FLOW.ini lines, incl. the active Release XML_CHECK_ITEM_PASS and PTest Test1 #PROXY lines).
- Legacy spellings (XML_CHECH_ITEM_PASS, HTML_CHECH_ITEM_PASS, PARSE_XML) intentionally NOT registered as separate rows: the app's Normalization rules treat typo spellings as import-time fixes; the legacy names are documented inside the corresponding descriptions instead.
- PTestFAIL.dll (deliberate-fail variant, [Dll_Version] PTestFAIL=1.0.0.0) has no call sites in any MAIN_FLOW.ini -> no row; noted here for traceability.
- Gate via app (POST /api/import/reset of merged file): functions=259, parameter_definitions=175, parameter_links=504, criteria_definitions=52, criteria_links=179, examples=143; orphan total 0. Normalization log = the 4 pre-existing fixes + 3 new expected PTest->PTEST device-scope uppercases (same documented device-scope-uppercase rule as GoldenKV->GOLDENKV).
- Option for future enrichment: decompile QSPR.dll/PTest.dll with ILSpy to extract the Execute() dispatch and replace pass-through descriptions with exact semantics (explore agent's recommendation).
