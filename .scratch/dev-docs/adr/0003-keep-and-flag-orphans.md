# Orphan policy: keep and flag, never drop or block

The workbook contains rows referencing targets that don't exist (verified: 25 parameter-links, 8 criteria-links, 16 examples reference absent Functions; 58 parameter names and 23 criteria names used by links are missing from the dictionaries). We decided the app imports everything as-is, visibly flags Orphans in the UI ("unresolved"), and lets the user resolve them by creating the missing Function/definition inline. Import is never blocked and no data is silently dropped or quarantined.

## Consequences

- Foreign keys are enforced at the API layer, not the schema level for link tables (orphan rows must be storable); the UI treats an orphan as first-class data with a warning state, not an error state.
- The website doubles as the tool that repairs the workbook's gaps.
