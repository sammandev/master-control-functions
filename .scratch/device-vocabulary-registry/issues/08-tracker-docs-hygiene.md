# 08: Tracker & docs hygiene

**What to build:** The paper trail matches reality. Tickets 03 and 05 have done status lines but unchecked acceptance boxes; ticket 04's status line still says "pending commit" although its commit landed; README's issue-tracker pointer still names the old feature's `.scratch/web-program-functions-db/issues/` and its E2E section references the old ticket numbering. Tick the boxes, fix the status lines, update README pointers, and note the sqlc v1.31.1 regen recipe (5-file split workaround) wherever the repo documents store regeneration.

**Blocked by:** None (can start immediately).

**Status:** done (see commit below)

- [x] Tickets 03/05 acceptance boxes ticked; ticket 04 status line references its real commit
- [x] README issue-tracker + E2E sections current (device vocabulary, /devices routes, schema version 6, regen recipe)
- [x] No doc still describes /scopes routes or schema version 5
- [x] Full pass green (docs-only change; spot-check builds still pass)

Note: verification ran with unrelated in-progress frontend work (hooks.server.ts, hooks.ts, paraglide i18n) present in the working tree; svelte-check's 10 errors are confined to those foreign files. Backend suite green (api/db/importer).
