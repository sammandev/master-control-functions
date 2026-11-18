# Spec: Workbook ecosystem expansion

Extend the script-definition knowledge workbook (`.dev-reference/Master_Control_V2_Script_Definition_Knowledge_merged.xlsx` — the merged file is the single canonical workbook as of 2026-10-02, superset of all earlier files; update it only) beyond the already-covered DLL device functions (WIFI_TEST, GOLDENKV, DMM, OSP, GPIB) with the BUILT-IN handler families of the test-execution engine, plus more Examples for under-covered Functions.

## Source of truth

The built-in dispatch is a single switch in the engine MasterControl hosts:

- `F:\!SVN\WIFI_TEST\trunk\General_DUT\Script.cs` → `cTestFlow.CreateCmd()` switch at line **1635** (`switch (getDevice().ToUpper() + "|" + getFunc().ToUpper())`, 168 case labels). MasterControl.exe hosts this as `General_DUT.dll` (`MasterControl\MasterContronForm.cs:249`).
- Script-line grammar: `DEVICE|FUNCTION|PARAM|CRITERIA|WAIT_TIME|PASS_PATH|FAIL_PATH|ERROR_CODE|DESCRIPTION`. Session/console families `@`-split PARAMS into `<device_name>@<payload or regex>`.
- Fallback prefixes (Script.cs:2281-2291): `DLL*` → `DLLControlCmd` (`General_DUT\DLLControl.cs:19`; `#PROXY` → pipe → `MasterControl\Proxy\ProxyControl.cs RunCMD`, `#QUICK` → in-proc), `MACRO*` → `cRunMacro` (`General_DUT\GeneralFunction.cs:5048`).
- MasterControl-local in-proc families: `AEROFLEX_PPM_16`, `ECHANNEL` (`MasterControl\Script.cs:130-152`).
- Live usage evidence: `MasterControl\bin\Release\Scripts\MAIN_FLOW.ini` (alias definitions) and `MACRO_LIST.ini` (macro sections).

## Conventions for new workbook rows

- `function_main.source` = `BUILT-IN` for all families in tickets 01-03 and 05; `DLL` for ticket 04 (QSPR/PTest DLLs). `interface` = the family device string exactly as dispatched (e.g. `TEST PROGRAM`, `SSH`, `CMD_CONSOLEV2`, `AEROFLEX_PPM_16`).
- Descriptions follow the existing verbose two-to-four sentence style; categories reuse the established vocabulary (`communication`, `utility`, `instrument_control`, `logging`, …).
- Examples must be plausible per the alias lines in `MAIN_FLOW.ini` (real device strings, `@`-split payloads) — never invented syntax.
- Function identity is `(source, interface, name)`; same-name functions on different interfaces are legitimate and must all be registered (e.g. OPEN on TELNET, RS232, CMD_CONSOLE, SSH).

## Tickets

| # | Ticket | Scope |
|---|---|---|
| 01 | Test Program family | ~75 functions, Script.cs:1638-1865 |
| 02 | Session & console families | TELNET / SSH / RS232 / CMD_SINGLE / CMD_CONSOLEV2 / CMD_CONSOLE |
| 03 | HTML, SFIS, IPLAS, FTP, SCRIPT, AEROFLEX, ECHANNEL | remaining named families |
| 04 | QSPR + PTest DLL functions | `DLLQSPR`, `DLLPTest` device functions |
| 05 | MACRO family + import/orphan gate | cRunMacro representation; re-import; orphan report must stay at zero; E2E green |

Ticket 06 in `web-program-functions-db` (orphan report) is the verification gate: after each content ticket, re-import the workbook and the new rows must introduce zero new orphans.
