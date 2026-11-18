# 06: Device-name normalization & validation hardening

**What to build:** Device names behave the same everywhere. Content paths that create *derived* device names (dictionary create/edit, link edit, example device field) apply the same TrimSpace+ToUpper normalization the registry does (ADR-0002), so a registered WIFI_TEST and a derived lowercase wifi_test can never coexist. Create-time validation rejects reserved/confusing names (`_GLOBAL`, `GLOBAL`) and control characters, and caps length (256). Detail-path GETs/DELETEs URL-decode the device segment server-side so encoded names (spaces etc.) resolve instead of 404ing.

**Blocked by:** None (can start immediately).

**Status:** done

- [x] Dictionary/link/example create+update paths normalize device_name exactly like registry create (shared normalizeDeviceName helper; importer now trims beside its uppercase rule)
- [x] CountDeviceNameUsages (and index union) can no longer miss a casing twin; existing ADR-0002 uppercase tests extended to cover the dictionary path
- [x] POST /api/devices rejects `_GLOBAL`, `GLOBAL`, control characters, and names over the cap with field errors on device_name
- [x] Go integration tests: casing-twin scenario (dictionary lowercase create then registry register → one device), reserved-name rejection, long-name rejection
- [x] Device segment decoded server-side (url.PathUnescape or equivalent) with a test hitting /api/devices/{name with space} and DELETE of the same
- [x] Full pass green (backend tests, pnpm check, build, E2E)
