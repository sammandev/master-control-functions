# Graph Report - Web_Program_Functions_Database  (2026-10-02)

## Corpus Check
- 117 files · ~63,913 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 4 file(s) not represented in the graph (top: (none) 2, .css 1, .exe 1)

## Summary
- 856 nodes · 2276 edges · 56 communities (34 shown, 22 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 229 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `cfe82325`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- testing.T
- importer/export.go
- Queries
- workbook.go
- links.go
- package.json
- orphans.go
- database/sql.NullInt64
- compose.go
- import.sql.go
- sequences.sql.go
- context.Context
- sequences.ts
- request
- definitions.ts
- compilerOptions
- orphans.sql.go
- imports.ts
- Function
- links.ts
- client.ts
- examples.ts
- orphans.ts
- Spec: Web Program Functions Database
- Development
- Language
- Agent skills
- Domain Docs
- Issue tracker: Local Markdown
- sv
- Stack: Go (Fiber) + Svelte 5, SQLite with a PostgreSQL exit path
- Device scope is an attribute on link rows, not an entity
- Workbook is a repeatable upsert source; SQLite is live truth between imports
- Sequences are web-native entities, outside the workbook
- Orphan policy: keep and flag, never drop or block
- Repository layout: backend/ and frontend/ directories
- +layout.svelte
- 00: Environment + app skeleton
- 01: Function CRUD end-to-end
- 02: Dictionary CRUD (parameter + criteria definitions)
- 03: Workbook import — upsert + normalization + change log
- triage-labels.md
- app.d.ts
- ssr
- 04-function-editor-links.md
- 05-orphan-report.md
- 06-examples-crud.md
- 07-script-composer.md
- 08-export-roundtrip.md
- 09-integrate-e2e.md
- 10-sequence-builder.md
- 11-repo-layout.md
- webprogfndb

## God Nodes (most connected - your core abstractions)
1. `postJSON()` - 54 edges
2. `doRequest()` - 52 edges
3. `Queries` - 48 edges
4. `newLinksTestApp()` - 38 edges
5. `serverError()` - 36 edges
6. `request()` - 35 edges
7. `seedFunctionAndParameterDefinition()` - 23 edges
8. `New()` - 22 edges
9. `Queries` - 20 edges
10. `newImportTestApp()` - 17 edges

## Surprising Connections (you probably didn't know these)
- `emitSequence()` --calls--> `validateAndEmit()`  [INFERRED]
  backend/internal/api/sequences.go → backend/internal/api/compose.go
- `New()` --calls--> `registerComposeRoutes()`  [INFERRED]
  backend/internal/api/server.go → backend/internal/api/compose.go
- `definitionWriteError()` --calls--> `serverError()`  [INFERRED]
  backend/internal/api/dictionary.go → backend/internal/api/functions.go
- `newDictionaryTestApp()` --calls--> `New()`  [INFERRED]
  backend/internal/api/dictionary_test.go → backend/internal/api/server.go
- `New()` --calls--> `registerExampleRoutes()`  [INFERRED]
  backend/internal/api/server.go → backend/internal/api/examples.go

## Import Cycles
- None detected.

## Communities (56 total, 22 thin omitted)

### Community 0 - "testing.T"
Cohesion: 0.06
Nodes (110): listBody, main(), compose1(), decodeScript(), fiber.App, seedComposeFunction(), TestComposeFieldsEndpoint(), TestComposeKnownFunctionScript() (+102 more)

### Community 1 - "importer/export.go"
Cohesion: 0.05
Nodes (58): exportWorkbook(), fiber.App, fiber.Handler, registerExportRoutes(), getImportRun(), fiber.App, fiber.Handler, fiber.Map (+50 more)

### Community 2 - "Queries"
Cohesion: 0.07
Nodes (67): exampleBody, exampleResponse, exampleUpdateBody, functionBody, listQuery, sequenceBody, sequenceResponse, sequenceStepBody (+59 more)

### Community 3 - "workbook.go"
Cohesion: 0.08
Nodes (60): compareRows(), exportWorkbookBytes(), fiber.App, T, importFixture(), TestExportEmptyDatabase(), TestExportIncludesOrphans(), TestExportReimportIdempotent() (+52 more)

### Community 4 - "links.go"
Cohesion: 0.07
Nodes (53): definitionBody, dictionaryRoutes, dictionaryRoutes[T], fieldValidationError, httpError, insertLinkParams, linkBody, linkStore (+45 more)

### Community 5 - "package.json"
Cohesion: 0.05
Nodes (37): ADR-0001, ADR-0005, E2E_DIR, FIXTURE, devDependencies, @playwright/test, svelte, svelte-check (+29 more)

### Community 6 - "orphans.go"
Cohesion: 0.12
Nodes (21): orphanExampleItem, orphanGroup, orphanLinkItem, fiber.App, fiber.Ctx, fiber.Handler, nullableAny(), nullablePosition() (+13 more)

### Community 7 - "database/sql.NullInt64"
Cohesion: 0.12
Nodes (14): Queries, CriteriaLink, ParameterLink, database/sql.NullInt64, GetCriteriaLinkRow, GetParameterLinkRow, InsertCriteriaLinkParams, InsertParameterLinkParams (+6 more)

### Community 8 - "compose.go"
Cohesion: 0.18
Nodes (22): composeField, composeRequest, composeSegments, checkAllowed(), checkType(), composeFieldFromCriteriaRow(), composeFieldFromRow(), dedupeCriteriaRows() (+14 more)

### Community 9 - "import.sql.go"
Cohesion: 0.10
Nodes (9): Queries, UpsertCriteriaDefinitionParams, ImportNormalization, ImportRun, InsertImportNormalizationParams, InsertImportRunParams, UpsertCriteriaLinkParams, UpsertExampleParams (+1 more)

### Community 10 - "sequences.sql.go"
Cohesion: 0.13
Nodes (9): ListSequenceStepsRow, Queries, SequenceStep, GetFunctionByNaturalKeyParams, InsertSequenceParams, InsertSequenceStepParams, ListSequencesRow, Sequence (+1 more)

### Community 11 - "context.Context"
Cohesion: 0.21
Nodes (4): Queries, ParameterDefinition, Queries, context.Context

### Community 12 - "sequences.ts"
Cohesion: 0.12
Nodes (14): createSequence(), EmitResult, emitSequence(), getSequence(), listSequences(), SequenceDetail, SequenceInput, SequenceListItem (+6 more)

### Community 13 - "request"
Cohesion: 0.18
Nodes (13): createFunction(), FunctionFacets, FunctionInput, FunctionRow, getFunction(), getFunctionFacets(), listFunctions(), ListParams (+5 more)

### Community 14 - "definitions.ts"
Cohesion: 0.14
Nodes (9): if(), if(), createDefinition(), DefinitionKind, DefinitionListResult, DefinitionRow, listDefinitions(), ADR-0003 (+1 more)

### Community 15 - "compilerOptions"
Cohesion: 0.14
Nodes (13): compilerOptions, allowJs, checkJs, esModuleInterop, forceConsistentCasingInFileNames, moduleResolution, resolveJsonModule, rewriteRelativeImportExtensions (+5 more)

### Community 16 - "orphans.sql.go"
Cohesion: 0.17
Nodes (4): Example, Queries, ListOrphanCriteriaLinksRow, ListOrphanParameterLinksRow

### Community 17 - "imports.ts"
Cohesion: 0.19
Nodes (9): getImportRun(), ImportNormalization, importReset(), ImportResult, ImportRun, importWorkbook(), listImportRuns(), upload() (+1 more)

### Community 18 - "Function"
Cohesion: 0.25
Nodes (4): cmpBy(), Queries, Function, UpsertFunctionNaturalParams

### Community 19 - "links.ts"
Cohesion: 0.20
Nodes (8): ADR-0002, DefinitionInput, attachLink(), LinkAttachInput, LinkKind, LinkRow, LinkUpdateInput, updateLink()

### Community 20 - "client.ts"
Cohesion: 0.22
Nodes (6): ApiError, ComposeField, ComposeFields, ComposeRequest, composeScript(), getComposeFields()

### Community 21 - "examples.ts"
Cohesion: 0.18
Nodes (9): createExample(), ExampleInput, ExampleListParams, ExampleListResult, ExampleRow, ExampleUpdateInput, listExamples(), ADR-0003 (+1 more)

### Community 22 - "orphans.ts"
Cohesion: 0.22
Nodes (7): listOrphans(), OrphanExampleRow, OrphanGroup, OrphanKind, OrphanLinkRow, OrphanReport, if()

### Community 23 - "Spec: Web Program Functions Database"
Cohesion: 0.22
Nodes (8): Further Notes, Implementation Decisions, Out of Scope, Problem Statement, Solution, Spec: Web Program Functions Database, Testing Decisions, User Stories

### Community 24 - "Development"
Cohesion: 0.25
Nodes (7): Database migrations, Development, End-to-end tests (ticket 09), Go toolchain notes, Issue tracker, One-command full test pass, Web_Program_Functions_Database

### Community 25 - "Language"
Cohesion: 0.29
Nodes (6): App-specific terms, Core entities, Data-quality terms, Language, Sequences, Web Program Functions Database

### Community 26 - "Agent skills"
Cohesion: 0.33
Nodes (5): Agent skills, Domain docs, Issue tracker, Triage labels, Web_Program_Functions_Database

### Community 27 - "Domain Docs"
Cohesion: 0.33
Nodes (5): Before exploring, read these, Domain Docs, File structure, Flag ADR conflicts, Use the glossary's vocabulary

### Community 28 - "Issue tracker: Local Markdown"
Cohesion: 0.33
Nodes (5): Conventions, Issue tracker: Local Markdown, Wayfinding operations, When a skill says "fetch the relevant ticket", When a skill says "publish to the issue tracker"

### Community 30 - "sv"
Cohesion: 0.40
Nodes (4): Building, Creating a project, Developing, sv

### Community 31 - "Stack: Go (Fiber) + Svelte 5, SQLite with a PostgreSQL exit path"
Cohesion: 0.50
Nodes (3): Consequences, Considered Options, Stack: Go (Fiber) + Svelte 5, SQLite with a PostgreSQL exit path

### Community 32 - "Device scope is an attribute on link rows, not an entity"
Cohesion: 0.50
Nodes (3): Consequences, Considered Options, Device scope is an attribute on link rows, not an entity

### Community 33 - "Workbook is a repeatable upsert source; SQLite is live truth between imports"
Cohesion: 0.50
Nodes (3): Consequences, Considered Options, Workbook is a repeatable upsert source; SQLite is live truth between imports

### Community 34 - "Sequences are web-native entities, outside the workbook"
Cohesion: 0.50
Nodes (3): Consequences, Considered options, Sequences are web-native entities, outside the workbook

## Knowledge Gaps
- **140 isolated node(s):** `webprogfndb`, `composeSegments`, `exampleBody`, `exampleUpdateBody`, `orphanGroup` (+135 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 235 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **22 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Import()` connect `importer/export.go` to `testing.T`, `workbook.go`, `context.Context`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `Queries` connect `Queries` to `importer/export.go`, `links.go`, `orphans.go`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Why does `Export()` connect `importer/export.go` to `testing.T`, `context.Context`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Are the 43 inferred relationships involving `postJSON()` (e.g. with `compose1()` and `seedComposeFunction()`) actually correct?**
  _`postJSON()` has 43 INFERRED edges - model-reasoned connections that need verification._
- **Are the 43 inferred relationships involving `doRequest()` (e.g. with `TestComposeFieldsEndpoint()` and `TestComposeSaveAsExample()`) actually correct?**
  _`doRequest()` has 43 INFERRED edges - model-reasoned connections that need verification._
- **What connects `webprogfndb`, `composeSegments`, `exampleBody` to the rest of the system?**
  _140 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.06490737038517926 - nodes in this community are weakly interconnected._