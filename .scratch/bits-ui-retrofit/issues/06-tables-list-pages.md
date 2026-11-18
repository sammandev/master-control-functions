# 06: Table/Pagination + list pages

**What to build:** The functions list, sequences list, and examples panel adopt the Table and Pagination chrome: one visual language for all tables, consistent pager with page-size control, and sort-state exposure on sortable headers. Sorting and paging logic stays exactly as built — only presentation migrates. Toolbars and banners across list pages become consistent.

**Blocked by:** 01 (Foundation), 02 (Dialog — functions list and examples panel host modals), 03 (Select — facet filters and page size already migrated on these pages).

**Status:** resolved

- [x] Sortable headers still sort correctly and expose sort state (aria-sort)
- [x] Pager: previous/next plus page-size select unified across list pages
- [x] Functions list, sequences list, examples panel fully migrated; scoped styles deleted
- [x] Functions filter e2e and examples e2e pass unchanged
- [x] Full suite green

**Resolution notes:**
- New owned wrappers: `ui/table/` namespace (Table/TableHeader/TableBody/TableRow/TableHead/TableCell,
  shadcn-svelte shape, compact density) and `ui/pagination.svelte` (bits-ui Pagination prev/next +
  "Page X of Y" + unified page-size select; page-size change resets to page 1, the shared pager
  behavior). Sort and paging state stays in the pages; only presentation migrated.
- Functions sortable headers pass `aria-sort` (ascending/descending, omitted when inactive) via
  TableHead attribute passthrough; sort toggle logic untouched.
- Sequences list gained the pager over the already-supported `page`/`page_size` API params (story
  21's consistent pager); its search debounce now routes through a `searchDebounced` snapshot so
  the reactive load effect owns all list params.
- ExamplesPanel adds the Pagination chrome to its pre-existing page/pageSize state (the effect now
  tracks pageSize explicitly).
- All three surfaces' scoped `<style>` blocks deleted; global shared classes (`.toolbar`,
  `.banner-error`, `.field`) remain until the polish step's dead-global cleanup, per spec.
- Wrapper-layer dependency fix from review: pagination indicator uses `text-xs text-muted` tokens
  instead of the global `.muted` class, so ticket 08's dead-global cleanup cannot unstyle it.
- Incidental gate fix: `src/hooks.ts` / `src/hooks.server.ts` (untracked paraglide WIP from other
  work) imported `Handle`/`Reroute` from `@sveltejs/kit/hooks`, which does not export them — 6
  svelte-check errors blocked the green gate. Imports corrected to `@sveltejs/kit`; those files
  stay uncommitted with the rest of that in-flight work.
