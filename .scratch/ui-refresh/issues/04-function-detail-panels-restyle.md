# 04: Function detail + shared panels restyle

**What to build:** The Function detail page and the shared components it hosts get the same treatment: detail header and tabs in Card sections, parameter/criteria LinksPanel and ExamplesPanel restyled, ComposeDialog and DefinitionDialog visually consistent with the new surfaces. The Composer's script output stays Cascadia mono and visually distinct. Edit/delete/link flows keep behavior; successes toast, persistent error banners stay. Breadcrumbs consistent. Token usages on these surfaces migrate to default names.

**Blocked by:** 02.

**Status:** done

- [x] Function detail: header, tabs, panels in Card sections; light + dark verified
- [x] LinksPanel, ExamplesPanel, ComposeDialog, DefinitionDialog restyled; dialog keyboard behavior (Escape, focus trap/restore) unchanged
- [x] Composer output remains mono; compose-and-save flow ends in a success toast
- [x] Editing/linking/deleting behavior unchanged; error banners persistent and announced
- [x] Token usages migrated; e2e (dialogs, tabs, combobox, selects, a11y) green with updates where selectors needed them
- [x] Full gate green
