# Task 4 Report: Reusable Inline-Create Registry Picker

## Implementation

- Added `frontend/src/lib/components/TroubleshootingRegistryPicker.svelte` using the installed Bits UI Popover and Combobox composition.
- Supports bindable registry keys, local label/key/description filtering, selected labels, exact-match duplicate prevention, inline creation, pending state, and inline errors without discarding failed text.
- Both category and severity modes are accessible by name. Severity creation accepts only the label, leaving tone selection to the backend default (`outline`).
- Trigger is 40px high; options and create action are at least 40px high.
- Reused existing Bits UI components; no dependencies or additional shadcn wrapper files were added.

## TDD evidence

RED: Added `frontend/tests/components/TroubleshootingRegistryPicker.svelte.spec.ts`, then ran:

`pnpm test:unit --run tests/components/TroubleshootingRegistryPicker.svelte.spec.ts`

The first test environment attempt found the missing Playwright browser. After installing the existing Playwright tool's Chromium browser, the recorded feature failure was the unresolved `TroubleshootingRegistryPicker.svelte` import.

GREEN: Implemented the component and reran the focused suite: 5/5 tests passed, covering matching selection/close, unmatched create exposure, successful creation/selection, failed-create term retention, and exact key/label duplicate suppression.

## Verification

- Full frontend browser/component unit suite: 10 files, 36 tests passed.
- `pnpm check`: 0 errors, 0 warnings.
- Prettier check on both new files: passed.
- `git diff --check`: passed.
- Official Svelte MCP `list-sections` and relevant Svelte 5 `$state`, `$derived`, `$effect`, `$props`, `$bindable`, markup, snippets, testing, TypeScript, and warning docs were consulted.
- Svelte MCP `svelte-autofixer`: 0 issues and 0 suggestions on the final component.
- `graphify update .`: completed successfully.

## Self-review

- Confirmed semantic tokens only, no manual overlay z-index, no `space-x/y-*`, and icon button conventions.
- Confirmed `onCreate` receives one trimmed label and the returned key becomes the selected bindable value.
- Confirmed error, pending, and duplicate-create guards do not clear typed text on failure.
- No scope expansion or wrapper additions.

## Review Fixes

- Changed `severityValues` so newly created severities default to `outline` while preserving explicit tones and update behavior.
- Added regression coverage in both the API registry test and the internal registry value unit test.
- Added a severity-mode browser component test that verifies the `Severity: Critical` and `Search severities` accessible names, no tone prompt, label-only `onCreate("Major")`, and selection of the returned severity carrying `tone: 'outline'`.

### Focused commands

- RED: `go test ./internal/api -run TestNewTroubleshootingSeverityDefaultsToOutlineTone -count=1`
  - `FAIL: new severity tone = "default", want outline`
- GREEN: same command
  - `ok webprogfndb/internal/api`
- `go test ./tests/api -run TestTroubleshootingRegistryValidationAndKeyGeneration -count=1`
  - `ok webprogfndb/tests/api`
- `pnpm test:unit --run tests/components/TroubleshootingRegistryPicker.svelte.spec.ts`
  - `1 passed`, `6 passed`, no failures
- `pnpm check`
  - `svelte-check found 0 errors and 0 warnings`
- Svelte autofixer remained applicable to the unchanged component from Task 4: `0 issues`, `0 suggestions`.
- `git diff --check`: passed (line-ending notices only).
