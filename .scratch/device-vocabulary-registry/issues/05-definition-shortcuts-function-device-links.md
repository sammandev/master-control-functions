# 05: Populate paths — definition shortcuts + function→device back-path

**What to build:** Populating a device without leaving the device context. The device page's Parameters and Criteria tabs offer "Add definition" actions that open the existing dictionary form with the device prefilled; on save, the new definition appears in that device's scoped list. The Functions tab, when empty, shows guidance that Functions are linked from Function pages. On a Function page, the links' device values are clickable and lead into the device browser.

**Blocked by:** 04.

**Status:** done (commit 5899008)

- [x] "Add definition" on device page opens the dictionary form with the device prefilled and read-only-in-context (or visibly locked to the device)
- [x] Saved definitions appear in the device page's scoped list without a stale cache
- [x] Empty Functions tab shows guidance text pointing to Function pages for linking
- [x] Function page's link device values navigate to the device browser
- [x] E2E covers: create device → add parameter definition from device page → definition visible; and function→device navigation; suite green
