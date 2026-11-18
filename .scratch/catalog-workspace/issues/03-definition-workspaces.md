# 03: Unify Parameter And Criteria Workspaces

**Spec:** [Consistent Catalog Workspaces](../spec.md)

**What to build:** Apply the established catalog browsing pattern to both definition dictionaries and add scoped actions within Device tabs. Users can inspect ownership, constraints, and usage before changing a definition, while existing forms and safety checks remain authoritative.

**Blocked by:** 02: Unify The Functions Workspace (shared shell and catalog browsing behavior).

**Status:** completed

- [x] Both global dictionaries provide searchable list/detail browsing with consistent controls, sorting and pagination. Device-specific and Global discovery are available and distinctly labelled.
- [x] Details expose the definition's name, Device/Global ownership, type, unit, default, allowed values, cardinality, pattern, required flag, and existing usage/orphan-usage information where applicable.
- [x] Filtering covers the complete matching result set. Identical names in different Devices or Global remain distinct selectable records.
- [x] Create/edit/delete actions reuse existing definition forms and validation. Saving or cancelling preserves list/filter context; deleting the selected row resolves selection safely.
- [x] Device Parameters and Criteria tabs provide contextual actions with the Device prefilled and protected against accidental changes to another Device.
- [x] A Global fallback is visibly shared. Creating a Device override creates a Device-specific definition; editing Global requires a deliberate shared action. Exact Device matching and Global fallback remain unchanged.
- [x] Usage guards and any existing explicit force-delete confirmation remain intact. Removing a Function link never silently deletes its definition.
- [x] Loading, empty, retry, validation, keyboard, and narrow-screen states follow the established pattern for both dictionaries, including local table scrolling and readable long names.
- [x] Mocked checks run against both dictionary endpoints and cover same-name scopes, later-page discovery, contextual create/edit/delete, override versus Global edit, and usage-protected deletion. Inspect desktop/mobile and relevant dark-theme screenshots.
- [x] Focused checks and the complete database-free frontend closeout gates pass. Obtain Standards/Spec review and commit only this ticket's changes; leave unrelated edits intact.

## Completion Evidence

- Commit: `67a59b7` (`feat(ui): unify definition workspaces`) on the existing `master` branch; only the eight ticket files are committed.
- Both dictionaries use URL-backed list/detail browsing and complete Device/Global discovery. Canonical details supply constraints and usage, including orphan usage.
- Device tabs reuse the definition form with protected ownership. Editing shared Global and creating a Device override are separate actions. Deletion retains links and explicitly warns about fallback or unresolved references.
- Gates: all 43 frontend-only mocked Playwright workflows and two existing database-free Vitest tests passed. Svelte check reports zero errors/warnings; scoped ESLint, Prettier and static build passed.
- Visual inspection: dictionary layouts at 1440px, 960px and 390px, plus dark theme; Device layouts at desktop and mobile widths, plus dark theme. Long content stays contained and keyboard selection works.
- Svelte analyzer: no issues in the changed components or full Device page. Suggestions concern existing form reset and Device lazy-load effects, retained without unrelated refactoring.
- Standards review: no documented violations; one optional duplicated-selection note. Spec review: fallback warning corrected after a failing regression, then independently confirmed resolved with no remaining blockers.
- No database engines, database tests, host access, push or deployment. Real database/deployment checks remain user-run off-machine. Existing Paraglide CDN-fetch warnings remain.
- Unrelated Examples and frontend ignore-file edits were excluded from the ticket commit and appear separately in `f353d35`. The final worktree is clean. Tickets 04-07 and other workspace roots are unchanged by this work.
