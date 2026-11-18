# Spec: UI Refresh (shadcn-default restyle + kit completeness)

Status: done

## Completion Record

Closed on 2026-10-15 with all ten implementation tickets marked done. [Ticket 10](issues/10-polish-contract.md#verification) records 154 passing API-mocked UI checks, seven passing unit tests, the frontend gates, and independent Standards/Spec reviews with no findings. This reconciliation does not rerun those checks.

Later approved decisions supersede the original capped-width, card-section, and density-toggle proposals below: production workspaces are full-width, page sections are unframed, and there is no density preference. The current [styling contract](../../frontend/README.md#styling-contract) and [ticket 10 completion evidence](issues/10-polish-contract.md#completion-evidence) record the delivered behavior. Live shared layout classes remain where they have consumers.

Completion uses the approved database-free verification boundary. Historical database-backed gate commands below are not authorization to run them on this workstation; database integration and deployment remain separate operator-run checks.

Second-stage frontend effort, building on the completed `bits-ui-retrofit` (ADR-0008, Status: done). That effort put every interactive surface onto bits-ui primitives through the owned shadcn-svelte wrapper layer, but deliberately deferred three things: toasts, a visual redesign (it preserved the compact data-dense identity), and new table features. This effort is those three things, plus one ecosystem-alignment decision: renaming the theme token vocabulary to shadcn-svelte's defaults. Governed by `CONTEXT.md` (glossary), ADR-0006 (repo layout), ADR-0007 (pnpm canonical), ADR-0008 (bits-ui/shadcn stack). No backend changes.

## Problem Statement

The app works and is accessible, but it reads as a prototype rather than a product. Pages are flat walls of tables and buttons with no visual hierarchy — nothing groups the toolbar, table, and panels into scannable sections. Rows that need several actions (edit, delete, open, reorder) show every button at once. Feedback is inline-only: after saving a Function or deleting a Device, the outcome sits in a banner the user may not be looking at. Several list tables cannot be sorted, and the fixed data-dense spacing is uncomfortable for long reading sessions. Beneath the surface, the token vocabulary (`surface`, `text`, `muted`, `accent`) diverges from the shadcn-svelte defaults the wrapper layer is built from, so every component added from the ecosystem needs its class names hand-edited.

## Solution

One coherent restyle pass to the shadcn-svelte default structure: rename the token layer to shadcn defaults (light + dark, the blue accent becomes `primary`), present page sections as card surfaces with subtle borders, 0.5rem radius, and a wider (~1280px) content column, with dark mode staying first-class. Finish kit adoption: row actions move behind dropdown menus, successes and completions get polite toast feedback (persistent banners stay for errors), and the missing wrappers (card, dropdown-menu, toast, separator) join the owned component layer. Table work: sorting extended to every list table plus a persisted density toggle. All of it rides the existing Playwright e2e seam, green per ticket.

## User Stories

1. As a test engineer, I want every page's sections (toolbar, table, detail panels) grouped into card surfaces with consistent spacing, so that I can scan the page by structure instead of by guessing at whitespace.
2. As a knowledge-base maintainer, I want the theme tokens named per shadcn-svelte defaults, so that any component added from the ecosystem drops in without class renames.
3. As a test engineer, I want to sort every list table by its columns (Functions, Parameter definitions, Criteria definitions, Examples, Devices, Orphans, Sequences), so that the row I need comes to the top without browser find.
4. As a test engineer, I want a density toggle between compact and comfortable table spacing, so that I can trade screen real estate for readability, and I want it remembered across visits.
5. As a test engineer, I want row actions (edit, delete, open, reorder) collected behind a per-row dropdown menu on crowded tables, so that rows stay scannable and the actions are one consistent click away.
6. As a keyboard-only user, I want row-action dropdown menus fully keyboard operable (open, arrow through, Enter, Escape), so that no action requires a mouse.
7. As a test engineer, I want a toast on successful saves, deletes, and import completion, so that the outcome is noticeable without hunting for an inline banner.
8. As a screen-reader user, I want validation errors and blocking failures to keep their persistent inline banners with live-region announcement, so that an error that must not disappear is never reduced to a fleeting toast.
9. As a test engineer, I want destructive confirms to stay as accessible AlertDialogs, visually consistent with the new surfaces, so that mistakes stay preventable.
10. As a test engineer, I want sortable headers to keep exposing their sort state through aria-sort, so that screen readers can navigate large sorted lists knowingly.
11. As a test engineer, I want the active route highlighted in the header nav, so that I always know which section I am in.
12. As a test engineer, I want script strings (Composer output, Example scripts, Sequence steps) to stay in the Cascadia mono stack, so that code-like content remains visually distinct.
13. As a test engineer, I want the Functions list and Function detail to share the new look end to end — toolbar, table, tabs, dialogs — so that the flagship flow feels finished.
14. As a knowledge-base maintainer, I want the Parameter definition and Criteria definition dictionary pages refreshed through their shared page component, so that both dictionaries stay identical in look and behavior.
15. As a test engineer, I want the Examples browser restyled with its search, source filter, and pagination intact, so that browsing usage records improves visually without changing how it filters.
16. As a test engineer, I want the Devices index and Device detail (including its `?tab=` deep links and definition shortcuts) restyled, so that device scoping workflows keep their behavior under the new skin.
17. As a test engineer, I want the Sequences list and Sequence builder (emit, reorder, composer steps) restyled, so that composing long sequences stays comfortable.
18. As a knowledge-base maintainer, I want the Orphans report and Import page restyled, so that data-quality and workbook workflows match the rest of the app.
19. As a test engineer, I want the landing page consistent with the new look, so that no page feels leftover.
20. As a screen-reader user, I want toasts announced politely (aria-live), so that success feedback reaches me without stealing focus.
21. As a test engineer, I want dark mode verified on every new and changed surface, so that the night shift looks as deliberate as the day one.
22. As a keyboard-only user, I want focus states, tab order, and focus-visible outlines preserved or improved on every new component (dropdown menus, toasts, cards), so that the restyle never costs keyboard access.
23. As a frontend developer, I want new wrappers (card, dropdown-menu, toast, separator) sourced from the shadcn-svelte registry into the owned wrapper layer, so that they stay consistent with the existing seventeen and remain app-owned.
24. As a frontend developer, I want the shared layout classes retired where a kit component now visually supersedes them on a touched surface, so that one styling vocabulary survives instead of two parallel ones.
25. As a frontend developer, I want an ADR recording the token-vocabulary rename and its trade-off, so that a future reader understands why ~40 files changed name-shaped edits at once.
26. As a frontend developer, I want the old token-name aliases removed at the polish step, so that no dead vocabulary survives to mislead future work.
27. As a knowledge-base maintainer, I want the Composer, Export, import, and all CRUD behavior byte-for-byte unchanged, so that the restyle cannot corrupt the knowledge workbook round-trip.
28. As a knowledge-base maintainer, I want the full verification suite green at the end of every ticket, so that regressions surface in the step that caused them.

## Implementation Decisions

- **Base**: the owned wrapper layer from the completed retrofit remains the only styling path — pages import wrappers, never bits-ui primitives directly. ADR-0008's contract stands; this effort restyles and extends it, never bypasses it.
- **Token rename** (ADR-worthy, recorded in a new ADR): root variables re-declared under shadcn-svelte default names — `background`, `foreground`, `card`, `popover`, `primary`, `secondary`, `muted`, `muted-foreground`, `accent`, `accent-foreground`, `destructive`, `border`, `input`, `ring`, `radius`. The blue accent becomes `primary`; `danger` maps to `destructive`; `ok`, `warn`, `crit-soft`, and the mono stack remain app-specific extensions declared in the same layer. Old custom names (`surface`, `text`, `muted`, `accent`, …) are aliased to the new variables during migration and the aliases are deleted at polish. Rationale: the shadcn-svelte registry pastes components referencing the default names; aligning removes per-addition hand edits. The one-time cost is a rename across the wrapper layer and pages, taken while everything is being restyled anyway.
- **New wrappers**: card, dropdown-menu, sonner (toast), separator — copied from the shadcn-svelte registry into the owned wrapper layer. Hand-copying matches the established practice (this repo keeps its SvelteKit config inline, so CLI alias resolution is not guaranteed); the CLI may be used where it works, but the wrappers are app-owned either way.
- **Tables**: sorting is extended to list tables that lack it (the Functions and dictionary tables already sort); the density toggle persists in localStorage and swaps a root-level class that adjusts table cell padding; data-dense remains the default identity; aria-sort is maintained on all sortable headers.
- **Toasts**: sonner wrapper with polite announcement; success/completion events (save, delete, import finished) toast. Validation errors and blocking failures keep persistent inline banners with `role="alert"` — a toast is never the only channel for an error that must persist. Existing a11y banner coverage must stay green.
- **Row actions**: per-row dropdown-menu on tables where actions crowd the row (Functions, definitions, Examples, Devices, Sequences); menus open on click, arrow-navigable, close on Escape. Where a table currently has one or two quiet actions, they may stay inline — the driver is crowding, not uniformity for its own sake.
- **Visual structure**: content width grows to ~1280px; page sections (toolbar + table, detail panels, form groups) sit in Card surfaces with 0.5rem radius and subtle borders; the header nav highlights the active route. Spacing and typography otherwise follow shadcn-svelte defaults; density stays data-dense by default.
- **Shared layout classes**: `.toolbar`, `.field`, `.banner-error`, `.banner-empty`, `.muted` are replaced on touched surfaces where a kit component visually supersedes them (Card, Form-field patterns, Alert); they are not churned across untouched surfaces — full retirement happens only at polish if no live usages remain.
- **Migration order**: foundation (token rename + ADR + shell width/radius/header nav) → new wrappers (card, dropdown-menu, sonner, separator) → per-area passes (Functions list + detail; dictionary pages; Examples; Devices index + detail; Sequences list + builder; Orphans + Import + home) → table features (sorting where missing + density toggle) → polish (dark audit, keyboard-only accessibility sweep, token-alias removal, dead-style removal). Each ticket ends with the whole suite green.
- **Frozen behavior**: Composer emission, Export, import/reset, CRUD APIs, and Sequence emission are untouched. E2E selector maintenance stays role-based per ADR-0008 practice; portal-mounted content interacts in the document body.

## Testing Decisions

- **What makes a good test here**: assert external behavior only — keyboard interaction, focus movement, ARIA roles and state, visible outcomes. Never assert Tailwind class names or internal markup; classes churn constantly during a restyle.
- **Which modules**: the frontend only. Backend Go tests pass untouched and prove no API or data change.
- **Single seam, existing**: the Playwright e2e suite (smoke, dialogs, tabs, combobox, selects, devices, examples, a11y) driving the real UI through the real Go server against the seeded database. No component unit-test layer is introduced; no visual-regression tooling.
- **New coverage added in-ticket**: sorting on newly sortable tables; density toggle toggling and persistence; dropdown-menu row actions opened and driven by keyboard; toast appearing on a save/delete success; active-route highlight; existing a11y assertions (aria-sort, role="alert" banners) stay green throughout.
- **Prior art**: the existing specs — especially a11y.spec.ts's aria-sort and banner-announcement assertions and the dialogs/tabs/combobox keyboard specs — are the pattern to extend.
- **Per-ticket gate**: `go -C backend test ./...`, `pnpm check`, `pnpm build`, `pnpm test:e2e` — all green before a ticket is done.

## Out of Scope

- Any backend change: API, schema, import/export, Composer emission.
- Responsive or mobile redesign; the desktop data-tool layout is kept.
- Table virtualization; column-visibility toggles; pagination changes beyond styling.
- Drag-and-drop reordering; up/down button reordering stays.
- Internationalization; themes beyond light and dark; Storybook stories.
- New pages, new data features, or content changes.
- Toasts replacing persistent error banners.

## Further Notes

- The prior retrofit's "Out of Scope" list is this effort's scope statement — the two specs are designed to compose, not overlap.
- Data-dense stays the default; the density toggle is additive comfort, not a redesign of the identity.
- bits-ui portal surfaces mount client-side only; e2e continues to interact with portaled content in the document body.
- pnpm 12 remains the only supported installer (ADR-0007); Node ≥ 20 already satisfied.
