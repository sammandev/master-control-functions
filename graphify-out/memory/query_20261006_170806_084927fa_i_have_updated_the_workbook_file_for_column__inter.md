---
type: "query"
date: "2026-10-06T17:08:06.833453+00:00"
question: "I have updated the workbook file for column ‘interface’ name to ‘device_name’ for ‘function_parameter’ sheet. Please hlep me to make the the ‘Import’, ‘Reset & import’, and ‘Export workbook’ buttons on the ‘Import’ page to implement this changes."
contributor: "graphify"
outcome: "useful"
source_nodes: ["Export()", "Workbook"]
---

# Q: I have updated the workbook file for column ‘interface’ name to ‘device_name’ for ‘function_parameter’ sheet. Please hlep me to make the the ‘Import’, ‘Reset & import’, and ‘Export workbook’ buttons on the ‘Import’ page to implement this changes.

## Answer

Expanded via vocab: import export workbook device source parameter type. Shared ParseWorkbook supports device_name and now legacy interface when canonical header is absent. Blank canonical scope stays Global. Export already writes device_name. Parser tests pass; DB round trips compiled only.

## Outcome

- Signal: useful

## Source Nodes

- Export()
- Workbook