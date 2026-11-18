# 04: Combobox — link picking on functions

**What to build:** The hand-rolled link picker inside the links panel becomes a Combobox: debounced async search over parameter and criteria definitions, a listbox navigable with arrow keys and committed with Enter, Escape to dismiss, proper expanded/active-option state for assistive tech, and the create-new-definition inline flow still reachable from an empty result list. Linking a Parameter or Criteria definition to a Function becomes fully keyboard operable.

**Blocked by:** 01 (Foundation), 02 (Dialog — the links panel is one of the modal sites).

**Status:** resolved

- [x] Arrow keys move through options; Enter commits; Escape closes without committing
- [x] Debounced async search preserved with loading and empty states handled
- [x] Create-new-definition flow reachable from empty results without leaving the dialog
- [x] Expanded state and active option announced via ARIA
- [x] E2e keyboard test added: open picker, search, navigate, commit link
- [x] Full suite green

## Comments

- New wrapper `src/lib/components/ui/combobox.svelte`: async search-as-you-type over a caller-supplied
  `search(term)` function. bits-ui 2.x Combobox does NO client-side filtering — the Root's `inputValue`
  is NOT bindable — so the wrapper tracks the term through its own `oninput` on `Combobox.Input`
  (bits-ui `mergeProps` composes, not overrides, handlers passed alongside its own) and owns the 250 ms
  debounce, a request-id race guard (old picker could resolve out of order; this fixes that), the
  loading indicator (aria-live="polite" inside the listbox) and the empty state.
- Selection semantics preserved deliberately: Enter/ArrowDown select and close the listbox (input shows
  the definition's label) but do NOT auto-commit the link — the device-scope/position/required fields
  stay editable, matching the old picker. Committing is the dialog's Attach action; a second Enter on
  the picker input falls through bits-ui's early return and submits the form implicitly, so the whole
  flow is keyboard-only.
- Escape layering (bits-ui layer stack): with the listbox open, Escape closes the listbox only; a
  second Escape closes the dialog. Covered by e2e.
- ARIA comes free from bits-ui: input has role=combobox, aria-expanded, aria-activedescendant,
  aria-autocomplete=list; the portaled listbox is role=listbox with role=option items.
- Empty results render the empty label plus an optional "Create new definition" action button
  (`emptyActionLabel`/`onEmptyAction`) which switches the dialog to inline mode in place. The
  pick/inline radio toggle stays as the non-empty path to inline mode.
- `allowDeselect={false}` so re-Enter on the selected item just closes the listbox instead of clearing
  the pick.
- e2e `combobox.spec.ts` (3 tests, first coverage of the attach flow at all): keyboard commit
  (click → type → ArrowDown → Enter select → Enter submit → row resolved + count text), Escape
  layering (listbox closes, dialog survives, second Escape closes dialog), and empty→inline creation
  (search a missing name → click "Create new definition" in the empty state → fill inline form →
  attach). All role/label selectors, options looked up at page level (listbox portals beside the
  dialog).
- Old `.options/.option/.opt-name` styles and the pickOptions/pickLoading state are gone; the 250 ms
  debounce now lives in the wrapper.
- Concurrent paraglide effort continues: pnpm check still reports its 6 pre-existing errors in
  src/hooks.ts/src/hooks.server.ts (wrong '@sveltejs/kit/hooks' import) — untouched, not staged.
- Verification record (full suite green per spec): go test ./... (5 packages), pnpm check (0
  errors / 0 warnings in ticket files), pnpm build, e2e 19/19 (3 new combobox + 16 prior) — all
  green.
- Review follow-ups applied: header comment corrected (Root.inputValue is NOT bindable — it
  contradicted the onValueChange comment); the "Searching…" live region is now persistent while
  the listbox is open (visible while searching, sr-only otherwise) so freshly-typed terms are
  actually announced; e2e keyboard test now asserts aria-expanded=true after open and a non-empty
  aria-activedescendant after ArrowDown.
- Deferred smell: the seed() fixture-upload helper is now copy-pasted into 5 e2e specs (smoke,
  dialogs, selects, examples, combobox) — extract a shared e2e/helpers.ts during ticket 08 polish.

