# Graph Report - Web_Program_Functions_Database  (2026-10-05)

## Corpus Check
- 237 files · ~129,969 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 19 file(s) not represented in the graph (top: (none) 7, .css 5, .example 3)

## Summary
- 1481 nodes · 3694 edges · 118 communities (81 shown, 37 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 320 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1287f513`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- utils.ts
- import_test.go
- functions.go
- @playwright/test
- blue-green.sh
- import.go
- devices.spec.ts
- testing.T
- links.go
- database/sql.NullString
- Svelte Core Best Practices
- dialogs.spec.ts
- compose.go
- function_links.sql.go
- context.Context
- package.json
- devDependencies
- workbook.go
- Configure your project (Storybook docs page)
- devices.ts
- 0006_baseline.sql
- Function
- sequences.sql.go
- +layout.svelte
- ref_node_path
- combobox.spec.ts
- examples.sql.go
- client.ts
- request
- import.sql.go
- api.ts
- AGENTS.md
- ref_node_fs
- ref_node_url
- smoke.spec.ts
- tabs.spec.ts
- scripts
- compilerOptions
- importer/export.go
- Q: I want to commit the project Web_Program_Functions_Database to github. Please help me to update the .gitignore file and make sure that files for AI are not commited to the github (excepts agents and the skills).
- aliases
- orphans.go
- Q: please create proper test folder for the backend and frontend folder and move all test related files accordingly
- definitions.ts
- examples.ts
- 0002_dictionaries.sql
- hooks.server.ts
- links.ts
- Q: start with Phase 1
- examples.go
- importHandler
- Q: Please continue to Phase 2.
- imports.ts
- Q: continue with Phase 3
- Q: please continue to Phase 4.
- README.md
- 0003_import.sql
- models.go
- Q: because I will be deploying the backend and frontend to the ubuntu server and the postgresql database is hosted on that server too, please help to update the environment files to setup the database configuration if it hosted on the ubuntu server (using localhost instead of server ip address). Please also create the nginx configurations for each backend (radom port) and frontend (random port). After that please help to update the backend and frontend README.md files on how to setup and run the projects locally and deployed in on the server.
- settings.json
- vitest
- vite.config.ts
- Accessibility Checker Results Screenshot
- Storybook UI (sidebar, controls, args table)
- PASS State Test Results Panel
- Storybook Theming Screenshot (theming.png)
- Queries
- Styling Showcase Image (styling.png)
- share.png — screenshot of a browser viewing a Storybook published to a chromatic.com URL (sidebar with Storybook logo, Find components search, EXAMPLE/Button; canvas showing a blue Button)
- main.ts
- devices.go
- 0001_functions.sql
- 0004_sequences.sql
- dependencies
- Browser Tab / Site Identity Asset
- Addon Library Icon Collage (Storybook Stories Asset)
- Storybook Assets Placeholder Image (assets.png)
- Discord Logo Mark
- YouTube Icon (SVG)
- 0005_device_scoped_definitions.sql
- DBTX
- .ListDeviceScopes
- en.json
- id.json
- id-id.json
- zh-hant.json
- opencode.json
- prettier.config.js
- Context Layered Stack Illustration
- GitHub Octocat Logo Icon (github.svg)
- Tutorials Icon (Open Book SVG)
- 0006_devices_registry.sql
- app.d.ts
- ssr
- Accessibility (universal access symbol: person with outstretched arms in a circle, purple/violet palette #A849FF / #A470D5)
- criteria_links
- parameter_links
- webprogfndb
- blue-green.test.mjs

## God Nodes (most connected - your core abstractions)
1. `postJSON()` - 68 edges
2. `doRequest()` - 64 edges
3. `Queries` - 55 edges
4. `request()` - 46 edges
5. `serverError()` - 42 edges
6. `newLinksTestApp()` - 40 edges
7. `newDictionaryTestApp()` - 29 edges
8. `New()` - 24 edges
9. `seedFunctionAndParameterDefinition()` - 23 edges
10. `Queries` - 20 edges

## Surprising Connections (you probably didn't know these)
- `sqlc Code Generation Workflow` --references--> `sqlc.yaml (backend)`  [INFERRED]
  README.md → backend/sqlc.yaml
- `internal/db/postgres_migrations (sqlc schema dir)` --shares_data_with--> `PostgreSQL Migrations (schema_migrations, baseline v6)`  [INFERRED]
  backend/sqlc.yaml → README.md
- `emitSequence()` --calls--> `validateAndEmit()`  [INFERRED]
  backend/internal/api/sequences.go → backend/internal/api/compose.go
- `New()` --calls--> `registerComposeRoutes()`  [INFERRED]
  backend/internal/api/server.go → backend/internal/api/compose.go
- `loadFunctionOr404()` --calls--> `serverError()`  [INFERRED]
  backend/internal/api/compose.go → backend/internal/api/functions.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Button example story group in sidebar** — frontend_src_stories_assets_docs_button_stories, frontend_src_stories_assets_docs_storybook_ui, frontend_src_stories_assets_docs_docs_page [EXTRACTED 0.75]
- **Storybook autodocs documentation illustration** — frontend_src_stories_assets_docs_png, frontend_src_stories_assets_docs_autodocs_tag, frontend_src_stories_assets_docs_csf_meta_button, frontend_src_stories_assets_docs_docs_page, frontend_src_stories_assets_docs_storybook_ui [EXTRACTED 0.85]
- **Async Svelte rendering pipeline** — frontend__github_skills_svelte_core_bestpractices_references_await_expressions_await_expressions, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_experimental_async_option, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_svelte_boundary, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_effect_pending, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_settled, frontend__github_skills_svelte_core_bestpractices_references_await_expressions_fork_api, frontend__github_skills_svelte_core_bestpractices_references_hydratable_hydratable_api [EXTRACTED 1.00]
- **Publishing a Storybook to Chromatic for team review and collaboration** — frontend_src_stories_assets_share, publish_to_chromatic, chromatic, storybook [EXTRACTED 1.00]
- **Svelte $effect escape-hatch alternatives** — frontend__github_skills_svelte_core_bestpractices_skill_effect_rune, frontend__github_skills_svelte_core_bestpractices_references_attach_attachments, frontend__github_skills_svelte_core_bestpractices_references_bind_function_bindings, frontend__github_skills_svelte_core_bestpractices_references_inspect_inspect_rune, frontend__github_skills_svelte_core_bestpractices_references_svelte_reactivity_create_subscriber [EXTRACTED 1.00]
- **Shared Interaction Test Steps Across States** — frontend_src_stories_assets_testing_histogram_default_step, frontend_src_stories_assets_testing_pass_state_panel, frontend_src_stories_assets_testing_fail_state_panel [EXTRACTED 1.00]
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

## Communities (118 total, 37 thin omitted)

### Community 0 - "utils.ts"
Cohesion: 0.05
Nodes (19): $schema, children(), Content, List, Root, Trigger, cn(), compare() (+11 more)

### Community 1 - "import_test.go"
Cohesion: 0.05
Nodes (93): main(), run(), main(), exportWorkbook(), fiber.App, fiber.Handler, registerExportRoutes(), CurrentVersion() (+85 more)

### Community 2 - "functions.go"
Cohesion: 0.23
Nodes (19): functionBody, listQuery, conflict(), createFunction(), deleteFunction(), functionFacets(), functionResponse(), getFunction() (+11 more)

### Community 3 - "@playwright/test"
Cohesion: 0.22
Nodes (3): E2E_DIR, FIXTURE, @playwright/test

### Community 4 - "blue-green.sh"
Cohesion: 0.44
Nodes (11): active_color(), compose(), fail(), finish(), owned_running(), require_free_port(), blue-green.sh script, slot_port() (+3 more)

### Community 5 - "import.go"
Cohesion: 0.14
Nodes (19): boolInt(), criteriaDefinitionParams(), DefinitionRow, Import(), nullInt64(), nullString(), parameterDefinitionParams(), go_pkg_context (+11 more)

### Community 6 - "devices.spec.ts"
Cohesion: 0.29
Nodes (6): DeviceSummary, E2E_DIR, FIXTURE, ADR-0002, ADR-0009, ADR-0010

### Community 7 - "testing.T"
Cohesion: 0.07
Nodes (113): deviceFunctionRow, deviceSummary, listBody, registryDevice, fiber.App, RegisterStatic(), compose1(), decodeScript() (+105 more)

### Community 8 - "links.go"
Cohesion: 0.09
Nodes (48): definitionBody, dictionaryRoutes, dictionaryRoutes[T], fieldValidationError, httpError, insertLinkParams, linkBody, linkStore (+40 more)

### Community 9 - "database/sql.NullString"
Cohesion: 0.12
Nodes (8): Queries, UpsertCriteriaDefinitionParams, CriteriaDefinition, ParameterDefinition, Queries, database/sql.NullString, InsertParameterDefinitionParams, UpdateParameterDefinitionParams

### Community 10 - "Svelte Core Best Practices"
Cohesion: 0.09
Nodes (35): Actions (use:action), Attachment factories, Attachments ({@attach}), createAttachmentKey, fromAction, Await expressions, $effect.pending(), experimental.async option (+27 more)

### Community 12 - "compose.go"
Cohesion: 0.14
Nodes (30): composeField, composeRequest, composeSegments, checkAllowed(), checkType(), composeFieldFromCriteriaRow(), composeFieldFromRow(), composeFieldsFor() (+22 more)

### Community 13 - "function_links.sql.go"
Cohesion: 0.10
Nodes (14): Queries, CriteriaLink, GetCriteriaLinkRow, GetParameterLinkRow, InsertCriteriaLinkParams, InsertParameterLinkParams, ListCriteriaLinksByFunctionParams, ListCriteriaLinksByFunctionRow (+6 more)

### Community 14 - "context.Context"
Cohesion: 0.14
Nodes (6): Queries, Queries, Queries, Queries, context.Context, ListOrphanCriteriaLinksRow

### Community 15 - "package.json"
Cohesion: 0.08
Nodes (29): gitignorePath, name, packageManager, private, type, version, clsx, eslint (+21 more)

### Community 16 - "devDependencies"
Cohesion: 0.06
Nodes (31): devDependencies, eslint, eslint-config-prettier, @eslint/js, eslint-plugin-svelte, globals, @inlang/paraglide-js, @lucide/svelte (+23 more)

### Community 17 - "workbook.go"
Cohesion: 0.12
Nodes (33): cell(), DefinitionRow, Workbook, indexHeader(), normalizePhase(), parseBool(), parseDefinitions(), parseExamples() (+25 more)

### Community 18 - "Configure your project (Storybook docs page)"
Cohesion: 0.10
Nodes (27): svelte-file-editor agent, svelte-code-writer skill, @sveltejs/mcp CLI, get-documentation tool, list-sections tool, playground-link tool, svelte-autofixer tool, Svelte MCP Server (+19 more)

### Community 19 - "devices.ts"
Cohesion: 0.11
Nodes (22): createDevice(), deleteDevice(), deviceBadgeLinkClass, deviceCountLinkClass, DeviceDefinitionKind, DeviceDefinitionRow, DeviceDefinitionsResult, DeviceFunctionRow (+14 more)

### Community 20 - "0006_baseline.sql"
Cohesion: 0.15
Nodes (20): criteria_definitions, criteria_links, devices, examples, functions, idx_criteria_links_function, idx_criteria_links_name, idx_examples_function (+12 more)

### Community 21 - "Function"
Cohesion: 0.17
Nodes (5): cmpBy(), Queries, CountFunctionsParams, Queries, Function

### Community 22 - "sequences.sql.go"
Cohesion: 0.13
Nodes (9): ListSequenceStepsRow, Queries, SequenceStep, GetFunctionByNaturalKeyParams, InsertSequenceParams, InsertSequenceStepParams, ListSequencesRow, Sequence (+1 more)

### Community 23 - "+layout.svelte"
Cohesion: 0.11
Nodes (5): frontend_src_app, frontend_src_lib_assets_favicon, density, frontend_src_routes_layout, $app/types

### Community 24 - "ref_node_path"
Cohesion: 0.40
Nodes (3): E2E_DIR, FIXTURE, ref_node_path

### Community 26 - "examples.sql.go"
Cohesion: 0.15
Nodes (9): ListExamplesParams, ListExamplesRow, Queries, Example, CountExamplesParams, FunctionNaturalKeyExistsParams, GetExampleRow, InsertExampleParams (+1 more)

### Community 27 - "client.ts"
Cohesion: 0.13
Nodes (12): ApiError, ComposeField, ComposeFields, ComposeRequest, composeScript(), getComposeFields(), listOrphans(), OrphanExampleRow (+4 more)

### Community 28 - "request"
Cohesion: 0.15
Nodes (15): request(), createSequence(), EmitResult, emitSequence(), getSequence(), listSequences(), SequenceDetail, SequenceInput (+7 more)

### Community 29 - "import.sql.go"
Cohesion: 0.09
Nodes (11): Queries, UpsertParameterDefinitionParams, ParameterLink, ImportNormalization, ImportRun, InsertImportNormalizationParams, InsertImportRunParams, UpsertCriteriaLinkParams (+3 more)

### Community 30 - "api.ts"
Cohesion: 0.14
Nodes (13): createFunction(), FunctionFacets, FunctionInput, FunctionRow, getFunction(), getFunctionFacets(), listFunctions(), ListParams (+5 more)

### Community 31 - "AGENTS.md"
Cohesion: 0.17
Nodes (11): Agent skills, Domain docs, Issue tracker, Agent Must Not Access Approved Ubuntu Host, CRUD Web App over Script-Definition Knowledge Workbook, Five-Role Triage Vocabulary, Local Markdown Issue Tracker (.scratch/<feature>/), PostgreSQL Production Database (+3 more)

### Community 32 - "ref_node_fs"
Cohesion: 0.40
Nodes (3): E2E_DIR, FIXTURE, ref_node_fs

### Community 33 - "ref_node_url"
Cohesion: 0.40
Nodes (3): E2E_DIR, FIXTURE, ref_node_url

### Community 34 - "smoke.spec.ts"
Cohesion: 0.40
Nodes (3): E2E_DIR, FIXTURE, ADR-0005

### Community 36 - "scripts"
Cohesion: 0.14
Nodes (14): scripts, build, build-storybook, check, check:watch, dev, format, lint (+6 more)

### Community 37 - "compilerOptions"
Cohesion: 0.13
Nodes (14): compilerOptions, allowJs, checkJs, esModuleInterop, forceConsistentCasingInFileNames, moduleResolution, resolveJsonModule, rewriteRelativeImportExtensions (+6 more)

### Community 38 - "importer/export.go"
Cohesion: 0.31
Nodes (13): boolCell(), criteriaDefinitionCells(), criteriaLinkCells(), definitionRow(), exampleCells(), Export(), functionCells(), linkRow() (+5 more)

### Community 39 - "Q: I want to commit the project Web_Program_Functions_Database to github. Please help me to update the .gitignore file and make sure that files for AI are not commited to the github (excepts agents and the skills)."
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: I want to commit the project Web_Program_Functions_Database to github. Please help me to update the .gitignore file and make sure that files for AI are not commited to the github (excepts agents and the skills)., Source Nodes

### Community 40 - "aliases"
Cohesion: 0.15
Nodes (12): aliases, components, hooks, lib, ui, utils, registry, $schema (+4 more)

### Community 41 - "orphans.go"
Cohesion: 0.29
Nodes (13): orphanExampleItem, orphanGroup, orphanLinkItem, fiber.App, fiber.Ctx, fiber.Handler, nullableAny(), nullablePosition() (+5 more)

### Community 42 - "Q: please create proper test folder for the backend and frontend folder and move all test related files accordingly"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: please create proper test folder for the backend and frontend folder and move all test related files accordingly, Source Nodes

### Community 43 - "definitions.ts"
Cohesion: 0.17
Nodes (10): if(), if(), createDefinition(), DefinitionKind, DefinitionListResult, DefinitionRow, listDefinitions(), ADR-0003 (+2 more)

### Community 44 - "examples.ts"
Cohesion: 0.17
Nodes (10): if(), createExample(), ExampleInput, ExampleListParams, ExampleListResult, ExampleRow, ExampleUpdateInput, listExamples() (+2 more)

### Community 45 - "0002_dictionaries.sql"
Cohesion: 0.27
Nodes (10): criteria_definitions, criteria_links, examples, idx_criteria_links_function, idx_criteria_links_name, idx_examples_function, idx_parameter_links_function, idx_parameter_links_name (+2 more)

### Community 46 - "hooks.server.ts"
Cohesion: 0.20
Nodes (7): handle, frontend_src_lib_paraglide_runtime, frontend_src_lib_paraglide_runtime_delocalizeurl, frontend_src_lib_paraglide_runtime_gettextdirection, frontend_src_lib_paraglide_server, frontend_src_lib_paraglide_server_paraglidemiddleware, @sveltejs/kit

### Community 47 - "links.ts"
Cohesion: 0.20
Nodes (8): DefinitionInput, attachLink(), LinkAttachInput, LinkKind, LinkRow, LinkUpdateInput, ADR-0002, updateLink()

### Community 48 - "Q: start with Phase 1"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: start with Phase 1, Source Nodes

### Community 49 - "examples.go"
Cohesion: 0.18
Nodes (19): exampleBody, exampleResponse, exampleUpdateBody, normalizeDeviceName(), createExample(), deleteExample(), exampleFilters(), exampleItem() (+11 more)

### Community 50 - "importHandler"
Cohesion: 0.36
Nodes (8): getImportRun(), fiber.App, fiber.Handler, fiber.Map, importHandler(), importRunResponse(), listImportRuns(), registerImportRoutes()

### Community 51 - "Q: Please continue to Phase 2."
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: Please continue to Phase 2., Source Nodes

### Community 52 - "imports.ts"
Cohesion: 0.28
Nodes (8): getImportRun(), ImportNormalization, importReset(), ImportResult, ImportRun, importWorkbook(), listImportRuns(), upload()

### Community 53 - "Q: continue with Phase 3"
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: continue with Phase 3, Source Nodes

### Community 54 - "Q: please continue to Phase 4."
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: please continue to Phase 4., Source Nodes

### Community 55 - "README.md"
Cohesion: 0.04
Nodes (44): Backend, Backend Nginx, Backend Service, Build And Secrets, Database Configuration, Local Build And Development, Requirements, Tests (+36 more)

### Community 56 - "0003_import.sql"
Cohesion: 0.38
Nodes (6): examples_new, idx_examples_function, idx_import_normalizations_run, import_normalizations, import_runs, examples

### Community 57 - "models.go"
Cohesion: 0.22
Nodes (5): Queries, ImportRun, Device, ImportNormalization, SequenceStep

### Community 58 - "Q: because I will be deploying the backend and frontend to the ubuntu server and the postgresql database is hosted on that server too, please help to update the environment files to setup the database configuration if it hosted on the ubuntu server (using localhost instead of server ip address). Please also create the nginx configurations for each backend (radom port) and frontend (random port). After that please help to update the backend and frontend README.md files on how to setup and run the projects locally and deployed in on the server."
Cohesion: 0.40
Nodes (4): Answer, Outcome, Q: because I will be deploying the backend and frontend to the ubuntu server and the postgresql database is hosted on that server too, please help to update the environment files to setup the database configuration if it hosted on the ubuntu server (using localhost instead of server ip address). Please also create the nginx configurations for each backend (radom port) and frontend (random port). After that please help to update the backend and frontend README.md files on how to setup and run the projects locally and deployed in on the server., Source Nodes

### Community 61 - "settings.json"
Cohesion: 0.29
Nodes (6): baseLocale, locales, modules, plugin.inlang.messageFormat, pathPattern, $schema

### Community 62 - "vitest"
Cohesion: 0.38
Nodes (3): greet(), vitest, vitest-browser-svelte

### Community 63 - "vite.config.ts"
Cohesion: 0.33
Nodes (5): ADR-0001, @inlang/paraglide-js, @sveltejs/adapter-static, @tailwindcss/vite, @vitest/browser-playwright

### Community 65 - "Accessibility Checker Results Screenshot"
Cohesion: 0.47
Nodes (6): Accessibility Person Icon Badge, Accessibility Checker Results Screenshot, Passes Count (11 Passes, active tab), A11y Rule Checklist (button text, unique ids, heading text), Violations Count (2 Violations), Warnings Count (0 Warnings)

### Community 66 - "Storybook UI (sidebar, controls, args table)"
Cohesion: 0.53
Nodes (6): tags: ['autodocs'] CSF tag, Button story variants (Primary, Secondary, Large, Small), Meta<Button> CSF3 meta object, Storybook Docs page, docs.png - Storybook autodocs illustration, Storybook UI (sidebar, controls, args table)

### Community 67 - "PASS State Test Results Panel"
Cohesion: 0.53
Nodes (6): Testing.png Illustration (Test Results UI Mockup), FAIL State Test Results Panel, Test Step: Histogram Default, PASS State Test Results Panel, Test Playback Controls (rewind, step, replay), Storybook Test Runner UI Concept

### Community 68 - "Storybook Theming Screenshot (theming.png)"
Cohesion: 0.47
Nodes (6): Acme Brand Namespace (Sidebar Header), Button Story (Example Section), Dark Theme Variant, Light Theme Variant, Storybook UI, Storybook Theming Screenshot (theming.png)

### Community 71 - "Queries"
Cohesion: 0.25
Nodes (20): sequenceBody, sequenceResponse, sequenceStepBody, sequenceStepResponse, createSequence(), deleteSequence(), emitSequence(), getSequence() (+12 more)

### Community 72 - "Styling Showcase Image (styling.png)"
Cohesion: 0.40
Nodes (5): Bootstrap, Styling Showcase Image (styling.png), Material UI (MUI), tailwindcss, Sass

### Community 73 - "share.png — screenshot of a browser viewing a Storybook published to a chromatic.com URL (sidebar with Storybook logo, Find components search, EXAMPLE/Button; canvas showing a blue Button)"
Cohesion: 0.80
Nodes (5): Chromatic — Storybook hosting/collaboration service (host of the chromatic.com URL shown in the screenshot), share.png — screenshot of a browser viewing a Storybook published to a chromatic.com URL (sidebar with Storybook logo, Find components search, EXAMPLE/Button; canvas showing a blue Button), Configure.mdx — Storybook 'Configure your project' onboarding docs page that imports share.png, Publish to Chromatic — 'Do more with Storybook' docs feature section illustrated by share.png, Storybook — component workshop/workbench tool whose UI (manager sidebar + canvas) is captured in the screenshot

### Community 75 - "main.ts"
Cohesion: 0.40
Nodes (3): config, preview, ref_storybook_sveltekit

### Community 76 - "devices.go"
Cohesion: 0.23
Nodes (19): definitionItem, deviceBody, createDevice(), criteriaItem(), deleteDevice(), deviceCriteria(), deviceFromSegment(), deviceFunctions() (+11 more)

### Community 77 - "0001_functions.sql"
Cohesion: 0.83
Nodes (3): functions, idx_functions_interface, idx_functions_name

### Community 78 - "0004_sequences.sql"
Cohesion: 0.83
Nodes (3): idx_sequence_steps_sequence, sequence_steps, sequences

### Community 79 - "dependencies"
Cohesion: 0.50
Nodes (4): dependencies, bits-ui, clsx, tailwind-merge

### Community 80 - "Browser Tab / Site Identity Asset"
Cohesion: 0.67
Nodes (4): Browser Tab / Site Identity Asset, Favicon SVG (frontend/src/lib/assets/favicon.svg), Svelte Framework, Svelte Logo Graphic (orange #ff3e00 S-curve mark)

### Community 81 - "Addon Library Icon Collage (Storybook Stories Asset)"
Cohesion: 0.67
Nodes (4): Addon Library Icon Collage (Storybook Stories Asset), Rounded-Tile Grid Collage Layout, Storybook Addon Ecosystem, Frontend Tooling and Framework Logos

### Community 82 - "Storybook Assets Placeholder Image (assets.png)"
Cohesion: 0.67
Nodes (4): Storybook Assets Placeholder Image (assets.png), Teal Picture/Photo Icon (image placeholder symbol), Storybook Sample Assets Directory Convention, Green Serif 'T' with Vertical Cursor (text placeholder symbol)

### Community 83 - "Discord Logo Mark"
Cohesion: 0.50
Nodes (4): Discord Blurple Brand Color (#5865F2), Discord Logo Mark, Storybook Static Asset, Discord Logo SVG Asset

### Community 84 - "YouTube Icon (SVG)"
Cohesion: 0.67
Nodes (4): YouTube Icon (SVG), Play Button Symbol, Storybook Story Asset, YouTube Brand Logo

### Community 86 - "DBTX"
Cohesion: 0.18
Nodes (7): New(), database/sql.Result, database/sql.Row, database/sql.Rows, database/sql.Stmt, Queries, DBTX

### Community 94 - "Context Layered Stack Illustration"
Cohesion: 1.00
Nodes (3): Context Layered Stack Illustration, Layered Context / Data Stack Concept, Storybook Decorative Asset Role

### Community 95 - "GitHub Octocat Logo Icon (github.svg)"
Cohesion: 0.67
Nodes (3): GitHub Octocat Logo Icon (github.svg), GitHub (Code Hosting Platform), SvelteKit Scaffolded Storybook Stories Assets

### Community 96 - "Tutorials Icon (Open Book SVG)"
Cohesion: 0.67
Nodes (3): Open Book / Learning Metaphor, Storybook UI Asset, Tutorials Icon (Open Book SVG)

### Community 137 - "blue-green.test.mjs"
Cohesion: 0.29
Nodes (7): deployment, runDeployment(), shellPath(), ref_node_assert, ref_node_child_process, ref_node_os, ref_node_test

## Ambiguous Edges - Review These
- `Layered Context / Data Stack Concept` → `Storybook Decorative Asset Role`  [AMBIGUOUS]
  frontend/src/stories/assets/context.png · relation: conceptually_related_to

## Knowledge Gaps
- **321 isolated node(s):** `webprogfndb`, `composeSegments`, `deviceBody`, `exampleBody`, `exampleUpdateBody` (+316 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 477 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **37 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Work-memory lessons

**Preferred sources** — corroborated by past sessions; start here.
- `CountsFrom()` (3× useful, score=2.995751523) _(code changed — re-verify)_
- `Workbook` (3× useful, score=2.995751523)

**Known dead ends** — questions that led nowhere; don't re-derive.
- "I want to commit the project Web_Program_Functions_Database to github. Please help me to update the .gitignore file and make sure that files for AI are not commited to the github (excepts agents and the skills)." -> `AGENTS.md`, `CONTEXT.md`, `Agent skills`

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Layered Context / Data Stack Concept` and `Storybook Decorative Asset Role`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `Migrate()` connect `import_test.go` to `context.Context`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Queries` connect `Queries` to `import_test.go`, `functions.go`, `import.go`, `links.go`, `orphans.go`, `compose.go`, `devices.go`, `examples.go`, `importHandler`, `DBTX`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Why does `request()` connect `request` to `ref_node_fs`, `ref_node_url`, `smoke.spec.ts`, `@playwright/test`, `tabs.spec.ts`, `definitions.ts`, `examples.ts`, `dialogs.spec.ts`, `links.ts`, `devices.ts`, `imports.ts`, `ref_node_path`, `combobox.spec.ts`, `client.ts`, `api.ts`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Are the 57 inferred relationships involving `postJSON()` (e.g. with `compose1()` and `seedComposeFunction()`) actually correct?**
  _`postJSON()` has 57 INFERRED edges - model-reasoned connections that need verification._
- **Are the 55 inferred relationships involving `doRequest()` (e.g. with `TestComposeFieldsEndpoint()` and `TestComposeSaveAsExample()`) actually correct?**
  _`doRequest()` has 55 INFERRED edges - model-reasoned connections that need verification._
- **What connects `webprogfndb`, `composeSegments`, `deviceBody` to the rest of the system?**
  _321 weakly-connected nodes found - possible documentation gaps or missing edges._