# 02: Dialog + AlertDialog — deduplicate all modals and confirms

**What to build:** Every modal site (compose dialog, examples panel, dictionary pickers, links panel, functions list, orphans page) uses one shared Dialog wrapper: Escape closes, focus is trapped while open, focus returns to the trigger on close, backdrop clicks close, content renders as a portal overlay with no layout shift. Every destructive action — delete Function, delete Parameter/Criteria definition, detach link, reset database — confirms through an accessible AlertDialog instead of the browser's `confirm()`. E2e selectors are updated for portaled content and new interaction tests cover Escape, focus restore, and the confirm flows.

**Blocked by:** 01 (Foundation — toolchain, theme, Button/Label wrappers).

**Status:** resolved

- [x] All six modal sites on the shared Dialog wrapper; per-site duplicated modal markup and styles deleted
- [x] Escape close, backdrop close, and close button all work on every modal
- [x] Focus trapped while open; initial focus sensible; focus restored to trigger on close
- [x] Modal semantics exposed to assistive tech; portal overlay causes no layout shift
- [x] All destructive confirms are AlertDialogs; no `window.confirm` remains in the frontend
- [x] E2e updated: dialog flows pass against portaled DOM; new tests for Escape close, focus restore, confirm flow
- [x] Full suite green

## Comments

- All six modal sites migrated onto `ui/dialog.svelte`; six `window.confirm`
  sites (function delete, definition delete, detach, force-delete-on-409,
  example delete, import reset) onto `ui/alert-dialog.svelte`. Each confirm
  keeps a pending-row state so the dialog description stays stable after
  close. ComposeDialog keeps its parent mount/unmount contract: internal
  `open = $state(true)`, `$effect` forwards close requests to `onClose`.
- bits-ui 2.x API notes for later tickets: composition snippet is
  `child({ props })` (spread onto a custom component), NOT `asChild` and
  NOT `children({ props })`; `on*`-named props must be passed as explicit
  arrows (`onOpenChange={(o) => ...}`) — svelte-check mis-types the
  shorthand. Controlled pattern everywhere: `{open}` + `onOpenChange`, no
  `$bindable` in the wrappers.
- New `frontend/e2e/dialogs.spec.ts`: 6 tests (Escape close + focus
  restore, focus trap under 12 Tabs, backdrop click, delete-confirm,
  cancel-keeps-row, functions delete-confirm). e2e suite now 13 tests,
  all green; Go 89 tests green; `pnpm build` green.
- `pnpm check`: 0 errors in every file this ticket touched. 6 remaining
  errors live in the concurrent paraglide i18n effort's `src/hooks.ts` /
  `src/hooks.server.ts` (it imports `Handle`/`Reroute` from
  `@sveltejs/kit/hooks`; they are exported from `@sveltejs/kit`), which
  that effort is still wiring (its deps landed mid-implementation). Not
  touched from here to avoid colliding with in-flight work.
- Legacy `.dialog-backdrop`/`.dialog` global classes are now unreferenced
  by application code; they stay in `@layer base` until the polish ticket
  deletes dead global styles (spec Further Notes).
