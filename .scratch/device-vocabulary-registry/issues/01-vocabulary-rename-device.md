# 01: Vocabulary rename — Scope → Device everywhere

**What to build:** The user-facing vocabulary for the device dimension becomes "Device": nav item "Devices", pages under `/devices` and `/devices/[name]`, read API under `/api/devices`, and every label, column head, badge, empty state, and deep-link across the app (dictionary pages, links panel, examples panel, compose dialog, orphans page, sequence editor). The blank value keeps rendering as "(Global)"; "device scope" survives only as an internal concept (code identifiers may keep `device`/`device_name`). Old wording is gone from every user-visible surface. Full test suite green after the rename.

**Blocked by:** None (can start immediately).

**Status:** done (commit 6de8dda)

- [x] Nav item reads "Devices"; device index and detail pages serve under `/devices` and `/devices/[name]`; old `/scopes` URLs no longer exist
- [x] Read API endpoints registered under `/api/devices`; no `/api/scopes` routes remain
- [x] Column heads, form labels, placeholders, badges, breadcrumbs, and empty states across all affected screens say "Device" (not "Device scope" / "Scope")
- [x] Blank value renders "(Global)" everywhere it did before
- [x] Dictionary "Device" values deep-link into the renamed device browser
- [x] Go integration tests and the Playwright E2E suite assert the new vocabulary and URLs; full pass green (backend tests, `pnpm check`, build, E2E)
