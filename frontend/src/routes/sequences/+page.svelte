<script lang="ts">
	import { resolve } from '$app/paths';
	import { listSequences, type SequenceListItem } from '#lib/sequences.ts';
	import { RefreshCw } from '@lucide/svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir } from '#lib/table-sort.ts';
	import {
		Table,
		TableHeader,
		TableBody,
		TableRow,
		TableCell
	} from '#lib/components/ui/table/index.ts';

	// Creation happens from the composer's "Add to sequence" action: a sequence
	// needs at least one step, and the composer is the step source.
	let items = $state<SequenceListItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let loadError = $state('');
	let search = $state('');
	let searchDebounced = $state('');
	let page = $state(1);
	let pageSize = $state(20);
	let searchTimer: ReturnType<typeof setTimeout> | undefined;

	// Reload whenever any list parameter changes; the search term enters
	// through its debounced snapshot so typing does not fire a request
	// per keystroke.
	$effect(() => {
		void searchDebounced;
		void page;
		void pageSize;
		load();
	});

	async function load() {
		loading = true;
		loadError = '';
		try {
			const res = await listSequences({ search: searchDebounced, page, page_size: pageSize });
			items = res.items;
			total = res.total;
		} catch (e) {
			items = [];
			total = 0;
			loadError = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	function scheduleSearch() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			searchDebounced = search.trim();
			page = 1;
		}, 300);
	}

	// ---- sorting (ticket 09) ------------------------------------------------
	// The sequences API paginates server-side without sort params, so
	// sorting applies to the loaded page through the shared sortRows
	// helper. '' keeps the API's default order until a header is clicked.
	const SEQUENCE_SORT_KEYS: Record<
		string,
		(row: SequenceListItem) => string | number | null | undefined
	> = {
		name: (r) => r.name,
		description: (r) => r.description || '',
		steps: (r) => r.step_count,
		updated: (r) => r.updated_at
	};

	let sortKey = $state('');
	let sortDir = $state<SortDir>('asc');

	function applySort(col: string) {
		({ key: sortKey, dir: sortDir } = nextSort(sortKey, sortDir, col));
	}

	const view = $derived(
		sortKey === '' ? items : sortRows(items, SEQUENCE_SORT_KEYS[sortKey]!, sortDir)
	);
</script>

<h1 class="mt-2 mb-4 text-xl font-semibold">Sequences</h1>

<p class="muted">
	Named, ordered multi-function scripts built from the
	<a href={resolve('functions')}>functions</a>
	composer. Sequences live only in this app; the workbook's six sheets are untouched. Create one via a
	function's "Compose script" → "Add to sequence".
</p>

<div class="filters" role="region" aria-label="Sequence filters">
	<Input
		class="search-input"
		type="search"
		placeholder="Search name or description…"
		aria-label="Search sequences"
		bind:value={search}
		oninput={scheduleSearch}
	/>
</div>
<div class="catalog-toolbar">
	<span class="muted">
		{loading
			? 'Loading…'
			: loadError
				? 'Count unavailable'
				: `${total} sequence${total === 1 ? '' : 's'}`}
	</span>
</div>

{#if loadError}
	<div class="banner-error" role="alert">
		{loadError}<Button variant="outline" onclick={() => load()}
			><RefreshCw size={16} /> Retry sequences</Button
		>
	</div>
{/if}

<div class="sequence-table">
	<Table class="border-0">
		<TableHeader>
			<TableRow>
				<SortHead
					label="Name"
					col="name"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('name')}
				/>
				<SortHead
					label="Description"
					col="description"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('description')}
				/>
				<SortHead
					label="Steps"
					col="steps"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('steps')}
				/>
				<SortHead
					label="Updated"
					col="updated"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('updated')}
				/>
			</TableRow>
		</TableHeader>
		<TableBody>
			{#each view as s (s.id)}
				<TableRow>
					<TableCell>
						<a
							class="text-primary hover:underline"
							href={resolve('/sequences/[id]', { id: String(s.id) })}>{s.name}</a
						>
					</TableCell>
					<TableCell>{s.description || 'Not set'}</TableCell>
					<TableCell>{s.step_count}</TableCell>
					<TableCell class="text-muted-foreground text-xs">{s.updated_at.slice(0, 10)}</TableCell>
				</TableRow>
			{:else}
				<TableRow>
					<TableCell colspan={4} class="banner-empty text-center" aria-live="polite">
						{#if loading}
							Loading…
						{:else if loadError}
							Sequences could not be loaded.
						{:else if total === 0}
							No sequences yet.
						{:else}
							No sequences match the search.
						{/if}
					</TableCell>
				</TableRow>
			{/each}
		</TableBody>
	</Table>

	<Pagination class="mt-0 border-t px-4 pt-2 pb-3" bind:page {total} bind:pageSize />
</div>

<style>
	.filters,
	.catalog-toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		margin-bottom: 16px;
	}
	.filters :global(.search-input) {
		flex: 1 1 220px;
		min-width: 0;
	}
	.sequence-table {
		min-width: 0;
		overflow-x: auto;
		border: 1px solid var(--border);
		border-radius: var(--radius);
	}
</style>
