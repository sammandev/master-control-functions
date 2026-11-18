# Consistent Catalog Workspaces

**Status:** done

## Completion Record

Closed on 2026-10-15 after all seven implementation tickets reached complete/completed status. [Ticket 04](issues/04-examples-workspace.md#verification-notes) records the completed post-fix Standards/Spec review and its remaining nonblocking test-synchronization gap. [Ticket 07](issues/07-import-presentation.md#verification) records the completed Import workflow and database-free closeout gates.

This is tracker reconciliation, not a new implementation or test run. Completion uses the approved mocked-API frontend boundary; real database integration and deployment remain separate operator-run verification.

## Problem Statement

The Device directory and Function browser now follow the approved Device hub design, but the other pages retain different styling and browsing patterns. Users switch pages to inspect or edit related records and lose their browsing context. Device discovery also needs a denser desktop directory and a searchable Device selector.

## Solution

Extend the approved Device visual language across the application. Use filterable list-and-detail workspaces for Functions, Parameter definitions, Criteria definitions, and Examples. Put explicit CRUD actions near the selected record and reuse existing forms.

Keep the Device workspace as the contextual entry point and keep the global catalogs accessible through the existing navigation. Give Sequences, Orphans, and Import the same styling while preserving their specialized workflows.

Change the Device directory to four cards per wide desktop row. Replace the native Device selector with a labelled trigger whose popup contains a search input above its options.

## User Stories

1. As a catalog user, I want consistent navigation and styling across pages, so that switching tasks does not require learning a different interface.
2. As a catalog user, I want all Devices shown on the landing page, so that I can choose the Device relevant to my task.
3. As a desktop user, I want four Device cards per row, so that I can scan more Devices at once.
4. As a mobile user, I want fewer columns and readable controls, so that content fits without document-level horizontal scrolling.
5. As a Device user, I want a search field inside the Device selector popup, so that I can locate a Device without scanning the whole list.
6. As a keyboard user, I want to open, search, select, and dismiss that popup, so that switching Devices does not require a mouse.
7. As a Device user, I want the selected Device and Global option clearly identified, so that I understand the current context.
8. As a Device user, I want switching Devices to update the selected content reliably, so that late responses cannot show another Device's records.
9. As a catalog user, I want global catalog pages to remain available, so that I can discover shared records and definitions unused by any Function.
10. As a Function user, I want a filterable list beside the selected Function, so that I can compare records without repeatedly opening separate pages.
11. As a Function user, I want the existing Source, Interface, Category, and execution filters retained, so that the new layout preserves current discovery capabilities.
12. As a Function user, I want filters to cover the full result set, so that a match on a later page remains discoverable.
13. As a Function user, I want to create, edit, and delete records through explicit actions, so that browsing does not accidentally become editing.
14. As a Function editor, I want shared ownership visible before mutation, so that I know an edit can affect multiple Devices.
15. As a Device user, I want contextual Function actions, so that I can work on the selected Function while retaining Device context.
16. As a Function user, I want link editing and the Composer still reachable, so that the redesign preserves existing authoring tools.
17. As a dictionary user, I want Parameter definitions and Criteria definitions to use the same browsing pattern, so that their common properties are easy to inspect.
18. As a dictionary user, I want search and Device/Global filtering, so that I can find a definition with the correct ownership.
19. As a dictionary user, I want constraints, defaults, and usage information beside the list, so that I can inspect a definition before editing it.
20. As a Device user, I want definition CRUD in the relevant Device tab, so that I can maintain Device-specific definitions without leaving the workspace.
21. As a Device user, I want a Global fallback distinguished from a Device definition, so that I can deliberately create an override or edit the shared definition.
22. As an editor, I want removing a Function link distinguished from deleting a definition, so that I do not destroy a record used elsewhere.
23. As an Example user, I want search and Function/Device filters beside script details, so that I can locate a useful usage record.
24. As an Example user, I want the complete script and meaning visible, so that I can inspect and copy the actual usage.
25. As a Device user, I want contextual Example CRUD, so that the selected Function and Device are prefilled correctly.
26. As an editor, I want saving or cancelling a form to preserve filters and selection where possible, so that I can continue the same task.
27. As an editor, I want server validation and destructive-action confirmations retained, so that the new presentation preserves existing safety checks.
28. As a Sequence author, I want the ordered-step editor kept spacious, so that I can manage Function references and step settings efficiently.
29. As a Sequence author, I want existing composition and step operations preserved, so that visual changes do not alter emitted text or execution settings.
30. As a data-quality user, I want Orphans shown as unresolved-reference work, so that I can inspect and repair missing targets.
31. As a data-quality user, I want unresolved rows retained after unsuccessful repair, so that incomplete data is never silently discarded.
32. As an import operator, I want upload, confirmation, results, and history to remain distinct stages, so that I can understand what happened during an import.
33. As an import operator, I want reset confirmation and normalization details preserved, so that destructive operations and corrections stay visible.
34. As a catalog user, I want loading, empty, and retry states consistent, so that failed requests do not look like missing data.
35. As a catalog user, I want navigation history and existing URLs preserved, so that bookmarks and back/forward navigation remain useful.
36. As a keyboard or dark-theme user, I want legible contrast and visible focus across the redesigned pages, so that every workflow remains usable.

## Implementation Decisions

- Keep the approved B Device hub with A-style Function list/detail browsing as the visual reference. The open prototype's variant C does not change that approval.
- Apply the existing Device visual tokens, restrained borders, identifier typography and wider workspace shell consistently. Preserve the product name, theme and density controls, navigation and skip link. Do not invent Device families or descriptions absent from the API.
- At a 1440-pixel desktop viewport, show four Device cards in a row. Reduce to two columns at intermediate widths and one on mobile. Card labels and all four counts must remain readable; the Global entry stays separate.
- Keep the Device selector trigger outside the popup. Put its search input inside the popup, above a scrollable list. Filter the already loaded Device summaries case-insensitively; support clearing the search, selected-state indication, no matches and arrow navigation. Enter selects a match; Escape dismisses the popup. Keep Global distinctly labelled and reachable.
- Keep global navigation entries for Functions, Parameters, Criteria, Examples, Sequences, Orphans and Import. The Device workspace shows contextual records and separately labelled Global references; global catalogs remain cross-Device discovery surfaces.
- Functions, Parameter definitions, Criteria definitions and Examples use list/detail browsing. Place search and compact filters above the list, with secondary filters collapsible where needed. Avoid adding another permanent sidebar to the Device workspace.
- Reuse suitable list controls and owned UI wrappers. Establish shared browsing behavior through the Functions slice before adapting dictionaries and Examples. Introduce shared components only for demonstrated reuse.
- Use explicit create/edit/delete actions and existing dialogs for routine CRUD. Keep advanced Function editing and link management reachable, with access to the Composer. Reuse their existing logic. Preserve browsing state on save/cancel and handle deletion of the selected record explicitly.
- A Function remains shared by its identity `(source, interface, name)`. Device context must not turn it into a Device-owned copy. Show shared ownership in Function mutation flows.
- Parameter and Criteria definitions retain identity `(name, device_name)`. Links carry Device context and Function association. A link removal must remain distinct from deleting its definition. Preserve usage information and existing delete/force-delete guards.
- Prefill contextual definition and Example forms with the current Device and, where relevant, Function. Creating a Device override from a Global definition must target a new Device-specific definition. Editing Global must be a deliberate, visibly shared operation.
- Preserve exact Device scope followed by Global fallback, including suppression of shadowed Global links and visibility of unresolved links. Resolve definitions through the server-resolved definition identity; the Composer's deduplicated field list is not a substitute for raw scoped links.
- Retain Function filters for Source, Interface, Category, execution role, and execution phase. Add Device/Global filtering where relevant to catalog discovery, using existing scoped endpoints where suitable. Global and selected-Device result sets must be distinguished clearly.
- Preserve server pagination or retrieve a complete result set before local filtering. Never filter only the first loaded page while presenting results as complete. Do not display fixture-derived or guessed counts.
- Preserve full Example scripts, copy feedback, meaning and ownership. Criteria examples remain Function usages. Keep Device-specific and Global examples separate in Device browsing.
- Sequences retain their list, ordered-step editor, Composer entry points, and existing step operations. Use the shared shell and controls while preserving Function references, values, timeout, flow, on-fail, Device, and note semantics.
- Orphans remain a repair workflow with grouped unresolved references and existing prefilled resolution forms. Import remains an operational workflow with upload, optional reset confirmation, run history and normalization details.
- Preserve existing routes, history behavior, loading/retry states, and stale-response guards. On narrow screens, stack panes or use existing route navigation; confine wide tables/scripts to their own scrolling surfaces.
- Frontend changes are the intended scope. Reuse existing API contracts and forms. If an acceptance condition cannot be met through existing endpoints, report the concrete gap before proposing backend or schema changes.

## Testing Decisions

- The user approved the existing mocked-API browser boundary. Exercise rendered workflows and request contracts against a frontend-only server. Keep unexpected API requests failing in mocks so missing coverage is visible.
- Extend the existing mocked UI suite and helpers instead of creating a parallel test system. Reuse existing frontend unit/component tests where they already own relevant behavior.
- Cover Device discovery and selection, catalog filters/selection, CRUD validation/save/cancel/delete, scope/override behavior, full-script copy, and preservation of specialized workflows.
- Include records beyond the first page in discovery fixtures. Test identical definition names in Device and Global contexts, multiple Devices sharing a Function, unresolved links, and out-of-order responses where the touched behavior depends on them.
- Make the four-card check discriminate four columns from three using at least four non-Global Devices at a fixed desktop viewport. Exercise a realistic 32-Device fixture and long identifiers.
- Assert visible presentation contracts alongside behavior: shared styling, pane geometry, identifier hierarchy, readable counts, focus and contrast. Inspect desktop/mobile screenshots and relevant dark-theme states; do not equate passing API-flow tests with visual fidelity.
- Run regular focused checks per ticket and the complete database-free frontend suite at its closeout. Typecheck, scoped lint/format checks, and a static build are required gates. Run browser tests against a freshly built static preview when cold development hydration is unreliable.
- Use only the frontend-only mocked browser configuration. The default E2E configuration starts a backend/database and is prohibited on this workstation. Do not run database tests, install engines, connect to a database, or access the deployment host.
- Real database integration and deployment remain separate operator-run verification. Mocked success establishes frontend behavior, not production database correctness or an updated deployment.

## Out of Scope

- Changes to Function identity, Device identity, definition resolution, workbook schema, or database schema.
- A physical-model hierarchy, remembered last-used Device, invented Device metadata, or removal of global catalogs.
- Import/reset behavior changes, Sequence business-logic changes, checker integration, or a client-owned script emitter.
- Wholesale rewrites of existing forms, unrelated cleanup and new backend functionality without a demonstrated gap and approval.
- Database tests or connections from this workstation, database engine installation, host access and deployment automation or execution.

## Further Notes

The user approved the recommendations, the mocked test boundary, and the seven-ticket dependency graph on 2026-10-06. Each ticket is a complete, independently demonstrable slice. Tickets with satisfied blockers can be implemented in a fresh session using the Ask Matt implementation flow: focused test-first work, database-free closeout gates, Standards/Spec review, and a scoped commit on the current branch.

The first ticket addresses the two explicit Device requests. The Functions ticket establishes the shared browsing pattern and app shell; dictionaries, Examples, and the specialized page tickets depend on that work. The first two tickets have no blockers between them.

This is a local tracker under the repository's declared `.scratch/<feature>/issues/` convention, using `ready-for-agent`. The referenced tracker/label guide files were absent at publication; no tracker infrastructure or domain glossary changes are part of this work.
