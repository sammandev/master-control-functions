# 03: Functions list restyle

**What to build:** The Functions list page reads as a finished product surface: toolbar, search, facet filters, and table grouped into Card sections with consistent spacing. Table rows stop being walls of buttons — crowded row actions move behind a per-row dropdown menu (open, arrow-through, Enter, Escape all keyboard operable). Successful saves and deletes toast politely; validation errors and blocking failures keep persistent `role="alert"` banners. Existing sorting and pagination behave exactly as before, aria-sort intact. The page migrates its token usages from old names to the default vocabulary.

**Blocked by:** 02.

**Status:** done

- [x] Functions list composed of Card sections; toolbar/filters/table hierarchy visually clear in light and dark
- [x] Row actions behind keyboard-operable dropdown menu where crowding warranted; remaining quiet actions may stay inline
- [x] Save/delete success toasts; error banners unchanged and still announced
- [x] Sorting, pagination, search, facets byte-for-byte behaviorally unchanged; aria-sort kept
- [x] Token usages on this surface migrated to default names
- [x] e2e updated in-ticket (dropdown row actions by keyboard, toast on success); full gate green
