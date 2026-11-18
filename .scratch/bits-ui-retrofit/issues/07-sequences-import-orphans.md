# 07: Sequences detail + import + orphans

**What to build:** The remaining surfaces migrate: the sequence detail page (copy-all, up/down step reordering kept as-is, its large scoped style block dissolved into the new system), the import page (styled file input, import history rows as Collapsible, banners), the orphans page, and the compose dialog's detail groups as Collapsible sections. Reordering, copy-all, and the import → report flow behave exactly as before.

**Blocked by:** 01 (Foundation), 02 (Dialog — orphans page and compose dialog host modals), 03 (Select — orphan source/type dropdowns).

**Status:** resolved

- [x] Collapsible sections: compose dialog groups and import history rows; keyboard operable
- [x] File input styled to match the form language; import flow works end-to-end (upload → change report)
- [x] Sequence step reordering and copy-all behavior unchanged
- [x] Orphans page fully migrated, including resolve flows from orphan rows
- [x] All touched scoped styles deleted
- [x] Breadcrumb: `sequences/[id]` old `nav.crumbs` → `ui/breadcrumb.svelte` (deferred from ticket 05 — only functions detail was migrated there)
- [x] Full suite green

**Resolution notes:**
- New wrapper: `ui/collapsible.svelte` over bits-ui Collapsible (Root/Trigger/Content), with
  bindable `open`, optional `onOpenChange`, rotating-chevron trigger, keyboard operable by default.
- Sequences detail: scoped style block (160 lines) dissolved into utilities; Breadcrumb replaces
  `nav.crumbs`; step cards, meta form, emission box all wrapper-based. The e2e class hooks
  (`.steps`, `.emitted-preview`) and every button name/aria-label are preserved; reorder, copy-all
  and emission behavior untouched. A `border-warn` conditional keeps the flagged-step cue.
- Import page: native file input styled to the form language (`file:` variants); the history table
  became one Collapsible per run — run #, file, reset badge, finished timestamp and row counts
  stay visible in the collapsed header, the normalization table (Table parts) and skipped rows
  expand on demand with the same lazy per-run fetch. Last-import card, buttons, badges migrated.
- Orphans page: Table parts, filter bar as Buttons (active = accent variant), Badge warnings,
  dialog form fields on Label/Input. Resolve flows byte-for-byte logic — a created blank-scope
  definition still resolves its link through ADR-0009's global fallback (device scope is not
  prefilled, consistent with the dictionary forms).
- ComposeDialog: the two `<details>` sections ("Save as example", "Add to sequence") are
  Collapsibles (the latter still reloads sequences on open); all raw inputs/buttons/labels moved
  to wrappers; scoped styles deleted. All e2e-pinned texts survive.
- Fixture re-sync (domain fallout from ADR-0009, outside the ticket's UI scope): the tracked
  importer fixture was stale relative to the merged master — it still held the intermediate
  comma-list device scopes (e.g. "IOT,WIFI_TEST") which the exact-scope-or-blank resolution can
  never match, producing 279 orphans and failing TestOrphanReportFromRealWorkbook + the orphan
  e2e (both pin 0). The merged master's one-row-per-device expansion imports clean (orphans 0,
  261 parameter / 70 criteria definitions), so the fixture was re-copied from it and the pinned
  counts updated (workbook_test 211→261, 56→70; import/export tests 211/56→261/70; OSP ANT row
  id 210→260). All pins ride on the ADR-0009 working-tree changes.
