# Task 5 Report: Manage Lists Dialog and Reassignment UX

## Summary

Implemented `TroubleshootingRegistryDialog.svelte` with:

- Accessible Category and Severity tabs in the shared dialog wrapper.
- Row-level label and description editing, usage counts, Save actions, and serialized up/down ordering updates.
- Severity-only tone editing restricted to `destructive`, `warn`, `default`, and `outline`.
- No delete control for built-in entries.
- Custom-entry AlertDialog deletion with a required, same-type, non-self replacement.
- Success and error toasts, inline failure messages, retained form values, and failed deletion remaining open.
- Existing registry update/delete client functions with stable keys and encoded reassignment parameters.

## TDD evidence

### RED

Created `frontend/tests/components/TroubleshootingRegistryDialog.svelte.spec.ts` before the component.

Focused command:

```text
pnpm test:unit -- --run --project client tests/components/TroubleshootingRegistryDialog.svelte.spec.ts
```

Result: failed as expected because `TroubleshootingRegistryDialog.svelte` did not exist. The command also selected unrelated client suites because of the extra npm argument separator; subsequent focused runs used `pnpm exec vitest --run --project client ...`.

### GREEN

Focused command:

```text
pnpm exec vitest --run --project client tests/components/TroubleshootingRegistryDialog.svelte.spec.ts
```

Result: 1 file passed, 4 tests passed.

Tests cover:

- Category/Severity tabs and hidden built-in deletion.
- Usage count rendering.
- Metadata save with stable key and severity tone.
- Serialized move updates with expected `sort_order` values.
- Required different replacement, self exclusion, and failed DELETE remaining open with inline error.

## Verification

Passed:

```text
pnpm exec vitest --run
11 files passed, 41 tests passed

pnpm check
svelte-check found 0 errors and 0 warnings

pnpm exec eslint src/lib/components/TroubleshootingRegistryDialog.svelte tests/components/TroubleshootingRegistryDialog.svelte.spec.ts
0 errors

git diff --check
clean
```

Svelte MCP `svelte-autofixer` returned no issues and no suggestions for the final component.

`graphify update .` completed and refreshed the repository graph for the new component and test.

## Self-review

- CRUD functions are used without duplicating request logic.
- Reorder writes are serialized and only update local order after both updates succeed.
- Failed saves/reorders preserve drafts and expose visible errors; failed deletes do not close the alert.
- Delete replacement options exclude the source and are constrained to the active registry type.
- Severity tones are exactly constrained to the backend-supported UI vocabulary.
- Shared dialog, alert-dialog, tabs, select, input, textarea, and button components are used; no raw z-index values were added.

## Review follow-up

Addressed the Task 5 review findings:

- Reorder requests now use the row's last persisted label, description, and tone and change only `sort_order`; unsaved drafts remain untouched.
- Save, move, and delete mutation handling is separated from the `onChanged` refresh callback. Refresh rejection produces an error toast without converting an already-successful mutation into an inline mutation failure.
- A successful DELETE now removes the local row and closes the alert before awaiting refresh, preventing repeat deletion while refresh is slow or fails.
- Reorder sequencing is proven with a deferred first response: the second request is absent until the first resolves.
- DELETE tests assert the method, the exact normal URL, and the exact percent-encoded source/replacement URL.
- Tests assert exactly four severity tone choices, no category tone control, and that severity reorder ignores an unsaved tone draft.

### Follow-up RED/GREEN evidence

The strengthened focused test run first produced 3 failures: one invalid Vitest Browser matcher in the new exact-count assertion, plus the two intended production failures (reorder submitted draft metadata, and successful DELETE remained open when refresh rejected). The matcher assertion was corrected to exercise the fourth option plus absence of a fifth option. The subsequent focused run passed after the production changes.

### Official Svelte MCP process evidence

Before implementation, `svelte/list-sections` was called and `get-documentation` was used for these Svelte 5 sections:

- `$props`, `$state`, `$derived`, `$effect`: typed callback props, local draft state, writable derived row collections, and avoiding unnecessary state synchronization effects.
- Basic markup, `{#if ...}`, `{#each ...}`: accessible event handlers, keyed row iteration, conditional severity-only controls, and conditional deletion/inline errors.
- `{#snippet ...}`, `{@render ...}`: composition expectations around installed Bits UI/snippet-based wrappers.
- `bind:`: row-local form bindings and controlled tab/select values.
- Testing and TypeScript: component-test behavior and typed props/state in Svelte 5.
- Best practices: avoiding state mutation in effects and keeping derived values side-effect free.

Autofixer evidence from the recorded MCP calls:

- First valid autofixer pass found an unkeyed each block and advised replacing `$state` + `$effect` row synchronization with writable `$derived`.
- Those findings were fixed by keying the tab loop and using derived `EditableRow` instances.
- The next valid pass, and the final review-fix pass, each returned `issues: []`, `suggestions: []`, and `require_another_tool_call_after_fixing: false`.

Limitation: the MCP transcript was not persisted as a separate repository artifact, so this report records the section names and returned results from the calls in this session rather than claiming unavailable raw command logs or timestamps.

### Follow-up verification commands and output

```text
pnpm exec vitest --run --project client tests/components/TroubleshootingRegistryDialog.svelte.spec.ts
Test Files  1 passed (1)
Tests       8 passed (8)

pnpm exec vitest --run --project client
Test Files  4 passed (4)
Tests      18 passed (18)

pnpm check
svelte-check found 0 errors and 0 warnings

pnpm exec eslint src/lib/components/TroubleshootingRegistryDialog.svelte tests/components/TroubleshootingRegistryDialog.svelte.spec.ts
(no output; exit 0)

git diff --check
(no output; exit 0)
```
