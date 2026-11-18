# 06: Examples + Devices restyle

**What to build:** The Examples browser and both Devices pages match the new look. Examples: search, source filter, pagination in Card sections; browsing/filtering behavior unchanged. Devices index and Device detail: card composition, tabs with `?tab=` deep links still working, definition shortcuts from a device page intact, registered-device create/delete flows with confirm dialogs and success toasts. Token usages migrate on all three surfaces.

**Blocked by:** 02.

**Status:** done

- [x] Examples browser in Card sections; search/filter/pagination behavior unchanged; empty-state live region intact
- [x] Devices index + detail restyled; `?tab=` deep links and definition shortcuts verified working
- [x] Device create/delete keeps accessible confirm; success toasts added; error banners persistent
- [x] Light + dark verified; token usages migrated
- [x] e2e (devices spec) green with in-ticket updates; full gate green
