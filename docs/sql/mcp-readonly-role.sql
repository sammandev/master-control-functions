-- SELECT-only login for script-audit-mcp (ADR-0012). User-run as a database
-- owner on the catalog host after migration 0008 has been applied. Supply the
-- password at run time; never commit a real password.
--   psql "$ADMIN_DATABASE_URL" -v mcp_password="'<strong password>'" -f docs/sql/mcp-readonly-role.sql
-- Then set the MCP's SCRIPT_AUDIT_DATABASE_URL to this role and database.
-- Adjust the database name below if the catalog is not master_control_functions.

CREATE ROLE script_audit_mcp LOGIN PASSWORD :mcp_password
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION;
ALTER ROLE script_audit_mcp SET default_transaction_read_only = on;

GRANT CONNECT ON DATABASE master_control_functions TO script_audit_mcp;
GRANT USAGE ON SCHEMA public TO script_audit_mcp;
GRANT SELECT ON
    mcp_v1_functions,
    mcp_v1_parameter_definitions,
    mcp_v1_criteria_definitions,
    mcp_v1_parameter_links,
    mcp_v1_criteria_links,
    mcp_v1_examples,
    mcp_v1_script_key_audits
TO script_audit_mcp;
