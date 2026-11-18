<script lang="ts">
	import { copyText } from '#lib/clipboard.ts';
	import Select from '#lib/components/ui/select.svelte';
	import { onMount } from 'svelte';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		CircuitBoard,
		Blocks,
		Layers,
		CircleHelp,
		Search,
		ChevronRight,
		RefreshCw,
		Copy,
		Check,
		Plus,
		Pencil,
		Trash2
	} from '@lucide/svelte';
	import {
		GLOBAL_DEVICE_SEGMENT,
		devicePath,
		deviceTypeLabel,
		listDevices,
		listDeviceFunctions,
		listDeviceDefinitions,
		type DeviceSummary,
		type DeviceFunctionRow,
		type DeviceDefinitionRow
	} from '#lib/devices.ts';
	import { deleteExample, listExamples, type ExampleRow } from '#lib/examples.ts';
	import type { DefinitionKind, DefinitionRow } from '#lib/definitions.ts';
	import Breadcrumb from '#lib/components/ui/breadcrumb.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import DeviceFunctionDetail from '#lib/components/DeviceFunctionDetail.svelte';
	import DeviceDefinitionDetail from '#lib/components/DeviceDefinitionDetail.svelte';
	import DefinitionDialog from '#lib/components/DefinitionDialog.svelte';
	import FunctionDialog from '#lib/components/FunctionDialog.svelte';
	import ExampleDialog from '#lib/components/ExampleDialog.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import * as Tabs from '#lib/components/ui/tabs/index.ts';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir, type SortValue } from '#lib/table-sort.ts';
	import {
		Table,
		TableHeader,
		TableBody,
		TableRow,
		TableHead,
		TableCell
	} from '#lib/components/ui/table/index.ts';
	import { toast } from 'svelte-sonner';

	// Device detail (read-only): the functions linked in the device and the
	// definitions scoped to it, with the global definitions its links
	// resolve to shown as a muted second group (exact-scope-then-global,
	// ADR-0009). Examples reuse the examples list API's device filter.
	let device = $derived(page.params.name === GLOBAL_DEVICE_SEGMENT ? '' : (page.params.name ?? ''));
	let isGlobal = $derived(device === '');
	let deviceRevision = 0;
	let summaries = $state<DeviceSummary[]>([]);
	let summaryError = $state('');
	const summary = $derived(summaries.find((item) => item.device_name === device));
	onMount(() => {
		void listDevices()
			.then((result) => {
				summaries = result.items;
			})
			.catch((error) => {
				summaryError = error instanceof Error ? error.message : String(error);
			});
	});

	// The active tab lives in the URL (?tab=) and is fully derived from it,
	// so the URL is the single source of truth: deep links, back/forward,
	// and tab clicks all change the URL and the tab follows. User clicks
	// go through setTab, which records a history entry; missing or unknown
	// values fall back to Functions.
	const DEVICE_TABS = ['functions', 'parameters', 'criteria', 'examples'] as const;
	type DeviceTab = (typeof DEVICE_TABS)[number];

	function parseTab(raw: string | null): DeviceTab {
		return (DEVICE_TABS as readonly string[]).includes(raw ?? '')
			? (raw as DeviceTab)
			: 'functions';
	}

	let tab = $derived(parseTab(page.url.searchParams.get('tab')));

	function setTab(next: string, functionId?: number) {
		const query = new SvelteURLSearchParams(page.url.searchParams.toString());
		if (next !== tab) query.delete('definition');
		query.set('tab', next);
		if (functionId !== undefined) query.set('function', String(functionId));
		void goto(resolve(`/devices/[name]?${query.toString()}`, { name: devicePath(device) }), {
			reset: false
		});
	}

	// ---- functions tab -----------------------------------------------------
	let fnItems = $state<DeviceFunctionRow[]>([]);
	let fnTotal = $state(0);
	let fnPage = $state(1);
	let fnPageSize = $state(20);
	let fnLoaded = $state(false);
	let fnLoading = $state(false);
	let fnError = $state('');
	let fnRevision = 0;
	let newFunctionOpen = $state(false);
	let fnSearch = $state('');
	const selectedFunction = $derived(
		fnItems.find((fn) => fn.id === Number(page.url.searchParams.get('function'))) ??
			(page.url.searchParams.has('function') ? undefined : fnItems[0])
	);

	// ---- parameters / criteria tabs ----------------------------------------
	let paramScoped = $state<DeviceDefinitionRow[]>([]);
	let paramGlobal = $state<DeviceDefinitionRow[]>([]);
	let paramTotal = $state(0);
	let paramGlobalTotal = $state(0);
	let paramPage = $state(1);
	let paramPageSize = $state(20);
	let paramLoaded = $state(false);
	let paramLoading = $state(false);
	let paramError = $state('');

	let critScoped = $state<DeviceDefinitionRow[]>([]);
	let critGlobal = $state<DeviceDefinitionRow[]>([]);
	let critTotal = $state(0);
	let critGlobalTotal = $state(0);
	let critPage = $state(1);
	let critPageSize = $state(20);
	let critLoaded = $state(false);
	let critLoading = $state(false);
	let critError = $state('');

	// ---- examples tab -------------------------------------------------------
	let exampleItems = $state<ExampleRow[]>([]);
	let exampleTotal = $state(0);
	let examplePage = $state(1);
	let examplePageSize = $state(20);
	let exampleLoaded = $state(false);
	let exampleLoading = $state(false);
	let exampleError = $state('');
	let copiedExample = $state<number | null>(null);
	let deviceCopyStatus = $state('');
	let exampleDialogOpen = $state(false);
	let exampleEditRow = $state<ExampleRow | null>(null);
	let exampleConfirmRow = $state<ExampleRow | null>(null);
	let deletingExample = $state(false);

	// ---- sorting (ticket 09) ------------------------------------------------
	// The device list APIs paginate server-side without sort params, so
	// sorting applies to the loaded page through the shared sortRows
	// helper. Each tab has one sort state; the Scoped and Resolved-from-
	// Global tables of a definition tab share it so both groups follow the
	// same header clicks. '' keeps the API's default order until a header
	// is clicked.
	const FN_SORT_KEYS: Record<string, (row: DeviceFunctionRow) => SortValue> = {
		name: (r) => r.name,
		source: (r) => r.source,
		interface: (r) => r.interface,
		category: (r) => r.category || '',
		updated: (r) => r.updated_at
	};

	const DEF_SORT_KEYS: Record<string, (row: DeviceDefinitionRow) => SortValue> = {
		name: (r) => r.name,
		type: (r) => r.type,
		cardinality: (r) => r.cardinality,
		allowed_values: (r) => r.allowed_values || '',
		default_value: (r) => r.default_value || '',
		value_pattern: (r) => r.value_pattern || ''
	};

	const EX_SORT_KEYS: Record<string, (row: ExampleRow) => SortValue> = {
		function: (r) => `${r.function_source}/${r.function_name}`,
		script: (r) => r.script,
		meaning: (r) => r.meaning || ''
	};

	let fnSortKey = $state('name');
	let fnSortDir = $state<SortDir>('asc');
	let paramSortKey = $state('');
	let paramSortDir = $state<SortDir>('asc');
	let critSortKey = $state('');
	let critSortDir = $state<SortDir>('asc');
	let exSortKey = $state('');
	let exSortDir = $state<SortDir>('asc');

	function applyParamSort(col: string) {
		({ key: paramSortKey, dir: paramSortDir } = nextSort(paramSortKey, paramSortDir, col));
	}

	function applyCritSort(col: string) {
		({ key: critSortKey, dir: critSortDir } = nextSort(critSortKey, critSortDir, col));
	}

	function applyExSort(col: string) {
		({ key: exSortKey, dir: exSortDir } = nextSort(exSortKey, exSortDir, col));
	}

	const fnFiltered = $derived(
		fnItems.filter((fn) =>
			`${fn.name} ${fn.source} ${fn.interface} ${fn.description || ''} ${fn.category || ''}`
				.toLowerCase()
				.includes(fnSearch.trim().toLowerCase())
		)
	);
	const fnView = $derived(
		sortRows(fnFiltered, FN_SORT_KEYS[fnSortKey]!, fnSortDir).slice(
			(fnPage - 1) * fnPageSize,
			fnPage * fnPageSize
		)
	);
	// Section searches filter the loaded page only; the APIs page server-side.
	let paramSearch = $state('');
	let critSearch = $state('');
	let exSearch = $state('');
	function matchesDefinition(row: DeviceDefinitionRow, term: string) {
		const needle = term.trim().toLowerCase();
		return (
			!needle ||
			`${row.name} ${row.type} ${row.allowed_values || ''} ${row.default_value || ''} ${row.description || ''}`
				.toLowerCase()
				.includes(needle)
		);
	}
	function viewDefinitions(rows: DeviceDefinitionRow[], term: string, key: string, dir: SortDir) {
		const matches = rows.filter((row) => matchesDefinition(row, term));
		return key === '' ? matches : sortRows(matches, DEF_SORT_KEYS[key]!, dir);
	}
	const paramScopedView = $derived(
		viewDefinitions(paramScoped, paramSearch, paramSortKey, paramSortDir)
	);
	const paramGlobalView = $derived(
		viewDefinitions(paramGlobal, paramSearch, paramSortKey, paramSortDir)
	);
	const critScopedView = $derived(
		viewDefinitions(critScoped, critSearch, critSortKey, critSortDir)
	);
	const critGlobalView = $derived(
		viewDefinitions(critGlobal, critSearch, critSortKey, critSortDir)
	);

	const selectedParameter = $derived(
		[...paramScoped, ...paramGlobal].find(
			(row) => row.id === Number(page.url.searchParams.get('definition'))
		) ?? (page.url.searchParams.has('definition') ? undefined : (paramScoped[0] ?? paramGlobal[0]))
	);
	const selectedCriterion = $derived(
		[...critScoped, ...critGlobal].find(
			(row) => row.id === Number(page.url.searchParams.get('definition'))
		) ?? (page.url.searchParams.has('definition') ? undefined : (critScoped[0] ?? critGlobal[0]))
	);

	const exampleMatches = $derived(
		exampleItems.filter((row) =>
			`${row.function_source} ${row.function_name} ${row.script} ${row.meaning || ''}`
				.toLowerCase()
				.includes(exSearch.trim().toLowerCase())
		)
	);
	const exampleView = $derived(
		exSortKey === ''
			? exampleMatches
			: sortRows(exampleMatches, EX_SORT_KEYS[exSortKey]!, exSortDir)
	);

	// ---- definition shortcuts (ticket 05) -----------------------------------
	// The Parameters and Criteria tabs can open the shared dictionary form
	// with the device prefilled and locked, so a definition is born scoped
	// to the device in context without leaving the page.
	let defDialogOpen = $state(false);
	let defTab = $state<'parameters' | 'criteria'>('parameters');

	const DEF_KIND: Record<'parameters' | 'criteria', DefinitionKind> = {
		parameters: 'parameter-definitions',
		criteria: 'criteria-definitions'
	};

	function openDefDialog(next: 'parameters' | 'criteria') {
		defTab = next;
		defDialogOpen = true;
	}

	function selectDefinition(next: 'parameters' | 'criteria', id?: number, replaceState = false) {
		const query = new SvelteURLSearchParams(page.url.searchParams.toString());
		query.set('tab', next);
		if (id === undefined) query.delete('definition');
		else query.set('definition', String(id));
		return goto(resolve(`/devices/[name]?${query.toString()}`, { name: devicePath(device) }), {
			reset: false,
			replace: replaceState
		});
	}
	function selectDefinitionFromRow(event: MouseEvent, next: 'parameters' | 'criteria', id: number) {
		if (event.target instanceof Element && event.target.closest('button, a')) return;
		void selectDefinition(next, id);
	}
	async function inspectDefinition(next: 'parameters' | 'criteria', id: number) {
		const scope = device;
		const revision = deviceRevision;
		if (next === 'parameters') paramSearch = '';
		else critSearch = '';
		await selectDefinition(next, id);
		const loaded =
			next === 'parameters' ? [...paramScoped, ...paramGlobal] : [...critScoped, ...critGlobal];
		if (loaded.some((row) => row.id === id)) return;
		const size = next === 'parameters' ? paramPageSize : critPageSize;
		try {
			for (let requested = 1; ; requested++) {
				const result = await listDeviceDefinitions(next, scope, requested, size);
				if (
					revision !== deviceRevision ||
					scope !== device ||
					tab !== next ||
					Number(page.url.searchParams.get('definition')) !== id
				)
					return;
				if ([...result.items, ...result.global_items].some((row) => row.id === id)) {
					if (next === 'parameters') {
						paramPage = requested;
						paramLoaded = false;
					} else {
						critPage = requested;
						critLoaded = false;
					}
					return;
				}
				if (!result.items.length || requested * size >= result.total)
					throw new Error('This definition is no longer available for this device.');
			}
		} catch (error) {
			if (revision !== deviceRevision || scope !== device) return;
			const message = error instanceof Error ? error.message : String(error);
			if (next === 'parameters') paramError = message;
			else critError = message;
		}
	}
	async function refreshDefinitions(next: 'parameters' | 'criteria', row: DefinitionRow | null) {
		const scope = device;
		await selectDefinition(next, row?.id, row === null);
		if (next === 'parameters') await loadParameters();
		else await loadCriteria();
		if (scope !== device) return;
		try {
			summaries = (await listDevices()).items;
			summaryError = '';
		} catch (error) {
			summaryError = error instanceof Error ? error.message : String(error);
		}
	}
	async function reloadDefinitions(row: DefinitionRow) {
		await selectDefinition(defTab, row.id);
		// Back to page 1: a fresh definition may sort onto any page.
		if (defTab === 'parameters') {
			paramPage = 1;
			paramLoaded = false;
			await loadParameters();
		} else {
			critPage = 1;
			critLoaded = false;
			await loadCriteria();
		}
		try {
			summaries = (await listDevices()).items;
			summaryError = '';
		} catch (error) {
			summaryError = error instanceof Error ? error.message : String(error);
		}
	}

	// Tabs load lazily on first activation and cache afterwards. Switching
	// device clears the per-tab caches and pagination so the active tab
	// reloads for the new device.
	$effect(() => {
		void device;
		deviceRevision++;
		copiedExample = null;
		deviceCopyStatus = '';
		fnItems = [];
		fnTotal = 0;
		fnSearch = '';
		fnError = '';
		paramScoped = [];
		paramGlobal = [];
		paramTotal = 0;
		paramGlobalTotal = 0;
		paramError = '';
		critScoped = [];
		critGlobal = [];
		critTotal = 0;
		critGlobalTotal = 0;
		critError = '';
		exampleItems = [];
		exampleTotal = 0;
		exampleError = '';
		fnLoaded = false;
		paramLoaded = false;
		critLoaded = false;
		exampleLoaded = false;
		fnPage = 1;
		paramPage = 1;
		critPage = 1;
		examplePage = 1;
	});
	// Pagination interacts with the lazy cache: changing a tab's page or page
	// size invalidates that tab's cache so the loader below refetches with
	// the new page. (Device switches above reset both together.)
	$effect(() => {
		void paramPage;
		void paramPageSize;
		paramLoaded = false;
	});
	$effect(() => {
		void critPage;
		void critPageSize;
		critLoaded = false;
	});
	$effect(() => {
		void examplePage;
		void examplePageSize;
		exampleLoaded = false;
	});
	$effect(() => {
		if (tab === 'functions' && !fnLoaded) loadFunctions();
	});
	$effect(() => {
		if (tab === 'parameters' && !paramLoaded) loadParameters();
	});
	$effect(() => {
		if (tab === 'criteria' && !critLoaded) loadCriteria();
	});
	$effect(() => {
		if (tab === 'examples' && !exampleLoaded) loadExamples();
	});

	async function loadFunctions() {
		const scope = device;
		const revision = deviceRevision;
		const attempt = ++fnRevision;
		const current = () => revision === deviceRevision && attempt === fnRevision;
		fnLoading = true;
		fnError = '';
		try {
			const res = await listDeviceFunctions(scope, 1, 100);
			const items = [...res.items];
			for (let next = 2; items.length < res.total; next++) {
				if (!current()) return;
				const more = await listDeviceFunctions(scope, next, 100);
				if (!more.items.length) throw new Error('The function list changed. Please retry.');
				items.push(...more.items);
			}
			if (!current()) return;
			fnItems = items;
			fnTotal = res.total;
			fnLoaded = true;
		} catch (e) {
			if (!current()) return;
			fnItems = [];
			fnTotal = 0;
			fnError = e instanceof Error ? e.message : String(e);
		} finally {
			if (current()) fnLoading = false;
		}
	}

	async function refreshFunctions(deletedId?: number) {
		const scope = device;
		if (deletedId !== undefined) {
			fnItems = fnItems.filter((row) => row.id !== deletedId);
			if (Number(page.url.searchParams.get('function')) === deletedId) {
				const query = new SvelteURLSearchParams(page.url.searchParams.toString());
				query.delete('function');
				await goto(resolve(`/devices/[name]?${query.toString()}`, { name: devicePath(scope) }), {
					reset: false,
					replace: true
				});
			}
		}
		await loadFunctions();
		if (scope !== device) return;
		fnPage = Math.min(fnPage, Math.max(1, Math.ceil(fnFiltered.length / fnPageSize)));
		try {
			summaries = (await listDevices()).items;
		} catch (error) {
			summaryError = error instanceof Error ? error.message : String(error);
		}
	}

	async function loadParameters() {
		const scope = device;
		const revision = deviceRevision;
		const requestedPage = paramPage;
		const requestedSize = paramPageSize;
		const current = () =>
			revision === deviceRevision && requestedPage === paramPage && requestedSize === paramPageSize;
		paramLoading = true;
		paramError = '';
		try {
			const res = await listDeviceDefinitions('parameters', scope, requestedPage, requestedSize);
			if (!current()) return;
			const last = Math.max(1, Math.ceil(res.total / requestedSize));
			if (requestedPage > last) {
				paramPage = last;
				return;
			}
			paramScoped = res.items;
			paramGlobal = res.global_items;
			paramTotal = res.total;
			paramGlobalTotal = res.global_total;
			paramLoaded = true;
		} catch (e) {
			if (!current()) return;
			paramScoped = [];
			paramGlobal = [];
			paramTotal = 0;
			paramGlobalTotal = 0;
			paramError = e instanceof Error ? e.message : String(e);
		} finally {
			if (current()) paramLoading = false;
		}
	}

	async function loadCriteria() {
		const scope = device;
		const revision = deviceRevision;
		const requestedPage = critPage;
		const requestedSize = critPageSize;
		const current = () =>
			revision === deviceRevision && requestedPage === critPage && requestedSize === critPageSize;
		critLoading = true;
		critError = '';
		try {
			const res = await listDeviceDefinitions('criteria', scope, requestedPage, requestedSize);
			if (!current()) return;
			const last = Math.max(1, Math.ceil(res.total / requestedSize));
			if (requestedPage > last) {
				critPage = last;
				return;
			}
			critScoped = res.items;
			critGlobal = res.global_items;
			critTotal = res.total;
			critGlobalTotal = res.global_total;
			critLoaded = true;
		} catch (e) {
			if (!current()) return;
			critScoped = [];
			critGlobal = [];
			critTotal = 0;
			critGlobalTotal = 0;
			critError = e instanceof Error ? e.message : String(e);
		} finally {
			if (current()) critLoading = false;
		}
	}

	async function loadExamples() {
		const scope = device;
		const revision = deviceRevision;
		const requestedPage = examplePage;
		const requestedSize = examplePageSize;
		const current = () =>
			revision === deviceRevision &&
			requestedPage === examplePage &&
			requestedSize === examplePageSize;
		exampleLoading = true;
		exampleError = '';
		try {
			const res = await listExamples({
				device: scope,
				page: requestedPage,
				page_size: requestedSize
			});
			if (!current()) return;
			exampleItems = res.items;
			exampleTotal = res.total;
			exampleLoaded = true;
		} catch (e) {
			if (!current()) return;
			exampleItems = [];
			exampleTotal = 0;
			exampleError = e instanceof Error ? e.message : String(e);
		} finally {
			if (current()) exampleLoading = false;
		}
	}

	async function copyDeviceExample(example: ExampleRow) {
		const success = await copyText(example.script);
		copiedExample = success ? example.id : null;
		deviceCopyStatus = success ? 'Script copied.' : 'Copy failed. The script text is selectable.';
	}

	function openCreateExample() {
		exampleEditRow = null;
		exampleDialogOpen = true;
	}

	function openEditExample(example: ExampleRow) {
		exampleEditRow = example;
		exampleDialogOpen = true;
	}

	async function refreshExamples() {
		await loadExamples();
		try {
			summaries = (await listDevices()).items;
			summaryError = '';
		} catch (error) {
			summaryError = error instanceof Error ? error.message : String(error);
		}
	}

	async function deleteDeviceExample() {
		if (!exampleConfirmRow || deletingExample) return;
		deletingExample = true;
		try {
			await deleteExample(exampleConfirmRow.id);
			exampleConfirmRow = null;
			toast.success('Example deleted');
			await refreshExamples();
		} catch (error) {
			exampleError = error instanceof Error ? error.message : String(error);
		} finally {
			deletingExample = false;
		}
	}

	function label(deviceName: string): string {
		return deviceName === '' ? '(Global)' : deviceName;
	}
</script>

<svelte:head
	><title>{device || 'Global'} | Devices | MasterControl Functions Database</title></svelte:head
>

{#snippet defToolbar(
	next: 'parameters' | 'criteria',
	total: number,
	globalTotal: number,
	term: string,
	setTerm: (value: string) => void
)}
	{@const noun = next === 'parameters' ? 'parameter' : 'criteria'}
	<div class="section-toolbar">
		<div class="section-title">
			<h2>{next === 'parameters' ? 'Parameter definitions' : 'Criteria definitions'}</h2>
			<p>
				{total}
				{noun} definition{total === 1 ? '' : 's'} scoped to {label(device)}, plus {globalTotal} resolved
				from Global.
			</p>
		</div>
		<div class="section-actions">
			<div class="section-search">
				<Search size={16} /><Input
					value={term}
					oninput={(event) => setTerm(event.currentTarget.value)}
					aria-label={`Search ${noun} definitions`}
					placeholder="Search name, type or value"
				/>
			</div>
			<Button onclick={() => openDefDialog(next)}
				><Plus data-icon="inline-start" /> Add {noun} definition</Button
			>
		</div>
	</div>
{/snippet}

<div class="device-workspace">
	<Breadcrumb items={[{ label: 'Devices', href: resolve('devices') }, { label: label(device) }]} />

	<header class="device-heading">
		<div class="device-identity">
			<div
				class="device-symbol"
				role="img"
				aria-label={`${summary ? deviceTypeLabel(summary) : 'Unassigned'} type`}
			>
				{#if summary?.source === 'BUILT-IN'}<CircuitBoard size={28} />
				{:else if summary?.source === 'DLL'}<Blocks size={28} />
				{:else if (summary?.sources?.length ?? 0) > 1}<Layers size={28} />
				{:else if summary?.sources?.[0] === 'BUILT-IN'}<CircuitBoard size={28} />
				{:else if summary?.sources?.[0] === 'DLL'}<Blocks size={28} />
				{:else}<CircleHelp size={28} />{/if}
			</div>
			<div>
				<span class="device-label"
					>{isGlobal ? 'Global definitions' : summary ? deviceTypeLabel(summary) : 'Device'}</span
				>
				<DevicePicker
					heading
					{device}
					blankLabel="Global definitions"
					label="Device"
					summaries={summaries.length ? summaries : undefined}
					onValueChange={(nextDevice) => {
						void goto(resolve('/devices/[name]', { name: devicePath(nextDevice) }), {
							reset: false
						});
					}}
				/>
			</div>
		</div>
		{#if tab === 'functions'}
			<a class="all-device-link" href={resolve('functions')}>
				All-Device discovery <ChevronRight size={16} />
			</a>
		{/if}
	</header>
	{#if summaryError}<p class="banner-error" role="alert">
			Device switcher unavailable: {summaryError}
		</p>{/if}

	<Tabs.Root value={tab} onValueChange={setTab}>
		<Tabs.List class="device-sections" aria-label="Device sections">
			{#each DEVICE_TABS as section (section)}
				<Tabs.Trigger class="device-section" value={section}>
					<span>{section[0].toUpperCase() + section.slice(1)}</span>
					{#if summary}<span class="section-count" aria-hidden="true">{summary[section]}</span>{/if}
				</Tabs.Trigger>
			{/each}
		</Tabs.List>

		<Tabs.Content value="functions">
			{#if fnError}
				<div class="banner-error" role="alert">
					{fnError}<Button size="sm" variant="outline" onclick={loadFunctions}
						><RefreshCw size={16} /> Retry functions</Button
					>
				</div>
			{/if}
			<div class="function-browser">
				<aside class="function-list" aria-label="Device functions">
					<div class="rail-controls">
						<div class="rail-heading">
							<div class="rail-title">
								<h2>Functions</h2>
								<span>{fnTotal}</span>
							</div>
							<Button
								variant="outline"
								size="sm"
								aria-label="New function"
								title="New function"
								onclick={() => (newFunctionOpen = true)}
							>
								<Plus data-icon="inline-start" /> Add
							</Button>
						</div>
						<div class="function-search">
							<Search size={16} /><Input
								bind:value={fnSearch}
								aria-label="Search device functions"
								placeholder="Search functions"
								oninput={() => {
									fnPage = 1;
								}}
							/>
						</div>
						<div class="function-sort">
							<Select
								ariaLabel="Sort functions"
								bind:value={fnSortKey}
								options={[
									{ value: 'name', label: 'Name' },
									{ value: 'source', label: 'Source' },
									{ value: 'interface', label: 'Device / Interface' },
									{ value: 'category', label: 'Category' },
									{ value: 'updated', label: 'Updated' }
								]}
							/>
							<Select
								ariaLabel="Function sort order"
								bind:value={fnSortDir}
								options={[
									{ value: 'asc', label: 'Ascending' },
									{ value: 'desc', label: 'Descending' }
								]}
							/>
						</div>
					</div>
					{#each fnView as fn (fn.id)}<button
							class="function-choice focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none"
							class:chosen={selectedFunction?.id === fn.id}
							aria-label={`Select ${fn.name}`}
							aria-pressed={selectedFunction?.id === fn.id}
							onclick={() => setTab('functions', fn.id)}
							><strong>{fn.name}</strong><span>{fn.source} / {fn.interface}</span><small
								title={fn.description || fn.category || 'Uncategorized'}
								>{fn.description || fn.category || 'Uncategorized'}</small
							><ChevronRight size={16} /></button
						>{:else}<p class="banner-empty" role="status">
							{fnLoading
								? 'Loading functions...'
								: fnSearch
									? 'No matching functions.'
									: 'No linked functions.'}
						</p>{/each}
					{#if fnFiltered.length > 10}<Pagination
							variant="rail"
							bind:page={fnPage}
							total={fnFiltered.length}
							bind:pageSize={fnPageSize}
						/>{/if}
				</aside>
				<div class="detail-pane">
					{#if selectedFunction}{#key `${device}/${selectedFunction.id}`}<DeviceFunctionDetail
								fn={selectedFunction}
								{device}
								onChanged={refreshFunctions}
								onInspectDefinition={inspectDefinition}
							/>{/key}{:else if !fnLoading && fnItems.length}<p class="banner-empty" role="status">
							This function is not linked to this device. Select a function from the list.
						</p>{/if}
				</div>
			</div>
		</Tabs.Content>

		<Tabs.Content value="parameters">
			<div class="definition-browser">
				<div class="definition-tables">
					{#if paramError}
						<div class="px-4">
							<div class="banner-error" role="alert">
								{paramError}<Button size="sm" variant="outline" onclick={loadParameters}
									><RefreshCw size={16} /> Retry parameters</Button
								>
							</div>
						</div>
					{/if}
					{@render defToolbar(
						'parameters',
						paramTotal,
						paramGlobalTotal,
						paramSearch,
						(value) => (paramSearch = value)
					)}
					<h3 class="group-heading">
						Scoped <span class="group-count">{paramScopedView.length}</span>
					</h3>
					<Table class="border-0">
						<TableHeader>
							<TableRow>
								<SortHead
									label="Name"
									col="name"
									active={paramSortKey}
									dir={paramSortDir}
									onclick={() => applyParamSort('name')}
								/>
								<SortHead
									label="Type"
									col="type"
									active={paramSortKey}
									dir={paramSortDir}
									onclick={() => applyParamSort('type')}
								/>
								<SortHead
									label="Cardinality"
									col="cardinality"
									active={paramSortKey}
									dir={paramSortDir}
									onclick={() => applyParamSort('cardinality')}
								/>
								<SortHead
									label="Allowed values"
									col="allowed_values"
									active={paramSortKey}
									dir={paramSortDir}
									onclick={() => applyParamSort('allowed_values')}
								/>
								<SortHead
									label="Default"
									col="default_value"
									active={paramSortKey}
									dir={paramSortDir}
									onclick={() => applyParamSort('default_value')}
								/>
								<SortHead
									label="Pattern"
									col="value_pattern"
									active={paramSortKey}
									dir={paramSortDir}
									onclick={() => applyParamSort('value_pattern')}
								/>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each paramScopedView as d (d.id)}
								<TableRow
									class="cursor-pointer"
									data-selected={selectedParameter?.id === d.id}
									onclick={(event) => selectDefinitionFromRow(event, 'parameters', d.id)}
								>
									<TableCell class="font-medium"
										><button
											class="definition-choice focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none"
											aria-label={`Inspect ${d.name} (${d.device_name || 'Global'})`}
											onclick={() => {
												void selectDefinition('parameters', d.id);
											}}>{d.name}</button
										></TableCell
									>
									<TableCell><Badge>{d.type}</Badge></TableCell>
									<TableCell>{d.cardinality}</TableCell>
									<TableCell class="font-mono text-xs">{d.allowed_values}</TableCell>
									<TableCell class="font-mono text-xs">{d.default_value}</TableCell>
									<TableCell class="font-mono text-xs">{d.value_pattern}</TableCell>
								</TableRow>
							{:else}
								<TableRow>
									<TableCell colspan={6} class="banner-empty text-center" aria-live="polite">
										{paramLoading
											? 'Loading…'
											: paramSearch
												? 'No scoped definitions match your search.'
												: 'No definitions scoped to this device.'}
									</TableCell>
								</TableRow>
							{/each}
						</TableBody>
					</Table>
					{#if paramGlobalView.length > 0}
						<h3 class="group-heading global">
							Resolved from Global <span class="group-count">{paramGlobalView.length}</span>
						</h3>
						<Table class="text-muted-foreground border-0">
							<TableHeader>
								<TableRow>
									<SortHead
										label="Name"
										col="name"
										active={paramSortKey}
										dir={paramSortDir}
										onclick={() => applyParamSort('name')}
									/>
									<SortHead
										label="Type"
										col="type"
										active={paramSortKey}
										dir={paramSortDir}
										onclick={() => applyParamSort('type')}
									/>
									<SortHead
										label="Cardinality"
										col="cardinality"
										active={paramSortKey}
										dir={paramSortDir}
										onclick={() => applyParamSort('cardinality')}
									/>
									<SortHead
										label="Allowed values"
										col="allowed_values"
										active={paramSortKey}
										dir={paramSortDir}
										onclick={() => applyParamSort('allowed_values')}
									/>
									<SortHead
										label="Default"
										col="default_value"
										active={paramSortKey}
										dir={paramSortDir}
										onclick={() => applyParamSort('default_value')}
									/>
									<SortHead
										label="Pattern"
										col="value_pattern"
										active={paramSortKey}
										dir={paramSortDir}
										onclick={() => applyParamSort('value_pattern')}
									/>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#each paramGlobalView as d (d.id)}
									<TableRow
										class="cursor-pointer"
										data-selected={selectedParameter?.id === d.id}
										onclick={(event) => selectDefinitionFromRow(event, 'parameters', d.id)}
									>
										<TableCell class="font-medium"
											><button
												class="definition-choice focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none"
												aria-label={`Inspect ${d.name} (${d.device_name || 'Global'})`}
												onclick={() => {
													void selectDefinition('parameters', d.id);
												}}>{d.name}</button
											></TableCell
										>
										<TableCell><Badge>{d.type}</Badge></TableCell>
										<TableCell>{d.cardinality}</TableCell>
										<TableCell class="font-mono text-xs">{d.allowed_values}</TableCell>
										<TableCell class="font-mono text-xs">{d.default_value}</TableCell>
										<TableCell class="font-mono text-xs">{d.value_pattern}</TableCell>
									</TableRow>
								{/each}
							</TableBody>
						</Table>
					{/if}
					<Pagination
						class="device-pagination"
						bind:page={paramPage}
						total={paramTotal}
						bind:pageSize={paramPageSize}
					/>
				</div>
				<div class="definition-pane">
					{#if selectedParameter && !paramLoading && !paramError}{#key `${device}/parameters/${selectedParameter.id}`}<DeviceDefinitionDetail
								definition={selectedParameter}
								kind="parameter-definitions"
								{device}
								onChanged={(row) => refreshDefinitions('parameters', row)}
							/>{/key}
					{:else if !paramLoading}<p class="banner-empty" role="status">
							Select a parameter definition.
						</p>{/if}
				</div>
			</div>
		</Tabs.Content>

		<Tabs.Content value="criteria">
			<div class="definition-browser">
				<div class="definition-tables">
					{#if critError}
						<div class="px-4">
							<div class="banner-error" role="alert">
								{critError}<Button size="sm" variant="outline" onclick={loadCriteria}
									><RefreshCw size={16} /> Retry criteria</Button
								>
							</div>
						</div>
					{/if}
					{@render defToolbar(
						'criteria',
						critTotal,
						critGlobalTotal,
						critSearch,
						(value) => (critSearch = value)
					)}
					<h3 class="group-heading">
						Scoped <span class="group-count">{critScopedView.length}</span>
					</h3>
					<Table class="border-0">
						<TableHeader>
							<TableRow>
								<SortHead
									label="Name"
									col="name"
									active={critSortKey}
									dir={critSortDir}
									onclick={() => applyCritSort('name')}
								/>
								<SortHead
									label="Type"
									col="type"
									active={critSortKey}
									dir={critSortDir}
									onclick={() => applyCritSort('type')}
								/>
								<SortHead
									label="Cardinality"
									col="cardinality"
									active={critSortKey}
									dir={critSortDir}
									onclick={() => applyCritSort('cardinality')}
								/>
								<SortHead
									label="Allowed values"
									col="allowed_values"
									active={critSortKey}
									dir={critSortDir}
									onclick={() => applyCritSort('allowed_values')}
								/>
								<SortHead
									label="Default"
									col="default_value"
									active={critSortKey}
									dir={critSortDir}
									onclick={() => applyCritSort('default_value')}
								/>
								<SortHead
									label="Pattern"
									col="value_pattern"
									active={critSortKey}
									dir={critSortDir}
									onclick={() => applyCritSort('value_pattern')}
								/>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each critScopedView as d (d.id)}
								<TableRow
									class="cursor-pointer"
									data-selected={selectedCriterion?.id === d.id}
									onclick={(event) => selectDefinitionFromRow(event, 'criteria', d.id)}
								>
									<TableCell class="font-medium"
										><button
											class="definition-choice focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none"
											aria-label={`Inspect ${d.name} (${d.device_name || 'Global'})`}
											onclick={() => {
												void selectDefinition('criteria', d.id);
											}}>{d.name}</button
										></TableCell
									>
									<TableCell><Badge>{d.type}</Badge></TableCell>
									<TableCell>{d.cardinality}</TableCell>
									<TableCell class="font-mono text-xs">{d.allowed_values}</TableCell>
									<TableCell class="font-mono text-xs">{d.default_value}</TableCell>
									<TableCell class="font-mono text-xs">{d.value_pattern}</TableCell>
								</TableRow>
							{:else}
								<TableRow>
									<TableCell colspan={6} class="banner-empty text-center" aria-live="polite">
										{critLoading
											? 'Loading…'
											: critSearch
												? 'No scoped definitions match your search.'
												: 'No definitions scoped to this device.'}
									</TableCell>
								</TableRow>
							{/each}
						</TableBody>
					</Table>
					{#if critGlobalView.length > 0}
						<h3 class="group-heading global">
							Resolved from Global <span class="group-count">{critGlobalView.length}</span>
						</h3>
						<Table class="text-muted-foreground border-0">
							<TableHeader>
								<TableRow>
									<SortHead
										label="Name"
										col="name"
										active={critSortKey}
										dir={critSortDir}
										onclick={() => applyCritSort('name')}
									/>
									<SortHead
										label="Type"
										col="type"
										active={critSortKey}
										dir={critSortDir}
										onclick={() => applyCritSort('type')}
									/>
									<SortHead
										label="Cardinality"
										col="cardinality"
										active={critSortKey}
										dir={critSortDir}
										onclick={() => applyCritSort('cardinality')}
									/>
									<SortHead
										label="Allowed values"
										col="allowed_values"
										active={critSortKey}
										dir={critSortDir}
										onclick={() => applyCritSort('allowed_values')}
									/>
									<SortHead
										label="Default"
										col="default_value"
										active={critSortKey}
										dir={critSortDir}
										onclick={() => applyCritSort('default_value')}
									/>
									<SortHead
										label="Pattern"
										col="value_pattern"
										active={critSortKey}
										dir={critSortDir}
										onclick={() => applyCritSort('value_pattern')}
									/>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#each critGlobalView as d (d.id)}
									<TableRow
										class="cursor-pointer"
										data-selected={selectedCriterion?.id === d.id}
										onclick={(event) => selectDefinitionFromRow(event, 'criteria', d.id)}
									>
										<TableCell class="font-medium"
											><button
												class="definition-choice focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none"
												aria-label={`Inspect ${d.name} (${d.device_name || 'Global'})`}
												onclick={() => {
													void selectDefinition('criteria', d.id);
												}}>{d.name}</button
											></TableCell
										>
										<TableCell><Badge>{d.type}</Badge></TableCell>
										<TableCell>{d.cardinality}</TableCell>
										<TableCell class="font-mono text-xs">{d.allowed_values}</TableCell>
										<TableCell class="font-mono text-xs">{d.default_value}</TableCell>
										<TableCell class="font-mono text-xs">{d.value_pattern}</TableCell>
									</TableRow>
								{/each}
							</TableBody>
						</Table>
					{/if}
					<Pagination
						class="device-pagination"
						bind:page={critPage}
						total={critTotal}
						bind:pageSize={critPageSize}
					/>
				</div>
				<div class="definition-pane">
					{#if selectedCriterion && !critLoading && !critError}{#key `${device}/criteria/${selectedCriterion.id}`}<DeviceDefinitionDetail
								definition={selectedCriterion}
								kind="criteria-definitions"
								{device}
								onChanged={(row) => refreshDefinitions('criteria', row)}
							/>{/key}
					{:else if !critLoading}<p class="banner-empty" role="status">
							Select a criteria definition.
						</p>{/if}
				</div>
			</div>
		</Tabs.Content>

		<Tabs.Content value="examples">
			<div class="examples-panel">
				{#if exampleError}
					<div class="px-4">
						<div class="banner-error" role="alert">
							{exampleError}<Button
								variant="outline"
								size="sm"
								disabled={exampleLoading}
								onclick={() => void loadExamples()}><RefreshCw size={16} /> Retry examples</Button
							>
						</div>
					</div>
				{/if}
				<div class="section-toolbar">
					<div class="section-title">
						<h2>Device examples</h2>
						<p>{exampleTotal} example{exampleTotal === 1 ? '' : 's'} carrying this device.</p>
					</div>
					<div class="section-actions">
						<div class="section-search">
							<Search size={16} /><Input
								bind:value={exSearch}
								aria-label="Search device examples"
								placeholder="Search function, script or meaning"
							/>
						</div>
						<Button onclick={openCreateExample}
							><Plus data-icon="inline-start" /> New example</Button
						>
					</div>
				</div>
				<div class="examples-table-scroll">
					<Table class="min-w-[560px] table-fixed border-0">
						<colgroup>
							<col style="width: 20%" />
							<col />
							<col />
							<col style="width: 96px" />
						</colgroup>
						<TableHeader>
							<TableRow>
								<SortHead
									label="Function"
									col="function"
									active={exSortKey}
									dir={exSortDir}
									onclick={() => applyExSort('function')}
								/>
								<SortHead
									label="Script"
									col="script"
									active={exSortKey}
									dir={exSortDir}
									onclick={() => applyExSort('script')}
								/>
								<SortHead
									label="Meaning"
									col="meaning"
									active={exSortKey}
									dir={exSortDir}
									onclick={() => applyExSort('meaning')}
								/>
								<TableHead class="text-right">Actions</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each exampleView as ex (ex.id)}
								<TableRow>
									<TableCell class="whitespace-nowrap">
										<Badge>{ex.function_source}</Badge>
										{ex.function_name}
									</TableCell>
									<TableCell>
										<pre class="device-script"><code>{ex.script}</code></pre>
										<Button
											variant="ghost"
											size="icon"
											aria-label={`Copy device example ${ex.id}`}
											title="Copy script"
											onclick={() => copyDeviceExample(ex)}
										>
											{#if copiedExample === ex.id}
												<Check size={16} />
											{:else}
												<Copy size={16} />
											{/if}
										</Button>
									</TableCell>
									<TableCell class="text-sm">{ex.meaning}</TableCell>
									<TableCell class="text-right whitespace-nowrap">
										<Button
											size="icon"
											variant="ghost"
											aria-label={`Edit example ${ex.id}`}
											title="Edit example"
											onclick={() => openEditExample(ex)}><Pencil size={16} /></Button
										>
										<Button
											size="icon"
											variant="ghost"
											aria-label={`Delete example ${ex.id}`}
											title="Delete example"
											disabled={deletingExample}
											onclick={() => (exampleConfirmRow = ex)}><Trash2 size={16} /></Button
										>
									</TableCell>
								</TableRow>
							{:else}
								<TableRow>
									<TableCell colspan={4} class="banner-empty text-center" aria-live="polite">
										{exampleLoading
											? 'Loading…'
											: exSearch
												? 'No examples on this page match your search.'
												: 'No examples carry this device.'}
									</TableCell>
								</TableRow>
							{/each}
						</TableBody>
					</Table>
				</div>
				<Pagination
					class="device-pagination"
					bind:page={examplePage}
					total={exampleTotal}
					bind:pageSize={examplePageSize}
				/>
			</div>
			<p role="status" class="text-muted-foreground text-sm">{deviceCopyStatus}</p>
		</Tabs.Content>
	</Tabs.Root>
</div>

<ExampleDialog
	bind:open={exampleDialogOpen}
	editRow={exampleEditRow}
	prefillDevice={device}
	lockDevice={exampleEditRow?.device_name === ''}
	onsaved={refreshExamples}
/>

<AlertDialog
	open={exampleConfirmRow !== null}
	onOpenChange={(open) => {
		if (!open) exampleConfirmRow = null;
	}}
	title="Delete example"
	description={exampleConfirmRow
		? `Delete this example of ${exampleConfirmRow.function_name}? This cannot be undone.`
		: ''}
	confirmLabel="Delete"
	onConfirm={deleteDeviceExample}
/>

{#if newFunctionOpen}<FunctionDialog
		mode="create"
		defaults={{ interface: device }}
		onClose={() => (newFunctionOpen = false)}
		onSaved={async (row) => {
			if (row) await goto(resolve('/functions/[id]', { id: String(row.id) }));
		}}
	/>{/if}

<DefinitionDialog
	bind:open={defDialogOpen}
	kind={DEF_KIND[defTab]}
	title={defTab === 'parameters' ? 'Parameter Definitions' : 'Criteria Definitions'}
	prefillDevice={device}
	lockDevice
	onsaved={reloadDefinitions}
/>

<style>
	.definition-browser {
		display: grid;
		grid-template-columns: minmax(0, 3fr) minmax(280px, 2fr);
		grid-template-rows: minmax(0, 1fr);
		height: max(600px, calc(100dvh - 240px));
		margin-top: 20px;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--card);
		overflow: clip;
	}
	.section-toolbar {
		flex-shrink: 0;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 12px 20px;
		padding: 16px;
		border-bottom: 1px solid var(--border);
		background: var(--card);
	}
	.examples-panel {
		display: flex;
		flex-direction: column;
		height: max(600px, calc(100dvh - 240px));
		margin-top: 20px;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--card);
		overflow: clip;
	}
	.section-title {
		min-width: 0;
		flex: 1 1 240px;
	}
	.examples-table-scroll {
		flex: 1;
		min-height: 0;
		max-width: 100%;
		overflow: auto;
		overscroll-behavior: contain;
	}
	.device-workspace :global(.device-pagination) {
		position: sticky;
		bottom: 0;
		z-index: 2;
		flex-shrink: 0;
		margin-top: auto;
		border-top: 1px solid var(--border);
		padding: 10px 16px;
		background: var(--card);
	}
	.definition-tables > :global(table),
	.group-heading {
		flex-shrink: 0;
	}
	.definition-tables :global(tr[data-selected='true']) {
		background: var(--device-soft);
		box-shadow: inset 3px 0 var(--device-accent);
	}
	.section-title h2 {
		font-size: 14px;
		font-weight: 650;
	}
	.section-title p {
		margin-top: 2px;
		color: var(--muted-foreground);
		font-size: 12px;
	}
	.section-actions {
		display: flex;
		flex-wrap: nowrap;
		align-items: center;
		gap: 8px;
		flex: 0 1 520px;
		max-width: 100%;
		margin-left: auto;
	}
	.section-actions :global(button) {
		flex-shrink: 0;
	}
	.section-search {
		position: relative;
		flex: 1 1 280px;
		min-width: 160px;
	}
	.section-search > :global(svg) {
		position: absolute;
		top: 8px;
		left: 10px;
		color: var(--muted-foreground);
		pointer-events: none;
	}
	.section-search :global(input) {
		padding-left: 32px;
	}
	.group-heading {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 14px 16px 6px;
		color: var(--muted-foreground);
		font-size: 11px;
		font-weight: 650;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	.group-heading.global {
		margin-top: 8px;
		border-top: 1px solid var(--border);
	}
	.group-count {
		min-width: 20px;
		padding: 1px 6px;
		border: 1px solid var(--border);
		border-radius: 999px;
		font-size: 10px;
		text-align: center;
		letter-spacing: 0;
	}
	.definition-tables,
	.definition-pane {
		min-width: 0;
	}
	.definition-tables {
		display: flex;
		flex-direction: column;
		min-height: 0;
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
		overflow-x: auto;
	}
	.definition-pane {
		min-height: 0;
		overflow-y: auto;
		border-left: 1px solid var(--border);
	}
	.definition-choice {
		color: var(--primary);
		font-family: var(--mono);
		text-align: left;
		text-decoration: underline;
		cursor: pointer;
	}
	.definition-choice:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 2px;
	}
	@media (max-width: 1100px) {
		.definition-browser {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto auto;
			height: auto;
		}
		.definition-tables {
			position: static;
			max-height: 360px;
		}
		.definition-pane {
			overflow: visible;
			border-left: 0;
			border-top: 1px solid var(--border);
		}
	}
	.all-device-link {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		color: var(--primary);
		font-size: 13px;
		text-decoration: none;
	}
	.all-device-link:hover {
		color: var(--foreground);
		text-decoration: underline;
		text-underline-offset: 3px;
	}
	.all-device-link :global(svg) {
		flex-shrink: 0;
	}
	.device-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 24px;
		margin: 24px 0 28px;
	}
	.device-workspace {
		min-width: 0;
	}
	.device-identity {
		display: flex;
		align-items: center;
		gap: 16px;
		min-width: 0;
	}
	.device-identity > :last-child {
		min-width: 0;
	}
	.device-symbol {
		display: grid;
		place-items: center;
		width: 60px;
		height: 60px;
		background: var(--device-soft);
		color: var(--device-accent);
		border-radius: 6px;
		flex-shrink: 0;
	}
	.device-label {
		display: block;
		font-size: 11px;
		font-weight: 600;
		color: var(--muted-foreground);
		margin-bottom: 6px;
	}
	:global(.device-workspace .device-sections) {
		width: 100%;
		justify-content: flex-start;
		gap: 24px;
		overflow-x: auto;
		padding: 0;
		border-bottom: 1px solid var(--border);
		border-radius: 0;
		background: transparent;
	}
	:global(.device-workspace .device-section) {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 12px 0;
		min-height: 44px;
		border-radius: 0;
		border-bottom: 2px solid transparent;
		background: transparent;
		box-shadow: none;
		font-size: 14px;
		color: var(--muted-foreground);
		flex-shrink: 0;
	}
	.section-count {
		min-width: 20px;
		padding: 2px 6px;
		border: 1px solid var(--border);
		border-radius: 999px;
		color: var(--muted-foreground);
		font-size: 11px;
		font-variant-numeric: tabular-nums;
		line-height: 1.2;
		text-align: center;
	}
	:global(.device-workspace .device-section[data-state='active']) {
		color: var(--device-accent);
		border-bottom-color: var(--device-accent);
	}
	.function-browser {
		display: grid;
		grid-template-columns: minmax(280px, 30%) minmax(0, 1fr);
		grid-template-rows: minmax(0, 1fr);
		height: max(600px, calc(100dvh - 240px));
		border-block: 1px solid var(--border);
		background: var(--card);
		min-height: 600px;
		margin-top: 24px;
	}
	.function-list {
		min-width: 0;
		min-height: 0;
		position: sticky;
		top: 84px;
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
		border-right: 1px solid var(--border);
	}
	.rail-controls {
		flex-shrink: 0;
		position: sticky;
		top: 0;
		z-index: 2;
		background: var(--card);
		border-bottom: 1px solid var(--border);
	}
	.function-list {
		display: flex;
		flex-direction: column;
	}
	.rail-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px 16px 12px;
	}
	.rail-title {
		display: flex;
		align-items: baseline;
		gap: 8px;
		min-width: 0;
	}
	.rail-heading :global(button) {
		min-height: 36px;
		flex-shrink: 0;
	}
	.rail-heading h2 {
		font-size: 14px;
		font-weight: 650;
	}
	.rail-heading span {
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.function-search {
		position: relative;
		margin: 0 12px 12px;
	}
	.function-search > :global(svg) {
		position: absolute;
		top: 11px;
		left: 12px;
		color: var(--muted-foreground);
	}
	.function-search :global(input) {
		padding-left: 36px;
	}
	.function-sort {
		display: flex;
		gap: 8px;
		margin: 0 12px 16px;
		font-size: 13px;
	}
	.function-sort :global(> button) {
		width: 50%;
		min-width: 0;
	}
	.function-choice {
		flex-shrink: 0;
		display: grid;
		position: relative;
		width: 100%;
		gap: 6px;
		text-align: left;
		padding: 18px 28px 18px 16px;
		border-top: 1px solid var(--border);
		border-left: 3px solid transparent;
	}
	.function-choice span,
	.function-choice small {
		color: var(--muted-foreground);
		font-size: 12px;
	}
	.function-choice small {
		display: -webkit-box;
		line-clamp: 2;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.function-choice strong {
		font-family: var(--mono);
		font-size: 12px;
		line-height: 1.6;
		overflow-wrap: anywhere;
	}
	.function-choice :global(svg) {
		position: absolute;
		top: 40px;
		right: 4px;
	}
	.function-choice:hover,
	.function-choice.chosen {
		background: var(--device-soft);
	}
	.function-choice.chosen {
		border-left-color: var(--device-accent);
	}
	.function-choice:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: -2px;
	}
	.detail-pane {
		min-width: 0;
		min-height: 0;
		overflow-y: auto;
	}
	.device-script {
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		font-family: var(--mono);
		font-size: 12px;
		max-width: 64ch;
		min-width: 200px;
	}
	button:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 2px;
	}
	@media (max-width: 800px) {
		.function-browser {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto auto;
			height: auto;
		}
		.function-list {
			position: static;
			border-right: 0;
			border-bottom: 1px solid var(--border);
			padding-right: 0;
			max-height: 380px;
			overflow-y: auto;
		}
		.detail-pane {
			overflow: visible;
		}
	}
	@media (max-width: 600px) {
		.section-actions {
			flex-basis: 100%;
			width: 100%;
			flex-wrap: wrap;
		}
		.section-search {
			flex-basis: 100%;
			min-width: 0;
		}
		.device-heading {
			align-items: flex-start;
			gap: 16px;
			margin: 20px 0;
		}
		.device-symbol {
			width: 48px;
			height: 48px;
		}
		:global(.device-workspace .device-sections) {
			gap: 18px;
		}
		:global(.detail-pane .function-detail) {
			padding: 18px 14px;
		}
		.device-heading {
			align-items: center;
		}
	}
</style>
