# 07: Align The Import Workflow

**Spec:** [Consistent Catalog Workspaces](../spec.md)

**What to build:** Apply the shared visual language to workbook upload, confirmation, import results, and run history. Keep the existing operational workflow and reset safeguards unchanged.

**Blocked by:** 02: Unify The Functions Workspace (shared app shell and visual conventions).

**Status:** complete

- [x] Upload, options/confirmation, results, and history are clearly structured with shared typography, controls, borders, and action conventions.
- [x] Existing workbook selection and upload request semantics remain unchanged. Running, success, and failure states are visible and prevent accidental duplicate submissions as currently supported.
- [x] Optional reset remains explicitly confirmed before the existing destructive request. Cancelling confirmation sends no reset/import request.
- [x] Actual API-reported results, prior run history, and collapsible normalization details remain available and readable; no import/business-rule changes are introduced.
- [x] Keyboard navigation, focus after confirmation, loading/retry/empty states, and mobile layout remain usable. Wide history or normalization details scroll locally.
- [x] Mocked checks cover upload without reset, reset confirmation/cancellation, pending/failure/success, history, and normalization details. Use intercepted responses only; no database-backed import is executed.
- [x] Inspect desktop/mobile and dark-theme presentation, including long workbook names and normalization records.
- [x] Focused checks and the complete database-free frontend closeout gates pass. Obtain Standards/Spec review and commit only this ticket's changes; leave unrelated edits intact.

## Verification

- User confirmed API-mocked browser boundary and review baseline `04d42a861c7076fb4feec928977fd2184fc03083`.
- Six Import checks cover multipart payload, reset confirmation/cancellation and focus, pending controls, failed upload retry, actual results, null arrays, history/detail retries, and local keyboard scrolling.
- Complete production-SPA mocked UI suite: 130 passed. Unit suite: 7 passed.
- Svelte/TypeScript: zero errors and warnings. Scoped ESLint, Prettier, static build, and whitespace gates passed. Svelte analyzer: no issues or suggestions.
- Desktop 1440px and mobile 390px screenshots inspected in light and dark themes with long names and normalization records.
- Independent Standards review: no violations. Independent Spec review: no findings.
- Default database-backed E2E suite excluded. No database, Ubuntu host, or deployment operations performed.
- This tracker remains local and ignored; only the Import view and mocked test file are committed.
