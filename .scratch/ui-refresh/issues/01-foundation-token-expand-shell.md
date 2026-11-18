# 01: Foundation — token expand, ADR, shell width/radius/nav

**What to build:** The theme layer gains the full shadcn-svelte default vocabulary (`background`, `foreground`, `card`, `popover`, `primary`, `secondary`, `muted`, `muted-foreground`, `accent`, `accent-foreground`, `destructive`, `border`, `input`, `ring`, `radius`) mapped from the current palette — blue accent becomes `primary`, `danger` maps to `destructive`, `ok`/`warn`/`crit-soft`/mono stay as app-specific extensions declared in the same layer. Old custom names remain working as aliases, so nothing breaks. The shell visibly changes: content column widens to ~1280px, 0.5rem radius baseline, header nav highlights the active route. A new ADR records the rename decision and its trade-off. Light and dark both verified on every page.

**Blocked by:** None (can start immediately).

**Status:** done

- [x] New default-named variables drive the theme; old names resolve through aliases; no visual regression beyond the intended shell changes
- [x] ADR written for the token-vocabulary alignment (why, alternatives, consequences)
- [x] Content width ~1280px; 0.5rem radius baseline applied to shared chrome
- [x] Header nav highlights the active route
- [x] Dark mode spot-checked page-by-page; e2e a11y spec still green
- [x] Full gate green: backend tests, svelte-check, build, e2e
