# 01: Improve Device Discovery

**Spec:** [Consistent Catalog Workspaces](../spec.md)

**What to build:** Show four Device cards per wide desktop row and let users search for a Device inside the selector popup. Keep existing directory search/sort, counts, CRUD, Device navigation, and separate Global access working.

**Blocked by:** None (can start immediately).

**Status:** completed

- [x] At a 1440-pixel desktop viewport, at least four non-Global fixture Devices occupy the first row; the check fails for three columns.
- [x] At intermediate widths the directory uses two columns, and at a 390-pixel mobile viewport it uses one. Long names and all four counts remain readable without document-level overflow.
- [x] Global remains separate from the Device cards; directory search, sort, count links, and existing create/delete behavior remain available.
- [x] The selector trigger displays the current Device. Its popup contains the search input above a scrollable options list, rather than using an external search input as the trigger.
- [x] Case-insensitive search filters already loaded Devices. Clearing search restores the list; selected and no-match states are visible, and Global remains distinctly labelled and reachable.
- [x] Keyboard users can open the popup, search, move through matches, select with Enter, and dismiss with Escape. Focus returns predictably after selection or dismissal.
- [x] Changing Device preserves existing route/tab behavior and never allows late results for the previous Device to overwrite the new content.
- [x] Mocked checks cover the four-column contract, a 32-Device list, popup-local search, keyboard behavior, and desktop/mobile fit. Inspect corresponding screenshots and relevant dark-theme states.
- [x] Focused checks and the complete database-free frontend closeout gates pass. Obtain Standards/Spec review and commit only this ticket's changes; leave unrelated edits intact.
