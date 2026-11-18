# Device vocabulary and a devices registry

The user-facing vocabulary for `device_name` values changes from "scope" to "Device" (nav, pages, routes, API), and a web-native `devices` registry is added so a Device can be created — and exist — before it has any content. This amends ADR-0002 and ADR-0009 in one narrow way: their rejection of a devices entity table applied to *workbook-derived* data; a **registry** alongside the derived data is a different thing and is now wanted.

## Context

The data model carries `device_name` as a free-text attribute on link rows, definition rows, and Examples (ADR-0002, ADR-0009). A "device" has no existence of its own: it appears only when some row references its name, and the device list is derived (`UNION` of distinct `device_name` values). Two problems follow:

- **Vocabulary**: users of the app think in devices, not in "scopes". The UI, routes (`/scopes`), and API (`/api/scopes`) all say "scope", which forces a translation step the domain never needed.
- **Onboarding**: a brand-new device cannot be reserved. Until its first link, definition, or Example is saved, it does not appear anywhere, so "create a device, then populate it" is impossible.

ADR-0002/0009 rejected a devices entity table because the *source* data is free text with casing variants and no source taxonomy. That rejection concerned derived data and still holds for it. A registry table — like Sequences under ADR-0005 — does not compete with that reasoning.

## Decision

- **Vocabulary**: "Device" is the user-facing noun everywhere — nav item "Devices", pages `/devices` and `/devices/[name]`, API `/api/devices`, column heads and form labels "Device". "Device scope" remains the internal concept for the row attribute; the blank value remains **Global** and renders as "(Global)".
- **Registered devices**: a `devices` table (`device_name` unique, timestamps). Web-native per the Sequences precedent: not part of the workbook's six sheets, untouched by import/reset, not exported.
- **Device list = registry ∪ derived names**, so imported/legacy data still appears and empty registered devices are visible. Creating a name that already exists (registered or derived) is rejected with a clear error.
- **Creation normalizes casing to uppercase** before insert (the ADR-0002 import rule), keeping `Osp` and `OSP` from diverging.
- **Deletion** is allowed for registered devices with zero references (links, definitions, Examples). Renaming is out of scope.
- **Populating from the device page**: parameter/criteria definition creation is reachable from the device page (existing dictionary modal, device prefilled). Functions remain link-driven from Function pages — Function identity is `(source, interface, name)` and device-independent by design (ADR-0009).

## Considered Options

- Keep "scope" vocabulary — rejected: the translation cost is the problem being fixed.
- No registry; "create device" = attach first content row — rejected: empty devices remain impossible and onboarding is clunky.
- Devices table + foreign keys on every `device_name` — rejected: breaks the free-text model, fights the import pipeline, and adds migration cost for no integrity gain (links must keep resolving to Orphans, ADR-0003).
- Rename UI labels only, keep `/scopes` URLs and API — rejected: a half-rename (UI says "Devices", URL says "scopes") is the confusing outcome. No external API consumers exist, so the full rename is cheap now.

## Consequences

- ADR-0002/0009's "no devices table" stance now holds for derived data only; registered devices are additive and both sources are unioned in the list.
- The route and API rename touches the E2E suite and dictionary deep-links; all updated in the same change set.
- The CONTEXT.md glossary retires its "_Avoid_: device" rule; "Device" becomes canonical.
- Registered-but-empty devices survive import/reset by design (web-native), like Sequences.
