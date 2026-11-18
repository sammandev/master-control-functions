---
type: "query"
date: "2026-10-04T14:20:52.887857+00:00"
question: "please create proper test folder for the backend and frontend folder and move all test related files accordingly"
contributor: "graphify"
outcome: "useful"
source_nodes: ["FIXTURE", "Workbook", "CountsFrom()"]
---

# Q: please create proper test folder for the backend and frontend folder and move all test related files accordingly

## Answer

Expanded graph vocabulary: test fixture workbook. The traversal located FIXTURE at frontend/e2e/a11y.spec.ts L13, Workbook at backend/internal/importer/workbook.go L18, ParseWorkbook at workbook.go L98, CountsFrom at backend/internal/importer/oracle.go L23, and importer and export test nodes. Output was truncated to 30 of 82 nodes, so it does not prove complete test inventory. These are useful sources for the shared workbook boundary; current file paths and private-symbol dependencies must be checked directly before moving tests.

## Outcome

- Signal: useful

## Source Nodes

- FIXTURE
- Workbook
- CountsFrom()