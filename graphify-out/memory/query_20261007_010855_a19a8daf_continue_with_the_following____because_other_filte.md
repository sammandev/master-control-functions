---
type: "query"
date: "2026-10-07T01:08:55.332895+00:00"
question: "continue with the following: - because other filter fields don't have label, for good UI/UX, can we remove the label for the divice picker searchable fields? - can we also improve the UI/UX of all scrollbar/scroll area in the page to be better? for example we can make the scrollbar to be thinner, change the color based on the light/dark mode, and hide the scrollbar when user don't perform any scrolling. Just like how it is implemented on the \"Bits-UI - Scroll Area\"."
contributor: "graphify"
outcome: "useful"
source_nodes: ["DevicePicker.svelte", "bits-ui"]
---

# Q: continue with the following: - because other filter fields don't have label, for good UI/UX, can we remove the label for the divice picker searchable fields? - can we also improve the UI/UX of all scrollbar/scroll area in the page to be better? for example we can make the scrollbar to be thinner, change the color based on the light/dark mode, and hide the scrollbar when user don't perform any scrolling. Just like how it is implemented on the "Bits-UI - Scroll Area".

## Answer

Expanded from original query via graph vocab: device picker filter theme. The graph localized DevicePicker and Bits UI; it had no scrollbar-specific node. Removed the automatic visible picker caption while retaining accessible names and external form labels. Shared native scrollbars are thin, use theme tokens, reveal only on the scrolled element, and hide after 600ms of inactivity like Bits UI scroll mode. A lifecycle-managed passive capturing listener in the root layout uses SvelteMap timers; existing overflow, keyboard and touch behavior remain native. Shared CSS overrides Bits UI 2.19.4 combobox/select viewport scrollbar hiding without editing the dependency. High-contrast forced-colors mode retains automatic scrollbars. All 86 API-mocked UI tests, typecheck, scoped lint, formatting, Svelte analyzers, static build and whitespace pass. Two scrollbar tests also pass headed, and actual desktop/mobile light/dark active/idle thumb screenshots were inspected. Playwright headless Chromium passes --hide-scrollbars, so headed verification is required for native scrollbar screenshot evidence. Preview4204 is frontend-only; no database, host, deployment or commit operations.

## Outcome

- Signal: useful

## Source Nodes

- DevicePicker.svelte
- bits-ui