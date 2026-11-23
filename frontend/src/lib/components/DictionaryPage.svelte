<script lang="ts">
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import { onMount, onDestroy } from 'svelte';
	import { page as routePage } from '$app/state';
	import { goto, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { listDefinitions, type DefinitionRow, type DefinitionKind } from '#lib/definitions.ts';
	import { listDevices, type DeviceSummary } from '#lib/devices.ts';
	import DefinitionDialog from '#lib/components/DefinitionDialog.svelte';
	import DefinitionDetail from '#lib/components/DefinitionDetail.svelte';
	import DefinitionDeleteDialog from '#lib/components/DefinitionDeleteDialog.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import { Plus, RefreshCw, ChevronRight } from '@lucide/svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import { toast } from 'svelte-sonner';

	// Shared CRUD page for the Parameter definition and Criteria definition
	// dictionaries. kind picks the endpoint; everything else is identical.
	// The create/edit form lives in DefinitionDialog (shared with the
	// device page's definition shortcuts).
	let { kind, title }: { kind: DefinitionKind; title: string } = $props();
	let workspace = $state<HTMLDivElement | null>(null);
	let pendingHeight = $state(600);

	// ---- list state -------------------------------------------------------
	let search = $derived(routePage.url.searchParams.get('search') ?? '');
	const params = $derived.by(() => {
		const query = routePage.url.searchParams;
		const device = query.get('device');
		return {
			search: query.get('search') || undefined,
			device: device === '_global' ? '' : (device ?? undefined),
			sort: ['name', 'type', 'updated_at'].includes(query.get('sort') ?? '')
				? query.get('sort')!
				: 'name',
			order: query.get('order') === 'desc' ? ('desc' as const) : ('asc' as const),
			page: Math.max(1, Math.floor(Number(query.get('page')) || 1)),
			page_size: [10, 20, 50].includes(Number(query.get('page_size')))
				? Number(query.get('page_size'))
				: 20
		};
	});
	let retry = $state(0);
	const deviceFilter = $derived(routePage.url.searchParams.get('device'));
	let devices = $state<DeviceSummary[]>([]);
	let deviceError = $state('');
	async function loadDevices() {
		deviceError = '';
		try {
			devices = (await listDevices()).items;
		} catch (error) {
			deviceError = error instanceof Error ? error.message : String(error);
		}
	}
	onMount(() => {
		void loadDevices();
	});
	const requestKey = $derived(JSON.stringify({ kind, params, retry }));
	let revision = 0;
	const content = $derived.by(() => {
		const request: { kind: DefinitionKind; params: typeof params } = JSON.parse(requestKey);
		return load(request.kind, request.params);
	});
	function setQuery(key: string, value: string, reset = true, replaceState = false) {
		pendingHeight = workspace?.getBoundingClientRect().height ?? pendingHeight;
		const query = new SvelteURLSearchParams(routePage.url.searchParams.toString());
		if (value) query.set(key, value);
		else query.delete(key);
		if (reset) {
			query.delete('page');
			query.delete('definition');
		}
		return goto(resolve(`${kind}?${query.toString()}`), {
			reset: false,
			replace: replaceState
		});
	}
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	function scheduleSearch(term: string) {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			if (term !== (routePage.url.searchParams.get('search') ?? '')) void setQuery('search', term);
		}, 250);
	}

	afterNavigate(({ shallow, type }) => {
		if (shallow && type === 'goto') return;

		return clearTimeout(searchTimer);
	});

	onDestroy(() => clearTimeout(searchTimer));
	// The page holding the definition under query and device, or null when they hide it.
	async function definitionPage(query: typeof params, id: number) {
		const result = await listDefinitions(kind, { ...query, definition_id: id });
		return result.items.some((item) => item.id === id) ? result.page : null;
	}
	async function selectCreatedDefinition(row: DefinitionRow) {
		clearTimeout(searchTimer);
		const from = routePage.url.href;
		const location = new SvelteURLSearchParams(routePage.url.searchParams.toString());
		const query = params;
		let found = await definitionPage(query, row.id);
		if (found === null) {
			location.delete('search');
			if (query.device !== undefined) location.set('device', row.device_name || '_global');
			found = await definitionPage(
				{
					...query,
					search: undefined,
					device: query.device === undefined ? undefined : row.device_name
				},
				row.id
			);
		}
		if (routePage.url.href !== from) return;
		if (found === null) throw new Error('The saved definition is no longer available.');

		if (found > 1) location.set('page', String(found));
		else location.delete('page');
		location.set('definition', String(row.id));
		pendingHeight = workspace?.getBoundingClientRect().height ?? pendingHeight;
		await goto(resolve(`${kind}?${location.toString()}`), { reset: false });
	}
	async function load(endpoint: DefinitionKind, query: typeof params) {
		const attempt = ++revision;
		const result = await listDefinitions(endpoint, query);
		if (attempt !== revision) return result;
		const last = Math.max(1, Math.ceil(result.total / query.page_size));
		if (query.page > last) {
			await setQuery('page', String(last), false, true);
			return result;
		}
		if (
			routePage.url.searchParams.has('definition') &&
			!result.items.some((row) => row.id === Number(routePage.url.searchParams.get('definition')))
		)
			await setQuery('definition', result.items[0] ? String(result.items[0].id) : '', false, true);
		return result;
	}

	// ---- dialog state -----------------------------------------------------
	// The row being edited (null = create); DefinitionDialog owns the form.
	let dialogOpen = $state(false);
	let editTarget = $state<DefinitionRow | null>(null);

	function openCreate() {
		editTarget = null;
		dialogOpen = true;
	}

	function openEdit(row: DefinitionRow) {
		editTarget = row;
		dialogOpen = true;
	}

	let confirmRow = $state<DefinitionRow | null>(null);

	function requestRemove(row: DefinitionRow) {
		confirmRow = row;
	}
</script>

<h1 class="mt-2 mb-4 text-xl font-semibold">{title}</h1>

<div class="filters">
	<Input
		class="h-10 min-w-48 flex-1"
		placeholder="Search name or description..."
		aria-label={`Search ${title.toLowerCase()}`}
		bind:value={search}
		oninput={(event) => scheduleSearch(event.currentTarget.value)}
	/>
	<DevicePicker
		device={deviceFilter === '_global' ? '' : (deviceFilter ?? '_all')}
		summaries={devices}
		label="Filter definitions by Device"
		allowAll
		compact
		comfortable
		onValueChange={(value) => {
			void setQuery('device', value === '_all' ? '' : value || '_global');
		}}
	/>
	<Select
		class="h-10 w-44"
		ariaLabel="Sort definitions by"
		value={params.sort}
		options={[
			{ value: 'name', label: 'Name' },
			{ value: 'type', label: 'Type' },
			{ value: 'updated_at', label: 'Updated' }
		]}
		onchange={(value) => {
			void setQuery('sort', value);
		}}
	/>
	<Select
		class="h-10 w-44"
		ariaLabel="Sort direction"
		value={params.order}
		options={[
			{ value: 'asc', label: 'Ascending' },
			{ value: 'desc', label: 'Descending' }
		]}
		onchange={(value) => {
			void setQuery('order', value);
		}}
	/>
</div>
<div class="catalog-toolbar">
	<div class="muted">
		<span
			>{#await content}Loading...{:then result}{result.total} entries{:catch}Count unavailable{/await}</span
		>
		<span
			>{deviceFilter === null
				? 'All-Device discovery'
				: deviceFilter === '_global'
					? 'Global definitions'
					: `Device: ${deviceFilter}`}</span
		>
	</div>
	<Button onclick={openCreate}><Plus size={16} /> New</Button>
</div>
{#if deviceError}<div class="banner-error" role="alert">
		{deviceError}<Button
			variant="outline"
			onclick={() => {
				void loadDevices();
			}}><RefreshCw size={16} /> Retry Devices</Button
		>
	</div>{/if}
{#await content}<div
		class="dictionary-workspace"
		aria-busy="true"
		style:min-height={`${pendingHeight}px`}
	>
		<p class="banner-empty" role="status">Loading definitions...</p>
	</div>
{:then result}
	{@const selected =
		result.items.find((row) => row.id === Number(routePage.url.searchParams.get('definition'))) ??
		result.items[0]}
	<div class="dictionary-workspace" bind:this={workspace}>
		<aside class="definition-list" aria-label="Definitions list">
			<div class="rail-heading">
				<h2>{kind === 'parameter-definitions' ? 'Parameters' : 'Criteria'}</h2>
				<span>{result.total}</span>
			</div>
			{#each result.items as row (row.id)}
				<button
					class:chosen={selected?.id === row.id}
					aria-pressed={selected?.id === row.id}
					aria-label={`Select ${row.name} (${row.device_name || 'Global'})`}
					onclick={() => {
						void setQuery('definition', String(row.id), false);
					}}
				>
					<strong>{row.name}</strong><span
						><Badge class="max-w-full break-all whitespace-normal"
							>{row.device_name || 'Global'}</Badge
						>
						{row.type}</span
					>
					{#if row.description}<small>{row.description}</small>{/if}
					<ChevronRight size={16} />
				</button>
			{:else}<p class="banner-empty" role="status">
					No entries match. Adjust the search or create one.
				</p>{/each}
			<Pagination
				variant="rail"
				page={params.page}
				pageSize={params.page_size}
				total={result.total}
				onPageChange={(value) => {
					void setQuery('page', String(value), false);
				}}
				onPageSizeChange={(value) => {
					void setQuery('page_size', String(value));
				}}
			/>
		</aside>
		<div class="detail-pane">
			{#if selected}<DefinitionDetail
					row={selected}
					onEdit={openEdit}
					onDelete={requestRemove}
				/>{:else}<p class="banner-empty" role="status">No definition selected.</p>{/if}
		</div>
	</div>
{:catch error}<div class="banner-error" role="alert">
		{error instanceof Error ? error.message : String(error)}<Button
			variant="outline"
			onclick={() => retry++}><RefreshCw size={16} /> Retry definitions</Button
		>
	</div>{/await}

<DefinitionDialog
	bind:open={dialogOpen}
	{kind}
	{title}
	editRow={editTarget}
	prefillDevice={deviceFilter === '_global' ? '' : (deviceFilter ?? '')}
	onsaved={async (row) => {
		if (editTarget) await setQuery('definition', String(row.id), false);
		else {
			try {
				await selectCreatedDefinition(row);
			} catch (error) {
				toast.error('Definition saved, but could not open it', {
					description: error instanceof Error ? error.message : String(error)
				});
			}
		}
		retry++;
	}}
/>

{#if confirmRow}
	<DefinitionDeleteDialog
		{kind}
		row={confirmRow}
		onClose={() => {
			confirmRow = null;
		}}
		onDeleted={() => {
			retry++;
		}}
	/>{/if}

<style>
	.filters {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.catalog-toolbar {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		padding: 16px 0;
	}
	.catalog-toolbar span {
		font-size: 12px;
	}
	.catalog-toolbar span + span {
		margin-left: 8px;
	}
	.dictionary-workspace {
		display: grid;
		grid-template-columns: minmax(240px, 310px) minmax(0, 1fr);
		grid-template-rows: minmax(0, 1fr);
		height: max(600px, calc(100dvh - 240px));
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--card);
	}
	.definition-list {
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
		border-right: 1px solid var(--border);
	}
	.definition-list > button {
		position: relative;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		width: 100%;
		border: 0;
		border-bottom: 1px solid var(--border);
		border-left: 3px solid transparent;
		padding: 18px 28px 18px 16px;
		gap: 6px;
		text-align: left;
		color: var(--foreground);
		background: transparent;
		cursor: pointer;
	}
	.definition-list > button:hover {
		background: var(--accent);
	}
	.definition-list > button.chosen {
		background: var(--device-soft);
		border-left-color: var(--device-accent);
	}
	.definition-list > button:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: -2px;
	}
	.definition-list strong {
		font: 600 12px/1.6 var(--mono);
		overflow-wrap: anywhere;
	}
	.definition-list > button > span,
	.definition-list small {
		font-size: 12px;
		overflow-wrap: anywhere;
		max-width: 100%;
	}
	.detail-pane {
		min-width: 0;
		min-height: 0;
		overflow-y: auto;
	}
	.rail-heading {
		position: sticky;
		top: 0;
		z-index: 2;
		display: flex;
		flex-shrink: 0;
		align-items: center;
		justify-content: space-between;
		padding: 20px 16px 12px;
		background: var(--card);
	}
	.rail-heading h2 {
		font-size: 14px;
		font-weight: 650;
	}
	.rail-heading span {
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.definition-list small {
		display: -webkit-box;
		line-clamp: 2;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		color: var(--muted-foreground);
	}
	.definition-list > button :global(svg) {
		position: absolute;
		top: 40px;
		right: 4px;
	}
	@media (max-width: 850px) {
		.dictionary-workspace {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto auto;
			height: auto;
		}
		.definition-list {
			position: static;
			border-right: 0;
			border-bottom: 1px solid var(--border);
			padding-right: 0;
			max-height: 360px;
			overflow-y: auto;
		}
		.detail-pane {
			overflow: visible;
		}
	}
</style>
