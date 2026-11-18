# Graph Report - Web_Program_Functions_Database  (2026-10-06)

## Corpus Check
- 21 files · ~149,633 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1629 nodes · 3740 edges · 155 communities (107 shown, 48 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 286 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- SQLite Migration & Server CLI
- Device Workspace Pages
- Import Runs API
- Definition & Compose Dialogs
- Dictionary API Tests
- Export API Tests
- Svelte Best Practices Skill
- Compose Engine
- Link API & Validation
- Function Links Store
- Frontend Dev Dependencies
- Device Definitions Store
- Frontend Package Config
- Import Store
- Dropdown & Combobox UI
- App Shell & Breadcrumb UI
- Functions API
- Card UI Components
- Frontend Deploy Docs
- Dictionary Routes API
- Workbook Importer
- Frontend Devices Lib
- Devices API Tests
- Sequences API
- Examples API
- Postgres Baseline Migration
- Devices API
- Sequences Store
- Examples Store
- Workbook Importer Tests
- Frontend Sequences Lib
- Criteria Definitions Store
- Dictionary Test Helpers
- Blue-Green Deploy Scripts
- Frontend Package Scripts
- Frontend TSConfig
- Orphans API
- Functions API Tests
- Definition Delete Dialog
- Frontend Functions Client
- Component Aliases Config
- Storybook Figma Plugin
- Functions Store
- Parameter Definitions Store
- DB & Orphans Tests
- Server Health Tests
- Deploy Runbook Docs
- Frontend Examples Lib
- Dictionaries Migration
- Deploy Nginx & Releases
- Paraglide i18n Hooks
- Table Sort Utils
- Links Lib
- Filters API Tests
- Store DB Layer
- Backend README Ops
- Compose API Tests
- Frontend Compose Lib
- Frontend Imports Lib
- Deploy Host Prep
- Blue-Green Deploy Tests
- Svelte MCP Agent Config
- Import Migration
- Project READMEs
- Backend Ops & Query Memory
- ADRs & Devices E2E
- Tables E2E & Playwright Config
- Inlang Settings
- Tabs UI Components
- Definitions & Polish E2E
- Orphans Lib
- Vitest Unit Tests
- Dialogs E2E
- Agent Skills Docs
- ADR-0001 & Vite Config
- Device Function Detail
- Storybook Accessibility Assets
- Storybook Docs Assets
- Storybook Testing Assets
- Storybook Theming Assets
- Styling Bootstrap
- Chromatic Publishing
- Smoke E2E & ADR-0005
- Storybook Config
- A11y E2E
- Combobox E2E
- Examples E2E
- Selects E2E
- Tabs E2E
- Memory: Gitignore Query
- Memory: Test Folders Query
- Memory: Phase 1 Query
- Memory: Phase 2 Query
- Memory: Phase 3 Query
- Memory: Phase 4 Query
- Memory: Deploy Env Query
- Memory: Repo Rename Query
- Memory: Server Permission Query
- Functions Migration
- Sequences Migration
- Core UI Dependencies
- App Shell & Theme
- Favicon Assets
- Storybook Addon Library Assets
- Storybook Sample Assets
- Discord Brand Assets
- YouTube Brand Assets
- Device-Scoped Definitions Migration
- sqlc Config
- Deploy Compose Env
- English Locale Messages
- Indonesian Locale Messages
- Indonesian Locale Duplicates
- Traditional Chinese Locale
- OpenCode Plugin Config
- Prettier Config
- Storybook Context Assets
- GitHub Brand Assets
- Storybook Tutorial Assets
- Project Charter & Memory
- Devices Registry Migration
- App Type Definitions
- Layout SSR Config
- Accessibility Concept Asset
- Importer Definition Row
- Backend TLS Policy Doc
- Systemd Native Deployment
- Devices Test App
- Export Test App
- Criteria Links Schema
- Blue-Green Compose Stack
- Host Network Hardening
- Memory: Rename Query Record
- Parameter Links Schema
- WebProgFndb Package
- README Project Identity

## God Nodes (most connected - your core abstractions)
1. `postJSON()` - 57 edges
2. `Queries` - 55 edges
3. `doRequest()` - 55 edges
4. `serverError()` - 42 edges
5. `newLinksTestApp()` - 37 edges
6. `request()` - 35 edges
7. `seedFunctionAndParameterDefinition()` - 23 edges
8. `Queries` - 20 edges
9. `Svelte Core Best Practices` - 19 edges
10. `New()` - 18 edges

## Surprising Connections (you probably didn't know these)
- `playwright` --references--> `POST /api/import/reset`  [EXTRACTED]
  frontend/package.json → README.md
- `SvelteKit Browser Application` --semantically_similar_to--> `Svelte 5/SvelteKit Frontend`  [INFERRED] [semantically similar]
  frontend/README.md → README.md
- `playwright` --references--> `PostgreSQL`  [EXTRACTED]
  frontend/package.json → README.md
- `Ubuntu Blue-Green Deployment` --cites--> `Q: docker pull golang:1.27.1-alpine succeeded, what next`  [EXTRACTED]
  deploy/README.md → graphify-out/memory/query_20261004_182704_9b5a8c3c_docker_pull_golang_1_27_1_alpine_succeeded__what_t.md
- `Query record: reorganize test folders` --conceptually_related_to--> `Backend README - Operations Guide`  [INFERRED]
  graphify-out/memory/query_20261004_142052_401bf592_please_create_proper_test_folder_for_the_backend_a.md → backend/README.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Button example story group in sidebar** — frontend_src_stories_assets_docs_button_stories, frontend_src_stories_assets_docs_storybook_ui, frontend_src_stories_assets_docs_docs_page [EXTRACTED 0.75]
- **Storybook autodocs documentation illustration** — frontend_src_stories_assets_docs_png, frontend_src_stories_assets_docs_autodocs_tag, frontend_src_stories_assets_docs_csf_meta_button, frontend_src_stories_assets_docs_docs_page, frontend_src_stories_assets_docs_storybook_ui [EXTRACTED 0.85]
- **Async Svelte rendering pipeline** — frontend__github_skills_svelte_core_bestpractices_references_await_expressions_await_expressions, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_experimental_async_option, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_svelte_boundary, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_effect_pending, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_settled, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_fork_api, frontend__github_skills_svelte_core_bestpractices_references_hydratable_hydratable_api [EXTRACTED 1.00]
- **Publishing a Storybook to Chromatic for team review and collaboration** — frontend_src_stories_assets_share, publish_to_chromatic, chromatic, storybook [EXTRACTED 1.00]
- **Svelte $effect escape-hatch alternatives** — frontend__github_skills_svelte_core_bestpractices_skill_effect_rune, frontend__github_skills_svelte_core_bestpractices_references_attach_attachments, frontend__github_skills_svelte_core_bestpractices_references_bind_function_bindings, frontend__github_skills_svelte_core_bestpractices_references_inspect_inspect_rune, frontend__github_skills_svelte_core_bestpractices_references_svelte_reactivity_create_subscriber [EXTRACTED 1.00]
- **Shared Interaction Test Steps Across States** — frontend_src_stories_assets_testing_histogram_default_step, frontend_src_stories_assets_testing_pass_state_panel, frontend_src_stories_assets_testing_fail_state_panel [EXTRACTED 1.00]
- **Blue-Green Deployment Topology** — deploy_readme_nginx_edge, deploy_readme_blue_release, deploy_readme_green_release, deploy_readme_postgresql, deploy_readme_host_networking [EXTRACTED 1.00]
- **Isolated Docker Build Toolchain and Network Access** — deploy_readme_bun_frontend_builder, deploy_readme_go_builder, deploy_readme_go_toolchain, deploy_readme_build_network_access, deploy_readme_alpine_runtime_image [EXTRACTED 1.00]
- **SQLite to PostgreSQL Cutover Pipeline** — deploy_readme_sqlite_snapshot, deploy_readme_migrate_sqlite, deploy_readme_postgresql, deploy_readme_pg_dump_backup, deploy_readme_sqlite_migration_tag [EXTRACTED 1.00]
- **Scaffolded Storybook template asset set backing the Configure.mdx docs page** — frontend_src_stories_assets_share, frontend_src_stories_configure_mdx, storybook [EXTRACTED 1.00]
- **Svelte agent documentation and validation workflow** — frontend__github_agents_svelte_file_editor_agent_svelte_file_editor, frontend__github_skills_svelte_code_writer_skill_svelte_code_writer, frontend__github_skills_svelte_code_writer_skill_sveltejs_mcp_cli, frontend_agents_list_sections, frontend_agents_get_documentation, frontend_agents_svelte_autofixer [EXTRACTED 1.00]
- **Storybook Visual Context Illustration Set** — frontend_src_stories_assets_context_illustration, frontend_src_stories_assets_context_layered_stack_concept, frontend_src_stories_assets_context_storybook_asset_role [INFERRED 0.65]
- **Storybook Addon Library Showcase Collage** — frontend_src_stories_assets_addon_library_collage, frontend_src_stories_assets_addon_library_storybook_addon_ecosystem, frontend_src_stories_assets_addon_library_tool_logos, frontend_src_stories_assets_addon_library_grid_collage_layout [INFERRED 0.75]
- **Scaffolded Storybook Template Assets (GitHub Icon)** — frontend_src_stories_assets_github_github_logo, frontend_src_stories_assets_github_github_platform, frontend_src_stories_storybook_stories [INFERRED 0.75]
- **Open Book Icon Composition (two page panels plus center spine)** — frontend_src_stories_assets_tutorials_tutorials_icon, frontend_src_stories_assets_tutorials_open_book_metaphor, frontend_src_stories_assets_tutorials_storybook_ui_asset [INFERRED 0.75]
- **Discord Brand Asset Served via Storybook** — frontend_src_stories_assets_discord_svg_file, frontend_src_stories_assets_discord_discord_logo, frontend_src_stories_assets_discord_blurple_fill, frontend_src_stories_assets_discord_storybook_asset [INFERRED 0.85]
- **Figma-Storybook Design Handoff** — frontend_src_stories_assets_figma_plugin_figma, frontend_src_stories_configure_storybook, frontend_src_stories_assets_figma_plugin_storybook_connect, frontend_src_stories_assets_figma_plugin_design_handoff [INFERRED 0.85]
- **SvelteKit Project Branding (favicon renders Svelte logo as site identity)** — frontend_src_lib_assets_favicon_favicon, frontend_src_lib_assets_favicon_svelte_logo, frontend_src_lib_assets_favicon_svelte_framework [INFERRED 0.85]
- **Accessibility Audit Results Panel** — frontend_src_stories_assets_accessibility_image, frontend_src_stories_assets_accessibility_violations_tab, frontend_src_stories_assets_accessibility_passes_tab, frontend_src_stories_assets_accessibility_warnings_tab, frontend_src_stories_assets_accessibility_rule_list, frontend_src_stories_assets_accessibility_icon_badge [INFERRED 0.85]
- **Image-and-Text Asset Placeholder Composition** — frontend_src_stories_assets_assets, frontend_src_stories_assets_assets_image_icon, frontend_src_stories_assets_assets_text_t_icon [INFERRED 0.85]
- **Component Test Pass/Fail State Mockups** — frontend_src_stories_assets_testing_pass_state_panel, frontend_src_stories_assets_testing_fail_state_panel, frontend_src_stories_assets_testing_storybook_test_runner_ui [INFERRED 0.85]
- **Storybook Theming Demonstration (Light vs Dark)** — frontend_src_stories_assets_theming_theming_screenshot, frontend_src_stories_assets_theming_storybook_ui, frontend_src_stories_assets_theming_button_story, frontend_src_stories_assets_theming_dark_theme [INFERRED 0.85]
- **Styling Technology Showcase** — frontend_src_stories_assets_styling_image, ref_tailwindcss, material_ui, bootstrap, sass [INFERRED 0.85]
- **YouTube Play Button Glyph** — frontend_src_stories_assets_youtube_icon, frontend_src_stories_assets_youtube_youtube_brand, frontend_src_stories_assets_youtube_play_symbol [INFERRED 0.85]
- **Frontend Deployment Stack (Nginx + adapter-static + Go API)** — frontend_readme_nginx_conf, readme_adapter_static, readme_go_fiber_backend, readme_nginx [EXTRACTED 1.00]
- **E2E Testing Pipeline (Playwright + PostgreSQL + Go Server)** — ref_playwright, readme_postgresql, readme_go_fiber_backend, frontend_readme_e2e_suite, readme_api_import_reset [EXTRACTED 1.00]
- **PostgreSQL Cutover (ADR-0012 + migrations + sqlc + SQLite baseline)** — readme_adr0012, readme_postgres_migrations, readme_sqlc, readme_postgresql [EXTRACTED 1.00]

## Communities (155 total, 48 thin omitted)

### Community 0 - "SQLite Migration & Server CLI"
Cohesion: 0.05
Nodes (63): main(), run(), main(), exportWorkbook(), fiber.App, fiber.Handler, registerExportRoutes(), CurrentVersion() (+55 more)

### Community 1 - "Device Workspace Pages"
Cohesion: 0.04
Nodes (46): resolve(), chosen, critError, critGlobalTotal, critGlobalView, critLoaded, critLoading, critPage (+38 more)

### Community 2 - "Import Runs API"
Cohesion: 0.10
Nodes (31): getImportRun(), fiber.App, fiber.Handler, fiber.Map, importHandler(), importRunResponse(), listImportRuns(), registerImportRoutes() (+23 more)

### Community 4 - "Dictionary API Tests"
Cohesion: 0.16
Nodes (35): TestCriteriaDefinitionCRUD(), TestParameterDefinitionUpdateAndNameConflict(), TestParameterDefinitionUpdateScopeSemantics(), fiber.App, seedTwoFunctionsWithExamples(), TestExamplesCRUDRoundTrip(), TestExamplesDeviceParamPresentButEmpty(), TestExamplesFunctionFilterAndSearch() (+27 more)

### Community 5 - "Export API Tests"
Cohesion: 0.15
Nodes (35): compareRows(), exportWorkbookBytes(), fiber.App, importFixture(), TestExportEmptyDatabase(), TestExportIncludesOrphans(), TestExportReimportIdempotent(), TestExportRoundTripCellByCell() (+27 more)

### Community 6 - "Svelte Best Practices Skill"
Cohesion: 0.09
Nodes (35): Actions (use:action), Attachment factories, Attachments ({@attach}), createAttachmentKey, fromAction, Await expressions, $effect.pending(), experimental.async option (+27 more)

### Community 7 - "Compose Engine"
Cohesion: 0.14
Nodes (31): composeField, composeRequest, composeSegments, checkAllowed(), checkType(), composeFieldFromCriteriaRow(), composeFieldFromRow(), composeFieldsFor() (+23 more)

### Community 8 - "Link API & Validation"
Cohesion: 0.14
Nodes (31): httpError, insertLinkParams, linkBody, linkStore, linkView, updateLinkAttrs, boolToInt64(), fiber.App (+23 more)

### Community 9 - "Function Links Store"
Cohesion: 0.11
Nodes (16): Queries, CriteriaLink, ParameterLink, database/sql.NullInt64, GetCriteriaLinkRow, GetParameterLinkRow, InsertCriteriaLinkParams, InsertParameterLinkParams (+8 more)

### Community 10 - "Frontend Dev Dependencies"
Cohesion: 0.06
Nodes (31): devDependencies, eslint, eslint-config-prettier, @eslint/js, eslint-plugin-svelte, globals, @inlang/paraglide-js, @lucide/svelte (+23 more)

### Community 11 - "Device Definitions Store"
Cohesion: 0.11
Nodes (8): Queries, Queries, Queries, Queries, Device, context.Context, ListOrphanParameterLinksRow, Queries

### Community 12 - "Frontend Package Config"
Cohesion: 0.08
Nodes (28): gitignorePath, name, packageManager, private, type, version, clsx, eslint (+20 more)

### Community 13 - "Import Store"
Cohesion: 0.09
Nodes (11): Queries, UpsertCriteriaDefinitionParams, UpsertParameterDefinitionParams, ImportNormalization, ImportRun, InsertImportNormalizationParams, InsertImportRunParams, UpsertCriteriaLinkParams (+3 more)

### Community 15 - "App Shell & Breadcrumb UI"
Cohesion: 0.08
Nodes (4): frontend_src_app, frontend_src_lib_assets_favicon, density, $app/types

### Community 16 - "Functions API"
Cohesion: 0.15
Nodes (22): functionBody, listQuery, conflict(), createFunction(), deleteFunction(), functionFacets(), functionResponse(), getFunction() (+14 more)

### Community 17 - "Card UI Components"
Cohesion: 0.16
Nodes (5): WithElementRef, WithoutChild, WithoutChildren, WithoutChildrenOrChild, stepEmit

### Community 18 - "Frontend Deploy Docs"
Cohesion: 0.10
Nodes (27): Device Workspace, Full E2E Test Suite, .env.example, nginx.conf (frontend), polish.spec.ts (UI test suite), Same-origin /api pattern, SvelteKit Browser Application, Vite API Proxy (/api → 127.0.0.1:8080) (+19 more)

### Community 19 - "Dictionary Routes API"
Cohesion: 0.18
Nodes (19): definitionBody, dictionaryRoutes, dictionaryRoutes[T], fieldValidationError, criteriaDefinitionResponse(), definitionResponseMap(), definitionWriteError(), fiber.Ctx (+11 more)

### Community 20 - "Workbook Importer"
Cohesion: 0.18
Nodes (23): cell(), DefinitionRow, Workbook, indexHeader(), normalizePhase(), parseBool(), parseDefinitions(), parseExamples() (+15 more)

### Community 21 - "Frontend Devices Lib"
Cohesion: 0.11
Nodes (22): createDevice(), deleteDevice(), deviceBadgeLinkClass, deviceCountLinkClass, DeviceDefinitionKind, DeviceDefinitionRow, DeviceDefinitionsResult, DeviceFunctionRow (+14 more)

### Community 22 - "Devices API Tests"
Cohesion: 0.22
Nodes (21): deviceFunctionRow, deviceSummary, registryDevice, registerDevice(), registryScopeCounts(), seedDeviceData(), TestDeviceDetailDefinitions(), TestDeviceDetailFunctions() (+13 more)

### Community 23 - "Sequences API"
Cohesion: 0.25
Nodes (20): sequenceBody, sequenceResponse, sequenceStepBody, sequenceStepResponse, createSequence(), deleteSequence(), emitSequence(), getSequence() (+12 more)

### Community 24 - "Examples API"
Cohesion: 0.18
Nodes (19): exampleBody, exampleResponse, exampleUpdateBody, normalizeDeviceName(), createExample(), deleteExample(), exampleFilters(), exampleItem() (+11 more)

### Community 25 - "Postgres Baseline Migration"
Cohesion: 0.15
Nodes (20): criteria_definitions, criteria_links, devices, examples, functions, idx_criteria_links_function, idx_criteria_links_name, idx_examples_function (+12 more)

### Community 26 - "Devices API"
Cohesion: 0.23
Nodes (19): definitionItem, deviceBody, createDevice(), criteriaItem(), deleteDevice(), deviceCriteria(), deviceFromSegment(), deviceFunctions() (+11 more)

### Community 27 - "Sequences Store"
Cohesion: 0.13
Nodes (9): ListSequenceStepsRow, Queries, SequenceStep, GetFunctionByNaturalKeyParams, InsertSequenceParams, InsertSequenceStepParams, ListSequencesRow, Sequence (+1 more)

### Community 28 - "Examples Store"
Cohesion: 0.15
Nodes (9): ListExamplesParams, ListExamplesRow, Queries, Example, CountExamplesParams, FunctionNaturalKeyExistsParams, GetExampleRow, InsertExampleParams (+1 more)

### Community 29 - "Workbook Importer Tests"
Cohesion: 0.21
Nodes (15): buildSyntheticWorkbook(), openFixture(), TestParseWorkbookNormalizations(), TestParseWorkbookRealFixture(), writeAndOpen(), writeWorkbook(), CountsFrom(), FixtureCounts() (+7 more)

### Community 30 - "Frontend Sequences Lib"
Cohesion: 0.15
Nodes (15): request(), createSequence(), EmitResult, emitSequence(), getSequence(), listSequences(), SequenceDetail, SequenceInput (+7 more)

### Community 31 - "Criteria Definitions Store"
Cohesion: 0.16
Nodes (6): Queries, CriteriaDefinition, ImportRun, Device, ImportNormalization, SequenceStep

### Community 32 - "Dictionary Test Helpers"
Cohesion: 0.26
Nodes (14): extractItems(), fiber.App, newDictionaryTestApp(), TestCreateParameterDefinitionRoundTrip(), TestCriteriaDefinitionValidation(), TestDictionaryListPaginationAndSearch(), TestDictionarySortUnknownColumnDoesNotPanic(), TestDictionaryUsageCountAndDeleteGuard() (+6 more)

### Community 33 - "Blue-Green Deploy Scripts"
Cohesion: 0.30
Nodes (14): active_color(), compose(), fail(), finish(), owned_running(), require_free_port(), blue-green.sh script, slot_port() (+6 more)

### Community 34 - "Frontend Package Scripts"
Cohesion: 0.13
Nodes (15): scripts, build, build-storybook, check, check:watch, dev, dev:device-prototype, format (+7 more)

### Community 35 - "Frontend TSConfig"
Cohesion: 0.13
Nodes (14): compilerOptions, allowJs, checkJs, esModuleInterop, forceConsistentCasingInFileNames, moduleResolution, resolveJsonModule, rewriteRelativeImportExtensions (+6 more)

### Community 36 - "Orphans API"
Cohesion: 0.29
Nodes (13): orphanExampleItem, orphanGroup, orphanLinkItem, fiber.App, fiber.Ctx, fiber.Handler, nullableAny(), nullablePosition() (+5 more)

### Community 37 - "Functions API Tests"
Cohesion: 0.31
Nodes (13): fiber.App, newFunctionsTestApp(), seedFunctions(), TestCreateFunctionDuplicateKeyRejected(), TestCreateFunctionRoundTrip(), TestCreateFunctionValidationFieldErrors(), TestFunctionCRUDRoundTrip(), TestListFunctionsFilter() (+5 more)

### Community 38 - "Definition Delete Dialog"
Cohesion: 0.15
Nodes (9): ADR-0003, deleting, error, impact, remove(), if(), createDefinition(), DefinitionListResult (+1 more)

### Community 39 - "Frontend Functions Client"
Cohesion: 0.15
Nodes (12): createFunction(), FunctionFacets, FunctionInput, FunctionRow, getFunction(), getFunctionFacets(), listFunctions(), ListParams (+4 more)

### Community 40 - "Component Aliases Config"
Cohesion: 0.15
Nodes (12): aliases, components, hooks, lib, ui, utils, registry, $schema (+4 more)

### Community 41 - "Storybook Figma Plugin"
Cohesion: 0.19
Nodes (13): Design-to-Code Handoff Workflow, Figma Design Tool, Figma Plugin Screenshot (Storybook Connect), Price Range Histogram Component Preview, Storybook Connect Figma Plugin Panel, Storybook accessibility testing, Storybook Autodocs, Configure your project (Storybook docs page) (+5 more)

### Community 42 - "Functions Store"
Cohesion: 0.23
Nodes (4): cmpBy(), Queries, Queries, Function

### Community 43 - "Parameter Definitions Store"
Cohesion: 0.21
Nodes (4): ParameterDefinition, Queries, CountParameterUsageScopedParams, InsertParameterDefinitionParams

### Community 44 - "DB & Orphans Tests"
Cohesion: 0.26
Nodes (9): openTemp(), TestCurrentVersionOnFreshDatabase(), TestFunctionsTableExistsWithNaturalKey(), TestMigrateCreatesSchemaVersionTable(), TestMigrateIsIdempotent(), TestPostgresRejectsPlaintextFallbackWithoutEchoingCredentials(), go_pkg_strings, go_pkg_testing (+1 more)

### Community 45 - "Server Health Tests"
Cohesion: 0.39
Nodes (11): doRequest(), fiber.App, newTestApp(), TestHealthReportsDatabaseState(), TestHealthReturnsOK(), TestRegisterStaticServesRootAndFallsBack(), TestRegisterStaticWithoutBundleReturnsFalse(), TestUnknownAPIRouteReturns404JSON() (+3 more)

### Community 46 - "Deploy Runbook Docs"
Cohesion: 0.17
Nodes (12): 2. Configure Credentials, 3. Transfer and Verify SQLite First, 4. Deploy the First Release, Layout, Routing-only Rollback, Routine Releases and Rollback, Ubuntu Blue-Green Deployment, Verification Boundaries (+4 more)

### Community 47 - "Frontend Examples Lib"
Cohesion: 0.17
Nodes (10): if(), createExample(), ExampleInput, ExampleListParams, ExampleListResult, ExampleRow, ExampleUpdateInput, listExamples() (+2 more)

### Community 48 - "Dictionaries Migration"
Cohesion: 0.27
Nodes (10): criteria_definitions, criteria_links, examples, idx_criteria_links_function, idx_criteria_links_name, idx_examples_function, idx_parameter_links_function, idx_parameter_links_name (+2 more)

### Community 49 - "Deploy Nginx & Releases"
Cohesion: 0.24
Nodes (11): Blue Release Slot (127.0.0.1:24993), Go 1.27.1 Toolchain (golang:1.27.1-alpine), Green Release Slot (127.0.0.1:24994), Linux Host Networking, master_control_functions_test Disposable Test Database, migrate-sqlite transfer tool (cmd/migrate-sqlite), Nginx Edge Listener (port 31627), pg_dump Pre-migration Backup (program-functions-before.dump) (+3 more)

### Community 50 - "Paraglide i18n Hooks"
Cohesion: 0.20
Nodes (7): handle, frontend_src_lib_paraglide_runtime, frontend_src_lib_paraglide_runtime_delocalizeurl, frontend_src_lib_paraglide_runtime_gettextdirection, frontend_src_lib_paraglide_server, frontend_src_lib_paraglide_server_paraglidemiddleware, @sveltejs/kit

### Community 51 - "Table Sort Utils"
Cohesion: 0.22
Nodes (5): compare(), SORT_BUTTON_CLASS, SortDir, sortRows(), SortValue

### Community 52 - "Links Lib"
Cohesion: 0.20
Nodes (8): DefinitionInput, attachLink(), LinkAttachInput, LinkKind, LinkRow, LinkUpdateInput, ADR-0002, updateLink()

### Community 53 - "Filters API Tests"
Cohesion: 0.47
Nodes (9): listBody, fiber.App, itoa(), listBodyAt(), seedFilterFunctions(), TestFunctionEnumFilters(), TestFunctionFacets(), TestFunctionFilterComposesWithSearchSortPaging() (+1 more)

### Community 54 - "Store DB Layer"
Cohesion: 0.22
Nodes (6): New(), database/sql.Result, database/sql.Row, database/sql.Stmt, Queries, DBTX

### Community 55 - "Backend README Ops"
Cohesion: 0.20
Nodes (10): Backend, Backend Nginx, Backend Service, Build And Secrets, Database Configuration, Local Build And Development, Requirements, Tests (+2 more)

### Community 56 - "Compose API Tests"
Cohesion: 0.47
Nodes (9): compose1(), decodeScript(), fiber.App, seedComposeFunction(), TestComposeFieldsEndpoint(), TestComposeKnownFunctionScript(), TestComposeNoLinksFunction(), TestComposeSaveAsExample() (+1 more)

### Community 57 - "Frontend Compose Lib"
Cohesion: 0.22
Nodes (6): ApiError, ComposeField, ComposeFields, ComposeRequest, composeScript(), getComposeFields()

### Community 58 - "Frontend Imports Lib"
Cohesion: 0.28
Nodes (8): getImportRun(), ImportNormalization, importReset(), ImportResult, ImportRun, importWorkbook(), listImportRuns(), upload()

### Community 59 - "Deploy Host Prep"
Cohesion: 0.32
Nodes (8): 1. Prepare the Host, Non-root Alpine Runtime Image, Build Network Access, Bun Frontend Builder (Docker build stage), Go Builder (PostgreSQL-client-only server), Backend-only GOPROXY ARG with proxy forwarding, Q: Go proxy/checksum TLS resets, GitHub OK, cached Go image missing, Host egress/TLS filtering on Go module endpoints

### Community 60 - "Blue-Green Deploy Tests"
Cohesion: 0.29
Nodes (7): deployment, runDeployment(), shellPath(), ref_node_assert, ref_node_child_process, ref_node_os, ref_node_test

### Community 61 - "Svelte MCP Agent Config"
Cohesion: 0.46
Nodes (8): svelte-file-editor agent, svelte-code-writer skill, @sveltejs/mcp CLI, get-documentation tool, list-sections tool, playground-link tool, svelte-autofixer tool, Svelte MCP Server

### Community 62 - "Import Migration"
Cohesion: 0.38
Nodes (6): examples_new, idx_examples_function, idx_import_normalizations_run, import_normalizations, import_runs, examples

### Community 63 - "Project READMEs"
Cohesion: 0.33
Nodes (5): Frontend README, Building, Creating a project, Developing, sv

### Community 64 - "Backend Ops & Query Memory"
Cohesion: 0.29
Nodes (7): Backend README - Operations Guide, Query record: reorganize test folders, Query record: Phase 1 test-layout baseline, Query record: Phase 2 backend test relocation, Query record: Phase 3 frontend test relocation, Query record: Phase 4 env templates + UI gate, Query record: env files, nginx, READMEs for Ubuntu deploy

### Community 65 - "ADRs & Devices E2E"
Cohesion: 0.29
Nodes (6): ADR-0002, ADR-0009, ADR-0010, DeviceSummary, E2E_DIR, FIXTURE

### Community 66 - "Tables E2E & Playwright Config"
Cohesion: 0.29
Nodes (3): E2E_DIR, FIXTURE, @playwright/test

### Community 67 - "Inlang Settings"
Cohesion: 0.29
Nodes (6): baseLocale, locales, modules, plugin.inlang.messageFormat, pathPattern, $schema

### Community 68 - "Tabs UI Components"
Cohesion: 0.29
Nodes (5): Content, List, Root, Trigger, bits-ui

### Community 69 - "Definitions & Polish E2E"
Cohesion: 0.38
Nodes (5): DefinitionKind, DefinitionRow, mockDeviceHub(), mockDictionary(), mockFunctionCatalog()

### Community 70 - "Orphans Lib"
Cohesion: 0.29
Nodes (6): listOrphans(), OrphanExampleRow, OrphanGroup, OrphanKind, OrphanLinkRow, OrphanReport

### Community 71 - "Vitest Unit Tests"
Cohesion: 0.38
Nodes (3): greet(), vitest, vitest-browser-svelte

### Community 73 - "Agent Skills Docs"
Cohesion: 0.33
Nodes (5): Agent skills, Domain docs, Issue tracker, Triage labels, Web_Program_Functions_Database

### Community 74 - "ADR-0001 & Vite Config"
Cohesion: 0.33
Nodes (5): ADR-0001, @inlang/paraglide-js, @sveltejs/adapter-static, @tailwindcss/vite, @vitest/browser-playwright

### Community 75 - "Device Function Detail"
Cohesion: 0.33
Nodes (3): byId, content, resolved

### Community 76 - "Storybook Accessibility Assets"
Cohesion: 0.47
Nodes (6): Accessibility Person Icon Badge, Accessibility Checker Results Screenshot, Passes Count (11 Passes, active tab), A11y Rule Checklist (button text, unique ids, heading text), Violations Count (2 Violations), Warnings Count (0 Warnings)

### Community 77 - "Storybook Docs Assets"
Cohesion: 0.53
Nodes (6): tags: ['autodocs'] CSF tag, Button story variants (Primary, Secondary, Large, Small), Meta<Button> CSF3 meta object, Storybook Docs page, docs.png - Storybook autodocs illustration, Storybook UI (sidebar, controls, args table)

### Community 78 - "Storybook Testing Assets"
Cohesion: 0.53
Nodes (6): Testing.png Illustration (Test Results UI Mockup), FAIL State Test Results Panel, Test Step: Histogram Default, PASS State Test Results Panel, Test Playback Controls (rewind, step, replay), Storybook Test Runner UI Concept

### Community 79 - "Storybook Theming Assets"
Cohesion: 0.47
Nodes (6): Acme Brand Namespace (Sidebar Header), Button Story (Example Section), Dark Theme Variant, Light Theme Variant, Storybook UI, Storybook Theming Screenshot (theming.png)

### Community 80 - "Styling Bootstrap"
Cohesion: 0.40
Nodes (5): Bootstrap, Styling Showcase Image (styling.png), Material UI (MUI), tailwindcss, Sass

### Community 81 - "Chromatic Publishing"
Cohesion: 0.80
Nodes (5): Chromatic — Storybook hosting/collaboration service (host of the chromatic.com URL shown in the screenshot), share.png — screenshot of a browser viewing a Storybook published to a chromatic.com URL (sidebar with Storybook logo, Find components search, EXAMPLE/Button; canvas showing a blue Button), Configure.mdx — Storybook 'Configure your project' onboarding docs page that imports share.png, Publish to Chromatic — 'Do more with Storybook' docs feature section illustrated by share.png, Storybook — component workshop/workbench tool whose UI (manager sidebar + canvas) is captured in the screenshot

### Community 82 - "Smoke E2E & ADR-0005"
Cohesion: 0.40
Nodes (3): ADR-0005, E2E_DIR, FIXTURE

### Community 83 - "Storybook Config"
Cohesion: 0.40
Nodes (3): config, preview, ref_storybook_sveltekit

### Community 84 - "A11y E2E"
Cohesion: 0.40
Nodes (3): E2E_DIR, FIXTURE, ref_node_url

### Community 86 - "Examples E2E"
Cohesion: 0.40
Nodes (3): E2E_DIR, FIXTURE, ref_node_path

### Community 87 - "Selects E2E"
Cohesion: 0.40
Nodes (3): E2E_DIR, FIXTURE, ref_node_fs

### Community 89 - "Memory: Gitignore Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: I want to commit the project Web_Program_Functions_Database to github. Please help me to update the .gitignore file and make sure that files for AI are not commited to the github (excepts agents and the skills)., Source Nodes

### Community 90 - "Memory: Test Folders Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: please create proper test folder for the backend and frontend folder and move all test related files accordingly, Source Nodes

### Community 91 - "Memory: Phase 1 Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: start with Phase 1, Source Nodes

### Community 92 - "Memory: Phase 2 Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: Please continue to Phase 2., Source Nodes

### Community 93 - "Memory: Phase 3 Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: continue with Phase 3, Source Nodes

### Community 94 - "Memory: Phase 4 Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: please continue to Phase 4., Source Nodes

### Community 95 - "Memory: Deploy Env Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: because I will be deploying the backend and frontend to the ubuntu server and the postgresql database is hosted on that server too, please help to update the environment files to setup the database configuration if it hosted on the ubuntu server (using localhost instead of server ip address). Please also create the nginx configurations for each backend (radom port) and frontend (random port). After that please help to update the backend and frontend README.md files on how to setup and run the projects locally and deployed in on the server., Source Nodes

### Community 96 - "Memory: Repo Rename Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: I rename the project in github repo to master-control-functions. Please help to update related files., Source Nodes

### Community 97 - "Memory: Server Permission Query"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: Please check why I got this issue on the server: Protect deploy/.env with chmod 600., Source Nodes

### Community 98 - "Functions Migration"
Cohesion: 0.83
Nodes (3): functions, idx_functions_interface, idx_functions_name

### Community 99 - "Sequences Migration"
Cohesion: 0.83
Nodes (3): idx_sequence_steps_sequence, sequence_steps, sequences

### Community 100 - "Core UI Dependencies"
Cohesion: 0.50
Nodes (4): dependencies, bits-ui, clsx, tailwind-merge

### Community 101 - "App Shell & Theme"
Cohesion: 0.50
Nodes (4): SvelteKit app shell (app.html), Paraglide i18n placeholders, Theme pre-paint script, Robots crawl policy (allow all)

### Community 102 - "Favicon Assets"
Cohesion: 0.67
Nodes (4): Browser Tab / Site Identity Asset, Favicon SVG (frontend/src/lib/assets/favicon.svg), Svelte Framework, Svelte Logo Graphic (orange #ff3e00 S-curve mark)

### Community 103 - "Storybook Addon Library Assets"
Cohesion: 0.67
Nodes (4): Addon Library Icon Collage (Storybook Stories Asset), Rounded-Tile Grid Collage Layout, Storybook Addon Ecosystem, Frontend Tooling and Framework Logos

### Community 104 - "Storybook Sample Assets"
Cohesion: 0.67
Nodes (4): Storybook Assets Placeholder Image (assets.png), Teal Picture/Photo Icon (image placeholder symbol), Storybook Sample Assets Directory Convention, Green Serif 'T' with Vertical Cursor (text placeholder symbol)

### Community 105 - "Discord Brand Assets"
Cohesion: 0.50
Nodes (4): Discord Blurple Brand Color (#5865F2), Discord Logo Mark, Storybook Static Asset, Discord Logo SVG Asset

### Community 106 - "YouTube Brand Assets"
Cohesion: 0.67
Nodes (4): YouTube Icon (SVG), Play Button Symbol, Storybook Story Asset, YouTube Brand Logo

### Community 108 - "sqlc Config"
Cohesion: 0.67
Nodes (3): sqlc.yaml (backend), internal/db/postgres_migrations (sqlc schema dir), internal/store/queries (sqlc queries dir)

### Community 109 - "Deploy Compose Env"
Cohesion: 0.67
Nodes (3): deploy/compose.yml, deploy/.env (credentials), deploy/.env.proxy (build proxy settings)

### Community 116 - "Storybook Context Assets"
Cohesion: 1.00
Nodes (3): Context Layered Stack Illustration, Layered Context / Data Stack Concept, Storybook Decorative Asset Role

### Community 117 - "GitHub Brand Assets"
Cohesion: 0.67
Nodes (3): GitHub Octocat Logo Icon (github.svg), GitHub (Code Hosting Platform), SvelteKit Scaffolded Storybook Stories Assets

### Community 118 - "Storybook Tutorial Assets"
Cohesion: 0.67
Nodes (3): Open Book / Learning Metaphor, Storybook UI Asset, Tutorials Icon (Open Book SVG)

## Ambiguous Edges - Review These
- `Layered Context / Data Stack Concept` → `Storybook Decorative Asset Role`  [AMBIGUOUS]
  frontend/src/stories/assets/context.png · relation: conceptually_related_to

## Knowledge Gaps
- **370 isolated node(s):** `DeviceDefinitionKind`, `DeviceDefinitionRow`, `DeviceFunctionRow`, `DeviceListResult`, `DeviceSummary` (+365 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 552 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **48 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Layered Context / Data Stack Concept` and `Storybook Decorative Asset Role`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `@playwright/test` connect `Tables E2E & Playwright Config` to `ADRs & Devices E2E`, `Definitions & Polish E2E`, `Dialogs E2E`, `Frontend Package Config`, `Frontend Deploy Docs`, `Smoke E2E & ADR-0005`, `A11y E2E`, `Combobox E2E`, `Examples E2E`, `Selects E2E`, `Tabs E2E`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Why does `Migrate()` connect `SQLite Migration & Server CLI` to `Device Definitions Store`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Why does `devDependencies` connect `Frontend Dev Dependencies` to `Frontend Package Config`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **Are the 46 inferred relationships involving `postJSON()` (e.g. with `compose1()` and `seedComposeFunction()`) actually correct?**
  _`postJSON()` has 46 INFERRED edges - model-reasoned connections that need verification._
- **Are the 46 inferred relationships involving `doRequest()` (e.g. with `exportWorkbookBytes()` and `listBodyAt()`) actually correct?**
  _`doRequest()` has 46 INFERRED edges - model-reasoned connections that need verification._
- **What connects `DeviceDefinitionKind`, `DeviceDefinitionRow`, `DeviceFunctionRow` to the rest of the system?**
  _370 weakly-connected nodes found - possible documentation gaps or missing edges._