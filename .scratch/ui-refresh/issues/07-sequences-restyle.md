# 07: Sequences restyle

**What to build:** The Sequences list and Sequence builder get the new look: list table in a Card section; builder page — Function reference picking, parameter/criteria value editing, per-step timeout/flow/on-fail, up/down reorder buttons, notes, and the emitted text block — visually consistent, with the emitted text staying Cascadia mono. Create/save/delete keep behavior; successes toast, persistent error banners stay. Reordering stays button-driven (no drag-and-drop). Token usages migrate.

**Blocked by:** 02.

**Status:** done

- [x] Sequences list + builder in Card sections; light + dark verified
- [x] Emit, reorder (up/down), add composer step, save flows behaviorally unchanged; emitted text mono
- [x] Success toasts; persistent announced error banners; confirm dialogs for deletes
- [x] Token usages migrated
- [x] e2e (sequences happy paths: emit, reorder, composer step) green with in-ticket updates; full gate green
