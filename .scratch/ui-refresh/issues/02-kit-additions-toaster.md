# 02: Kit additions — card, dropdown-menu, sonner, separator wrappers + global Toaster

**What to build:** Four wrappers join the owned component layer, copied from the shadcn-svelte registry and styled through the default-named tokens: Card (section surfaces), DropdownMenu (row actions), Sonner (toast), Separator. The Sonner Toaster mounts once in the root layout so any surface can raise a toast. Everything else unchanged — this is the enablement step the area tickets build on. Wrappers are app-owned and may be edited freely, matching the established practice.

**Blocked by:** 01 (wrappers reference the new token vocabulary).

**Status:** done

- [x] Card, DropdownMenu, Sonner, Separator wrappers exist in the owned layer, styled from default tokens, no hardcoded colors
- [x] Toaster mounted globally; a test toast renders politely and dismisses (manual or dev-only proof, not a permanent UI element)
- [x] No dead styling: wrappers ship with shadcn defaults adapted to the token layer, zero ad-hoc hex/rgb
- [x] Full gate green
