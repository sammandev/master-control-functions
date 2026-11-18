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
	import { Plus, RefreshCw } from '@lucide/svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';

	// Shared CRUD page for the Parameter definition and Criteria definition
	// dictionaries. kind picks the endpoint; everything else is identical.
	// The create/edit form lives in DefinitionDialog (shared with the
	// device page's definition shortcuts).
	let { kind, title }: { kind: DefinitionKind; title: string } = $props();

	// ---- list state -------------------------------------------------------
	let search = $derived(routePage.url.searchParams.get('search') ?? '');
	const params = $derived.by(() => {
		const query = routePage.url.searchParams;
		return {
			search: query.get('search') || undefined,
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
	const requestKey = $derived(JSON.stringify({ kind, params, deviceFilter, retry }));
	let revision = 0;
	const content = $derived.by(() => {
		const request: { kind: DefinitionKind; params: typeof params; deviceFilter: string | null } =
			JSON.parse(requestKey);
		return load(request.kind, request.params, request.deviceFilter);
	});
	function setQuery(key: string, value: string, reset = true, replaceState = false) {
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
	async function load(endpoint: DefinitionKind, query: typeof params, device: string | null) {
		const attempt = ++revision;
		let result;
		if (device === null) result = await listDefinitions(endpoint, query);
		else {
			const first = await listDefinitions(endpoint, { ...query, page: 1, page_size: 100 });
			const rows = [...first.items];
			for (let next = 2; rows.length < first.total; next++) {
				if (attempt !== revision) return { items: [], total: 0 };
				const more = await listDefinitions(endpoint, { ...query, page: next, page_size: 100 });
				if (!more.items.length) throw new Error('The definition list changed. Please retry.');
				rows.push(...more.items);
			}
			const matches = rows.filter(
				(row) => row.device_name === (device === '_global' ? '' : device)
			);
			const start = (query.page - 1) * query.page_size;
			result = { items: matches.slice(start, start + query.page_size), total: matches.length };
		}
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
		class="min-w-48 flex-1"
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
		onValueChange={(value) => {
			void setQuery('device', value === '_all' ? '' : value || '_global');
		}}
	/>
	<Select
		class="w-36"
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
		class="w-36"
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
{#await content}<p class="banner-empty" role="status">Loading definitions...</p>
{:then result}
	{@const selected =
		result.items.find((row) => row.id === Number(routePage.url.searchParams.get('definition'))) ??
		result.items[0]}
	<div class="dictionary-workspace">
		<aside class="definition-list" aria-label="Definitions list">
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
				</button>
			{:else}<p class="banner-empty" role="status">
					No entries match. Adjust the search or create one.
				</p>{/each}
			<Pagination
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
		await setQuery('definition', String(row.id), false);
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
		grid-template-columns: 300px minmax(0, 1fr);
		border-top: 1px solid var(--border);
	}
	.definition-list {
		min-width: 0;
		position: sticky;
		top: 16px;
		align-self: start;
		max-height: min(640px, calc(100dvh - 32px));
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
		border-right: 1px solid var(--border);
		padding: 16px 16px 16px 0;
	}
	.definition-list button {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		width: 100%;
		border: 0;
		border-bottom: 1px solid var(--border);
		border-left: 3px solid transparent;
		padding: 14px 12px;
		gap: 8px;
		text-align: left;
		color: var(--foreground);
		background: transparent;
		cursor: pointer;
	}
	.definition-list button:hover {
		background: var(--accent);
	}
	.definition-list button.chosen {
		background: var(--device-soft);
		border-left-color: var(--device-accent);
	}
	.definition-list button:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: -2px;
	}
	.definition-list strong {
		font: 600 13px/1.5 var(--mono);
		overflow-wrap: anywhere;
	}
	.definition-list span,
	.definition-list small {
		font-size: 12px;
		overflow-wrap: anywhere;
		max-width: 100%;
	}
	.detail-pane {
		min-width: 0;
	}
	@media (max-width: 850px) {
		.dictionary-workspace {
			grid-template-columns: minmax(0, 1fr);
		}
		.definition-list {
			position: static;
			border-right: 0;
			border-bottom: 1px solid var(--border);
			padding-right: 0;
			max-height: 360px;
			overflow-y: auto;
		}
	}
</style>
