# 07: Extract shared UI helpers (error formatting + device links)

**What to build:** Two duplicated shapes collapse into shared helpers. (1) The ApiError-formatting ternary — currently 4 copies (ComposeDialog ×2, sequences editor, devices page) — becomes `formatApiError(e)` in the shared client module and every copy calls it. (2) The accent-link style block used by deep links into the device browser (devices index count cells, device detail function links, LinksPanel device cells, DictionaryPage badges) moves to one shared constant or small component, and the raw `encodeURIComponent` call sites (DictionaryPage, LinksPanel) switch to the shared `devicePath()` helper so Global-handling stays in one place.

**Blocked by:** None (can start immediately).

**Status:** done (commit a25b45b)

- [ ] `formatApiError(e)` exported from the shared client module; all 4 copies replaced; error display unchanged in every dialog/page
- [ ] Device-browser link markup (classes + href + title shape) shared; all four call sites use it; rendered output visually identical
- [ ] DictionaryPage and LinksPanel build device hrefs via `devicePath()` (no raw encodeURIComponent)
- [ ] pnpm check 0 errors; build green; E2E suite green (no behavior change intended)
