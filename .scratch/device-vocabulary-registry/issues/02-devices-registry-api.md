# 02: Devices registry — schema + create/delete API

**What to build:** A Device can be registered before it has any content. New web-native `devices` table (unique name, timestamps; appended migration, ANSI-compatible SQL). Create endpoint normalizes casing to uppercase before insert and rejects names that already exist (registered or derived). Delete endpoint allowed only when no link, definition, or Example references the name, else a clear error. Device list = registry ∪ derived names with the same counts as today. Registry rows survive workbook import/reset (Sequences precedent, ADR-0005) and are not exported.

**Blocked by:** 01 (lands on the final `/api/devices` paths).

**Status:** done (commit 6c96c3a)

- [x] Migration appends the `devices` table; fresh databases apply it; existing databases upgrade without data loss
- [x] Create: uppercase-normalized insert; duplicate (registered or derived) rejected with a clear error
- [x] Delete: succeeds only for registered devices with zero references; referenced devices rejected with a clear error
- [x] Device list includes registered empty devices (zero counts) alongside derived ones; blank/Global is never a registry entry (not registrable, not deletable) and stays the reserved first row as before
- [x] Registry rows survive import/reset; Export unchanged
- [x] Go integration tests cover create (normalization, duplicates), delete (empty vs referenced), and the union list; suite green
