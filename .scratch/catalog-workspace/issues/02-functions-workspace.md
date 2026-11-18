# 02: Unify The Functions Workspace

**Spec:** [Consistent Catalog Workspaces](../spec.md)

**What to build:** Extend the approved Device styling to the app shell and make Functions a filterable list/detail workspace with explicit CRUD. Add contextual actions to the Device Function browser while preserving shared Function ownership, link editing and the Composer. Establish reusable browsing behavior for the later catalog tickets.

**Blocked by:** None (can start immediately).

**Status:** completed

- [x] The shared header, navigation, colors, identifier typography, focus treatment and workspace sizing follow the approved Device visual language. Product identity, theme/density controls, skip navigation and all existing navigation destinations remain available.
- [x] The global Functions catalog presents a selectable list beside the Function details, with search and existing Source, Interface, Category, execution-role, and execution-phase filters.
- [x] Device-context browsing is clearly distinguished from all-Device discovery. Device filtering uses suitable existing endpoints or complete result sets; a later-page match is discoverable and counts represent real results.
- [x] Users can create, edit, and delete Functions through explicit actions and existing forms. The selected Device Function has reachable contextual actions, clearly identifying the Function as shared across Devices.
- [x] Save/cancel preserves browsing context where possible. Server validation is visible, deletion remains confirmed, and deleting the selected Function produces a valid next selection or clear empty state.
- [x] Existing advanced Function editing, parameter/criteria link operations, and Composer workflows remain reachable. Link removal and definition deletion are not conflated.
- [x] Device-scoped details retain exact-scope/Global fallback, resolved-definition joins, unresolved-link visibility, and separate Global references. Changing Device or selection cannot show stale details.
- [x] Existing Function detail URLs, Device tab/Function history, and back/forward behavior remain useful. Filters stay compact above the list; no additional permanent filter sidebar is added to the Device workspace.
- [x] Desktop list/detail browsing and narrow-screen navigation remain readable with long identifiers and scripts, including loading, empty, validation, and retry states.
- [x] Reuse owned controls and existing forms; establish shared browsing behavior only where it removes real duplication. If an existing API cannot satisfy an acceptance condition, report the gap before expanding scope.
- [x] Mocked workflows cover filter request semantics, later-page results, shared ownership, CRUD save/cancel/validation/delete, preserved link/Composer access, and Device fallback behavior. Inspect desktop/mobile and dark-theme presentation.
- [x] Focused checks and the complete database-free frontend closeout gates pass. Obtain Standards/Spec review and commit only this ticket's changes; leave unrelated edits intact.

## Completion Evidence

- Commit: `39c18f7` (`feat: unify Functions workspace`) on the existing `master` branch; only the ten ticket files are committed.
- Reused Function details and CRUD forms across global, standalone and Device browsing. Device creation opens the new shared Function detail; Back restores the Device context without inventing ownership or links.
- Gates: zero Svelte check errors/warnings, scoped ESLint and Prettier, two existing database-free Vitest tests, static build, and all 25 frontend-only mocked Playwright tests passed.
- Visual inspection: 1440px desktop, 960px intermediate, 390px mobile, dark theme and mobile form; long identifiers and tables remain contained.
- Standards review: no documented violations; optional pagination duplication retained because cancellation/cache ownership differs. Spec review: creation handoff gap reproduced by a failing regression, corrected, then independently reviewed with no remaining findings.
- No database engines, database tests, host access, push or deployment. Existing Paraglide CDN-fetch warnings remain. Tickets 03-07 are not implemented by this commit.
