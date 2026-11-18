# Theme tokens speak the shadcn-svelte default vocabulary

The token layer introduced with the bits-ui retrofit used a compact custom vocabulary (`bg`, `surface`, `accent`, `danger`, ...). It fit the original palette, but every component copied from the shadcn-svelte registry arrives referencing the shadcn default names (`background`, `card`, `primary`, `destructive`, `ring`, ...), so each new component needs its classes manually translated before it works, each future registry diff re-translates them again, and contributors familiar with shadcn-svelte must learn a private dialect first. Meanwhile the two vocabularies disagree on two names outright: shadcn's `muted` and `accent` are subtle *backgrounds* while ours were muted *text* and brand *blue*.

Decision: rename the token layer to the shadcn-svelte default vocabulary — `background`, `foreground`, `card`, `popover`, `primary`, `secondary`, `muted`, `muted-foreground`, `accent`, `accent-foreground`, `destructive`, `border`, `input`, `ring`, `radius` — mapped from the existing palette (brand blue becomes `primary`, `danger` becomes `destructive`). `ok`, `warn`, `crit-soft`, the hover surface, and the mono font stack stay as app-specific extensions in the same layer. The rename rides the ui-refresh effort as an expand–contract: the canonical variables and utilities land first, the old names survive as aliases so nothing breaks, each area ticket migrates its own usages, and the polish ticket deletes the aliases once a grep proves zero usages remain.

## Considered Options

- **Keep the custom vocabulary** — rejected: it taxes every future registry copy with manual translation and makes the codebase harder to navigate for anyone who knows shadcn-svelte; the vocabulary exists to serve the copied-component workflow, and it was fighting it.
- **Dual vocabulary forever (shadcn names beside custom names)** — rejected: two true names for one color invites drift between them and doubles the surface a dark-mode or palette change must touch.
- **Rename with temporary aliases (expand–contract)** — chosen: registry components drop in unmodified, migration cost is spread across the area tickets that were already restyling those files, and no step leaves the app broken.

## Consequences

- Components copied from the shadcn-svelte registry work as written; wrapper-layer diffs against upstream become meaningful.
- The two semantic collisions (`muted`, `accent`) cannot alias, so their canonical variables are withheld until the legacy utilities (`text-muted`, `text-accent`, ...) are fully migrated; until then subtle backgrounds use `secondary`, hover backgrounds use the extension hover surface, and brand blue uses `primary`. The polish ticket flips both names to shadcn semantics.
- `primary` inherits the existing blue rather than shadcn's near-black: the visual identity is unchanged; only the names align.
- Dark mode, the `.dark` class strategy, and the pre-paint script are untouched by the rename.
- The alias layer is temporary debt with a scheduled removal: keeping it past the polish ticket would defeat the rename's purpose.

## Resolution (ui-refresh polish ticket, contract executed)

The expand–contract completed as planned. All legacy usages (wrapper layer, pages, `devices.ts` link classes) were migrated to canonical names and the alias variables and alias utilities were deleted after a grep proved zero remaining usages. The two collided names now carry shadcn semantics app-wide:

- `muted` is a subtle background (same value as `secondary`, matching the shadcn default theme where the two are equal); muted *text* is `muted-foreground`.
- `accent` is the hover/highlight background, taking the values of the former `surface-hover` extension (the app's blue-tinted hover). That extension variable and its utility were folded into `accent` and deleted.
- Wrapper-layer stand-ins from the migration (dropdown-menu `focus:bg-surface-hover`) were restored to registry-exact `focus:bg-accent`.

The extension set is now `ok`, `warn`, `crit-soft`, and the mono font stack; the hover surface is no longer an extension.
