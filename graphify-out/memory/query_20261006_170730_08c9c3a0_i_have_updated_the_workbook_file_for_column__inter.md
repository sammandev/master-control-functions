---
type: "query"
date: "2026-10-06T17:07:30.804728+00:00"
question: "I have updated the workbook file for column ‘interface’ name to ‘device_name’ for ‘function_parameter’ sheet. Please hlep me to make the the ‘Import’, ‘Reset & import’, and ‘Export workbook’ buttons on the ‘Import’ page to implement this changes. For the ‘Devices’ page, on the ‘Create device’ modal form, please add ‘Type’ field that have 2 options: ‘BUILT-IN’ and ‘DLL’. The ‘Type’ data will be the same as ‘Function Source’. Currently all devices is using the same icon which is `CircuitBoard`, please differentiate the icon for type ‘BUILT-IN’ and ‘DLL’ devices and change the text ‘Device’ to their respective type (example: ‘BUILT-IN’ or ‘DLL’)."
contributor: "graphify"
outcome: "useful"
source_nodes: ["Export()", "Workbook"]
---

# Q: I have updated the workbook file for column ‘interface’ name to ‘device_name’ for ‘function_parameter’ sheet. Please hlep me to make the the ‘Import’, ‘Reset & import’, and ‘Export workbook’ buttons on the ‘Import’ page to implement this changes. For the ‘Devices’ page, on the ‘Create device’ modal form, please add ‘Type’ field that have 2 options: ‘BUILT-IN’ and ‘DLL’. The ‘Type’ data will be the same as ‘Function Source’. Currently all devices is using the same icon which is `CircuitBoard`, please differentiate the icon for type ‘BUILT-IN’ and ‘DLL’ devices and change the text ‘Device’ to their respective type (example: ‘BUILT-IN’ or ‘DLL’).

## Answer

Expanded via graph vocabulary: import export workbook device source parameter type. Export() and Workbook identified the shared import reader and canonical export headers. Reader already accepts device_name; added parameter-only legacy interface fallback by header presence, preserving canonical blanks. Device registry lacked source metadata; added nullable source migration0007, source-aware catalog inference and rename preservation, plus BUILT-IN/DLL UI choices and distinct icons. Parser and pure validation tests, Go build/vet, frontend check/build and 26 mocked Device UI tests passed. PostgreSQL integration tests compiled only; actual import/export/migration verification remains off-machine.

## Outcome

- Signal: useful

## Source Nodes

- Export()
- Workbook