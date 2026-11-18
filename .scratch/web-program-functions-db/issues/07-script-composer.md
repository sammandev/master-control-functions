# 07: Script composer

**What to build:** On a Function's page, a test engineer opens the composer: a form generated from that Function's parameter links in position order — required fields enforced, SINGLE vs MULTIPLE cardinality drives single/multi-value inputs, allowed values validated per definition. Submitting emits the pipe-delimited script string per the spec's grammar (source+interface prefix, function, parameter assignments, criteria or NULL, timeout-ms, trailing segments), with a segment for criteria assignments resolved from the function's criteria links (or NULL). The emitted script can be saved directly as an Example of that Function. The multi-function sequence mode stays deferred (story 37) but the composer's data flow does not preclude it.

**Blocked by:** 04 (Function editor — links), 06 (Examples CRUD).

**Status:** done

- [x] API: compose-preview endpoint (function id + parameter values → emitted script string + validation result) and save-as-example action
- [x] Form generation honors position order, required, SINGLE/MULTIPLE cardinality, allowed-values sets, and type parsing (int/double/list[int]/string)
- [x] Validation rejections return field-level messages before any script is emitted
- [x] UI: composer dialog/view on the function page with live preview of the script string and copy-to-clipboard
- [x] Save-as-example populates function + device scope + script text, editable before save
- [x] Integration tests through the HTTP seam: known function → expected script string; invalid inputs rejected per constraint; saved example matches emitted script

Annotations:
- Grammar derived from the workbook's real example scripts (7 segments: `source+interface|function|p=v&p=v|criteria-or-NULL|timeout-ms|flow|on-fail`); server-owned in compose.go, UI preview calls the endpoint (single source of truth).
- Save-as-example reuses the existing POST /api/examples (ticket 06); the compose dialog prefills function + device scope (from the links' device scope) + script, all editable before save.
- MULTIPLE cardinality uses a comma-separated text input — matches the workbook's own value_pattern (`<freq1>,<freq2>…`).
- allowed_values: the definitions API only creates `{set}` form; range form (`[a..b]`) arrives via workbook import and is validated in the composer too.
- Sequence mode (story 37) not precluded: emission is a pure function of (function, fields, name-keyed values).
