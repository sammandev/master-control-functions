<script lang="ts">
	import { onMount } from 'svelte';
	import { Blocks, ChevronRight, CircuitBoard, Globe, Menu, Moon, Pencil, Plus, Search, Sun, Trash2 } from '@lucide/svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import * as DropdownMenu from '#lib/components/ui/dropdown-menu/index.js';
	import appIcon from '#lib/assets/mac_10.ico?url';

	type Device = { name: string; source: 'BUILT-IN' | 'DLL'; functions: number; parameters: number; criteria: number; examples: number };
	type SortKey = 'name' | 'functions' | 'parameters' | 'criteria' | 'examples';
	type DialogMode = 'create' | 'edit' | 'delete' | 'details';
	let devices = $state<Device[]>([
		{ name: 'CELLULAR', source: 'BUILT-IN', functions: 1, parameters: 1, criteria: 0, examples: 1 },
		{ name: 'CMD_CONSOLE', source: 'BUILT-IN', functions: 41, parameters: 6, criteria: 1, examples: 6 },
		{ name: 'CMD_CONSOLEV2', source: 'BUILT-IN', functions: 40, parameters: 7, criteria: 2, examples: 5 },
		{ name: 'CMD_SINGLE', source: 'BUILT-IN', functions: 18, parameters: 3, criteria: 1, examples: 4 },
		{ name: 'GOLDENKV', source: 'DLL', functions: 12, parameters: 8, criteria: 4, examples: 9 },
		{ name: 'GPIB', source: 'DLL', functions: 9, parameters: 5, criteria: 2, examples: 3 },
		{ name: 'TELNET', source: 'DLL', functions: 15, parameters: 6, criteria: 2, examples: 7 },
		{ name: 'WIFI_TEST', source: 'DLL', functions: 22, parameters: 14, criteria: 8, examples: 12 }
	]);
	const nav = ['Devices', 'Functions', 'Parameters', 'Criteria', 'Examples', 'Sequences', 'Orphans', 'Import'];
	const metrics = ['functions', 'parameters', 'criteria', 'examples'] as const;
	let search = $state('');
	let sort = $state<SortKey>('name');
	let order = $state('asc');
	let dark = $state(false);
	let mode = $state<DialogMode | null>(null);
	let selected = $state('');
	let name = $state('');
	let source = $state<'BUILT-IN' | 'DLL'>('BUILT-IN');
	let error = $state('');
	let message = $state('');
	let topbar: HTMLDivElement;
	let brand: HTMLDivElement;
	let navigation: HTMLElement;
	let headerActions: HTMLDivElement;
	let collapsed = $state(false);
	let visible = $derived.by(() => {
		const rows = devices.filter((device) => device.name.toLowerCase().includes(search.trim().toLowerCase()));
		return rows.toSorted((first, second) => {
			const comparison = sort === 'name' ? first.name.localeCompare(second.name) : first[sort] - second[sort];
			return (order === 'asc' ? 1 : -1) * comparison;
		});
	});
	let current = $derived(devices.find((device) => device.name === selected));
	let title = $derived(mode === 'create' ? 'Create device' : mode === 'edit' ? `Edit ${selected}` : mode === 'delete' ? `Delete ${selected}?` : selected);
	const sortOptions: { value: SortKey; label: string }[] = [
		{ value: 'name', label: 'Device name' },
		{ value: 'functions', label: 'Functions' },
		{ value: 'parameters', label: 'Parameters' },
		{ value: 'criteria', label: 'Criteria' },
		{ value: 'examples', label: 'Examples' }
	];
	const orderOptions = [{ value: 'asc', label: 'Ascending' }, { value: 'desc', label: 'Descending' }];
	const sourceOptions: { value: 'BUILT-IN' | 'DLL'; label: string }[] = [{ value: 'BUILT-IN', label: 'BUILT-IN' }, { value: 'DLL', label: 'DLL' }];

	onMount(() => {
		const originalDark = document.documentElement.classList.contains('dark');
		dark = originalDark;
		const measureNavigation = () => {
			const gap = Number.parseFloat(getComputedStyle(topbar).columnGap) || 0;
			const itemGap = Number.parseFloat(getComputedStyle(navigation).columnGap) || 0;
			const navigationWidth = Array.from(navigation.children).reduce((width, item) => width + item.getBoundingClientRect().width, 0) + itemGap * (navigation.children.length - 1);
			const required = brand.getBoundingClientRect().width + navigationWidth + headerActions.getBoundingClientRect().width + gap * 2;
			collapsed = window.matchMedia('(max-width: 680px)').matches || required > topbar.getBoundingClientRect().width;
		};
		const observer = new ResizeObserver(measureNavigation);
		observer.observe(topbar);
		observer.observe(brand);
		observer.observe(headerActions);
		window.addEventListener('resize', measureNavigation);
		measureNavigation();
		return () => {
			observer.disconnect();
			window.removeEventListener('resize', measureNavigation);
			document.documentElement.classList.toggle('dark', originalDark);
		};
	});

	function toggleTheme() {
		dark = !dark;
		document.documentElement.classList.toggle('dark', dark);
	}

	function openDialog(next: DialogMode, device?: Device) {
		selected = device?.name ?? '';
		name = device?.name ?? '';
		source = device?.source ?? 'BUILT-IN';
		error = '';
		mode = next;
	}

	function saveDevice() {
		const normalized = name.trim().toUpperCase();
		if (!normalized) { error = 'Enter a device name.'; return; }
		if (devices.some((device) => device.name === normalized && device.name !== selected)) { error = 'A device with this name already exists.'; return; }
		if (mode === 'create') devices = [...devices, { name: normalized, source, functions: 0, parameters: 0, criteria: 0, examples: 0 }];
		if (mode === 'edit') devices = devices.map((device) => device.name === selected ? { ...device, name: normalized } : device);
		message = mode === 'create' ? `${normalized} created in the demo.` : `${normalized} updated in the demo.`;
		mode = null;
	}

	function deleteDevice() {
		devices = devices.filter((device) => device.name !== selected);
		message = `${selected} deleted from the demo.`;
		mode = null;
	}
</script>

<div class={['directory-demo', { 'dark-demo': dark }]}>
	<header class="demo-header">
		<div class="demo-topbar" bind:this={topbar}>
			<div class="demo-brand" bind:this={brand}><img src={appIcon} width="26" height="26" alt="" /><div><strong>MasterControl</strong><span>Functions Database</span></div></div>
			<nav class={['demo-nav', { collapsed }]} bind:this={navigation} inert={collapsed} aria-label="Main navigation">{#each nav as item (item)}<button class={{ active: item === 'Devices' }} aria-current={item === 'Devices' ? 'page' : undefined} onclick={() => { if (item !== 'Devices') message = `${item} is not available in this demo.`; }}>{item}</button>{/each}</nav>
			{#if collapsed}<div class="menu-spacer"></div><DropdownMenu.Root><DropdownMenu.Trigger>{#snippet child({ props })}<Button {...props} variant="outline" size="icon" aria-label="Open navigation menu" title="Open navigation menu"><Menu size={16} data-icon="inline-start" /></Button>{/snippet}</DropdownMenu.Trigger><DropdownMenu.Content align="end" class="w-48"><DropdownMenu.Group>{#each nav as item (item)}<DropdownMenu.Item onclick={() => { if (item !== 'Devices') message = `${item} is not available in this demo.`; }}>{item}</DropdownMenu.Item>{/each}</DropdownMenu.Group></DropdownMenu.Content></DropdownMenu.Root>{/if}
			<div class="demo-header-actions" bind:this={headerActions}><span class="demo-label">MOCK DATA</span><Button variant="outline" size="icon" title={dark ? 'Switch to light mode' : 'Switch to dark mode'} aria-label={dark ? 'Switch to light mode' : 'Switch to dark mode'} onclick={toggleTheme}>{#if dark}<Sun size={16} data-icon="inline-start" />{:else}<Moon size={16} data-icon="inline-start" />{/if}</Button></div>
		</div>
	</header>
	<div class="demo-main">
		<div class="page-heading"><div><div class="eyebrow">REFERENCE CATALOG <span>/</span> DEVICES</div><h1>Devices <span>{devices.length}</span></h1></div><Button onclick={() => openDialog('create')}><Plus size={16} data-icon="inline-start" /> Create device</Button></div>
		<div class="toolbar"><div class="search-field"><Search size={17} /><input type="search" bind:value={search} placeholder="Search devices" aria-label="Search devices" /></div><div class="sort-controls"><div class="sort-control"><label for="demo-sort">Sort by</label><Select id="demo-sort" bind:value={sort} ariaLabel="Sort devices" options={sortOptions} class="h-[42px]" /></div><div class="sort-control"><label for="demo-order">Order</label><Select id="demo-order" bind:value={order} ariaLabel="Sort order" options={orderOptions} class="h-[42px]" /></div></div></div>
		<div class="list-heading"><span>ALL DEVICES</span><span>{visible.length} {visible.length === 1 ? 'device' : 'devices'}</span></div>
		<div class="cards">{#each visible as device (device.name)}
			<article class="device-card" aria-label={device.name}>
				<div class="card-top"><div class="device-icon">{#if device.source === 'DLL'}<Blocks size={20} />{:else}<CircuitBoard size={20} />{/if}</div><Badge variant="outline" class="ml-2.5">{device.source}</Badge><div class="card-actions"><Button variant="ghost" size="icon" aria-label={`Edit ${device.name}`} title={`Edit ${device.name}`} onclick={() => openDialog('edit', device)}><Pencil size={16} data-icon="inline-start" /></Button><Button variant="ghost" size="icon" aria-label={`Delete ${device.name}`} title={`Delete ${device.name}`} onclick={() => openDialog('delete', device)}><Trash2 size={16} data-icon="inline-start" /></Button></div></div>
				<button class="card-name" onclick={() => openDialog('details', device)}><span>{device.name}</span><ChevronRight size={18} /></button>
				<div class="card-metrics">{#each metrics as metric (metric)}<button aria-label={`${device[metric]} ${metric} for ${device.name}`} onclick={() => openDialog('details', device)}><strong>{device[metric]}</strong><span>{metric}</span></button>{/each}</div>
			</article>
		{:else}<p class="empty" role="status">No devices match your search.</p>{/each}</div>
		<section class="global-row" aria-label="Global definitions"><div class="global-icon"><Globe size={19} /></div><div><strong>Global definitions</strong><span>Shared across devices</span></div><div class="global-counts"><span>4 functions</span><span>12 parameters</span><span>3 criteria</span><span>6 examples</span></div><button onclick={() => { message = 'Global definitions are not available in this demo.'; }} aria-label="Open global definitions" title="Open global definitions"><ChevronRight size={19} /></button></section>
		<p class="demo-status" role="status">{message}</p>
	</div>
</div>

<Dialog open={mode !== null} onOpenChange={(open) => { if (!open) mode = null; }} {title}>
	{#if mode === 'create' || mode === 'edit'}
		<form class="demo-dialog" onsubmit={(event) => { event.preventDefault(); saveDevice(); }}>
			<label for="demo-device-name">Device name</label><input id="demo-device-name" bind:value={name} placeholder="e.g. WIFI_TEST" />
			{#if mode === 'create'}<label for="demo-device-source">Type</label><Select id="demo-device-source" bind:value={source} ariaLabel="Type" options={sourceOptions} />{/if}
			{#if error}<p role="alert">{error}</p>{/if}
			<div class="dialog-actions"><button type="button" onclick={() => (mode = null)}>Cancel</button><button type="submit">{mode === 'create' ? 'Create' : 'Save'}</button></div>
		</form>
	{:else if mode === 'delete'}
		<div class="demo-dialog"><p>Remove {selected} from the mock directory?</p><div class="dialog-actions"><button onclick={() => (mode = null)}>Cancel</button><button class="delete-button" onclick={deleteDevice}>Delete device</button></div></div>
	{:else if current}
		<div class="demo-dialog"><p>{current.source}</p><div class="dialog-metrics">{#each metrics as metric (metric)}<span><strong>{current[metric]}</strong> {metric}</span>{/each}</div><div class="dialog-actions"><button onclick={() => (mode = null)}>Close</button></div></div>
	{/if}
</Dialog>

<style>
	.directory-demo { --surface: #fff; --canvas: #f5f5f4; --ink: #181a19; --subtle: #686e6a; --line: #dedfdd; --soft: #efefed; --primary: var(--ink); --primary-foreground: var(--surface); --foreground: var(--ink); --background: var(--canvas); --card: var(--surface); --border: var(--line); --ring: var(--ink); --accent: var(--soft); min-height: 100%; background: var(--canvas); color: var(--ink); font-family: 'Segoe UI', sans-serif; }
	.directory-demo.dark-demo { --surface: #202221; --canvas: #151716; --ink: #f4f4f2; --subtle: #b6bbb7; --line: #414542; --soft: #303330; }
	.demo-header { position: sticky; top: 0; z-index: 10; background: var(--surface); border-bottom: 1px solid var(--line); }
	.demo-topbar, .demo-main { width: min(100% - 64px, 1440px); margin-inline: auto; }
	.demo-topbar { display: flex; align-items: center; min-height: 68px; gap: 20px; }
	.demo-brand, .demo-header-actions { display: flex; align-items: center; flex-shrink: 0; gap: 12px; }
	.demo-brand img { object-fit: contain; flex-shrink: 0; }
	.demo-brand strong, .demo-brand span { display: block; }
	.demo-brand strong { font-size: 15px; line-height: 1.2; }
	.demo-brand span { font-size: 11px; color: var(--subtle); }
	.demo-label { padding: 5px 8px; background: var(--soft); color: var(--subtle); font-size: 10px; font-weight: 700; }
	.demo-nav { display: flex; flex: 1; min-width: 0; gap: 4px; overflow: hidden; }
	.demo-nav.collapsed { position: absolute; left: -10000px; width: max-content; visibility: hidden; }
	.menu-spacer { flex: 1; }
	.demo-nav button { flex: 0 0 auto; padding: 12px 14px; border-bottom: 2px solid transparent; color: var(--subtle); font-size: 13px; }
	.demo-nav button:hover { color: var(--ink); background: var(--soft); }
	.demo-nav button.active { color: var(--ink); border-bottom-color: var(--ink); font-weight: 700; }
	.demo-main { padding: 38px 0 120px; }
	.page-heading { display: flex; align-items: end; justify-content: space-between; gap: 24px; }
	.eyebrow { color: var(--subtle); font-size: 10px; font-weight: 700; }
	.eyebrow span { padding-inline: 5px; }
	.page-heading h1 { margin: 8px 0 2px; font-size: 30px; line-height: 1.2; font-weight: 700; }
	.page-heading h1 span { display: inline-block; vertical-align: middle; padding: 4px 8px; margin-left: 5px; border-radius: 4px; background: var(--soft); color: var(--subtle); font-size: 12px; }
	.toolbar { display: flex; align-items: end; gap: 18px; margin-top: 32px; }
	.search-field { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; height: 42px; padding: 0 12px; border: 1px solid var(--line); border-radius: 4px; background: var(--surface); color: var(--subtle); }
	.search-field input { width: 100%; min-width: 0; padding: 0; border: 0; box-shadow: none; background: transparent; color: var(--ink); outline: none; font-size: 13px; }
	.search-field:focus-within { outline: 2px solid var(--ink); outline-offset: 2px; }
	.sort-controls { display: flex; gap: 12px; }
	.sort-control { display: grid; gap: 5px; min-width: 145px; color: var(--subtle); font-size: 11px; }
	.list-heading { display: flex; justify-content: space-between; margin: 32px 0 12px; color: var(--subtle); font-size: 11px; font-weight: 700; }
	.cards { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; }
	.device-card { min-width: 0; padding: 18px; border: 1px solid var(--line); border-radius: 6px; background: var(--surface); }
	.device-card:hover, .device-card:focus-within { border-color: var(--subtle); }
	.card-top, .card-actions { display: flex; align-items: center; }
	.device-icon, .global-icon { display: grid; place-items: center; width: 38px; height: 38px; border-radius: 4px; background: var(--soft); }
	.card-actions { gap: 2px; margin-left: auto; }
	.card-name { display: flex; align-items: center; justify-content: space-between; gap: 10px; width: 100%; padding: 20px 0 18px; text-align: left; font-size: 16px; font-weight: 700; overflow-wrap: anywhere; }
	.card-name :global(svg) { flex-shrink: 0; color: var(--subtle); }
	.card-metrics { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 6px; padding-top: 14px; border-top: 1px solid var(--line); }
	.card-metrics button { min-width: 0; text-align: left; }
	.card-metrics button:hover strong { text-decoration: underline; text-underline-offset: 3px; }
	.card-metrics strong, .card-metrics span { display: block; }
	.card-metrics strong { font-size: 17px; }
	.card-metrics span { margin-top: 3px; color: var(--subtle); font-size: 11px; overflow-wrap: anywhere; }
	.global-row { display: flex; align-items: center; gap: 12px; margin-top: 24px; padding: 16px 0; border-top: 1px solid var(--line); border-bottom: 1px solid var(--line); }
	.global-row strong, .global-row span { display: block; }
	.global-row strong { font-size: 13px; }
	.global-row span { margin-top: 3px; color: var(--subtle); font-size: 11px; }
	.global-counts { display: flex; flex-wrap: wrap; gap: 16px; margin-left: auto; }
	.global-row > button { display: grid; place-items: center; margin-left: auto; width: 38px; height: 38px; }
	.empty { margin: 0; padding: 30px 0; color: var(--subtle); font-size: 13px; }
	.demo-status { min-height: 20px; color: var(--subtle); font-size: 12px; }
	.directory-demo button:focus-visible { outline: 2px solid var(--ink); outline-offset: 2px; }
	.demo-dialog { display: grid; gap: 10px; font-size: 13px; }
	.demo-dialog label { font-weight: 600; }
	.demo-dialog input { width: 100%; height: 40px; padding: 0 10px; border: 1px solid #aaa; border-radius: 4px; }
	.demo-dialog p { margin: 0; }
	.dialog-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 14px; }
	.dialog-actions button { min-height: 38px; padding: 0 14px; border: 1px solid #999; border-radius: 4px; }
	.dialog-actions button[type='submit'], .dialog-actions .delete-button { background: #222; color: white; border-color: #222; }
	.dialog-metrics { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; padding: 14px 0; }
	@media (max-width: 1050px) { .cards { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
	@media (max-width: 680px) { .demo-topbar, .demo-main { width: calc(100% - 32px); } .demo-topbar { gap: 10px; } .demo-brand > div, .demo-label { display: none; } .demo-main { padding-top: 28px; } .page-heading { align-items: flex-start; flex-direction: column; } .toolbar { align-items: stretch; flex-direction: column; gap: 12px; } .search-field { flex: none; } .sort-control { flex: 1; min-width: 0; } .cards { grid-template-columns: 1fr; } .global-row { flex-wrap: wrap; } .global-counts { order: 1; width: 100%; margin-left: 50px; } }
</style>