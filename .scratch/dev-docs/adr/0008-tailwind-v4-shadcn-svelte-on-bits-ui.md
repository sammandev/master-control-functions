# Frontend UI stack: Tailwind CSS v4 + shadcn-svelte components on bits-ui primitives

The frontend had no CSS tooling or component library: every interactive element was hand-rolled against per-component scoped CSS and a small global token sheet. This produced six near-identical modal implementations with no Escape handling or focus management, a link picker and tab set without keyboard navigation, destructive actions behind the browser's bare `confirm()`, and a token system where six of fourteen color variables were never actually declared — only referenced through fallbacks. Fixing any of this by hand means re-fixing it on every surface.

Decision: adopt **bits-ui** (headless, accessible Svelte 5 primitives) as a direct dependency, distributed the **shadcn-svelte** way — Tailwind CSS v4, a theme token layer established through the components manifest, and component source copied into the app and owned by it. Application code imports a wrapper layer of these copied components, never raw bits-ui primitives. The existing color tokens map onto the new CSS-variable theme, preserving the compact data-dense identity and enabling dark mode.

## Considered Options

- **Status quo (hand-rolled + scoped CSS)** — rejected: the accessibility gaps (no Escape close, no focus trap, no keyboard nav on tabs/combobox) are exactly the class of bug that recurs per implementation, and the six-way modal duplication is the evidence it already has.
- **Raw bits-ui with the existing scoped CSS (no Tailwind)** — the genuine alternative: bits-ui is framework-agnostic about styling and its data-attribute selectors can be driven from global CSS. Rejected because every element would still be styled by hand across eleven scoped style blocks, the undeclared ad-hoc tokens would persist, data-attribute selectors fight Svelte scoping (requiring `:global()` or child snippets everywhere), and there would be no coherent theme layer — more work than shadcn-svelte for a weaker result.
- **shadcn-svelte (bits-ui + Tailwind v4 + copied components)** — chosen: copy-paste ownership means no black-box dependency and free customization; the manifest-driven copy workflow establishes the token/theme layer and dark mode with one bootstrap (the CLI can generate it when a `svelte.config.js` alias file exists; this repo keeps its SvelteKit config inline in `vite.config.ts`, so the first delivery was hand-copied against the manifest); the pre-styled component set covers the entire retrofit scope (Dialog, AlertDialog, Combobox, Tabs, Select, form fields, Badge, Pagination, Table, Collapsible, Breadcrumb, Tooltip).

## Consequences

- Tailwind CSS v4 becomes a build dependency; existing global CSS migrates to utility classes plus the CSS-variable theme layer, and superseded styles are deleted as surfaces migrate.
- The app owns the copied component source: upgrades against the shadcn-svelte registry are manual diffs, and bug fixes happen in-tree — there is no library to bump.
- bits-ui is pinned as a direct dependency (2.x line, Svelte 5 peer); wrapper components keep raw primitives out of application code.
- Reversal is expensive — unwinding Tailwind and rewriting every migrated surface — which is why this decision is recorded. The status-quo option remains reachable only by redoing the retrofit, not by config.
- Dark mode now exists via the `.dark` class over the shared variables; density stays compact by mapping the original tokens rather than adopting the default palette.
