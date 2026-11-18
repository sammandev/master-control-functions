# 01: Foundation — UI toolchain, theme tokens, dark mode, Examples page tracer

**What to build:** The app gains its UI stack end-to-end: bits-ui as a direct dependency, Tailwind CSS v4, the shadcn-svelte CLI distribution (components manifest + copied, app-owned component wrappers for Button, Badge, Input, Label, Textarea, Checkbox, Radio Group), a CSS-variable theme mapping the existing color tokens, and dark mode (system preference by default, manual toggle in the app header). The thinnest surface — the examples page — is migrated completely as the tracer proof that the toolchain works through build, check, and e2e.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] Standard commands all green: backend Go tests, svelte-check, production build, e2e suite
- [x] bits-ui + Tailwind v4 installed; component manifest present; wrappers copied and owned in the UI component directory
- [x] Theme layer: existing 8 root tokens mapped to theme variables; 6 undeclared ad-hoc tokens become declared variables; compact density preserved
- [x] Dark mode works via class variant; header toggle present; system preference respected by default
- [x] Examples page renders entirely through new wrappers; its scoped styles deleted
- [x] Examples e2e happy path passes with role/label-based selectors (no class assertions)

## Comments

- Implemented 2026-10-02 (ticket 01 of the bits-ui retrofit). Deps: bits-ui 2.19.4, tailwindcss 4.3.3 + @tailwindcss/vite, clsx, tailwind-merge. shadcn-svelte CLI was NOT used — this project keeps its SvelteKit config inline in vite.config.ts (no svelte.config.js), which the CLI's alias detection requires; the 7 wrappers (button, badge, input, label, textarea, checkbox, radio-group + radio-group-item) were hand-copied instead (they are app-owned source either way). bits-ui 2.x has no Checkbox.Indicator/RadioGroup.Indicator sub-components — indicators render through the children snippet. components.json manifest added for future CLI compatibility.
- Smoke-suite data drift fixed alongside (pre-existing, unrelated to wrappers): the fixture workbook now ships 289 functions (not 179) and zero orphans (not 1 parameter link), so smoke.spec count/orphan-report assertions were updated to the current fixture truth. New e2e/examples.spec.ts covers the tracer page (seeded row + search filter, role/label selectors only).
