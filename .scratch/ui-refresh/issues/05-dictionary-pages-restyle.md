# 05: Dictionary pages restyle

**What to build:** Both dictionaries — Parameter definitions and Criteria definitions — refresh through their shared page component so they stay identical in look and behavior: Card composition, toolbar, table, pagination; row actions behind dropdown menu; save/delete success toasts with persistent error banners. Existing sorting keeps working with aria-sort. Token usages migrate to default names.

**Blocked by:** 02.

**Status:** done

- [x] Both dictionary pages visually identical, composed of Card sections, light + dark verified
- [x] Row-action dropdown keyboard operable; create/edit/delete flows behaviorally unchanged
- [x] Success toasts; persistent announced error banners retained
- [x] Sorting + pagination unchanged, aria-sort kept; token usages migrated
- [x] e2e updated in-ticket (dictionary flows now driving dropdown row actions); full gate green
