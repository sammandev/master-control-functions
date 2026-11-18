<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page as routePage } from '$app/state';
	import { goto, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { ChevronRight, Plus, RefreshCw, Search } from '@lucide/svelte';
	import {
		listDevices,
		listDeviceFunctions,
		devicePath,
		type DeviceSummary
	} from '#lib/devices.ts';
	import FunctionDetail from '#lib/components/FunctionDetail.svelte';
	import FunctionDialog from '#lib/components/FunctionDialog.svelte';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import {
		listFunctions,
		getFunctionFacets,
		type FunctionRow,
		type FunctionFacets,
		type ListParams
	} from '#lib/api.ts';
	import Select from '#lib/components/ui/select.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import { sortRows } from '#lib/table-sort.ts';

	let search = $derived(routePage.url.searchParams.get('search') ?? '');
	const params = $derived.by((): ListParams => {
		const query = routePage.url.searchParams;
		return {
			search: query.get('search') || undefined,
			source: query.get('source') || undefined,
			category: query.get('category') || undefined,
			execution_role: query.get('execution_role') || undefined,
			execution_phase: query.get('execution_phase') || undefined,
			sort: ['name', 'interface', 'updated_at'].includes(query.get('sort') ?? '')
				? query.get('sort')!
				: 'name',
			order: query.get('order') === 'desc' ? 'desc' : 'asc',
			page: Math.max(1, Math.floor(Number(query.get('page')) || 1)),
			page_size: [10, 20, 50].includes(Number(query.get('page_size')))
				? Number(query.get('page_size'))
				: 20
		};
	});
	const deviceFilter = $derived(
		routePage.url.searchParams.get('device') ?? routePage.url.searchParams.get('interface')
	);
	let retry = $state(0);
	const requestKey = $derived(JSON.stringify({ params, scope: deviceFilter, retry }));
	let revision = 0;
	let facets = $state<FunctionFacets | null>(null);
	let devices = $state<DeviceSummary[]>([]);
	let metadataError = $state('');
	const content = $derived.by(() => {
		const request: { params: ListParams; scope: string | null } = JSON.parse(requestKey);
		return load(request.params, request.scope);
	});
	function setQuery(key: string, value: string, reset = true, replaceState = false) {
		const query = new SvelteURLSearchParams(routePage.url.searchParams.toString());
		if (key === 'device') query.delete('interface');
		if (value) query.set(key, value);
		else query.delete(key);
		if (reset) {
			query.delete('page');
			query.delete('function');
		}

		return goto(resolve(`functions?${query.toString()}`), {
			reset: false,
			replace: replaceState
		});
	}
	function selectFunction(id: number) {
		void setQuery('function', String(id), false);
	}
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	function scheduleSearch(term: string) {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			if (term !== (routePage.url.searchParams.get('search') ?? '')) void setQuery('search', term);
		}, 250);
	}
	function clearFilters() {
		clearTimeout(searchTimer);
		search = '';
		const query = new SvelteURLSearchParams(routePage.url.searchParams.toString());
		for (const key of [
			'search',
			'source',
			'category',
			'execution_role',
			'execution_phase',
			'device',
			'interface',
			'page',
			'function'
		])
			query.delete(key);
		void goto(resolve(`functions?${query.toString()}`), { reset: false });
	}

	afterNavigate(({ shallow, type }) => {
		if (shallow && type === 'goto') return;

		return clearTimeout(searchTimer);
	});

	onDestroy(() => clearTimeout(searchTimer));
	async function loadMetadata() {
		metadataError = '';
		try {
			const [values, result] = await Promise.all([getFunctionFacets(), listDevices()]);
			facets = values;
			devices = result.items;
		} catch (error) {
			metadataError = error instanceof Error ? error.message : String(error);
		}
	}
	onMount(() => {
		void loadMetadata();
	});
	async function load(query: ListParams, scope = deviceFilter) {
		const attempt = ++revision;
		let res;
		if (scope === null) {
			res = await listFunctions(query);
		} else {
			const device = scope === '_global' ? '' : scope;
			const first = await listDeviceFunctions(device, 1, 100);
			const rows = [...first.items];
			for (let next = 2; rows.length < first.total; next++) {
				if (attempt !== revision) return { items: [], total: 0 };
				const more = await listDeviceFunctions(device, next, 100);
				if (!more.items.length) throw new Error('The function list changed. Please retry.');
				rows.push(...more.items);
			}
			const matches = rows.filter(
				(row) =>
					(!query.search ||
						`${row.name} ${row.description ?? ''}`
							.toLowerCase()
							.includes(query.search.trim().toLowerCase())) &&
					(!query.source || row.source === query.source) &&
					(!query.interface || row.interface === query.interface) &&
					(!query.category || row.category === query.category) &&
					(!query.execution_role || row.execution_role === query.execution_role) &&
					(!query.execution_phase || row.execution_phase === query.execution_phase)
			);
			const sorted = sortRows(
				matches,
				(row) =>
					query.sort === 'interface'
						? row.interface
						: query.sort === 'updated_at'
							? row.updated_at
							: row.name,
				query.order ?? 'asc'
			);
			const start = ((query.page ?? 1) - 1) * (query.page_size ?? 20);
			res = {
				total: matches.length,
				items: sorted.slice(start, start + (query.page_size ?? 20)).map((row): FunctionRow => ({
					...row,
					description: row.description ?? '',
					category: row.category ?? '',
					execution_role: row.execution_role ?? '',
					execution_phase: row.execution_phase ?? '',
					created_at: ''
				}))
			};
		}
		if (attempt !== revision) return res;
		const last = Math.max(1, Math.ceil(res.total / (query.page_size ?? 20)));
		if ((query.page ?? 1) > last) {
			await setQuery('page', String(last), false, true);
			return res;
		}
		if (
			routePage.url.searchParams.has('function') &&
			!res.items.some((row) => row.id === Number(routePage.url.searchParams.get('function')))
		)
			await setQuery('function', res.items[0] ? String(res.items[0].id) : '', false, true);
		return res;
	}

	let dialogMode = $state<'create' | 'edit' | 'delete' | null>(null);
	let editingFunction = $state<FunctionRow | undefined>();
	function openCreate() {
		editingFunction = undefined;
		dialogMode = 'create';
	}
	function openEdit(row: FunctionRow) {
		editingFunction = row;
		dialogMode = 'edit';
	}
	function requestRemove(row: FunctionRow) {
		editingFunction = row;
		dialogMode = 'delete';
	}
</script>

<svelte:head><title>Functions | MasterControl Functions Database</title></svelte:head>

<div class="directory-heading">
	<div>
		<p class="eyebrow">Reference catalog</p>
		<div class="directory-title">
			<h1>Functions</h1>
			<span role="status"
				>{#await content}Loading functions...{:then result}{result.total} functions{:catch}0
					functions{/await}</span
			>
		</div>
	</div>
	<Button onclick={openCreate}><Plus size={16} /> New function</Button>
</div>

<div class="filters" role="region" aria-label="Function filters">
	<div class="filter-field">
		<span>Source</span><Select
			ariaLabel="Filter by source"
			placeholder="All sources"
			value={params.source ?? ''}
			options={[
				{ value: '', label: 'All sources' },
				...(facets?.sources ?? []).map((v) => ({ value: v, label: v }))
			]}
			onchange={(value) => {
				void setQuery('source', value);
			}}
		/>
	</div>
	<div class="filter-field">
		<span>Category</span><Select
			ariaLabel="Filter by category"
			placeholder="All categories"
			value={params.category ?? ''}
			options={[
				{ value: '', label: 'All categories' },
				...(facets?.categories ?? []).map((v) => ({ value: v, label: v }))
			]}
			onchange={(value) => {
				void setQuery('category', value);
			}}
		/>
	</div>
	<div class="filter-field">
		<span>Execution role</span><Select
			ariaLabel="Filter by execution role"
			placeholder="All roles"
			value={params.execution_role ?? ''}
			options={[
				{ value: '', label: 'All roles' },
				...(facets?.execution_roles ?? []).map((v) => ({ value: v, label: v }))
			]}
			onchange={(value) => {
				void setQuery('execution_role', value);
			}}
		/>
	</div>
	<div class="filter-field">
		<span>Execution phase</span><Select
			ariaLabel="Filter by execution phase"
			placeholder="All phases"
			value={params.execution_phase ?? ''}
			options={[
				{ value: '', label: 'All phases' },
				...(facets?.execution_phases ?? []).map((v) => ({ value: v, label: v }))
			]}
			onchange={(value) => {
				void setQuery('execution_phase', value);
			}}
		/>
	</div>
	<div class="filter-field">
		<span>Device / Interface</span><DevicePicker
			device={deviceFilter === '_global' ? '' : (deviceFilter ?? '_all')}
			summaries={devices}
			label="Filter by Device"
			allowAll
			fullWidth
			compact
			onValueChange={(value) => {
				void setQuery('device', value === '_all' ? '' : value || '_global');
			}}
		/>
	</div>
	<Button class="clear-filters" variant="outline" onclick={clearFilters}>Clear filters</Button>
</div>
<div class="catalog-toolbar">
	{#if deviceFilter !== null}<a
			href={resolve('/devices/[name]?tab=functions', {
				name: devicePath(deviceFilter === '_global' ? '' : deviceFilter)
			})}>Open Device workspace</a
		>{:else}<span class="muted">All-Device discovery</span>{/if}
</div>
{#if metadataError}<div class="banner-error" role="alert">
		{metadataError}<Button
			variant="outline"
			onclick={() => {
				void loadMetadata();
			}}><RefreshCw size={16} /> Retry filters</Button
		>
	</div>{/if}

<div class="functions-workspace">
	<aside class="function-list" aria-label="Functions list">
		<div class="rail-heading">
			<h2>Functions</h2>
			<span>{#await content}...{:then result}{result.total}{:catch}0{/await}</span>
		</div>
		<div class="catalog-controls">
			<div class="function-search">
				<Search size={16} /><Input
					placeholder="Search name or description…"
					aria-label="Search functions"
					bind:value={search}
					oninput={(event) => scheduleSearch(event.currentTarget.value)}
				/>
			</div>
			<div class="catalog-sorting">
				<span>Catalog sorting</span>
				<div>
					<Select
						ariaLabel="Sort functions"
						value={params.sort ?? 'name'}
						options={[
							{ value: 'name', label: 'Name' },
							{ value: 'interface', label: 'Device / Interface' },
							{ value: 'updated_at', label: 'Updated' }
						]}
						onchange={(value) => {
							void setQuery('sort', value);
						}}
					/>
					<Select
						ariaLabel="Sort direction"
						value={params.order ?? 'asc'}
						options={[
							{ value: 'asc', label: 'Ascending' },
							{ value: 'desc', label: 'Descending' }
						]}
						onchange={(value) => {
							void setQuery('order', value);
						}}
					/>
				</div>
			</div>
		</div>
		{#await content}<p class="banner-empty" role="status">Loading functions...</p>
		{:then result}
			{@const selectedFunction =
				result.items.find((row) => row.id === Number(routePage.url.searchParams.get('function'))) ??
				result.items[0]}
			{#each result.items as row (row.id)}
				<button
					class:chosen={selectedFunction?.id === row.id}
					aria-pressed={selectedFunction?.id === row.id}
					aria-label={`Select ${row.name}`}
					onclick={() => selectFunction(row.id)}
				>
					<strong>{row.name}</strong><span><Badge>{row.source}</Badge> {row.interface}</span>
					{#if row.description}<small>{row.description}</small>{/if}
					<ChevronRight size={16} />
				</button>
			{:else}<p role="status" class="banner-empty">
					No functions match. Adjust the search or create one.
				</p>{/each}
			<Pagination
				page={params.page}
				total={result.total}
				pageSize={params.page_size}
				onPageChange={(value) => {
					void setQuery('page', String(value), false);
				}}
				onPageSizeChange={(value) => {
					void setQuery('page_size', String(value));
				}}
			/>
		{:catch error}<div class="banner-error" role="alert">
				{error instanceof Error ? error.message : String(error)}<Button
					variant="outline"
					onclick={() => retry++}><RefreshCw size={16} /> Retry functions</Button
				>
			</div>{/await}
	</aside>
	<div class="detail-pane">
		{#await content}<p class="banner-empty" role="status">Loading function details...</p>
		{:then result}
			{@const selectedFunction =
				result.items.find((row) => row.id === Number(routePage.url.searchParams.get('function'))) ??
				result.items[0]}
			{#if selectedFunction}{#key selectedFunction}<FunctionDetail
						fn={selectedFunction}
						onEdit={openEdit}
						onDelete={requestRemove}
					/>{/key}
			{:else}<p role="status" class="banner-empty">No Function selected.</p>{/if}
		{:catch}<p class="banner-empty">Unable to load function details.</p>{/await}
	</div>
</div>

{#if dialogMode}<FunctionDialog
		mode={dialogMode}
		fn={editingFunction}
		defaults={{
			source: params.source ?? 'DLL',
			interface: deviceFilter && deviceFilter !== '_global' ? deviceFilter : '',
			category: params.category ?? '',
			execution_role: params.execution_role ?? '',
			execution_phase: params.execution_phase ?? ''
		}}
		onClose={() => (dialogMode = null)}
		onSaved={async (row) => {
			await setQuery('function', row ? String(row.id) : '', false);
			retry++;
		}}
	/>{/if}

<style>
	.directory-heading,
	.directory-title {
		display: flex;
		align-items: center;
		gap: 16px;
	}
	.directory-heading {
		justify-content: space-between;
		flex-wrap: wrap;
		margin: 8px 0 28px;
	}
	.directory-title {
		gap: 8px;
		align-items: baseline;
	}
	h1 {
		font-size: 28px;
		font-weight: 650;
		line-height: 1.3;
	}
	.eyebrow {
		font-size: 11px;
		font-weight: 600;
		color: var(--muted-foreground);
		margin-bottom: 6px;
		text-transform: uppercase;
	}
	.directory-title span {
		color: var(--muted-foreground);
		font-size: 13px;
	}
	.filters,
	.catalog-toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
		margin-bottom: 24px;
	}
	.filters {
		align-items: end;
		gap: 10px;
	}
	.filter-field {
		display: grid;
		gap: 5px;
		flex: 1 1 150px;
		min-width: 0;
		color: var(--muted-foreground);
		font-size: 11px;
		font-weight: 600;
	}
	.filters :global(.clear-filters) {
		flex: 1 1 130px;
	}
	.function-search {
		position: relative;
		min-width: 0;
	}
	.catalog-controls {
		display: grid;
		gap: 12px;
		padding: 14px 16px;
		border-bottom: 1px solid var(--border);
	}
	.catalog-sorting {
		display: grid;
		gap: 5px;
		font-size: 11px;
		font-weight: 600;
		color: var(--muted-foreground);
	}
	.catalog-sorting > div {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
		gap: 8px;
	}
	.function-search :global(svg) {
		position: absolute;
		left: 12px;
		top: 11px;
		color: var(--muted-foreground);
	}
	.function-search :global(input) {
		padding-left: 36px;
	}
	.catalog-toolbar a {
		color: var(--primary);
	}
	.functions-workspace {
		display: grid;
		grid-template-columns: minmax(240px, 310px) minmax(0, 1fr);
		border-block: 1px solid var(--border);
		background: var(--card);
		min-height: 600px;
	}
	.function-list {
		min-width: 0;
		position: sticky;
		top: 16px;
		align-self: start;
		max-height: min(640px, calc(100dvh - 32px));
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
		border-right: 1px solid var(--border);
	}
	.rail-heading {
		display: flex;
		justify-content: space-between;
		padding: 20px 16px 12px;
	}
	.rail-heading h2 {
		font-size: 14px;
		font-weight: 650;
	}
	.rail-heading span {
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.function-list > button {
		display: grid;
		position: relative;
		gap: 6px;
		width: 100%;
		min-width: 0;
		text-align: left;
		padding: 18px 28px 18px 16px;
		border-left: 3px solid transparent;
		border-top: 1px solid var(--border);
	}
	.function-list strong {
		font-family: var(--mono);
		font-size: 12px;
		line-height: 1.6;
		overflow-wrap: anywhere;
	}
	.function-list span,
	small {
		font-size: 12px;
		color: var(--muted-foreground);
		overflow-wrap: anywhere;
	}
	.function-list small {
		display: -webkit-box;
		line-clamp: 2;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.function-list > button :global(svg) {
		position: absolute;
		top: 40px;
		right: 4px;
	}
	.function-list button:hover,
	.function-list button.chosen {
		background: var(--device-soft);
	}
	.function-list button.chosen {
		border-left-color: var(--device-accent);
	}
	.function-list button:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: -2px;
	}
	.detail-pane {
		min-width: 0;
	}
	@media (max-width: 850px) {
		.functions-workspace {
			grid-template-columns: minmax(0, 1fr);
		}
		.function-list {
			position: static;
			border-right: 0;
			border-bottom: 1px solid var(--border);
			max-height: 360px;
			overflow-y: auto;
		}
	}
</style>
