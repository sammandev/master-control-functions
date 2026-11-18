<script lang="ts">
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import ExampleDetail from '#lib/components/ExampleDetail.svelte';
	import ExampleDialog from '#lib/components/ExampleDialog.svelte';
	import { onDestroy, onMount } from 'svelte';
	import { page as routePage } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { Plus, RefreshCw, ChevronRight, Search } from '@lucide/svelte';
	import { listExamples, deleteExample, type ExampleRow } from '#lib/examples.ts';
	import { listDevices, type DeviceSummary } from '#lib/devices.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import { toast } from 'svelte-sonner';

	// The global Examples catalog (ticket 04): filterable list beside the
	// selected example's complete script, following the Functions and
	// dictionary workspaces. The examples API filters device, function and
	// search server-side, so pagination always represents the full matching
	// result set — a match on a later page stays discoverable.
	// Device scope meanings: no ?device= param is All-Device discovery,
	// ?device=_global is Global-only, ?device=<name> is that Device's rows.
	let search = $derived(routePage.url.searchParams.get('search') ?? '');
	let workspace = $state<HTMLDivElement | null>(null);
	let pendingHeight = $state(600);
	let functionName = $derived(routePage.url.searchParams.get('function_name') ?? '');
	const params = $derived.by(() => {
		const query = routePage.url.searchParams;
		return {
			search: query.get('search') || undefined,
			function_name: query.get('function_name') || undefined,
			page: Math.max(1, Math.floor(Number(query.get('page')) || 1)),
			page_size: [10, 20, 50].includes(Number(query.get('page_size')))
				? Number(query.get('page_size'))
				: 20
		};
	});
	const deviceFilter = $derived(routePage.url.searchParams.get('device'));
	let retry = $state(0);
	let devices = $state<DeviceSummary[]>([]);
	let deviceError = $state('');
	let functionNames = $state<string[]>([]);
	let functionError = $state('');
	let functionLoading = $state(false);
	const functionOptions = $derived([
		{ value: '', label: 'All functions' },
		...Array.from(new Set([...functionNames, ...(functionName ? [functionName] : [])]))
			.sort((first, second) => first.localeCompare(second))
			.map((name) => ({ value: name, label: name }))
	]);
	async function loadFunctionNames() {
		functionLoading = true;
		functionError = '';
		try {
			const result = await listExamples({ page: 1, page_size: 100 });
			const names = result.items.map((row) => row.function_name);
			let loaded = result.items.length;
			for (let next = 2; loaded < result.total; next++) {
				const more = await listExamples({ page: next, page_size: 100 });
				if (!more.items.length) throw new Error('The example list changed. Please retry.');
				names.push(...more.items.map((row) => row.function_name));
				loaded += more.items.length;
			}
			functionNames = names;
		} catch (error) {
			functionError = error instanceof Error ? error.message : String(error);
		} finally {
			functionLoading = false;
		}
	}
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
		void loadFunctionNames();
	});

	const requestKey = $derived(JSON.stringify({ params, deviceFilter, retry }));
	let revision = 0;
	const content = $derived.by(() => {
		const request: { params: typeof params; deviceFilter: string | null } = JSON.parse(requestKey);
		return load(request.params, request.deviceFilter);
	});
	function setQuery(key: string, value: string, reset = true, replaceState = false) {
		pendingHeight = workspace?.getBoundingClientRect().height ?? pendingHeight;
		const query = new SvelteURLSearchParams(routePage.url.searchParams.toString());
		query.delete('source');
		if (value) query.set(key, value);
		else query.delete(key);
		if (reset) {
			query.delete('page');
			query.delete('example');
		}

		return goto(resolve(`examples?${query.toString()}`), {
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

	onDestroy(() => {
		revision++;
		clearTimeout(searchTimer);
	});
	async function load(query: typeof params, device: string | null) {
		const attempt = ++revision;
		const result = await listExamples({
			function_name: query.function_name,
			search: query.search,
			// Absent = every Device; '' = Global-only (the API sends the key
			// whenever it is defined, so the blank string filters blank-scope
			// examples instead of meaning "no filter").
			device: device === null ? undefined : device === '_global' ? '' : device,
			page: query.page,
			page_size: query.page_size
		});
		if (attempt !== revision) return result;
		const last = Math.max(1, Math.ceil(result.total / query.page_size));
		if (query.page > last) {
			await setQuery('page', String(last), false, true);
			return result;
		}
		if (
			routePage.url.searchParams.has('example') &&
			!result.items.some((row) => row.id === Number(routePage.url.searchParams.get('example')))
		)
			await setQuery('example', result.items[0] ? String(result.items[0].id) : '', false, true);
		return result;
	}

	// ---- dialog + confirm state -------------------------------------------
	let dialogOpen = $state(false);
	let editTarget = $state<ExampleRow | null>(null);
	let confirmRow = $state<ExampleRow | null>(null);

	function openCreate() {
		editTarget = null;
		dialogOpen = true;
	}
	function openEdit(row: ExampleRow) {
		editTarget = row;
		dialogOpen = true;
	}
	async function confirmRemove() {
		if (!confirmRow) return;
		try {
			await deleteExample(confirmRow.id);
			toast.success('Example deleted');
			confirmRow = null;
			retry++;
			void loadFunctionNames();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : String(e));
		}
	}
</script>

<div class="filters">
	<div class="examples-search">
		<Search size={16} />
		<Input
			class="h-10 pl-9"
			placeholder="Search examples..."
			aria-label="Search examples"
			bind:value={search}
			oninput={(event) => scheduleSearch(event.currentTarget.value)}
		/>
	</div>
	<Select
		class="h-10 w-60 max-w-full"
		ariaLabel="Function name filter"
		value={functionName}
		options={functionOptions}
		disabled={functionLoading}
		onchange={(value) => {
			void setQuery('function_name', value);
		}}
	/>
	<DevicePicker
		device={deviceFilter === '_global' ? '' : (deviceFilter ?? '_all')}
		summaries={devices}
		label="Filter examples by Device"
		allowAll
		compact
		comfortable
		onValueChange={(value) => {
			void setQuery('device', value === '_all' ? '' : value || '_global');
		}}
	/>
</div>
{#if functionError}<div class="banner-error" role="alert">
		Function filter: {functionError}<Button variant="outline" onclick={loadFunctionNames}
			><RefreshCw size={16} /> Retry functions</Button
		>
	</div>{/if}
<div class="catalog-toolbar">
	<div class="muted">
		<span
			>{#await content}Loading...{:then result}{result.total}
				{result.total === 1 ? 'example' : 'examples'}{:catch}Count unavailable{/await}</span
		>
		<span
			>{deviceFilter === null
				? 'All-Device discovery'
				: deviceFilter === '_global'
					? 'Global examples'
					: `Device: ${deviceFilter}`}</span
		>
	</div>
	<Button onclick={openCreate}><Plus size={16} /> New example</Button>
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
		class="examples-workspace"
		aria-busy="true"
		style:min-height={`${pendingHeight}px`}
	>
		<p class="banner-empty" role="status">Loading examples...</p>
	</div>
{:then result}
	{@const selected =
		result.items.find((row) => row.id === Number(routePage.url.searchParams.get('example'))) ??
		result.items[0]}
	<div class="examples-workspace" bind:this={workspace}>
		<aside class="example-list" aria-label="Examples list">
			<div class="rail-heading">
				<h2>Examples</h2>
				<span>{result.total}</span>
			</div>
			{#each result.items as row (row.id)}
				<button
					class:chosen={selected?.id === row.id}
					aria-pressed={selected?.id === row.id}
					aria-label={`Select example of ${row.function_name} (${row.device_name || 'Global'})`}
					onclick={() => {
						void setQuery('example', String(row.id), false);
					}}
				>
					<strong>{row.function_source}/{row.function_name}</strong>
					<span
						><Badge class="max-w-full break-all whitespace-normal"
							>{row.device_name || 'Global'}</Badge
						>
						{#if !row.function_exists}<Badge variant="warn">orphan</Badge>{/if}</span
					>
					{#if row.meaning}<small>{row.meaning}</small>{/if}
					<ChevronRight size={16} />
				</button>
			{:else}<p class="banner-empty" role="status">
					No examples match. Adjust the search or create one.
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
			{#if selected}{#key selected}<ExampleDetail
						row={selected}
						onEdit={openEdit}
						onDelete={(row) => (confirmRow = row)}
					/>
				{/key}
			{:else}
				<p class="banner-empty" role="status">No example selected.</p>
			{/if}
		</div>
	</div>
{:catch error}<div class="banner-error" role="alert">
		{error instanceof Error ? error.message : String(error)}<Button
			variant="outline"
			onclick={() => retry++}><RefreshCw size={16} /> Retry examples</Button
		>
	</div>
{/await}

<ExampleDialog
	bind:open={dialogOpen}
	editRow={editTarget}
	prefillDevice={deviceFilter && deviceFilter !== '_global' ? deviceFilter : ''}
	onsaved={async (row) => {
		await setQuery('example', String(row.id), false);
		retry++;
		void loadFunctionNames();
	}}
/>

<AlertDialog
	open={confirmRow !== null}
	onOpenChange={(o) => {
		if (!o) confirmRow = null;
	}}
	title="Delete example"
	description={confirmRow
		? `Delete this example of ${confirmRow.function_name}? This cannot be undone.`
		: ''}
	confirmLabel="Delete"
	onConfirm={confirmRemove}
/>

<style>
	.filters,
	.catalog-toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		margin-bottom: 16px;
	}
	.examples-search {
		position: relative;
		flex: 1 1 240px;
		min-width: 0;
	}
	.examples-search > :global(svg) {
		position: absolute;
		top: 12px;
		left: 12px;
		color: var(--muted-foreground);
		pointer-events: none;
	}
	.catalog-toolbar {
		justify-content: space-between;
		padding-bottom: 16px;
	}
	.catalog-toolbar span {
		font-size: 12px;
	}
	.catalog-toolbar span + span {
		margin-left: 8px;
	}
	.examples-workspace {
		display: grid;
		grid-template-columns: minmax(240px, 310px) minmax(0, 1fr);
		grid-template-rows: minmax(0, 1fr);
		height: max(600px, calc(100dvh - 240px));
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--card);
	}
	.example-list {
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
		border-right: 1px solid var(--border);
	}
	.example-list > button {
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
	.example-list > button:hover {
		background: var(--accent);
	}
	.example-list > button.chosen {
		background: var(--device-soft);
		border-left-color: var(--device-accent);
	}
	.example-list > button:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: -2px;
	}
	.example-list strong {
		font: 600 12px/1.6 var(--mono);
		overflow-wrap: anywhere;
	}
	.example-list > button > span,
	.example-list small {
		font-size: 12px;
		overflow-wrap: anywhere;
		max-width: 100%;
	}
	.example-list > button :global(svg) {
		position: absolute;
		top: 40px;
		right: 4px;
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
	.example-list small {
		display: -webkit-box;
		line-clamp: 2;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		color: var(--muted-foreground);
	}
	@media (max-width: 850px) {
		.examples-workspace {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto auto;
			height: auto;
		}
		.example-list {
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
