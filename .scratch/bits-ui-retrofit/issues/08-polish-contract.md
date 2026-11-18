# 08: Polish + contract — delete dead styles, dark audit, accessibility sweep

**What to build:** The strangler's contract step: every migrated surface is done, so the superseded global styles (shared modal, field, banner, pager, badge classes) are deleted from the global stylesheet, the token layer is the single source of styling truth, and the whole app gets its final pass — density check (compact, data-dense preserved), dark-mode audit on every page, and an accessibility sweep in keyboard-only mode: tab order, focus visibility, live-region announcements, sort-state exposure, dialog focus behavior.

**Blocked by:** 02 (Dialog), 03 (Select), 04 (Combobox), 05 (Tabs), 06 (Table), 07 (Sequences/import/orphans).

**Status:** resolved

- [x] No dead global classes remain; no undeclared ad-hoc tokens; theme layer is sole styling source
- [x] Dark mode verified page-by-page with no unstyled remnants
- [x] Keyboard-only pass over every core flow: browse, filter, edit, link, compose, import, orphans
- [x] Live-region announcements and aria-sort verified
- [x] Full suite green from the standard commands; README/docs updated where they describe UI conventions

**Resolution notes:**
- Straggler surfaces finished: LinksPanel (raw buttons/inputs/checkbox/radios/table → wrappers,
  scoped style block deleted; mode toggle now the RadioGroup wrapper, badges the Badge wrapper),
  DictionaryPage (raw table → Table parts, `.pager` → unified Pagination, aria-sort added to the
  three sortable headers — it had none), layout header (app-header/app-title/nav styles →
  utilities), ComposeDialog + LinksPanel orphan badges → Badge.
- Global stylesheet contract (app.css): deleted the superseded classes — element defaults for
  button/select/input/textarea/table/th/td/tr:hover (no raw tables remain; the few remaining raw
  buttons — sort headers, dictionary link-name buttons — and the spec-mandated native file input
  are fully utility-styled), button.primary/danger/:disabled, .dialog-backdrop/.dialog (dead
  since ticket 02), .pager, .badge/.badge.warn, .app-header/.app-title/nav, and the dead
  `.field label` descendant rule. Kept the still-live shared layout vocabulary (.field,
  .field-error, .banner-error, .banner-empty, .toolbar, .spacer, .muted) — these are token-based
  and no wrapper supersedes them; converting ~130 usages would be churn without visual change.
  Header comment in app.css now states the styling contract.
- ARIA state regression coverage (spec Testing Decisions): new e2e/a11y.spec.ts asserts
  aria-sort toggling on the functions table (ascending → descending → moves to another column),
  role="alert" announcement on the import-without-file banner, and the polite empty-state live
  region on the examples search — stories 30/31 are now protected by the suite, not a manual
  pass. Also fixed the last focus-visible gap (combobox empty-action button) and removed the
  redundant `.toolbar .spacer` rule.
- Dark audit: the one hardcoded rgba (LinksPanel .editing td) is gone with its scoped block; a
  grep proves no hex/rgb colors remain outside the token layer, and var(--…) usage only touches
  declared tokens. Dark mode is a pure variable swap on every page.
- Accessibility sweep: focus-visible outlines added to the last custom interactive buttons
  (functions sort headers, dictionary sort/link-name buttons); aria-sort now on functions AND
  dictionary sortables; role="alert"/aria-live banners verified on every surface; e2e already
  covers keyboard flows (Escape/focus-trap dialogs, arrow-key tabs, combobox + select keyboard,
  page-size keyboard).
- README carries no UI-convention prose to update; the convention record is ADR-0008 plus the
  rewritten styling-contract comment at the top of app.css.
