<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import {
		CircuitBoard,
		Blocks,
		Layers,
		CircleHelp,
		Plus,
		Search,
		Trash2,
		Pencil,
		RefreshCw,
		ChevronRight,
		Globe
	} from '@lucide/svelte';
	import {
		listDevices,
		createDevice,
		deleteDevice,
		renameDevice,
		devicePath,
		deviceTypeLabel,
		type DeviceSummary,
		type DeviceSource
	} from '#lib/devices.ts';
	import { formatApiError } from '#lib/client.ts';
	import { toast } from 'svelte-sonner';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Label from '#lib/components/ui/label.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import { sortRows, type SortDir } from '#lib/table-sort.ts';

	let items = $state<DeviceSummary[]>([]);
	let loading = $state(true);
	let loadError = $state('');
	let search = $state('');

	onMount(() => {
		void load();
	});

	async function load() {
		loading = true;
		loadError = '';
		try {
			const res = await listDevices();
			items = res.items;
		} catch (e) {
			items = [];
			loadError = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	// Count cells deep-link to the matching tab on the device page (?tab=);
	// the device name opens the default (Functions) tab.
	const COUNT_COLUMNS: {
		key: 'functions' | 'parameters' | 'criteria' | 'examples';
		tab: string;
		label: string;
	}[] = [
		{ key: 'functions', tab: 'functions', label: 'functions' },
		{ key: 'parameters', tab: 'parameters', label: 'parameters' },
		{ key: 'criteria', tab: 'criteria', label: 'criteria' },
		{ key: 'examples', tab: 'examples', label: 'examples' }
	];

	const DEVICE_SORT_KEYS: Record<
		string,
		(row: DeviceSummary) => string | number | null | undefined
	> = {
		device: (r) => r.device_name,
		functions: (r) => r.functions,
		parameters: (r) => r.parameters,
		criteria: (r) => r.criteria,
		examples: (r) => r.examples
	};

	let sortKey = $state('device');
	let sortDir = $state<SortDir>('asc');

	const deviceItems = $derived(items.filter((row) => row.device_name !== ''));
	const global = $derived(items.find((row) => row.device_name === ''));
	const view = $derived(
		sortRows(
			deviceItems.filter((row) =>
				row.device_name.toLowerCase().includes(search.trim().toLowerCase())
			),
			DEVICE_SORT_KEYS[sortKey]!,
			sortDir
		)
	);

	// ---- create dialog -------------------------------------------------------
	// A Device can be registered before it has any content (ADR-0010); the
	// API uppercase-normalizes the name (ADR-0002), so the dialog previews
	// the normalized form before submit.
	let createOpen = $state(false);
	let createName = $state('');
	let createSource = $state<DeviceSource>('BUILT-IN');
	const sourceOptions: { value: DeviceSource; label: string }[] = [
		{ value: 'BUILT-IN', label: 'BUILT-IN' },
		{ value: 'DLL', label: 'DLL' }
	];
	let createError = $state('');
	let creating = $state(false);
	let editOpen = $state(false);
	let editOriginal = $state('');
	let editName = $state('');
	let editError = $state('');
	let renaming = $state(false);
	const normalizedEditName = $derived(editName.trim().toUpperCase());

	function openEdit(name: string) {
		editOriginal = name;
		editName = name;
		editError = '';
		editOpen = true;
	}
	async function submitRename() {
		if (renaming) return;
		renaming = true;
		editError = '';
		try {
			await renameDevice(editOriginal, editName);
			toast.success('Device renamed');
			editOpen = false;
			await load();
		} catch (cause) {
			editError = formatApiError(cause);
		} finally {
			renaming = false;
		}
	}

	// ---- delete confirm -------------------------------------------------------
	let confirmName = $state('');
	let confirmOpen = $state(false);
	let confirmation = $state('');
	let deleteError = $state('');
	let deleting = $state(false);
	const expectedConfirmation = $derived(`DELETE ${confirmName}`);

	function openDelete(name: string) {
		confirmName = name;
		confirmation = '';
		deleteError = '';
		confirmOpen = true;
	}

	function openCreate() {
		createName = '';
		createSource = 'BUILT-IN';
		createError = '';
		createOpen = true;
	}

	const normalizedName = $derived(createName.trim().toUpperCase());

	async function submitCreate() {
		creating = true;
		createError = '';
		try {
			await createDevice(createName, createSource);
			toast.success('Device created');
			createOpen = false;
			await load();
		} catch (e) {
			createError = formatApiError(e);
		} finally {
			creating = false;
		}
	}

	async function confirmRemove() {
		if (deleting || confirmation !== expectedConfirmation) return;
		deleting = true;
		deleteError = '';
		try {
			await deleteDevice(confirmName, confirmation);
			toast.success('Device deleted');
			confirmOpen = false;
			await load();
		} catch (e) {
			deleteError = formatApiError(e);
		} finally {
			deleting = false;
		}
	}
</script>

<svelte:head><title>Devices | MasterControl Functions Database</title></svelte:head>

<div class="directory-heading">
	<div>
		<div class="directory-title">
			<h1>Devices</h1>
			<span>{deviceItems.length} devices</span>
		</div>
	</div>
	<Button onclick={openCreate}><Plus size={16} /> Create device</Button>
</div>

<div class="directory-tools">
	<div class="device-search">
		<Search size={16} /><Input
			class="h-10"
			bind:value={search}
			aria-label="Search devices"
			placeholder="Search devices"
		/>
	</div>
	<div class="sort-field">
		<Label for="device-sort">Sort by</Label><Select
			class="h-10"
			id="device-sort"
			bind:value={sortKey}
			ariaLabel="Sort devices"
			options={[
				{ value: 'device', label: 'Device name' },
				...COUNT_COLUMNS.map((col) => ({
					value: col.key,
					label: col.label[0].toUpperCase() + col.label.slice(1)
				}))
			]}
		/>
	</div>
	<div class="sort-field">
		<Label for="device-order">Order</Label><Select
			class="h-10"
			id="device-order"
			bind:value={sortDir}
			ariaLabel="Sort order"
			options={[
				{ value: 'asc', label: 'Ascending' },
				{ value: 'desc', label: 'Descending' }
			]}
		/>
	</div>
</div>

{#if loadError}
	<div class="banner-error" role="alert">
		{loadError}<Button variant="outline" size="sm" onclick={load}
			><RefreshCw size={16} /> Retry</Button
		>
	</div>
{/if}

{#if loading}<p role="status" class="banner-empty">Loading devices...</p>
{:else if !loadError}
	<div class="list-heading">
		<span>All devices</span><span>{view.length} {view.length === 1 ? 'device' : 'devices'}</span>
	</div>
	<div class="device-grid">
		{#each view as row (row.device_name)}
			<article class="device-tile" aria-label={row.device_name}>
				<div class="tile-header">
					<span
						class="device-icon"
						role="img"
						aria-label={`${deviceTypeLabel(row)} type`}
						title={deviceTypeLabel(row)}
					>
						{#if row.source === 'BUILT-IN'}<CircuitBoard size={20} />
						{:else if row.source === 'DLL'}<Blocks size={20} />
						{:else if (row.sources?.length ?? 0) > 1}<Layers size={20} />
						{:else if row.sources?.[0] === 'BUILT-IN'}<CircuitBoard size={20} />
						{:else if row.sources?.[0] === 'DLL'}<Blocks size={20} />
						{:else}<CircleHelp size={20} />{/if}
					</span>
					<Badge variant="outline">{deviceTypeLabel(row)}</Badge>
					<div class="tile-actions">
						<Button
							size="icon"
							variant="ghost"
							aria-label={`Edit ${row.device_name}`}
							title={`Edit ${row.device_name}`}
							disabled={renaming}
							onclick={() => openEdit(row.device_name)}><Pencil size={16} /></Button
						>
						<Button
							size="icon"
							variant="ghost"
							aria-label={`Delete ${row.device_name}`}
							title={`Delete ${row.device_name}`}
							disabled={deleting}
							onclick={() => openDelete(row.device_name)}><Trash2 size={16} /></Button
						>
					</div>
				</div>
				<a
					class="device-open"
					href={resolve('/devices/[name]', { name: devicePath(row.device_name) })}
					aria-label={row.device_name}
				>
					<div class="tile-name">
						<h2>{row.device_name}</h2>
						<ChevronRight size={18} />
					</div>
				</a>
				<div class="tile-counts">
					{#each COUNT_COLUMNS as col (col.key)}<a
							href={resolve(`/devices/[name]?tab=${col.tab}`, {
								name: devicePath(row.device_name)
							})}
							aria-label={`${row[col.key]} ${col.label}`}
							><strong>{row[col.key]}</strong><span>{col.label}</span></a
						>{/each}
				</div>
			</article>
		{:else}<p class="banner-empty" role="status">
				{search ? 'No devices match your search.' : 'No devices in the dataset yet.'}
			</p>{/each}
	</div>
	{#if global}
		<div class="list-heading">
			<span>Global scope</span><span>Shared across devices</span>
		</div>
		<div class="device-grid">
			<section class="device-tile global-tile" aria-label="Global definitions">
				<div class="tile-header">
					<span class="device-icon" role="img" aria-label="Global type" title="Global">
						<Globe size={20} />
					</span>
					<Badge variant="outline">GLOBAL</Badge>
				</div>
				<a class="device-open" href={resolve('/devices/[name]', { name: '_global' })}>
					<div class="tile-name">
						<h2>Global definitions</h2>
						<ChevronRight size={18} />
					</div>
				</a>
				<div class="tile-counts">
					{#each COUNT_COLUMNS as col (col.key)}<a
							href={resolve(`/devices/[name]?tab=${col.tab}`, { name: '_global' })}
							aria-label={`${global[col.key]} ${col.label}`}
							><strong>{global[col.key]}</strong><span>{col.label}</span></a
						>{/each}
				</div>
			</section>
		</div>
	{/if}
{/if}

<Dialog open={createOpen} onOpenChange={(o) => (createOpen = o)} title="Create device">
	{#if createError}
		<div class="banner-error" role="alert">{createError}</div>
	{/if}
	<form
		onsubmit={(e) => {
			e.preventDefault();
			submitCreate();
		}}
	>
		<div class="field">
			<Label for="device-name">Device name</Label>
			<Input id="device-name" bind:value={createName} placeholder="e.g. WIFI_TEST" />
			{#if normalizedName}
				<span class="muted">Will be created as {normalizedName}</span>
			{/if}
		</div>
		<div class="field mt-4">
			<Label for="device-source">Type</Label>
			<Select
				id="device-source"
				ariaLabel="Type"
				bind:value={createSource}
				options={sourceOptions}
				disabled={creating}
				required
			/>
		</div>
		<div class="mt-4 flex justify-end gap-2">
			<Button variant="outline" onclick={() => (createOpen = false)}>Cancel</Button>
			<Button type="submit" disabled={creating || !normalizedName}>
				{creating ? 'Creating…' : 'Create'}
			</Button>
		</div>
	</form>
</Dialog>

<Dialog
	open={editOpen}
	onOpenChange={(next) => {
		if (!renaming) editOpen = next;
	}}
	title={`Edit ${editOriginal}`}
>
	{#if editError}<div class="banner-error" role="alert">{editError}</div>{/if}
	<form
		onsubmit={(event) => {
			event.preventDefault();
			void submitRename();
		}}
	>
		<div class="field">
			<Label for="edit-device-name">Device name</Label>
			<Input id="edit-device-name" bind:value={editName} disabled={renaming} />
			{#if normalizedEditName}<span class="muted">Will be saved as {normalizedEditName}</span>{/if}
		</div>
		<div class="mt-4 flex justify-end gap-2">
			<Button variant="outline" disabled={renaming} onclick={() => (editOpen = false)}
				>Cancel</Button
			>
			<Button
				type="submit"
				disabled={renaming || !normalizedEditName || normalizedEditName === editOriginal}
				>{renaming ? 'Saving...' : 'Save'}</Button
			>
		</div>
	</form>
</Dialog>

<Dialog
	open={confirmOpen}
	onOpenChange={(next) => {
		if (!deleting) confirmOpen = next;
	}}
	title={`Delete ${confirmName}?`}
	description="Permanently deletes this device, its Functions, scoped definitions, links, Examples, and grammar/key audits. Shared data belonging to other devices is preserved. A device used by Sequence steps cannot be deleted. This cannot be undone."
>
	{#if deleteError}<div class="banner-error" role="alert">{deleteError}</div>{/if}
	<form
		onsubmit={(event) => {
			event.preventDefault();
			void confirmRemove();
		}}
	>
		<div class="field">
			<Label for="delete-device-confirmation">Type {expectedConfirmation} to confirm</Label>
			<Input
				id="delete-device-confirmation"
				bind:value={confirmation}
				disabled={deleting}
				autocomplete="off"
				spellcheck={false}
			/>
		</div>
		<div class="mt-4 flex justify-end gap-2">
			<Button variant="outline" disabled={deleting} onclick={() => (confirmOpen = false)}
				>Cancel</Button
			>
			<Button
				type="submit"
				variant="destructive"
				disabled={deleting || confirmation !== expectedConfirmation}
				>{deleting ? 'Deleting...' : 'Delete'}</Button
			>
		</div>
	</form>
</Dialog>

<style>
	.directory-heading,
	.directory-tools,
	.directory-title,
	.tile-header,
	.tile-name,
	.tile-actions {
		display: flex;
		align-items: center;
		gap: 16px;
	}
	.directory-heading {
		justify-content: space-between;
		flex-wrap: wrap;
		margin: 8px 0 32px;
	}
	h1 {
		font-size: 28px;
		font-weight: 650;
		line-height: 1.3;
	}
	.directory-title {
		gap: 8px;
		align-items: baseline;
	}
	.directory-title span {
		color: var(--muted-foreground);
		font-size: 12px;
		background: var(--muted);
		border-radius: 4px;
		padding: 4px 8px;
	}
	.directory-tools {
		align-items: end;
		flex-wrap: wrap;
		margin-bottom: 24px;
	}
	.device-search {
		position: relative;
		flex: 1;
		min-width: 180px;
	}
	.device-search :global(svg) {
		position: absolute;
		left: 12px;
		top: 12px;
		color: var(--muted-foreground);
	}
	.device-search :global(input) {
		padding-left: 36px;
	}
	.sort-field {
		display: grid;
		gap: 5px;
		min-width: 168px;
		font-size: 11px;
	}
	.list-heading {
		display: flex;
		justify-content: space-between;
		margin: 32px 0 12px;
		color: var(--muted-foreground);
		font-size: 11px;
		font-weight: 700;
		text-transform: uppercase;
	}
	.device-grid {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 16px;
	}
	.device-tile {
		display: flex;
		flex-direction: column;
		padding: 18px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--card);
		min-width: 0;
		transition:
			border-color 150ms ease,
			box-shadow 150ms ease;
	}
	.device-tile:hover,
	.device-tile:focus-within {
		border-color: var(--muted-foreground);
		box-shadow: 0 2px 8px color-mix(in srgb, var(--foreground) 6%, transparent);
	}
	.device-open {
		display: block;
		color: var(--foreground);
		margin: 20px 0 18px;
		text-decoration: none;
	}
	.tile-header {
		gap: 10px;
	}
	.device-icon {
		display: grid;
		place-items: center;
		width: 38px;
		height: 38px;
		border-radius: 6px;
		background: var(--device-soft);
		color: var(--device-accent);
	}
	.tile-actions {
		gap: 4px;
		margin-left: auto;
	}
	.tile-actions :global(button) {
		width: 44px;
		height: 44px;
	}
	.tile-name {
		justify-content: space-between;
		gap: 8px;
	}
	.tile-name :global(svg) {
		flex-shrink: 0;
		color: var(--muted-foreground);
		transition: transform 150ms ease;
	}
	.device-open:hover .tile-name :global(svg) {
		transform: translateX(2px);
		color: var(--foreground);
	}
	h2 {
		font-size: 16px;
		font-weight: 650;
		line-height: 1.4;
		overflow-wrap: anywhere;
	}
	.tile-counts {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 8px;
		margin-top: auto;
		padding-top: 14px;
		border-top: 1px solid var(--border);
	}
	.tile-counts a {
		display: flex;
		gap: 4px;
		flex-direction: column;
		font-size: 11px;
		min-height: 44px;
		overflow-wrap: anywhere;
		text-decoration: none;
		border-radius: 4px;
		padding: 4px;
		transition: background-color 150ms ease;
	}
	.tile-counts a:hover {
		background: var(--muted);
	}
	.tile-counts strong {
		font-size: 17px;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}
	.tile-counts span {
		color: var(--muted-foreground);
		text-transform: capitalize;
	}
	a:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 4px;
	}
	@media (max-width: 1050px) {
		.device-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
	@media (max-width: 680px) {
		.device-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
