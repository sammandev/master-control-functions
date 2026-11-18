# 05: Align Sequence Presentation

**Spec:** [Consistent Catalog Workspaces](../spec.md)

**What to build:** Bring the Sequence list and ordered-step editor into the shared visual language while retaining a dedicated, spacious authoring workflow and existing operations.

**Blocked by:** 02: Unify The Functions Workspace (shared app shell and visual conventions).

**Status:** complete

- [x] Sequence list, detail, and step-editing surfaces use the shared typography, borders, controls, action placement, focus, and theme conventions.
- [x] The ordered-step editor remains the primary editing surface. It is not forced into the Function browser or reduced to a generic record-detail pane.
- [x] Existing Sequence CRUD, step add/edit/delete/order operations, and Composer-to-Sequence entry points remain available wherever currently supported.
- [x] Step Function identity, values, timeout, flow, on-fail, Device, and note remain intact through edits. Preserve existing request semantics, including full replacement when required.
- [x] Emitted text remains server-generated from ordered steps; presentation changes do not alter Sequence semantics or workbook boundaries.
- [x] Save/cancel, validation, empty/loading/retry states, and destructive confirmations remain usable. Long step content and editor controls fit desktop and mobile without page-level overflow.
- [x] Mocked checks exercise existing supported Sequence and step operations, retained step fields/order, and Composer entry points. Inspect desktop/mobile and dark-theme presentation.
- [x] Focused checks and the complete database-free frontend closeout gates pass. Obtain Standards/Spec review and commit only this ticket's changes; leave unrelated edits intact.

## Verification Notes

- Reconciled on 2026-10-15 from existing verification records; no implementation changes or new build/test run.
- Commit: `fed1197` (`feat(frontend): align sequence presentation with catalog workspace`), containing the three ticket-specific frontend files. The implementation and commit requirements are satisfied.
- Original gates: svelte-check 0 errors and 0 warnings; Vitest 7/7; ESLint and Prettier clean on the changed files; `pnpm build` exit 0; frontend-only mocked suite 112 passed and 2 failed.
- Historical hold: the 2 failures were outside this ticket and also failed at ticket 04's commit: `app uses the supplied ICO for the tab and navigation brand` (`.brand img` never renders) and `prototype styling handles large catalogs, long content and dark mode` (30s timeout). This was a failed full-suite gate, not an accepted exception.
- Review: Standards and Spec sub-agent reviews. Fixed: an open step editor now follows its step through reorder and removal; malformed `name=value` lines are rejected beside the form; list and detail error states no longer read as an empty catalog. Left as judgement calls: the detail page's size, per-page copies of the shared CSS, and unlabelled per-step Edit/Remove buttons.
- Later regression evidence: [ticket 07 verification](07-import-presentation.md#verification), committed as `ef63f8e`, records 130/130 production-SPA mocked UI tests, 7/7 unit tests, zero Svelte/TypeScript errors or warnings, and passing scoped lint, formatting, build, and whitespace gates. That descendant run includes Sequence coverage and both previously failing tests, clearing the historical full-suite hold without claiming the original run passed or that the underlying causes were fixed.
- Completion reconciles the existing behavior checks, visual inspection, Standards/Spec review, committed implementation, and later successful database-free closeout. Ticket 04's separate post-fix review requirement does not block this ticket; its declared blocker is ticket 02.
- Evidence: implementation/review session `b5637cfa-3538-4d5c-9faf-460208c605da` and ticket 07's verification. Default database-backed E2E remains excluded; no database, Ubuntu host, or deployment verification is claimed.
