---
type: "query"
date: "2026-10-07T00:29:45.881140+00:00"
question: "Take a look at this image: C:\\Users\\Samuel_Halomoan\\AppData\\Roaming\\Code\\User\\workspaceStorage\\vscode-chat-images\\image-1791331619244.png, the searchable filter dropdown option list field seems to have different size than the fields that don't have searchable function (example like filter fields on the \"Functions\", \"Parameters\", \"Criteria\" pages and so on). Please help to that field."
contributor: "graphify"
outcome: "useful"
source_nodes: ["DevicePicker.svelte", "select.svelte", "DictionaryPage.svelte"]
---

# Q: Take a look at this image: C:\Users\Samuel_Halomoan\AppData\Roaming\Code\User\workspaceStorage\vscode-chat-images\image-1791331619244.png, the searchable filter dropdown option list field seems to have different size than the fields that don't have searchable function (example like filter fields on the "Functions", "Parameters", "Criteria" pages and so on). Please help to that field.

## Answer

Expanded from original query via graph vocab: device select dictionary picker combobox filter. DevicePicker used a 40px minimum height, 280px width and a separate caption, unlike the 32px Select trigger. Added opt-in compact sizing for Functions and shared Parameters/Criteria filters: 32px high, 144px wide, matching padding, font, chevron and focus styles. Long selections truncate with a full-value title. Default form and header pickers and popup search remain unchanged. Verified 6 new desktop/mobile light/dark sizing and search checks plus 8 existing mocked picker/dropdown workflows, typecheck, scoped lint, Svelte analyzer, production build, whitespace and screenshots. No database, deployment or host operations.

## Outcome

- Signal: useful

## Source Nodes

- DevicePicker.svelte
- select.svelte
- DictionaryPage.svelte