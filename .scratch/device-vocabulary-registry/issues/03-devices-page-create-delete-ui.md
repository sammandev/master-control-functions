# 03: Devices page — create + delete UI

**What to build:** "Create device" on the devices page: a dialog taking a device name, showing the uppercase-normalized name before submit, surfacing duplicate errors from the API. Registered devices with zero references offer a delete action (with confirmation); deletion of a referenced device surfaces the API's clear error. Registered empty devices appear in the list with zero counts, visibly reserved-but-unpopulated.

**Blocked by:** 02.

**Status:** done (commit ac97108)

- [x] Create dialog normalizes and previews the name before submit; success navigates to (or reveals) the new device
- [x] Duplicate creation shows the API error message inline, state stays editable
- [x] Delete available only for registered devices with zero references; confirmation before delete; success removes the row
- [x] Delete of a referenced device is blocked with a clear error
- [x] Empty registered devices render in the list with zero counts
- [x] E2E covers create-through-UI (including duplicate error) and delete of an empty registered device; suite green
