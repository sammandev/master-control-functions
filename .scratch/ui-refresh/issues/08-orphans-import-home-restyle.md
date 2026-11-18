# 08: Orphans + Import + home restyle

**What to build:** The Orphans report, Import page, and landing page match the new look. Orphans: report table and filters in Card sections; orphan flags stay visually loud (Badge language consistent with danger/warn tokens). Import: workbook upload and import/reset flows restyled; the import-without-file banner keeps its persistent `role="alert"` announcement (a11y spec asserts it) and import completion gains a success toast. Home: consistent, no leftover look. Token usages migrate on all three.

**Blocked by:** 02.

**Status:** done

- [x] Orphans report + filters in Card sections; orphan/warn flag badges consistent
- [x] Import page restyled; import-without-file `role="alert"` banner unchanged and e2e-verified; completion toast added
- [x] Home consistent; light + dark verified
- [x] Token usages migrated
- [x] e2e (a11y banner assertions) green with in-ticket updates; full gate green (user-directed skip of full e2e run: backend/check/build green, new import-completion test green in isolation; full-suite verification deferred to ticket 09 gate)
