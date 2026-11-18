# Hosted PostgreSQL for Both Catalog Readers

Accepted on 2026-10-04. The user chose to move the web backend and Script Forge together to the existing Ubuntu-hosted `program_functions` PostgreSQL database. This replaces the SQLite deployment choice in ADR-0001 while retaining Go, Fiber, Svelte, and sqlc. A write pause was approved; the target was confirmed empty and safe to overwrite.

Use a separate PostgreSQL version-6 baseline with identity IDs. Preserve integer flags and text timestamps so the HTTP and workbook contracts stay unchanged. Keep orphan-capable links free of foreign keys per ADR-0003. The transfer retains every application row and verifies content before committing; the original SQLite data and binaries remain the pre-write rollback path.

Script Forge reads the same catalog through read-only repeatable-read transactions. Credentials come from the deployment environment, with verified TLS by default. Go's legacy SQLite driver and transfer fixtures require the off-machine-only `sqlite_migration` build tag; normal builds exclude that engine. The agent prepares and statically verifies code here; the user performs backups, database tests, transfer, and deployment on Ubuntu. See [the cutover procedure](../postgresql-migration.md).
