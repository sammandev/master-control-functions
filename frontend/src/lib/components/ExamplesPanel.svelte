<script lang="ts">
	import ExampleDialog from '#lib/components/ExampleDialog.svelte';
	import ExampleFieldUsage from '#lib/components/ExampleFieldUsage.svelte';
	import { getComposeFields } from '#lib/compose.ts';
	import { copyText } from '#lib/clipboard.ts';
	import { createLatestRequestGuard } from '#lib/request-guard.ts';
	import {
		listExamples,
		deleteExample,
		type ExampleRow,
		type ExampleListParams
	} from '#lib/examples.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import Tooltip from '#lib/components/ui/tooltip.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir } from '#lib/table-sort.ts';
	import { toast } from 'svelte-sonner';
	import { Copy, Check, Pencil, Plus, RefreshCw, Search, Trash2 } from '@lucide/svelte';
	import {
		Table,
		TableHeader,
		TableBody,
		TableRow,
		TableHead,
		TableCell
	} from '#lib/components/ui/table/index.ts';

	// In fixed mode this panel shows one Function's examples. The shared
	// ExampleDialog owns create/edit validation for every Examples surface.
	let {
		fixed = null,
		defaultDevice = '',
		functionId
	}: {
		fixed?: { source: string; name: string } | null;
		defaultDevice?: string;
		functionId?: number;
	} = $props();

	let items = $state<ExampleRow[]>([]);
	let total = $state(0);
	let page = $state(1);
	let pageSize = $state(20);
	let loading = $state(false);
	let loadError = $state('');
	const listRequestGuard = createLatestRequestGuard();
	const usageFields = $derived(
		functionId === undefined
			? null
			: getComposeFields(functionId).then((fields) => ({
					parameters: fields.parameters.map((field) => field.name),
					criteria: fields.criteria.map((field) => field.name)
				}))
	);

	let fnSource = $state('');
	let fnName = $state('');
	let deviceFilter = $state('');
	let search = $state('');

	function currentFilters(): ExampleListParams {
		return {
			function_source: fixed?.source ?? fnSource.trim(),
			function_name: fixed?.name ?? fnName.trim(),
			device: deviceFilter.trim() || undefined,
			search: search.trim(),
			page
		};
	}

	async function load(filters: ExampleListParams) {
		const request = listRequestGuard.begin();
		loading = true;
		loadError = '';
		try {
			const result = await listExamples({ ...filters, page_size: pageSize });
			if (!listRequestGuard.isCurrent(request)) return;
			items = result.items;
			total = result.total;
		} catch (error) {
			if (!listRequestGuard.isCurrent(request)) return;
			items = [];
			total = 0;
			loadError = error instanceof Error ? error.message : String(error);
		} finally {
			if (listRequestGuard.isCurrent(request)) loading = false;
		}
	}

	$effect(() => {
		// Read every filter input synchronously so changes re-trigger the
		// load; listExamples builds the query from the snapshot. pageSize
		// is added inside load(), so it is tracked here explicitly.
		const f = currentFilters();
		void pageSize;
		load(f);
	});

	// ---- sorting (ticket 09) ------------------------------------------------
	// The examples list API paginates server-side without sort params, so
	// sorting applies to the loaded page through the shared sortRows
	// helper. '' keeps the API's default order until a header is clicked.
	const EXAMPLE_SORT_KEYS: Record<string, (row: ExampleRow) => string | number | null | undefined> =
		{
			function: (r) => `${r.function_source}/${r.function_name}`,
			device: (r) => r.device_name || '',
			script: (r) => r.script,
			meaning: (r) => r.meaning
		};

	let sortKey = $state('');
	let sortDir = $state<SortDir>('asc');

	function applySort(col: string) {
		({ key: sortKey, dir: sortDir } = nextSort(sortKey, sortDir, col));
	}

	const view = $derived(
		sortKey === '' ? items : sortRows(items, EXAMPLE_SORT_KEYS[sortKey]!, sortDir)
	);

	// ---- dialog state ------------------------------------------------------
	let dialogOpen = $state(false);
	// null = create, otherwise the row being edited (the function reference
	// is immutable on edit and the form lives in ExampleDialog).
	let editTarget = $state<ExampleRow | null>(null);
	let deleting = $state(false);

	// ---- delete confirm state ----------------------------------------------
	// The row awaiting confirmation; kept after the dialog closes so the
	// description stays stable until the next confirm opens.
	let confirmRow = $state<ExampleRow | null>(null);
	let confirmOpen = $state(false);

	function requestRemove(row: ExampleRow) {
		confirmRow = row;
		confirmOpen = true;
	}

	async function confirmRemove() {
		if (!confirmRow) return;
		deleting = true;
		try {
			await deleteExample(confirmRow.id);
			await load(currentFilters());
			toast.success('Example deleted');
		} catch (e) {
			loadError = e instanceof Error ? e.message : String(e);
		} finally {
			deleting = false;
		}
	}

	function resetPage() {
		page = 1;
	}

	function openCreateDialog() {
		editTarget = null;
		dialogOpen = true;
	}

	function openEditDialog(row: ExampleRow) {
		editTarget = row;
		dialogOpen = true;
	}

	// ---- full-script copy ---------------------------------------------------
	let copied = $state<number | null>(null);
	let copyStatus = $state('');

	async function copyRow(row: ExampleRow) {
		const success = await copyText(row.script);
		copied = success ? row.id : null;
		copyStatus = success ? 'Script copied.' : 'Copy failed. The script text is selectable.';
	}
</script>

<!-- Mounts flush inside a Card (the function detail tabs Card in fixed
     mode; the /examples page wraps it in a Card in standalone mode): the
     toolbar and table carry the card's horizontal padding, the table
     itself is borderless so the Card frames it. -->
<div>
	{#if fixed}
		<div class="example-heading">
			<div>
				<strong>Examples</strong>
				<p>{loading ? 'Loading…' : `${total} ${total === 1 ? 'example' : 'examples'}`}</p>
			</div>
			<div class="example-actions">
				<Input
					class="w-36"
					placeholder="Device…"
					aria-label="Device filter"
					bind:value={deviceFilter}
					oninput={resetPage}
				/>
				<div class="example-search">
					<Search size={16} /><Input
						placeholder="Search script or meaning…"
						aria-label="Search examples"
						bind:value={search}
						oninput={resetPage}
					/>
				</div>
				<Button onclick={openCreateDialog}><Plus data-icon="inline-start" /> New example</Button>
			</div>
		</div>
	{:else}
		<div class="toolbar mb-0 px-4 pt-4 pb-3">
			<Input
				class="w-28"
				placeholder="Source…"
				aria-label="Function source filter"
				bind:value={fnSource}
				oninput={resetPage}
			/>
			<Input
				class="w-44"
				placeholder="Function name…"
				aria-label="Function name filter"
				bind:value={fnName}
				oninput={resetPage}
			/>
			<Input
				class="w-36"
				placeholder="Device…"
				aria-label="Device filter"
				bind:value={deviceFilter}
				oninput={resetPage}
			/>
			<Input
				class="min-w-56"
				placeholder="Search script or meaning…"
				aria-label="Search examples"
				bind:value={search}
				oninput={resetPage}
			/>
			<span class="spacer"></span>
			<span class="muted">{loading ? 'Loading…' : `${total} example(s)`}</span>
			<Button onclick={openCreateDialog}>New example</Button>
		</div>
	{/if}

	{#if loadError}
		<div class="px-4">
			<div class="banner-error" role="alert">
				{loadError}<Button
					variant="outline"
					size="sm"
					disabled={loading}
					onclick={() => void load(currentFilters())}><RefreshCw size={16} /> Retry examples</Button
				>
			</div>
		</div>
	{/if}

	<Table class="border-0">
		<TableHeader>
			<TableRow>
				{#if !fixed}
					<SortHead
						label="Function"
						col="function"
						active={sortKey}
						dir={sortDir}
						onclick={() => applySort('function')}
					/>
				{/if}
				<SortHead
					label="Device"
					col="device"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('device')}
				/>
				<SortHead
					label="Script"
					col="script"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('script')}
				/>
				<SortHead
					label="Meaning"
					col="meaning"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('meaning')}
				/>
				<TableHead></TableHead>
			</TableRow>
		</TableHeader>
		<TableBody>
			{#each view as row (row.id)}
				<TableRow>
					{#if !fixed}
						<TableCell>
							{row.function_source}/{row.function_name}
							{#if !row.function_exists}
								<Tooltip content="No function with this source and name exists">
									{#snippet children(tp)}
										<Badge variant="warn" {...tp} tabindex={0}>orphan</Badge>
									{/snippet}
								</Tooltip>
							{/if}
						</TableCell>
					{/if}
					<TableCell class="text-muted-foreground text-xs">{row.device_name || 'Not set'}</TableCell
					>
					<TableCell class="max-w-[420px] font-mono text-xs break-words whitespace-pre-wrap">
						{row.script}
						{#if usageFields}
							{#await usageFields}
								<ExampleFieldUsage script={row.script} loading />
							{:then fields}
								<ExampleFieldUsage
									script={row.script}
									parameters={fields.parameters}
									criteria={fields.criteria}
								/>
							{:catch}
								<ExampleFieldUsage script={row.script} unavailable />
							{/await}
						{/if}
					</TableCell>
					<TableCell class="max-w-80 text-[13px] break-words whitespace-pre-wrap"
						>{row.meaning}</TableCell
					>
					<TableCell class="text-right whitespace-nowrap">
						<Button
							size="sm"
							variant="ghost"
							aria-label={`Copy example ${row.id}`}
							title="Copy script"
							onclick={() => copyRow(row)}
							>{#if copied === row.id}<Check size={16} />{:else}<Copy size={16} />{/if}</Button
						>
						<Button size="sm" variant="ghost" onclick={() => openEditDialog(row)}
							><Pencil data-icon="inline-start" /> Edit</Button
						>
						<Button
							size="sm"
							variant="ghost"
							disabled={deleting}
							onclick={() => requestRemove(row)}
						>
							<Trash2 data-icon="inline-start" /> Delete
						</Button>
					</TableCell>
				</TableRow>
			{:else}
				<TableRow>
					<TableCell colspan={fixed ? 4 : 5} class="banner-empty text-center" aria-live="polite">
						{loading ? 'Loading…' : 'No examples yet.'}
					</TableCell>
				</TableRow>
			{/each}
		</TableBody>
	</Table>

	<Pagination class="mt-0 border-t px-4 pt-2 pb-3" bind:page {total} bind:pageSize />
	<p role="status" class="text-muted-foreground px-4 pb-3 text-sm">{copyStatus}</p>
</div>

<ExampleDialog
	bind:open={dialogOpen}
	editRow={editTarget}
	fixedFunction={fixed}
	prefillDevice={editTarget?.device_name ?? defaultDevice}
	onsaved={async () => {
		await load(currentFilters());
	}}
/>

<AlertDialog
	open={confirmOpen}
	onOpenChange={(o) => (confirmOpen = o)}
	title="Delete example"
	description={confirmRow
		? `Delete this example of ${confirmRow.function_name}? This cannot be undone.`
		: ''}
	confirmLabel="Delete"
	onConfirm={confirmRemove}
/>

<style>
	.example-heading {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 12px 20px;
		padding: 20px 0 14px;
	}
	.example-actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		margin-left: auto;
		max-width: 100%;
		flex: 0 1 560px;
	}
	.example-actions > :global(button) {
		flex-shrink: 0;
	}
	.example-actions > :global(input) {
		flex: 0 1 120px;
		min-width: 0;
	}
	.example-search {
		position: relative;
		flex: 1 1 200px;
		min-width: 0;
	}
	.example-search > :global(svg) {
		position: absolute;
		top: 8px;
		left: 10px;
		color: var(--muted-foreground);
		pointer-events: none;
	}
	.example-search :global(input) {
		padding-left: 32px;
	}
	.example-heading strong {
		font-size: 13px;
	}
	.example-heading p {
		color: var(--muted-foreground);
		font-size: 13px;
		margin: 3px 0 0;
	}
	@media (max-width: 600px) {
		.example-heading {
			align-items: start;
			flex-wrap: wrap;
		}
		.example-actions {
			flex-basis: 100%;
		}
	}
</style>
