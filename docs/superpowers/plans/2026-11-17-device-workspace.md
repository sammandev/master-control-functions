# Device Workspace Implementation Plan

> **For agentic workers:** Follow this plan task-by-task and verify each UI change with the frontend's API-mocked Playwright configuration.

**Goal:** Improve the shared shell and device pages while keeping navigation, API behavior, and scoped/Global data intact.

**Architecture:** Keep the existing SvelteKit routes and shadcn-svelte components. Organize the device detail content in a named workspace, make section tabs the primary navigation, expose per-section counts, and use responsive CSS for directory cards and list/detail panes.

**Tech Stack:** Svelte 5, SvelteKit, shadcn-svelte, Tailwind CSS, Playwright.

**Spec:** User-approved device-page reorganization in the conversation dated 2026-11-17.

## Global Constraints

- Do not change API contracts, mutations, URL tab history, or device/Global scoping.
- Preserve the shared full-width layout, sticky header, and footer.
- Use API-mocked frontend tests only; do not run database-backed tests or access deployment hosts.
- Keep existing uncommitted changes in `frontend/src/routes/+layout.svelte` and `frontend/src/routes/devices/+page.svelte`.

## Review Focus

- Desktop widths around 1104px must show four directory columns without card overflow; verify at 1104px and mobile.
- Tab count labels must retain accessible tab names and match device summary values.
- Sticky list and definition panes must remain below the app header and become non-sticky when stacked on narrow screens.
- Global card, URL-backed tab selection, mobile navigation, and footer must remain available.
- Dialogs, sort/search controls, and CRUD actions must continue to function with the existing mocked APIs.

---

### Task 1: Lock in UI behavior

**Files:**

- Modify: `frontend/tests/e2e/polish.spec.ts`

- [x] Add a mocked test for four columns at 1104px, the Global device card, shared footer, and sticky navigation.
- [x] Add a mocked device-detail test for count-bearing tabs, responsive list/detail layout, and tab URL behavior.
- [ ] Run only those named tests with `playwright.ui.config.ts` and confirm each fails on the current UI. Blocked because Playwright's Chromium executable is not installed; the shared browser was used for layout checks instead.

### Task 2: Reorganize the device workspace

**Files:**

- Modify: `frontend/src/routes/devices/[name]/+page.svelte`
- Modify: `frontend/src/routes/devices/+page.svelte`
- Modify: `frontend/src/routes/+layout.svelte`

- [x] Add the missing workspace wrapper; put section counts in the tabs while preserving tab labels and URL behavior.
- [x] Adjust responsive rail/table widths and sticky offsets; stack the detail pane at mobile widths.
- [x] Keep the existing Global card and CRUD/dialog behavior; ensure the directory has four columns at 1104px.
- [x] Remove the duplicate hidden locale-link block from the shared layout.
- [x] Run focused checks, `pnpm check`, scoped ESLint, and Svelte autofixer on changed components.

### Task 3: Verify the browser result

**Files:**

- No further files unless a regression is found.

- [x] Verify desktop/mobile layout, count tabs, Global card, shared footer, sticky header, and no horizontal overflow in the shared browser.
- [x] Record that a shared-browser dialog click could not be completed, the local Playwright runner lacks Chromium, and API requests return 502; no database-backed tests were run.
