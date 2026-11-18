# 10: Polish + contract — dark audit, a11y sweep, alias removal, dead-style retirement

**What to build:** The contract step. Every surface is restyled, so: page-by-page dark-mode audit with no unstyled remnants; keyboard-only accessibility sweep over every core flow (browse, filter, edit, link, compose, import, orphans, sequences) — tab order, focus visibility, live regions, dialog focus behavior, dropdown-menu keyboard paths; token-alias removal once zero old-name usages remain (grep-proven); shared layout classes retired where kit components superseded them on touched surfaces; the styling-contract comment updated; spec/ADR cross-references checked.

**Blocked by:** 09.

**Status:** done

- [x] Dark mode verified page-by-page; no hardcoded production colors outside the token layer (grep-proven; Storybook and the dev-only prototype excluded)
- [x] Keyboard-only pass over every core flow; focus, menus, dialog restoration, feedback, and payload assertions pass at the mocked UI boundary
- [x] Zero old token-name usages; no aliases remain; ADR record matches reality
- [x] Shared-class consumers audited; live field, feedback, and layout classes retained rather than deleting required CSS
- [x] Approved database-free gates pass; UI conventions documented; database-backed gates remain user-run off-machine

## Completion Evidence

- Run after completed Import commit `ef63f8e12559bdf41415649b41d2ad28452dce16`, the user-approved review baseline.
- Eleven routes audited in light/dark at 1920px and 390px: Devices, Device hub, Functions, Function detail, both dictionaries, Examples, Sequences, Sequence detail, Orphans, and Import. Twenty-two independently budgeted route tests preserve all four route/theme/viewport combinations.
- Manually inspected dark screenshots for all eleven routes, plus mobile Import, reset confirmation, and Composer. No new layout or readability defects found.
- Added light/dark keyboard regression coverage for browse, filter, edit, link menus, attachment, Composer, and Import reset cancellation. Existing Sequence reorder/edit and Orphan repair checks now exercise keyboard commands while preserving request assertions.
- Fixed dark destructive-button text contrast, verified at least 4.5:1 in both themes. Dialog overlays now use a central semantic token with unchanged opacity.
- Static searches found no legacy token aliases/usages or fixed production UI colors outside the token layer. Storybook and the development prototype remain intentional exclusions. Shared classes with live consumers were not removed.
- Durable conventions are in `frontend/README.md#styling-contract`; the application stylesheet header links to them. Local decision record: `docs/adr/0011-shadcn-theme-vocabulary.md`. Later full-width, unframed, no-density decisions supersede the original cap/card/density proposal.

## Verification

- Full API-mocked Playwright suite: **154 passed**, 27.4 minutes. Invocation from frontend: `playwright test --config=playwright.ui.config.ts --timeout=120000 --output=test-results/ui-refresh-10-gate`, with `PLAYWRIGHT_UI_PORT=4224`, `CI` unset, and loopback in `NO_PROXY`.
- The final invocation used a two-minute total scenario budget; individual assertion timeouts were unchanged and no retries were enabled. Earlier runs exposed long-scenario timeouts and a transient scrollbar-observation race. Route tests were isolated and active scrollbar state is captured on actual scroll events; product timing was not changed. Earlier failure artifacts are retained separately.
- Focused route audit: **22 passed**. Focused scrollbar repeat: **4 passed** across desktop/mobile. Import and Function keyboard/contrast checks, Sequence keyboard payload checks, and Orphan keyboard repair checks passed.
- Database-free Vitest server/client projects: **3 files, 7 tests passed**.
- `pnpm check`: **0 errors, 0 warnings**. Production build passed. Scoped ESLint and Prettier checks passed. Both edited Svelte wrappers returned no autofixer issues or suggestions.
- Independent final review against the Import baseline: **Standards 0 findings; Spec 0 findings**. The earlier full-suite verification condition was resolved by the complete 154-case run.
- No database engine was installed or accessed, no database-backed tests were run locally, and no deployment-host access or deployment was performed. Default database-backed E2E and backend database gates require user-authorized off-machine execution.
- Existing ignore policy is unchanged: this ticket and the local ADR remain local; only the five tracked frontend source/test/documentation files are committed.
