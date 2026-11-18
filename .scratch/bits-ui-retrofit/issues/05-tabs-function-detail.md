# 05: Tabs + function detail page

**What to build:** The function detail page runs entirely on the new system: its parameter links, criteria links, and examples sections become Tabs with proper tab semantics and arrow-key switching; breadcrumbs render consistently; hints that ride on title attributes become keyboard-reachable Tooltips. The page's scoped styles are deleted.

**Blocked by:** 01 (Foundation), 04 (Combobox — same page as the links panel; sequenced to avoid parallel edits to one page).

**Status:** resolved

- [x] Tabs: arrow keys switch tabs; tablist/tab/tabpanel roles correct; active tab announced
- [x] Breadcrumbs consistent on detail pages
- [x] Tooltips replace title-attribute hints and are keyboard reachable
- [x] Function detail e2e paths (view, edit, link, examples tab) pass
- [x] Function detail scoped styles deleted
- [x] Full suite green

## Comments

- **Wrappers added**: `ui/tabs/` namespace (shadcn-svelte shape — `index.ts` re-exports `Root`/`Content` from bits-ui directly, styled `tabs-list.svelte` + `tabs-trigger.svelte`; pages import `$lib/components/ui/tabs`, never raw bits-ui), `ui/tooltip.svelte` (composite: caller's snippet IS the trigger and receives trigger props to spread onto its own element — a Button or a span badge — so buttons are never nested), `ui/breadcrumb.svelte` (items prop; last crumb gets `aria-current="page"`).
- **bits-ui 2.19.4 Tooltip gotchas (carry forward)**: (1) `Tooltip.Root` throws `Context "Tooltip.Provider" not found` at render time unless wrapped in `Tooltip.Provider` — the wrapper embeds its own Provider (`delayDuration` lives on Provider, not Root). (2) `Tooltip.Content` sets NO `role` attribute in 2.19.4 — wrapper passes `role="tooltip"` explicitly (also what the e2e asserts). (3) Tooltip.Root prop is `delayDuration` (radix-style), not `openDelay`.
- **bits-ui Tabs behavior change**: `Tabs.Content` does NOT unmount inactive panels — it renders all of them and toggles the `hidden` attribute (no forceMount involved). Consequences: hidden panels stay mounted (state preserved across tab switches — better UX), their mount effects still fire (criteria/examples fetch eagerly), and their inputs stay in the DOM, so page-level locators can hit strict-mode collisions. `smoke.spec.ts` edit test now scopes to `getByRole('tabpanel', { name: 'Parameters' })` (role queries exclude hidden panels) with `exact: true` on 'Device scope'.
- **Tooltip scope**: title-attribute hints on the detail page's surfaces were exactly the three orphan badges (LinksPanel unresolved, ExamplesPanel orphan, ComposeDialog orphan). All three converted to keyboard-reachable Tooltips (badge spans get `tabindex="0"`; tooltip opens on focus; `svelte-ignore a11y_no_noninteractive_tabindex` with rationale). Remaining title attrs live on other pages: functions list, dictionary, import, orphans, sequences detail, dark-mode toggle — covered by tickets 06/07/08.
- **e2e `tabs.spec.ts` (new, 3 tests)**: arrow-key switching with aria-selected assertions (Parameters→Criteria→Examples→Criteria, automatic activation via roving focus); orphan-hint tooltip opens on keyboard FOCUS (not just hover) — seeds the chain definition→link→force-delete because the API refuses links to unknown names (422); breadcrumb nav with labelled link + `aria-current="page"`. All role/label-based.
- `selects.spec.ts` page-size keyboard test flaked once in a full-suite run (passed on rerun + in isolation, twice more since) — timing-sensitive Enter→ArrowUp→Enter against the portaled listbox; not touched by this ticket.
- Mojibake guard: all edits via edit tool; no PowerShell string writes. Byte-checks clean (0xEF 0xBF 0xBD scan on touched files).
- Concurrent paraglide i18n effort's files remain uncommitted/untouched (its 6 svelte-check errors in hooks.ts/hooks.server.ts are pre-existing and not ours).
- **Review follow-ups (post-review fixes)**: (1) `tab` state restored to literal-union type `$state<'parameters' | 'criteria' | 'examples'>('parameters')` (reviewer: Primitive Obsession — the rewrite had silently dropped the union). (2) `breadcrumb.svelte` hardened: `aria-current="page"` now set ONLY on the last item (a middle item without `href` previously got it wrongly). (3) Breadcrumb AC clarification: "consistent on detail pages" here means the functions detail page; `sequences/[id]` still uses old `.crumbs` and its migration (including breadcrumbs) belongs to ticket 07. (4) e2e coverage note: no test drives the Examples *tab's panel content* (tabs.spec asserts aria-selected on switch; panel content is exercised via the compose flow) — accepted. (5) Tooltip badge conversion duplicated across LinksPanel/ComposeDialog/ExamplesPanel (3 near-copies, ExamplesPanel diverges via Badge vs span) — deferred to ticket 08 dedupe ledger.

## Verification

- `pnpm check`: 0 errors, 0 warnings in all files touched by this ticket (6 pre-existing errors in the concurrent paraglide effort's `src/hooks.ts`/`src/hooks.server.ts`).
- `pnpm build`: green.
- `pnpm test:e2e`: 22/22 (3 new tabs + 19 prior).
- `go test ./...`: 4 packages ok (no backend changes).
