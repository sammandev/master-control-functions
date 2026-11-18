# Spec: bits-ui Retrofit (shadcn-svelte UI System)

Status: done

Retrofit of the entire frontend onto bits-ui headless primitives, distributed through shadcn-svelte (Tailwind CSS v4 + owned copied components + CSS-variable theming). Governed by `CONTEXT.md` (glossary) and ADR-0008. Respects ADR-0006 (backend/frontend directories) and ADR-0007 (pnpm canonical). No backend changes.

## Problem Statement

Every interactive surface of the app was hand-rolled, and the same mistakes were made six times: six near-identical modal implementations that do not close on Escape, do not trap focus, and do not return focus to where the user came from. The link picker has no keyboard navigation. The function detail tabs cannot be operated with arrow keys. Destructive actions rely on the browser's bare `confirm()` popup. Styling is split across a global stylesheet, eleven per-component style blocks, and a set of undeclared color tokens referenced only through CSS fallbacks — so the look drifts from page to page and every fix must be applied repeatedly. None of this is visible until a keyboard-only or screen-reader user actually tries to use the tool.

## Solution

Adopt bits-ui — headless, accessible Svelte 5 primitives — through the shadcn-svelte distribution: a CLI installs Tailwind CSS v4, a token/theme layer, and copied component source the app owns. All modals become Dialog, destructive confirms become AlertDialog, the link picker becomes Combobox, function detail tabs become Tabs, and form controls plus chrome (selects, inputs, badges, pagination, table markup, collapsibles, breadcrumbs, tooltips) take the shadcn-svelte equivalents. The refresh is moderate: the current compact, data-dense identity is preserved by mapping the existing color tokens onto the new theme variables, and a dark mode comes along for free. Custom business logic (sortable headers, pager, file upload, reorder buttons) stays; only the presentation migrates.

## User Stories

1. As a test engineer, I want every modal to close with Escape, so that I can dismiss it without hunting for a close button.
2. As a test engineer, I want focus to move into a dialog when it opens and return to the trigger when it closes, so that keyboard flow is never broken.
3. As a test engineer, I want focus trapped inside an open dialog, so that Tab cannot wander into the obscured page behind it.
4. As a screen-reader user, I want dialogs marked with modal semantics, so that I know interaction is confined to the dialog.
5. As a test engineer, I want backdrop clicks to still close dialogs, so that quick dismissal keeps working.
6. As a test engineer, I want the composer dialog, examples panel, dictionary pickers, links panel, and orphan-page dialogs to share one dialog behavior, so that the app feels like one product.
7. As a knowledge-base maintainer, I want destructive actions (delete function or definition, detach a link, reset the database) to confirm through an accessible dialog, so that mistakes are preventable without leaving the keyboard.
8. As a test engineer, I want dialogs rendered as overlay portals, so that opening one never shifts the page layout underneath.
9. As a test engineer, I want link-picking search to present a listbox I can navigate with arrow keys and commit with Enter, so that linking parameter and criteria definitions does not require a mouse.
10. As a screen-reader user, I want the picker to announce whether its list is open and which option is active, so that I can pick links reliably.
11. As a test engineer, I want the picker's debounced async search preserved, so that linking stays fast against the full dataset.
12. As a test engineer, I want the create-new-definition inline flow to remain reachable from an empty picker result, so that linking does not force a detour to the dictionary.
13. As a test engineer, I want the function detail tabs (parameter links, criteria links, examples) switchable with arrow keys, so that detail pages are fully keyboard operable.
14. As a screen-reader user, I want the active tab and its panel to be announced through proper tab semantics, so that I know which section I am reading.
15. As a test engineer, I want every dropdown (facet filters, dictionary type and cardinality, composer target, page size, orphan filters, examples source) rendered through one select component, so that the app looks and behaves uniformly.
16. As a keyboard-only user, I want every dropdown operable without a mouse, so that filtering works on any input device.
17. As a test engineer, I want text inputs, textareas, checkboxes, and radios to share one form-field language with consistent labels, so that every edit form reads the same way.
18. As a test engineer, I want field validation errors presented consistently under their fields, so that invalid edits are immediately locatable.
19. As a knowledge-base maintainer, I want the file picker on the import page to stay simple but match the new form styling, so that import does not look like a leftover.
20. As a test engineer, I want status badges (unresolved/orphan flags, warning states, counts) visually consistent everywhere, so that data quality signals are scannable at a glance.
21. As a test engineer, I want the list pager and its page-size control consistent across pages, so that moving between long lists requires no relearning.
22. As a test engineer, I want all tables to share one visual language while sorting keeps working exactly as before, so that familiarity is preserved.
23. As a test engineer, I want collapsible sections (composer detail groups, import history rows) to behave and look consistently and respond to the keyboard, so that progressive disclosure is uniform.
24. As a test engineer, I want hints that currently ride on title attributes replaced by real tooltips, so that they appear reliably and are reachable by keyboard.
25. As a test engineer, I want detail pages to use consistent breadcrumbs, so that navigation context is always visible.
26. As a test engineer, I want a dark mode that follows my system preference by default and can be toggled manually, so that long editing sessions are comfortable.
27. As a knowledge-base maintainer, I want the compact, data-dense look preserved through the restyle, so that the refresh remains a refresh and not a redesign.
28. As a knowledge-base maintainer, I want color roles (danger, warn, ok, accent) used consistently across all surfaces, so that color meaning is never ambiguous.
29. As a keyboard-only user, I want every interactive surface reachable and operable in a sensible tab order, so that no workflow requires a mouse.
30. As a screen-reader user, I want error banners announced when they appear, so that failed saves and validation failures are never silent.
31. As a screen-reader user, I want sortable column headers to expose their sort state, so that I can navigate large lists knowingly.
32. As a frontend developer, I want one owned source of truth for UI components, so that a fix to dialog or select behavior applies app-wide instead of six times.
33. As a frontend developer, I want a single declared token layer replacing ad-hoc fallback tokens, so that styling changes are predictable and centralized.
34. As a frontend developer, I want superseded global styles removed once migration completes, so that dead CSS cannot mislead future work.
35. As a frontend developer, I want every migration step to leave the full test suite green, so that regressions surface at the step that caused them.

## Implementation Decisions

- **Stack**: bits-ui 2.x as a direct dependency (latest 2.19.x verified; Svelte 5 peer requirement satisfied by the current frontend). Components are distributed the shadcn-svelte way: a components manifest (plus the CLI where a `svelte.config.js` alias file exists — this repo keeps its SvelteKit config inline in `vite.config.ts`, so wrappers are hand-copied) installs Tailwind CSS v4, the theme token layer, and copied component source that the app owns and may edit freely.
- **Wrapper layer**: a dedicated UI component directory holds the copied components; application pages and panels import these wrappers, never bits-ui primitives directly. Raw primitives stay an implementation detail behind the wrappers.
- **Component scope — full set**: Dialog (six modal sites), AlertDialog (destructive confirms), Combobox (link picking), Tabs (function detail), Select (all dropdowns), Input, Label, Textarea, Checkbox, Radio Group, Button, Badge, Pagination, Table, Collapsible, Breadcrumb, Tooltip. The file input stays a native element, styled to match.
- **Theming**: the existing eight root color tokens map onto the shadcn CSS-variable theme; the six undeclared ad-hoc tokens become declared variables in the same layer. Density stays compact and data-dense. Dark mode is a `.dark` class variant of the same variables: system preference by default, manual toggle in the app header.
- **Logic that does not migrate**: sortable table headers (sort logic stays; sort-state exposure added), the pager, file upload handling, and up/down row reordering keep their hand-rolled logic — only presentation migrates. These have no bits-ui equivalent and rewriting them is not the point.
- **No toasts**: the inline banner pattern remains; Sonner (shadcn-svelte's toast) is excluded. Error banners gain live-region announcement instead.
- **Migration order**: foundation (toolchain, theme, button/badge/form-field wrappers) → dialog + confirm (deduplicates all six modals) → dictionary pages and all selects → link-picker combobox → function-detail tabs → tables/pager and list pages → sequences/import/orphans surfaces → polish (density, dark audit, accessibility sweep, dead-style removal). Each step must end with the whole verification suite green: Go backend tests, svelte-check, production build, and the e2e suite.
- **E2E selector maintenance**: portal-rendered dialogs change DOM location; e2e selectors are updated in the dialog step. Role-based selectors are preferred so markup changes stay cheap.
- **Recorded decision**: the stack choice is captured in ADR-0008; package management stays pnpm-canonical per ADR-0007; the static adapter is unaffected because portal surfaces mount client-side only.

## Testing Decisions

- **What makes a good test here**: assert external behavior only — keyboard interaction, focus movement, ARIA roles and state, visible outcomes. Never assert utility class names or internal component markup; Tailwind classes are implementation detail and churn constantly during a restyle.
- **What is tested**: every migrated surface through the browser e2e seam, with new interaction coverage added alongside existing happy paths: Escape-close and focus restore on dialogs, focus trapping, arrow-key tab switching, combobox keyboard navigation and async result rendering, confirm-dialog flows replacing `window.confirm`, and select-driven filtering.
- **Which modules**: the frontend only. Backend Go tests continue to pass untouched and prove no API or data change.
- **Prior art**: the existing Playwright smoke suite (five tests driving the real app against a seeded database through the four happy paths) is the pattern to extend. Svelte-check remains the type gate run per migration step.
- **Seam count stays at one**: the browser e2e suite remains the only behavioral seam; no unit-test layer is introduced for this work.

## Out of Scope

- Toasts / Sonner / any new notification pattern; inline banners remain.
- Drag-and-drop reordering; up/down button reordering stays.
- Table virtualization or new sorting/pagination logic.
- Responsive or mobile redesign; the desktop data-tool layout is unchanged.
- Any backend change: API, database schema, import/export behavior are untouched.
- New pages, new features, or content changes anywhere.
- A full visual redesign; the refresh is explicitly moderate and density-preserving.
- Internationalization; themes beyond light and dark.

## Further Notes

- bits-ui portal surfaces mount client-side only; this is irrelevant to the static adapter but means e2e must interact with portaled content in the document body, not in-place.
- Scoped styles are deleted per surface as it migrates; the final polish step removes the now-dead global classes so nothing stale survives.
- The dark-mode toggle (system default plus manual override) is an interpretation of the accepted "CSS-variable theming + dark mode" package; it is cheap to drop the toggle and follow the system preference alone if preferred.
- Node ≥ 20 is required by bits-ui; the toolchain already satisfies this. pnpm 12 remains the only supported installer.
