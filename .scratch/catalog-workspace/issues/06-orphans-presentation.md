# 06: Align The Orphan Repair Workflow

**Spec:** [Consistent Catalog Workspaces](../spec.md)

**What to build:** Present Orphans as consistent, scannable unresolved-reference work with existing repair actions. Preserve grouped issue information and resolution forms rather than treating these rows as ordinary catalog CRUD.

**Blocked by:** 02: Unify The Functions Workspace (shared app shell and visual conventions).

**Status:** completed

- [x] The grouped Orphan view uses the shared styling and clearly identifies missing Functions, Parameter definitions, Criteria definitions and affected usage records.
- [x] Users can inspect the affected identity and Device/Global context and reach the existing prefilled creation/resolution flow.
- [x] Resolution actions retain existing server validation and request contracts. Successful repair refreshes the affected view from the API rather than assuming every related Orphan disappeared.
- [x] Cancelled or failed repair retains the unresolved rows and shows actionable feedback. No row is silently discarded or deleted as a presentation change.
- [x] Preserve existing loading/retry, empty, and grouping behavior. Repair controls remain keyboard-accessible; long identities and tables fit desktop/mobile with local scrolling.
- [x] Mocked checks cover representative missing-target groups, prefilled repair context, successful refresh, validation/failure, cancellation, and no-Orphans state. Inspect desktop/mobile and dark-theme presentation.
- [x] Focused checks and the complete database-free frontend closeout gates pass. Obtain Standards/Spec review and commit only this ticket's changes; leave unrelated edits intact.

## Verification Notes

- Commit: `d1cc643` (`feat(frontend): align orphan repair workflow with catalog workspace`); unrelated `.gitignore` edit left intact.
- Gates: `svelte-check` 0 errors and 0 warnings; Vitest 7/7; ESLint and Prettier clean on the changed files; `pnpm build` exit 0; focused Orphan mocked suite 10/10 passed.
- Visual inspection: desktop (1440px) and mobile (390px) light and dark screenshots plus mobile dark repair dialog (`orphans-1440.png`, `orphans-390.png`, `orphans-1440-dark.png`, `orphans-390-dark.png`, `orphan-repair-390-dark.png`). Stacked `.repair-cell` so multi-target secondary notes ("Also needs a Parameter definition") sit cleanly under the repair action button without crowding.
- Review: parallel Standards and Spec sub-agent reviews passed with 0 documented-standard violations, 0 Fowler-baseline issues, and 0 missing or incorrect spec requirements.
