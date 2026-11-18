# 02: Session & console families

**Status:** done

Add the missing functions of the session/console built-in families â€” `General_DUT\Script.cs` line ranges noted per family. The workbook already has OPEN/CLOSE/SENDSTR/WAITSTR/SENDSTR_AND_WAITSTR rows for TELNET, RS232, CMD_CONSOLE, SSH; complete each family and register the missing ones.

- **TELNET** (1962-2022): add SENDSTR_AND_RECV_ONE_REG, WAIT_RESP_AND_SEND, RECV_LENGTH_STR, RECV_MULTI_REG (SENDSTR/CLOSE/RECV_ONE_REG/OPEN/WAITSTR exist).
- **SSH** (2029-2079): add SENDSTR_AND_RECV_ONE_REG, RECV_LENGTH_STR, RECV_MULTI_REG, RECV_ONE_REG, CLOSE (others exist).
- **RS232** (1902-1951): add RECV_LENGTH_STR, RECV_ONE_REG, RECV_MULTI_REG.
- **CMD_SINGLE** (2084-2095): new family â€” SENDSTR, SENDSTR_AND_WAITSTR, SENDSTR_AND_RECV_ONE_REG (single-shot, PARAM is NOT `@`-split).
- **CMD_CONSOLEV2** (2099-2138): new family â€” OPEN, CLOSE, SENDSTR, WAITSTR, SENDSTR_AND_WAITSTR, SENDSTR_AND_RECV_ONE_REG, RECV_LENGTH_STR, RECV_ONE_REG, RECV_MULTI_REG.
- **CMD_CONSOLE** (2142-2189): add CTRLC (hardcoded `Break(0)` payload), SENDSTR_AND_RECV_ONE_REG, RECV_LENGTH_STR, RECV_MULTI_REG, RECV_ONE_REG (OPEN/CLOSE/SENDSTR/WAITSTR/SENDSTR_AND_WAITSTR exist).

**Conventions:**
- PARAMS for these families is `@`-split: `<device_name>@<payload or regex>` â€” parameter definitions must reflect P[0] (session device name) and P[1] (payload/expected regex). Reuse existing criteria (EXPECTED_RESPONSE, REGEX_CAPTURE, VARIABLEâ€¦) before inventing new ones.
- One `function_main` row per (interface, name); the existing TELNET/RS232/CMD_CONSOLE/SSH OPEN rows are the style reference.

**Blocked by:** none

**Acceptance:** zero new orphans after import; no duplicate (source, interface, name); examples use real `DUT@...` / `CMD@...` payload shapes from MAIN_FLOW.ini aliases.

## Comments

- Done 2026-10-02. First-hand from General_DUT\Script.cs:1902-2189. Applied to the NEW canonical workbook `Master_Control_V2_Script_Definition_Knowledge_merged.xlsx` (user merged all files; superset of the old canonical, 179→208 functions).
- Added 29 function_main rows: RS232 CLOSE/RECV_LENGTH_STR/RECV_ONE_REG/RECV_MULTI_REG; TELNET WAIT_RESP_AND_SEND/RECV_LENGTH_STR/RECV_MULTI_REG; SSH CLOSE/SENDSTR_AND_RECV_ONE_REG/RECV_LENGTH_STR/RECV_ONE_REG/RECV_MULTI_REG; CMD_SINGLE SENDSTR/SENDSTR_AND_WAITSTR/SENDSTR_AND_RECV_ONE_REG; CMD_CONSOLEV2 all 9; CMD_CONSOLE CTRLC/RECV_LENGTH_STR/SENDSTR_AND_RECV_ONE_REG/RECV_ONE_REG/RECV_MULTI_REG.
- +3 parameter definitions (SESSION_LENGTH, CONSOLE_WINDOW, CONSOLE_LAUNCH_COMMAND), +45 parameter links (mirroring existing TELNET/SSH conventions: SESSION_COMMAND / SESSION_EXPECTED_RESPONSE / SESSION_REGEX_CAPTURE + VARIABLE / REGEX_CAPTURE), +9 criteria links (EXPECTED_RESPONSE, only where the switch wires getCriteria), +14 examples.
- Name-vs-switch check: missing=[] for all 6 families. RS232 extra SENDSTR_AND_WAITSTR_NOLINE is pre-existing and NOT in the main switch (handled elsewhere) - left as-is.
- Repair carried over from the old canonical: deleted the pre-existing malformed function_parameter row WAVEFORM_NAME->WIFI_TEST (survived the user merge).
- Gate via app (POST /api/import/reset of the merged file): functions=208, parameter_definitions=161, parameter_links=443, criteria_definitions=51, criteria_links=161, examples=123; orphan total 0; normalization log = the 4 pre-existing expected fixes only.
