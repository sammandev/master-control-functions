<script lang="ts">
	import { onMount, tick } from 'svelte';
	import {
		Download,
		Upload,
		RotateCcw,
		RefreshCw,
		LoaderCircle,
		CircleCheck
	} from '@lucide/svelte';
	import {
		importWorkbook,
		importReset,
		listImportRuns,
		getImportRun,
		ApiError,
		type ImportResult,
		type ImportRun,
		type ImportNormalization
	} from '#lib/imports.ts';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Collapsible from '#lib/components/ui/collapsible.svelte';
	import { toast } from 'svelte-sonner';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir, type SortValue } from '#lib/table-sort.ts';
	import {
		Table,
		TableHeader,
		TableBody,
		TableRow,
		TableCell
	} from '#lib/components/ui/table/index.ts';

	let selectedFile = $state<File | null>(null);
	const selectedFileName = $derived(selectedFile?.name ?? '');
	let importing = $state(false);
	let lastResult = $state<ImportResult | null>(null);
	let actionError = $state('');
	let resetOpen = $state(false);
	let resetTrigger: HTMLElement | null = null;
	let resultDetailsOpen = $state(false);

	let runs = $state<ImportRun[]>([]);
	let loadingRuns = $state(true);
	let runsError = $state('');
	let expandedRun = $state<number | null>(null);
	let runNormalizations = $state<ImportNormalization[]>([]);
	let loadingRun = $state(false);
	let detailsError = $state('');
	let detailRevision = 0;
	let normalizationSortKey = $state('');
	let normalizationSortDir = $state<SortDir>('asc');
	const normalizationColumns = [
		{ key: 'table', label: 'Table' },
		{ key: 'row_key', label: 'Row' },
		{ key: 'column', label: 'Column' },
		{ key: 'source', label: 'Source' },
		{ key: 'normalized', label: 'Normalized' },
		{ key: 'rule', label: 'Rule' }
	];
	const normalizationKeys: Record<
		string,
		(record: ImportResult['normalizations'][number]) => SortValue
	> = {
		table: (record) => record.table,
		row_key: (record) => record.row_key,
		column: (record) => record.column,
		source: (record) => record.source,
		normalized: (record) => record.normalized,
		rule: (record) => record.rule
	};
	function sortNormalizations(column: string) {
		({ key: normalizationSortKey, dir: normalizationSortDir } = nextSort(
			normalizationSortKey,
			normalizationSortDir,
			column
		));
	}

	onMount(() => {
		void refreshRuns();
		return () => {
			detailRevision += 1;
		};
	});

	async function refreshRuns() {
		loadingRuns = true;
		runsError = '';
		try {
			runs = (await listImportRuns()).items;
		} catch (e) {
			runsError = e instanceof Error ? e.message : String(e);
		} finally {
			loadingRuns = false;
		}
	}

	async function run(reset: boolean, event?: MouseEvent) {
		if (importing) return;
		const file = selectedFile;
		if (!file) {
			actionError = 'Choose a workbook file first.';
			return;
		}
		actionError = '';
		if (reset) {
			resetTrigger = event?.currentTarget instanceof HTMLElement ? event.currentTarget : null;
			resetOpen = true;
			return;
		}
		await performImport(reset);
	}

	async function performImport(reset: boolean) {
		const file = selectedFile;
		if (!file || importing) return;
		importing = true;
		actionError = '';
		try {
			lastResult = reset ? await importReset(file) : await importWorkbook(file);
			resultDetailsOpen = false;
			toast.success(`Import complete - run #${lastResult.run_id}`);
			await refreshRuns();
		} catch (e) {
			if (e instanceof ApiError) {
				actionError = e.message;
			} else {
				actionError = e instanceof Error ? e.message : String(e);
			}
		} finally {
			importing = false;
			if (reset) {
				await tick();
				resetTrigger?.focus();
			}
		}
	}

	async function changeResetOpen(open: boolean) {
		resetOpen = open;
		if (!open && !importing) {
			await tick();
			resetTrigger?.focus();
		}
	}

	async function toggleRun(id: number, open: boolean) {
		const revision = ++detailRevision;
		if (!open) {
			if (expandedRun === id) expandedRun = null;
			loadingRun = false;
			return;
		}
		expandedRun = id;
		runNormalizations = [];
		detailsError = '';
		loadingRun = true;
		try {
			const detail = await getImportRun(id);
			if (revision === detailRevision) runNormalizations = detail.normalizations ?? [];
		} catch (e) {
			if (revision === detailRevision) detailsError = e instanceof Error ? e.message : String(e);
		} finally {
			if (revision === detailRevision) loadingRun = false;
		}
	}

	function parseCounts(raw: string): Record<string, number> {
		try {
			return JSON.parse(raw) as Record<string, number>;
		} catch {
			return {};
		}
	}

	function countsSummary(raw: string): string {
		return Object.entries(parseCounts(raw))
			.map(([table, count]) => `${table}: ${count}`)
			.join(' · ');
	}

	function parseSkipped(raw: string): { table: string; row_key: string; reason: string }[] {
		try {
			const records: unknown = JSON.parse(raw);
			return Array.isArray(records) ? records : [];
		} catch {
			return [];
		}
	}
</script>

{#snippet normalizationTable(records: ImportResult['normalizations'])}
	{@const rows = normalizationSortKey
		? sortRows(records, normalizationKeys[normalizationSortKey], normalizationSortDir)
		: records}
	<div class="normalization-scroll" role="region" aria-label="Normalization records">
		<Table class="border-0">
			<TableHeader
				><TableRow>
					{#each normalizationColumns as column (column.key)}<SortHead
							label={column.label}
							col={column.key}
							active={normalizationSortKey}
							dir={normalizationSortDir}
							onclick={() => sortNormalizations(column.key)}
						/>{/each}
				</TableRow></TableHeader
			>
			<TableBody>
				{#each rows as record, index (index)}
					<TableRow>
						<TableCell><span class="record-value">{record.table}</span></TableCell>
						<TableCell><span class="record-key">{record.row_key}</span></TableCell>
						<TableCell><span class="record-value">{record.column}</span></TableCell>
						<TableCell><span class="record-key">{record.source}</span></TableCell>
						<TableCell><span class="record-key">{record.normalized}</span></TableCell>
						<TableCell><span class="record-value">{record.rule}</span></TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/snippet}

{#snippet skippedRows(records: ImportResult['skipped'])}
	{#if records.length > 0}
		<div class="skipped-rows">
			<h3>{records.length} rows skipped</h3>
			<ul>
				{#each records as record, index (index)}
					<li>
						<span class="record-key">{record.table}: {record.row_key}</span><span
							>{record.reason}</span
						>
					</li>
				{/each}
			</ul>
		</div>
	{/if}
{/snippet}

<div class="import-workspace">
	<div class="import-heading">
		<h1>Workbook import</h1>
		<!-- eslint-disable svelte/no-navigation-without-resolve -->
		<a class="export-action" href="/api/export" download="knowledge_workbook.xlsx"
			><Download size={16} /> Export workbook</a
		>
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
	</div>
	<section class="import-section" aria-labelledby="upload-heading" aria-busy={importing}>
		<div class="section-heading">
			<h2 id="upload-heading">Upload workbook</h2>
			<Badge>Excel .xlsx</Badge>
		</div>
		<p class="import-description">
			Import upserts all six knowledge tables by natural key: matching rows are updated from the
			workbook, rows edited only in this app survive. Orphan rows are kept and flagged, never
			dropped.
		</p>
		<div class="upload-field">
			<label for="workbook-file">Workbook file</label>
			<Input
				id="workbook-file"
				type="file"
				accept=".xlsx"
				disabled={importing}
				class="text-muted-foreground file:border-border file:bg-secondary file:text-foreground h-9 px-0 file:mr-3 file:h-full file:cursor-pointer file:border-0 file:border-r file:px-3 file:font-medium"
				onchange={(event) => {
					selectedFile = event.currentTarget.files?.[0] ?? null;
					actionError = '';
				}}
			/>
			{#if selectedFileName}<p class="workbook-name">{selectedFileName}</p>{/if}
		</div>
		<div class="import-actions">
			<Button disabled={importing} onclick={() => run(false)}
				>{#if importing}<LoaderCircle size={16} class="animate-spin" /> Importing{:else}<Upload
						size={16}
					/> Import{/if}</Button
			>
			<Button variant="destructive" disabled={importing} onclick={(event) => run(true, event)}
				><RotateCcw size={16} /> Reset &amp; import</Button
			>
		</div>
		{#if importing}<p class="operation-status" role="status">Importing workbook...</p>{/if}
		{#if actionError}<div class="banner-error" role="alert">{actionError}</div>{/if}
	</section>

	{#if lastResult}
		<section class="import-section" aria-label="Import results">
			<div class="section-heading">
				<h2>Last import - run #{lastResult.run_id}</h2>
				<Badge><CircleCheck size={14} /> Complete</Badge>{#if lastResult.reset}<Badge variant="warn"
						>Reset import</Badge
					>{/if}
			</div>
			<p class="workbook-name">{lastResult.file_name}</p>
			<dl class="result-counts">
				{#each Object.entries(lastResult.counts) as [table, count] (table)}<div>
						<dt>{table}</dt>
						<dd>{count}</dd>
					</div>{/each}
			</dl>
			{#if lastResult.normalizations.length > 0}
				<Collapsible bind:open={resultDetailsOpen}>
					{#snippet title()}Normalization details ({lastResult?.normalizations.length}){/snippet}
					{@render normalizationTable(lastResult.normalizations)}
				</Collapsible>
			{:else}<p class="muted">No normalizations needed.</p>{/if}
			{@render skippedRows(lastResult.skipped)}
		</section>
	{/if}

	<section class="import-section" aria-labelledby="history-heading" aria-busy={loadingRuns}>
		<div class="section-heading">
			<h2 id="history-heading">Import history</h2>
			<span class="muted">{runs.length} runs</span>
		</div>
		{#if runsError}<div class="banner-error" role="alert">
				{runsError}<Button variant="outline" onclick={refreshRuns}
					><RefreshCw size={16} /> Retry history</Button
				>
			</div>{/if}
		{#if loadingRuns}<p class="banner-empty" role="status">Loading import history...</p>
		{:else if runs.length === 0 && !runsError}<p class="banner-empty" role="status">
				No imports yet.
			</p>
		{:else}
			<div class="history-runs">
				{#each runs as historyRun (historyRun.id)}
					<Collapsible
						open={expandedRun === historyRun.id}
						onOpenChange={(open) => toggleRun(historyRun.id, open)}
					>
						{#snippet title()}
							<span class="run-summary">
								<span class="run-identity"
									><strong>#{historyRun.id}</strong><span class="workbook-name"
										>{historyRun.file_name}</span
									></span
								>
								<span class="run-metadata"
									>{#if historyRun.reset}<Badge variant="warn">reset</Badge>{/if}<time
										datetime={historyRun.finished_at}
										>{historyRun.finished_at.slice(0, 19).replace('T', ' ')}</time
									></span
								>
								<span class="run-counts">{countsSummary(historyRun.counts)}</span>
							</span>
						{/snippet}
						{#if loadingRun}<p class="operation-status" role="status">
								Loading normalization details...
							</p>
						{:else if detailsError}<div class="banner-error" role="alert">
								{detailsError}<Button
									variant="outline"
									onclick={() => toggleRun(historyRun.id, true)}
									><RefreshCw size={16} /> Retry details</Button
								>
							</div>
						{:else}
							{#if runNormalizations.length === 0}<p class="muted">
									No normalizations in this run.
								</p>
							{:else}{@render normalizationTable(
									runNormalizations.map((record) => ({
										table: record.table_name,
										row_key: record.row_key,
										column: record.column_name,
										source: record.source_value,
										normalized: record.normalized_value,
										rule: record.rule
									}))
								)}{/if}
							{@render skippedRows(parseSkipped(historyRun.skipped))}
						{/if}
					</Collapsible>
				{/each}
			</div>
		{/if}
	</section>
</div>

<AlertDialog
	open={resetOpen}
	onOpenChange={changeResetOpen}
	title="Reset &amp; import"
	description="Reset clears ALL six entity tables and re-imports from this workbook. Rows edited only in this app are lost. Continue?"
	confirmLabel="Reset &amp; import"
	onConfirm={() => performImport(true)}
/>

<style>
	.import-workspace {
		min-width: 0;
	}
	.import-heading,
	.section-heading,
	.import-actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
	}
	.import-heading {
		justify-content: space-between;
		margin-bottom: 1rem;
	}
	h1 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}
	h2 {
		margin: 0;
		font-size: 1rem;
		font-weight: 600;
	}
	h3 {
		margin: 0 0 0.5rem;
		font-size: 0.875rem;
		font-weight: 600;
	}
	.import-section {
		min-width: 0;
		padding: 1.25rem 0;
		border-top: 1px solid var(--border);
	}
	.section-heading {
		margin-bottom: 0.75rem;
	}
	.import-description {
		max-width: 80ch;
		margin: 0 0 1rem;
		color: var(--muted-foreground);
		font-size: 0.875rem;
		line-height: 1.6;
	}
	.export-action {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		min-height: 2rem;
		padding: 0 0.75rem;
		border: 1px solid var(--border);
		border-radius: 4px;
		background: var(--card);
		font-size: 0.875rem;
		font-weight: 500;
	}
	.export-action:hover {
		background: var(--accent);
	}
	.export-action:focus-visible,
	.normalization-scroll:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 2px;
	}
	.upload-field {
		display: grid;
		gap: 0.5rem;
		max-width: 40rem;
		margin-bottom: 1rem;
	}
	label {
		font-size: 0.875rem;
		font-weight: 500;
	}
	.workbook-name {
		min-width: 0;
		margin: 0;
		overflow-wrap: anywhere;
		font-family: var(--font-mono);
		font-size: 0.8125rem;
	}
	.operation-status {
		margin: 0.75rem 0;
		color: var(--muted-foreground);
		font-size: 0.875rem;
	}
	.banner-error {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
		margin-top: 0.75rem;
		overflow-wrap: anywhere;
	}
	.result-counts {
		display: flex;
		flex-wrap: wrap;
		margin: 1rem 0;
		border-block: 1px solid var(--border);
	}
	.result-counts > div {
		display: grid;
		gap: 0.25rem;
		min-width: 0;
		padding: 0.75rem 1.5rem 0.75rem 0;
	}
	dt {
		color: var(--muted-foreground);
		font-family: var(--font-mono);
		font-size: 0.8125rem;
		overflow-wrap: anywhere;
	}
	dd {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}
	.history-runs {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		min-width: 0;
		gap: 0.5rem;
	}
	.run-summary {
		display: grid;
		min-width: 0;
		flex: 1;
		gap: 0.5rem;
	}
	.run-identity {
		display: flex;
		align-items: baseline;
		min-width: 0;
		gap: 0.75rem;
	}
	.run-identity strong {
		flex-shrink: 0;
		font-family: var(--font-mono);
		color: var(--primary);
	}
	.run-metadata {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--muted-foreground);
	}
	.run-counts {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--muted-foreground);
		overflow-wrap: anywhere;
	}
	.normalization-scroll {
		min-width: 0;
		max-width: 100%;
		overflow-x: auto;
		margin-top: 0.75rem;
		border: 1px solid var(--border);
		border-radius: 4px;
	}
	.normalization-scroll :global(table) {
		min-width: 840px;
	}
	.record-key,
	.record-value {
		display: block;
		max-width: 28rem;
		overflow-wrap: anywhere;
	}
	.record-key {
		font-family: var(--font-mono);
		font-size: 0.8125rem;
	}
	.skipped-rows {
		margin-top: 1rem;
	}
	.skipped-rows ul {
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.skipped-rows li {
		display: grid;
		gap: 0.25rem;
		padding: 0.5rem 0;
		border-top: 1px solid var(--border);
		overflow-wrap: anywhere;
		font-size: 0.8125rem;
	}
	@media (max-width: 600px) {
		.import-heading {
			align-items: flex-start;
		}
		.run-identity {
			flex-direction: column;
			gap: 0.25rem;
		}
	}
</style>
