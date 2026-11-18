# 03: HTML, SFIS, IPLAS, FTP, SCRIPT, AEROFLEX, ECHANNEL

**Status:** done

Register the remaining named built-in families â€” all cases in `General_DUT\Script.cs` unless noted:

- **HTML** (1874-1894): 21 functions (URL, WAITURL, READ, WRITE, JAVASCRIPT, BUTTONCLICK, CLICK, RADIOBUTTON, CHECKBOX, COMBOBOX, XPATH, XPATHANDCOMPARE, XPATHANDCOMPARE_DYNAMIC, XPATHANDRECEIVE, READINI, POPUP, READANDCOMPARE, CLEARSESSION, PF_URL, PF_LOGIN, REFRESH).
- **SFIS** (2200-2242): 8 functions (WRITE_DATA_TO_SFIS, TSP_GETIMAC, TSP_GETIMACNEW, TSP_GETVERSION, TSP_GETVERSION_ZYXEL, TSP_CHECKROUTE, TSP_REPAIR, TSP_SSD_INPUTDATA); note the login guard (`Global.sbSFISHasLogined` â†’ SFIS_CAN_NOT_LOGIN) in descriptions.
- **IPLAS** (2255-2263): add WRITE_TIME + links for WRITE_DATA_TO_CSV / WRITE_DATA (WRITE_DATA exists in function_main as IPLAS BUILT-IN).
- **FTP** (2267-2272): 6 functions (COPY_TEST_RESULTS_ALL, COPY_TEST_RESULTS, COPY_TEST_RESULTS_FOLDER, COPY_PATHLOSS, COPY_FILE, MARS_EAP).
- **SCRIPT** (1824-1834): MACRO_CMD â€” runs a MACRO_LIST.ini section of full script lines; parameter = section name.
- **AEROFLEX_PPM_16** (MasterControl\Script.cs:134-140, impl `MasterControl\AEROFLEX_PPM_16.cs`): CONNECTION, MEASURE_SETTING, MEASURE_PULSE_POWER, MEASURE_AVERAGE_POWER (last two take criteria).
- **ECHANNEL** (MasterControl\RF_Switch_eChannel.cs): RF_SWITCH_ETHERNET.

**Blocked by:** none

**Acceptance:** zero new orphans after import; descriptions note the login guard and web-browser context where the switch implies them; examples only where usage is statically clear (IPLAS WRITE_DATA_TO_CSV and SFIS TSP_* have real shapes in MAIN_FLOW.ini aliases).

## Comments

- Done 2026-10-02. First-hand from General_DUT\Script.cs (1824-1834, 1874-1894, 2200-2272) and MasterControl\Script.cs (134-146). Applied to the merged canonical workbook.
- Added 41 function_main rows: SCRIPT MACRO_CMD; HTML all 21; SFIS TSP_GETIMAC/GETIMACNEW/GETVERSION/GETVERSION_ZYXEL/REPAIR/SSD_INPUTDATA (all descriptions carry the Global.sbSFISHasLogined -> SFIS_CAN_NOT_LOGIN guard); IPLAS WRITE_DATA_TO_CSV/WRITE_TIME; FTP all 6; AEROFLEX_PPM_16 all 4; ECHANNEL RF_SWITCH_ETHERNET.
- +5 parameter definitions (PAGE_URL, HTML_ELEMENT, ELEMENT_VALUE, MACRO_SECTION, SFIS_DATA), +43 parameter links (IPLAS rows mirror the existing PROJECT/TSP convention with CSV_REC/CSV_FILE_NAME), +1 criteria definition (EXPECTED_VALUE for HTML/SFIS content comparisons), +13 criteria links (HTML compare items, SFIS version/SSD, IPLAS/AEROFLEX via CRITERIA_ARGUMENTS), +10 examples.
- Name-vs-switch check: missing=[] and extra=[] for all 7 families.
- SFIS_TO_VAR: verified ABSENT from both the General_DUT and MasterControl dispatch switches - the ticket's retarget idea was dropped (any retarget would be a guess). The row keeps interface TEST PROGRAM and is flagged here as unresolvable in the inspected engine (likely handled by a device DLL or stale).
- Gate via app (POST /api/import/reset of merged file): functions=249, parameter_definitions=166, parameter_links=486, criteria_definitions=52, criteria_links=174, examples=133; orphan total 0; normalization log = the 4 pre-existing expected fixes only.
