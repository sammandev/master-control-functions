# 04: Device page flow — deep-linkable tabs + clickable counts

**What to build:** The device page's active tab (Functions / Parameters / Criteria / Examples) is reflected in the URL, so tabs are shareable, bookmarkable, and back/forward-friendly. On the devices list, the count columns are clickable and open that device's page directly on the matching tab.

**Blocked by:** 01.

**Status:** done (commit 9753f2d)

- [x] Active tab encoded in the URL on the device page; opening a shared link lands on that tab with data loaded
- [x] Browser back/forward moves between tabs; direct page loads default sensibly (Functions)
- [x] Each count column cell on the devices list links to that device with the matching tab preselected
- [x] Tab lazy-loading/caching behavior preserved (no duplicate fetch loops)
- [x] E2E asserts deep-link landing and count-column navigation; suite green
