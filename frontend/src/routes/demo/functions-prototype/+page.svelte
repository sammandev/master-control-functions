<script lang="ts">
	import { dev } from '$app/env';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { ChevronLeft, ChevronRight, Code2, ExternalLink, Filter, Menu, Moon, MoreHorizontal, Pencil, Plus, Search, Sun, Trash2, X } from '@lucide/svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Checkbox from '#lib/components/ui/checkbox.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir } from '#lib/table-sort.ts';
	import * as Tabs from '#lib/components/ui/tabs/index.ts';
	import * as DropdownMenu from '#lib/components/ui/dropdown-menu/index.js';
	import appIcon from '#lib/assets/mac_10.ico?url';

	type Fn = { name: string; source: string; device: string; description: string; category: string; role: string; phase: string; convention: string; updated: string };
	type Link = { name: string; device: string; required: boolean };
	const initial: Fn[] = [
		{ name: 'ARITHMETIC', source: 'BUILT-IN', device: 'TEST PROGRAM', description: 'Performs a simple arithmetic operation using two operands and stores the calculated result in a temporary variable.', category: 'data_processing', role: 'calculation', phase: 'test_execution', convention: '1', updated: '2026-10-06' },
		{ name: 'ARITHMETICVAR', source: 'BUILT-IN', device: 'TEST PROGRAM', description: 'Evaluates the arithmetic expression carried in the PARAM string over variables and stores the result via the criteria.', category: 'data_processing', role: 'calculation', phase: 'test_execution', convention: '1', updated: '2026-10-06' },
		{ name: 'AUTO_RUN', source: 'BUILT-IN', device: 'TEST PROGRAM', description: 'Toggles the auto-run flag, making the station execute automatically on its normal trigger.', category: 'station_control', role: 'configuration', phase: 'setup', convention: '1', updated: '2026-09-28' },
		{ name: 'AVERAGE_REAL_POWER_CSV', source: 'DLL', device: 'WIFI_TEST', description: 'Transposes recorded transmit power results and computes the average for each frequency.', category: 'measurement', role: 'calculation', phase: 'test_execution', convention: '2', updated: '2026-09-20' },
		{ name: 'TX_MULTI_VERIFICATION', source: 'DLL', device: 'WIFI_TEST', description: 'Measures transmit power and EVM for the selected antenna, frequency, and wireless standard.', category: 'measurement', role: 'verification', phase: 'test_execution', convention: '2', updated: '2026-09-18' },
		{ name: 'CONNECT_DEVICE', source: 'DLL', device: 'GOLDENKV', description: 'Initializes the device connection before executing the test sequence.', category: 'device_control', role: 'configuration', phase: 'setup', convention: '1', updated: '2026-08-30' }
	];
	const layouts = [{ key: 'A', name: 'Focus workspace' }, { key: 'B', name: 'Catalog table' }, { key: 'C', name: 'Filter explorer' }];
	const navigation = ['Devices', 'Functions', 'Parameters', 'Criteria', 'Examples', 'Sequences', 'Orphans', 'Import'];
	const sourceOptions = [{ value: '', label: 'All sources' }, { value: 'BUILT-IN', label: 'BUILT-IN' }, { value: 'DLL', label: 'DLL' }];
	const deviceOptions = [{ value: '', label: 'All devices' }, ...['TEST PROGRAM', 'WIFI_TEST', 'GOLDENKV'].map((value) => ({ value, label: value }))];
	const scopedDeviceOptions = [{ value: '', label: 'Global (no Device / Interface)' }, ...deviceOptions.slice(1)];
	const categoryOptions = [{ value: '', label: 'All categories' }, ...['data_processing', 'station_control', 'measurement', 'device_control'].map((value) => ({ value, label: value }))];
	const roleOptions = [{ value: '', label: 'All roles' }, ...['calculation', 'configuration', 'verification'].map((value) => ({ value, label: value }))];
	const phaseOptions = [{ value: '', label: 'All phases' }, ...['test_execution', 'setup'].map((value) => ({ value, label: value }))];
	const sortOptions = [{ value: 'name', label: 'Name' }, { value: 'device', label: 'Device / Interface' }, { value: 'updated', label: 'Updated' }];
	const orderOptions = [{ value: 'asc', label: 'Ascending' }, { value: 'desc', label: 'Descending' }];
	const pageSizeOptions = ['10', '20', '50'].map((value) => ({ value, label: value }));
	const encodingOptions = [{ value: 'named', label: 'Named' }, { value: 'raw', label: 'Raw' }];
	const scopeOptions = [{ value: 'bundle', label: 'Bundle' }, { value: 'line', label: 'Line' }];
	let dark = $state(false);
	onMount(() => {
		const originalDark = document.documentElement.classList.contains('dark');
		dark = originalDark;
		return () => document.documentElement.classList.toggle('dark', originalDark);
	});
	function toggleTheme() {
		dark = !dark;
		document.documentElement.classList.toggle('dark', dark);
	}
	let rows = $state<Fn[]>(initial.map((row) => ({ ...row })));
	let selectedName = $state('ARITHMETIC');
	let search = $state(''); let source = $state(''); let device = $state(''); let category = $state(''); let role = $state(''); let phase = $state('');
	let sort = $state('name'); let order = $state('asc'); let pageSize = $state('10'); let currentPage = $state(1);
	let activeTab = $state('parameters'); let modal = $state(''); let form = $state<Fn>({ ...initial[0] });
	let links = $state<Link[]>([{ name: 'LEFT_OPERAND', device: 'TEST PROGRAM', required: true }, { name: 'OPERATOR', device: 'TEST PROGRAM', required: true }, { name: 'RIGHT_OPERAND', device: 'TEST PROGRAM', required: true }]);
	let criteria = $state<Link[]>([{ name: 'RESULT', device: 'TEST PROGRAM', required: true }]);
	let linkName = $state(''); let linkDevice = $state('TEST PROGRAM'); let linkRequired = $state(true); let linkIndex = $state(-1);
	let example = $state('ARITHMETIC, LEFT_OPERAND=1, OPERATOR=+, RIGHT_OPERAND=2'); let exampleMeaning = $state('Add two operands and store the result.');
	let auditor = $state(''); let encoding = $state('named'); let scope = $state('bundle'); let notice = $state('');
	let variant = $derived(layouts.find((layout) => layout.key === page.url.searchParams.get('variant')) ?? layouts[0]);
	let filtered = $derived.by(() => {
		const term = search.trim().toLowerCase();
		return rows.filter((row) => (!term || `${row.name} ${row.description}`.toLowerCase().includes(term)) && (!source || row.source === source) && (!device || row.device === device) && (!category || row.category === category) && (!role || row.role === role) && (!phase || row.phase === phase)).toSorted((left, right) => {
			const key = sort === 'device' ? 'device' : sort === 'updated' ? 'updated' : 'name';
			return left[key].localeCompare(right[key]) * (order === 'asc' ? 1 : -1);
		});
	});
	let selected = $derived(filtered.find((row) => row.name === selectedName) ?? filtered[0]);
	let pages = $derived(Math.max(1, Math.ceil(filtered.length / Number(pageSize))));
	let visible = $derived(filtered.slice((Math.min(currentPage, pages) - 1) * Number(pageSize), Math.min(currentPage, pages) * Number(pageSize)));
	let currentLinks = $derived(activeTab === 'criteria' ? criteria : links);
	let parameterSort = $state<{ key: string; dir: SortDir }>({ key: 'position', dir: 'asc' });
	let criteriaSort = $state<{ key: string; dir: SortDir }>({ key: 'position', dir: 'asc' });
	let exampleSort = $state<{ key: string; dir: SortDir }>({ key: '', dir: 'asc' });
	let linkSort = $derived(activeTab === 'criteria' ? criteriaSort : parameterSort);
	let sortedLinks = $derived(sortRows(currentLinks.map((link, index) => ({ link, index })), (row) => {
		if (linkSort.key === 'name') return row.link.name;
		if (linkSort.key === 'device') return row.link.device;
		if (linkSort.key === 'required') return Number(row.link.required);
		return row.index;
	}, linkSort.dir));
	function sortLinks(key: string) {
		if (activeTab === 'criteria') criteriaSort = nextSort(criteriaSort.key, criteriaSort.dir, key);
		else parameterSort = nextSort(parameterSort.key, parameterSort.dir, key);
	}
	function switchVariant(direction: number) {
		const index = layouts.findIndex((layout) => layout.key === variant.key);
		void goto(resolve(`demo/functions-prototype?variant=${layouts[(index + direction + layouts.length) % layouts.length].key}`), { replaceState: true, reset: false });
	}
	function onKeydown(event: KeyboardEvent) {
		if (modal || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey || (event.target instanceof Element && event.target.closest('input, textarea, select, [contenteditable], [role="tablist"]'))) return;
		if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') { event.preventDefault(); switchVariant(event.key === 'ArrowLeft' ? -1 : 1); }
	}
	function clearFilters() { search = ''; source = ''; device = ''; category = ''; role = ''; phase = ''; currentPage = 1; }
	function openFunction(mode: string) {
		form = mode === 'new' ? { name: '', source: 'DLL', device: '', description: '', category: '', role: '', phase: '', convention: '', updated: '2026-11-16' } : { ...selected };
		modal = mode;
	}
	function saveFunction() {
		if (!form.name.trim()) return;
		if (modal === 'edit') rows = rows.map((row) => row.name === selected?.name ? { ...form } : row); else rows = [...rows, { ...form }];
		selectedName = form.name; modal = ''; notice = 'Changes saved in this prototype only.';
	}
	function openLink(index = -1) {
		linkIndex = index; linkName = index < 0 ? '' : currentLinks[index].name; linkDevice = index < 0 ? (selected?.device ?? '') : currentLinks[index].device; linkRequired = index < 0 || currentLinks[index].required; modal = 'link';
	}
	function saveLink() {
		if (!linkName.trim()) return;
		const updated = [...currentLinks]; const value = { name: linkName, device: linkDevice, required: linkRequired };
		if (linkIndex < 0) updated.push(value); else updated[linkIndex] = value;
		if (activeTab === 'criteria') criteria = updated; else links = updated;
		modal = ''; notice = 'Link saved in this prototype only.';
	}
</script>

<svelte:window onkeydown={dev ? onKeydown : undefined} />

{#snippet filters()}
	<div class="filters" aria-label="Function filters">
		{#if variant.key !== 'A'}<label class="search"><Search size={16} /><span class="sr-only">Search functions</span><input bind:value={search} placeholder="Search name or description..." /></label>{/if}
		<div class="filter-field"><span>Source</span><Select ariaLabel="Filter by source" bind:value={source} options={sourceOptions} /></div>
		<div class="filter-field"><span>Device / Interface</span><Select ariaLabel="Filter by Device / Interface" bind:value={device} options={deviceOptions} /></div>
		<div class="filter-field"><span>Category</span><Select ariaLabel="Filter by category" bind:value={category} options={categoryOptions} /></div>
		<div class="filter-field"><span>Execution role</span><Select ariaLabel="Filter by execution role" bind:value={role} options={roleOptions} /></div>
		<div class="filter-field"><span>Execution phase</span><Select ariaLabel="Filter by execution phase" bind:value={phase} options={phaseOptions} /></div>
		<Button class="filter-clear" variant="outline" onclick={clearFilters}>Clear filters</Button>
	</div>
{/snippet}

{#snippet list()}
	<div class="list-heading"><div><strong>Functions</strong><span>{filtered.length} results</span></div></div>
	{#if variant.key === 'A'}<div class="catalog-controls"><label class="search"><Search size={16} /><span class="sr-only">Search functions</span><input bind:value={search} placeholder="Search name or description..." /></label><div class="catalog-sorting"><span>Catalog sorting</span><div><Select ariaLabel="Sort functions" bind:value={sort} options={sortOptions} /><Select ariaLabel="Sort direction" bind:value={order} options={orderOptions} /></div></div></div>{/if}
	<div class="function-list">{#each visible as row (row.name)}<button class:chosen={selected?.name === row.name} aria-pressed={selected?.name === row.name} onclick={() => selectedName = row.name}><span class="row-name">{row.name}</span><span class="row-meta"><span class="source-dot"></span>{row.source} <span class="dot-divider">/</span> {row.device}</span><span class="row-description">{row.description}</span></button>{:else}<p class="empty">No functions match these filters.</p>{/each}</div>
	<div class="pagination"><div class="page-size"><span>Rows</span><Select ariaLabel="Rows per page" bind:value={pageSize} options={pageSizeOptions} onchange={() => currentPage = 1} /></div><span>{Math.min(currentPage, pages)} / {pages}</span><button aria-label="Previous page" disabled={currentPage <= 1} onclick={() => currentPage--}><ChevronLeft size={16} /></button><button aria-label="Next page" disabled={currentPage >= pages} onclick={() => currentPage++}><ChevronRight size={16} /></button></div>
{/snippet}

{#snippet detail()}
	{#if selected}<section class="detail" aria-label="Function details">
		<div class="detail-top"><div><div class="eyebrow">FUNCTION REFERENCE <span>/</span> {selected.device || 'GLOBAL'}</div><h2>{selected.name}</h2><div class="identity"><Badge>{selected.source}</Badge><Badge>{selected.device || 'GLOBAL'}</Badge></div></div><div class="detail-actions"><Button variant="outline" onclick={() => openFunction('edit')}><Pencil data-icon="inline-start" /> Edit</Button><Button variant="outline" size="icon" aria-label="Delete function" title="Delete function" onclick={() => modal = 'delete'}><Trash2 /></Button><Button onclick={() => modal = 'compose'}><Code2 data-icon="inline-start" /> Compose script</Button></div></div>
		<p class="description">{selected.description}</p>
		<dl class="metadata"><div><dt>Category</dt><dd>{selected.category || '-'}</dd></div><div><dt>Execution role</dt><dd>{selected.role || '-'}</dd></div><div><dt>Execution phase</dt><dd>{selected.phase || '-'}</dd></div><div><dt>Convention</dt><dd>{selected.convention || '-'}</dd></div><div><dt>Updated</dt><dd>{selected.updated}</dd></div></dl>
		<Tabs.Root bind:value={activeTab}>
			<Tabs.List aria-label="Function sections"><Tabs.Trigger value="parameters">Parameters <span class="tab-count">{links.length}</span></Tabs.Trigger><Tabs.Trigger value="criteria">Criteria <span class="tab-count">{criteria.length}</span></Tabs.Trigger><Tabs.Trigger value="examples">Examples</Tabs.Trigger><Tabs.Trigger value="mcp">MCP readiness</Tabs.Trigger></Tabs.List>
			{#each ['parameters', 'criteria'] as kind (kind)}<Tabs.Content value={kind}><div class="section-toolbar"><div><strong>{kind === 'parameters' ? 'Parameter links' : 'Criteria links'}</strong><p>{currentLinks.length} linked definitions</p></div><Button size="sm" onclick={() => openLink()}><Plus data-icon="inline-start" /> Attach {kind === 'parameters' ? 'parameter' : 'criteria'}</Button></div><div class="table-scroll"><table><thead><tr><SortHead label="Position" col="position" active={linkSort.key} dir={linkSort.dir} onclick={() => sortLinks('position')} /><SortHead label="Name" col="name" active={linkSort.key} dir={linkSort.dir} onclick={() => sortLinks('name')} /><SortHead label="Device" col="device" active={linkSort.key} dir={linkSort.dir} onclick={() => sortLinks('device')} /><SortHead label="Required" col="required" active={linkSort.key} dir={linkSort.dir} onclick={() => sortLinks('required')} /><th><span class="sr-only">Actions</span></th></tr></thead><tbody>{#each sortedLinks as row (row.index)}<tr><td>{row.index + 1}</td><td><strong class="mono">{row.link.name}</strong></td><td>{row.link.device || 'Global'}</td><td><Checkbox aria-label={`Required: ${row.link.name}`} bind:checked={row.link.required} /></td><td><button class="icon-action" aria-label={`Edit ${row.link.name}`} title="Edit link" onclick={() => openLink(row.index)}><MoreHorizontal size={17} /></button></td></tr>{:else}<tr><td colspan="5">No links attached.</td></tr>{/each}</tbody></table></div></Tabs.Content>{/each}
			<Tabs.Content value="examples"><div class="section-toolbar"><div><strong>Examples</strong><p>1 example</p></div><Button size="sm" onclick={() => modal = 'example'}><Plus data-icon="inline-start" /> New example</Button></div><div class="table-scroll"><table><thead><tr><SortHead label="Device" col="device" active={exampleSort.key} dir={exampleSort.dir} onclick={() => exampleSort = nextSort(exampleSort.key, exampleSort.dir, 'device')} /><SortHead label="Script" col="script" active={exampleSort.key} dir={exampleSort.dir} onclick={() => exampleSort = nextSort(exampleSort.key, exampleSort.dir, 'script')} /><SortHead label="Meaning" col="meaning" active={exampleSort.key} dir={exampleSort.dir} onclick={() => exampleSort = nextSort(exampleSort.key, exampleSort.dir, 'meaning')} /><th></th></tr></thead><tbody><tr><td>{selected.device}</td><td><code>{example}</code></td><td>{exampleMeaning}</td><td><button class="icon-action" aria-label="Edit example" onclick={() => modal = 'example'}><Pencil size={16} /></button></td></tr></tbody></table></div></Tabs.Content>
			<Tabs.Content value="mcp"><div class="section-toolbar"><div><strong>MCP readiness</strong><p>Review grammar and script keys for this function.</p></div><Badge>Review needed</Badge></div><div class="audit-grid"><label>Device scope<input value={selected.device} /></label><label>Audited by<input bind:value={auditor} placeholder="Your name" /></label><div class="modal-field"><span>Encoding</span><Select ariaLabel="Encoding" bind:value={encoding} options={encodingOptions} /></div><div class="modal-field"><span>Scope</span><Select ariaLabel="Scope" bind:value={scope} options={scopeOptions} /></div><Button variant="outline" onclick={() => notice = 'Readiness checked in the prototype.'}>Check readiness</Button><Button onclick={() => notice = auditor ? 'Audit saved in the prototype.' : 'Enter an auditor name to save.'}>Save grammar</Button></div></Tabs.Content>
		</Tabs.Root>
		<div class="detail-footer"><span>Function reference</span><button onclick={() => notice = 'Manage function is represented by this prototype.'}>Manage function <ExternalLink size={14} /></button></div>
	</section>{:else}<div class="empty detail">Select a function to see its details.</div>{/if}
{/snippet}

{#if dev}<div class={['demo-shell', { 'dark-demo': dark }]}>
<header class="demo-header"><div class="demo-topbar"><div class="demo-brand"><img src={appIcon} width="26" height="26" alt="" /><div><strong>MasterControl</strong><span>Functions Database</span></div></div><nav class="demo-nav" aria-label="Main navigation">{#each navigation as item (item)}<button class={{ active: item === 'Functions' }} aria-current={item === 'Functions' ? 'page' : undefined} onclick={() => { if (item !== 'Functions') notice = `${item} is not available in this demo.`; }}>{item}</button>{/each}</nav><div class="demo-menu"><DropdownMenu.Root><DropdownMenu.Trigger>{#snippet child({ props })}<Button {...props} variant="outline" size="icon" aria-label="Open navigation menu" title="Open navigation menu"><Menu data-icon="inline-start" /></Button>{/snippet}</DropdownMenu.Trigger><DropdownMenu.Content align="end" class="w-48"><DropdownMenu.Group>{#each navigation as item (item)}<DropdownMenu.Item onclick={() => { if (item !== 'Functions') notice = `${item} is not available in this demo.`; }}>{item}</DropdownMenu.Item>{/each}</DropdownMenu.Group></DropdownMenu.Content></DropdownMenu.Root></div><div class="demo-header-actions"><span class="demo-label">MOCK DATA</span><Button variant="outline" size="icon" title={dark ? 'Switch to light mode' : 'Switch to dark mode'} aria-label={dark ? 'Switch to light mode' : 'Switch to dark mode'} onclick={toggleTheme}>{#if dark}<Sun data-icon="inline-start" />{:else}<Moon data-icon="inline-start" />{/if}</Button></div></div></header>
<div class="prototype" class:table-layout={variant.key === 'B'} class:explorer-layout={variant.key === 'C'}>
	<header class="page-heading"><div><div class="eyebrow">REFERENCE CATALOG <span>/</span> FUNCTIONS</div><h1>Functions <span class="heading-count">{rows.length}</span></h1></div><Button onclick={() => openFunction('new')}><Plus data-icon="inline-start" /> New function</Button></header>
	{#if notice}<div class="notice" role="status">{notice}<button aria-label="Dismiss" onclick={() => notice = ''}><X size={15} /></button></div>{/if}
	{#if variant.key !== 'C'}<div class="filter-band">{@render filters()}</div>{/if}
	{#if variant.key === 'A'}<div class="workspace"><aside class="catalog">{@render list()}</aside>{@render detail()}</div>
	{:else if variant.key === 'B'}<div class="catalog-bar"><div><strong>Function catalog</strong><span>{filtered.length} results</span></div><div class="sorting"><label>Sort <select bind:value={sort}><option value="name">Name</option><option value="device">Device / Interface</option><option value="updated">Updated</option></select></label><select aria-label="Sort direction" bind:value={order}><option value="asc">Ascending</option><option value="desc">Descending</option></select></div></div><div class="table-scroll catalog-table"><table><thead><tr><th>FUNCTION</th><th>SOURCE</th><th>DEVICE / INTERFACE</th><th>CATEGORY</th><th>UPDATED</th><th></th></tr></thead><tbody>{#each visible as row (row.name)}<tr class:selected-row={selected?.name === row.name}><td><button class="table-name" onclick={() => selectedName = row.name}>{row.name}</button><small>{row.description}</small></td><td><Badge>{row.source}</Badge></td><td>{row.device}</td><td>{row.category}</td><td>{row.updated}</td><td><button class="icon-action" aria-label={`Inspect ${row.name}`} onclick={() => selectedName = row.name}><ChevronRight size={17} /></button></td></tr>{:else}<tr><td colspan="6">No matching functions.</td></tr>{/each}</tbody></table></div><div class="table-pagination"><label>Rows <select aria-label="Rows per page" bind:value={pageSize}><option>10</option><option>20</option><option>50</option></select></label><span>{Math.min(currentPage, pages)} / {pages}</span><button aria-label="Previous page" disabled={currentPage <= 1} onclick={() => currentPage--}><ChevronLeft size={16} /></button><button aria-label="Next page" disabled={currentPage >= pages} onclick={() => currentPage++}><ChevronRight size={16} /></button></div>{@render detail()}
	{:else}<div class="explorer"><aside class="filter-sidebar"><div class="list-heading"><div><strong>Refine results</strong><span>Find a function</span></div><Filter size={16} /></div>{@render filters()}<div class="sorting"><label>Sort by<select bind:value={sort}><option value="name">Name</option><option value="device">Device / Interface</option><option value="updated">Updated</option></select></label><label>Direction<select bind:value={order}><option value="asc">Ascending</option><option value="desc">Descending</option></select></label></div></aside><div class="explorer-content"><div class="explorer-list">{@render list()}</div>{@render detail()}</div></div>{/if}
	<footer class="page-footer"><span>MasterControl Functions Database</span><span>Function catalog <span class="dot-divider">/</span> Sample data</span></footer>
</div>
<nav class="switcher" aria-label="Prototype layouts"><button aria-label="Previous layout" title="Previous layout" onclick={() => switchVariant(-1)}><ChevronLeft size={18} /></button><div aria-live="polite"><span>PROTOTYPE</span><strong>{variant.key} / {variant.name}</strong></div><button aria-label="Next layout" title="Next layout" onclick={() => switchVariant(1)}><ChevronRight size={18} /></button></nav>
{#if modal === 'new' || modal === 'edit'}<Dialog open={true} onOpenChange={(open) => { if (!open) modal = ''; }} title={modal === 'new' ? 'New function' : 'Edit function'} description="Changes are kept in this prototype only."><form class="modal-form" onsubmit={(event) => { event.preventDefault(); saveFunction(); }}><div class="modal-field"><span>Source</span><Select ariaLabel="Source" bind:value={form.source} options={sourceOptions.slice(1)} /></div><div class="modal-field"><span>Device / Interface</span><Select ariaLabel="Device / Interface" bind:value={form.device} options={scopedDeviceOptions} /></div><label>Name<input bind:value={form.name} required /></label><label>Description<textarea bind:value={form.description}></textarea></label><label>Convention<input type="number" bind:value={form.convention} /></label><label>Category<input bind:value={form.category} /></label><label>Execution role<input bind:value={form.role} /></label><label>Execution phase<input bind:value={form.phase} /></label><div class="modal-actions"><Button variant="outline" onclick={() => modal = ''}>Cancel</Button><Button type="submit">Save function</Button></div></form></Dialog>{/if}
{#if modal === 'delete'}<Dialog open={true} onOpenChange={(open) => { if (!open) modal = ''; }} title="Delete function" description={`Delete ${selected?.name}? Links and examples can become unresolved. This action only affects the prototype.`}><div class="modal-actions"><Button variant="outline" onclick={() => modal = ''}>Cancel</Button><Button variant="destructive" onclick={() => { rows = rows.filter((row) => row.name !== selected?.name); modal = ''; notice = 'Function removed from the prototype.'; }}>Delete function</Button></div></Dialog>{/if}
{#if modal === 'link'}<Dialog open={true} onOpenChange={(open) => { if (!open) modal = ''; }} title={`${linkIndex < 0 ? 'Attach' : 'Edit'} ${activeTab === 'criteria' ? 'criteria' : 'parameter'}`} description="Choose a definition and its scope."><form class="modal-form" onsubmit={(event) => { event.preventDefault(); saveLink(); }}><label>Definition<input bind:value={linkName} placeholder="Search or enter a name" required /></label><div class="modal-field"><span>Device / Interface</span><Select ariaLabel="Device / Interface" bind:value={linkDevice} options={scopedDeviceOptions} /></div><label>Position<input type="number" min="1" value={linkIndex < 0 ? currentLinks.length + 1 : linkIndex + 1} /></label><div class="check-label"><Checkbox id="link-required" bind:checked={linkRequired} /><label for="link-required">Required</label></div><div class="modal-actions">{#if linkIndex >= 0}<Button variant="destructive" onclick={() => { if (activeTab === 'criteria') criteria = criteria.filter((_, index) => index !== linkIndex); else links = links.filter((_, index) => index !== linkIndex); modal = ''; }}>Detach</Button>{/if}<Button variant="outline" onclick={() => modal = ''}>Cancel</Button><Button type="submit">Save link</Button></div></form></Dialog>{/if}
{#if modal === 'example'}<Dialog open={true} onOpenChange={(open) => { if (!open) modal = ''; }} title="Example" description="Edit the sample script and its meaning."><form class="modal-form" onsubmit={(event) => { event.preventDefault(); modal = ''; notice = 'Example saved in the prototype.'; }}><label>Device / Interface<input value={selected?.device} /></label><label>Script<textarea bind:value={example}></textarea></label><label>Meaning<textarea bind:value={exampleMeaning}></textarea></label><div class="modal-actions"><Button variant="outline" onclick={() => modal = ''}>Cancel</Button><Button type="submit">Save example</Button></div></form></Dialog>{/if}
{#if modal === 'compose'}<Dialog open={true} onOpenChange={(open) => { if (!open) modal = ''; }} title="Compose script" description={`Compose a sample line for ${selected?.name}.`}><form class="modal-form" onsubmit={(event) => { event.preventDefault(); notice = 'Script composed in the prototype.'; modal = ''; }}><label>Device / Interface<input value={selected?.device} /></label>{#each links as link, index (index)}<label>{link.name}<input placeholder={link.required ? 'Required value' : 'Optional value'} /></label>{/each}<label>Criteria<input placeholder="Result" /></label><div class="modal-actions"><Button variant="outline" onclick={() => modal = ''}>Cancel</Button><Button type="submit">Compose</Button></div></form></Dialog>{/if}
</div>
{:else}<p>Not found.</p>{/if}

<style>
	.demo-shell { --surface: #fff; --canvas: #f5f5f4; --ink: #181a19; --subtle: #686e6a; --line: #dedfdd; --soft: #efefed; --primary: var(--ink); --primary-foreground: var(--surface); --foreground: var(--ink); --background: var(--canvas); --card: var(--surface); --border: var(--line); --ring: var(--ink); --accent: var(--soft); --muted: var(--soft); --muted-foreground: var(--subtle); min-height: 100vh; background: var(--canvas); color: var(--ink); font-family: 'Segoe UI', sans-serif; }
	.demo-shell.dark-demo { --surface: #202221; --canvas: #151716; --ink: #f4f4f2; --subtle: #b6bbb7; --line: #414542; --soft: #303330; }
	:global(body:has(.demo-shell) [data-dialog-content]),
	:global(body:has(.demo-shell) [data-dropdown-menu-content]),
	:global(body:has(.demo-shell) [data-popover-content]) { --primary: #181a19; --primary-foreground: #fff; --foreground: #181a19; --background: #f5f5f4; --card: #fff; --popover: #fff; --border: #dedfdd; --ring: #181a19; --accent: #efefed; --muted-foreground: #686e6a; }
	:global(.dark body:has(.demo-shell) [data-dialog-content]),
	:global(.dark body:has(.demo-shell) [data-dropdown-menu-content]),
	:global(.dark body:has(.demo-shell) [data-popover-content]) { --primary: #f4f4f2; --primary-foreground: #202221; --foreground: #f4f4f2; --background: #151716; --card: #202221; --popover: #202221; --border: #414542; --ring: #f4f4f2; --accent: #303330; --muted-foreground: #b6bbb7; }
	.demo-header { position: sticky; top: 0; z-index: 10; background: var(--surface); border-bottom: 1px solid var(--line); }
	.demo-topbar, .prototype { width: min(100% - 64px, 1440px); margin-inline: auto; }
	.demo-topbar { display: flex; align-items: center; min-height: 68px; gap: 20px; }
	.demo-brand, .demo-header-actions { display: flex; align-items: center; flex-shrink: 0; gap: 12px; }
	.demo-brand strong, .demo-brand span { display: block; }
	.demo-brand strong { font-size: 15px; line-height: 1.2; }
	.demo-brand span { font-size: 11px; color: var(--subtle); }
	.demo-label { padding: 5px 8px; background: var(--soft); color: var(--subtle); font-size: 10px; font-weight: 700; }
	.demo-nav { display: flex; flex: 1; min-width: 0; gap: 4px; }
	.demo-nav button { flex: 0 0 auto; padding: 12px 14px; border-bottom: 2px solid transparent; color: var(--subtle); font-size: 13px; }
	.demo-nav button:hover { color: var(--ink); background: var(--soft); }
	.demo-nav button.active { color: var(--ink); border-bottom-color: var(--ink); font-weight: 700; }
	.demo-menu { display: none; margin-left: auto; }
	.prototype { padding-top: 38px; color: var(--foreground); }
	.page-heading { display: flex; justify-content: space-between; align-items: end; gap: 24px; margin-bottom: 32px; }
	.page-heading h1 { font-size: 30px; line-height: 1.2; font-weight: 700; margin: 8px 0 2px; }
	.page-heading p, .section-toolbar p { color: var(--muted-foreground); font-size: 13px; margin: 3px 0 0; }
	.eyebrow { font-size: 10px; font-weight: 700; color: var(--muted-foreground); }
	.eyebrow span, .dot-divider { margin: 0 5px; color: var(--border); }
	.heading-count { display: inline-block; vertical-align: middle; padding: 4px 8px; margin-left: 5px; border-radius: 4px; background: var(--soft); color: var(--subtle); font-size: 12px; }
	.filter-band { margin-bottom: 32px; }
	.filters { display: flex; flex-wrap: wrap; align-items: end; gap: 10px; }
	.filters label, .filter-field, .modal-field, .sorting label, .audit-grid label, .modal-form label { display: flex; flex-direction: column; gap: 5px; color: var(--muted-foreground); font-size: 11px; font-weight: 600; }
	.filter-field { flex: 1 1 150px; min-width: 0; }
	.filters :global(.filter-clear) { flex: 1 1 130px; }
	input:not([type='checkbox']), select, textarea { min-width: 0; border: 1px solid var(--border); background: var(--background); color: var(--foreground); border-radius: 4px; padding: 7px 9px; font-size: 13px; min-height: 34px; }
	input:focus-visible, select:focus-visible, textarea:focus-visible, button:focus-visible { outline: 2px solid var(--ring); outline-offset: 2px; }
	.search { flex: 1 1 230px; position: relative; justify-content: end; }
	.search :global(svg) { position: absolute; left: 10px; bottom: 9px; color: var(--muted-foreground); }
	.search input { padding-left: 32px; width: 100%; }
	.workspace { display: grid; grid-template-columns: minmax(240px, 310px) minmax(0, 1fr); background: var(--card); border: 1px solid var(--border); border-radius: 6px; overflow: hidden; }
	.catalog { border-right: 1px solid var(--border); min-width: 0; }
	.catalog-controls { display: grid; gap: 12px; padding: 14px 16px; border-bottom: 1px solid var(--border); }
	.catalog-controls .search { display: flex; flex-direction: column; }
	.catalog-sorting { display: grid; gap: 5px; color: var(--muted-foreground); font-size: 11px; font-weight: 600; }
	.catalog-sorting > div { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 8px; }
	.list-heading { min-height: 61px; display: flex; justify-content: space-between; align-items: center; padding: 14px 18px; border-bottom: 1px solid var(--border); }
	.list-heading strong, .catalog-bar strong { display: block; font-size: 13px; }
	.list-heading span, .catalog-bar span { display: block; font-size: 11px; color: var(--muted-foreground); margin-top: 3px; }
	.function-list { max-height: 540px; overflow: auto; }
	.function-list > button { display: flex; flex-direction: column; gap: 7px; width: 100%; text-align: left; padding: 14px 16px; border-bottom: 1px solid var(--border); border-left: 3px solid transparent; cursor: pointer; }
	.function-list > button:hover, .function-list > button.chosen { background: var(--accent); }
	.function-list > button.chosen { border-left-color: var(--primary); }
	.row-name, .mono, .table-name { font-family: var(--mono); font-weight: 700; font-size: 12px; overflow-wrap: anywhere; }
	.row-meta { color: var(--muted-foreground); font-size: 10px; font-weight: 600; }
	.source-dot { display: inline-block; width: 6px; height: 6px; background: var(--primary); border-radius: 50%; margin-right: 5px; }
	.row-description { color: var(--muted-foreground); font-size: 12px; line-height: 1.45; display: -webkit-box; line-clamp: 2; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
	.pagination, .table-pagination { display: flex; gap: 10px; align-items: center; justify-content: end; padding: 12px; color: var(--muted-foreground); font-size: 11px; }
	.pagination label, .table-pagination label, .page-size { margin-right: auto; }
	.page-size { display: flex; align-items: center; gap: 6px; }
	.page-size :global(button) { width: 64px; }
	.table-pagination select { min-height: 28px; padding: 3px; }
	.pagination button, .table-pagination button, .icon-action { display: inline-grid; place-items: center; width: 30px; height: 30px; cursor: pointer; }
	button:disabled { opacity: .4; cursor: default; }
	.detail { min-width: 0; padding: 24px; }
	.detail-top { display: flex; align-items: start; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
	.detail h2 { font-family: var(--mono); font-size: 20px; font-weight: 700; margin: 7px 0 10px; overflow-wrap: anywhere; }
	.identity, .detail-actions { display: flex; flex-wrap: wrap; gap: 7px; align-items: center; }
	.description { font-size: 13px; line-height: 1.6; margin: 17px 0; max-width: 780px; }
	.metadata { display: flex; flex-wrap: wrap; gap: 16px 30px; padding: 15px 0 22px; }
	.metadata dt { font-size: 11px; color: var(--muted-foreground); margin-bottom: 4px; }
	.metadata dd { font-size: 12px; font-weight: 600; overflow-wrap: anywhere; }
	.detail :global([role='tablist']) { width: 100%; max-width: 100%; justify-content: flex-start; gap: 24px; overflow-x: auto; padding: 0; border-bottom: 1px solid var(--border); background: transparent; }
	.detail :global([role='tab']) { flex-shrink: 0; min-height: 44px; padding: 12px 0; border-radius: 0; border-bottom: 2px solid transparent; background: transparent; color: var(--muted-foreground); box-shadow: none; }
	.detail :global([role='tab'][data-state='active']) { color: var(--foreground); border-bottom-color: var(--foreground); }
	.tab-count { color: var(--muted-foreground); font-size: 10px; margin-left: 4px; }
	.section-toolbar { display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 22px 0 14px; }
	.section-toolbar strong { font-size: 13px; }
	.table-scroll { max-width: 100%; overflow-x: auto; }
	table { width: 100%; border-collapse: collapse; text-align: left; font-size: 12px; }
	th { color: var(--muted-foreground); font-size: 10px; font-weight: 600; background: var(--muted); }
	th, td { padding: 12px; border-bottom: 1px solid var(--border); vertical-align: middle; }
	td code { font-size: 11px; overflow-wrap: anywhere; }
	.detail-footer { margin-top: 30px; border-top: 1px solid var(--border); padding-top: 14px; display: flex; justify-content: space-between; color: var(--muted-foreground); font-size: 11px; }
	.detail-footer button { display: flex; align-items: center; gap: 5px; color: var(--primary); cursor: pointer; }
	.catalog-bar { display: flex; justify-content: space-between; align-items: end; gap: 12px; margin-bottom: 12px; }
	.sorting { display: flex; flex-wrap: wrap; align-items: end; gap: 8px; }
	.catalog-table { border: 1px solid var(--border); background: var(--card); }
	.catalog-table table { min-width: 800px; }
	.catalog-table td:first-child { width: 32%; }
	.table-name { color: var(--primary); cursor: pointer; text-align: left; }
	.catalog-table small { display: block; color: var(--muted-foreground); margin-top: 5px; max-width: 380px; }
	.selected-row { background: var(--accent); }
	.table-layout > .detail { background: var(--card); border: 1px solid var(--border); border-top: 3px solid var(--primary); margin-top: 14px; }
	.explorer { display: grid; grid-template-columns: 220px minmax(0, 1fr); border-top: 1px solid var(--border); }
	.filter-sidebar { border-right: 1px solid var(--border); padding: 0 16px 20px 0; }
	.filter-sidebar .list-heading { padding-left: 0; }
	.filter-sidebar .filters { display: flex; flex-direction: column; align-items: stretch; padding: 18px 0; }
	.filter-sidebar .filters label, .filter-sidebar .filters .filter-field { width: 100%; max-width: none; }
	.filter-sidebar .search { flex: auto; }
	.filter-sidebar .sorting { display: grid; }
	.explorer-content { min-width: 0; padding-left: 20px; }
	.explorer-list .function-list { display: flex; gap: 8px; overflow-x: auto; max-height: none; padding: 12px 0; }
	.explorer-list .function-list > button { flex: 0 0 210px; border: 1px solid var(--border); border-bottom: 3px solid transparent; border-radius: 4px; background: var(--card); }
	.explorer-list .function-list > button.chosen { border-bottom-color: var(--primary); background: var(--accent); }
	.explorer-list .pagination { padding: 0 0 8px; }
	.explorer-content .detail { background: var(--card); border: 1px solid var(--border); }
	.page-footer { display: flex; justify-content: space-between; gap: 10px; color: var(--muted-foreground); font-size: 11px; padding: 25px 0 75px; }
	.notice { display: flex; justify-content: space-between; padding: 10px 12px; background: var(--accent); font-size: 12px; margin-bottom: 12px; }
	.empty { padding: 20px; color: var(--muted-foreground); font-size: 13px; }
	.modal-form { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
	.modal-form label, .modal-form .modal-field { font-size: 12px; color: var(--foreground); }
	.modal-form textarea { min-height: 70px; resize: vertical; }
	.modal-form .check-label { display: flex; gap: 8px; align-items: center; }
	.modal-actions { grid-column: 1 / -1; display: flex; justify-content: end; flex-wrap: wrap; gap: 8px; margin-top: 12px; }
	.audit-grid { display: flex; flex-wrap: wrap; align-items: end; gap: 12px; padding: 10px 0; }
	.switcher { position: fixed; bottom: 18px; left: 50%; transform: translateX(-50%); z-index: 100; display: flex; align-items: center; gap: 12px; padding: 8px; width: 300px; max-width: calc(100vw - 24px); background: var(--foreground); color: var(--background); border-radius: 8px; box-shadow: 0 4px 20px #0003; }
	.switcher div { flex: 1; text-align: center; }
	.switcher span, .switcher strong { display: block; }
	.switcher span { font-size: 10px; opacity: .7; }
	.switcher strong { font-size: 13px; }
	.switcher button { display: grid; place-items: center; width: 40px; height: 40px; cursor: pointer; }
	.switcher button:hover { background: #ffffff20; }
	@media (max-width: 1200px) { .demo-nav { display: none; } .demo-menu { display: block; } }
	@media (max-width: 850px) { .workspace { grid-template-columns: 1fr; } .catalog { border-right: 0; border-bottom: 1px solid var(--border); } .function-list { max-height: 250px; } .explorer { grid-template-columns: 1fr; } .filter-sidebar { border-right: 0; padding-right: 0; } .filter-sidebar .filters { display: flex; flex-direction: row; } .filter-sidebar .filters label, .filter-sidebar .filters .filter-field { flex: 1 1 150px; } .explorer-content { padding-left: 0; } }
	@media (max-width: 600px) { .demo-topbar, .prototype { width: calc(100% - 32px); } .demo-topbar { gap: 10px; } .demo-brand > div, .demo-label { display: none; } .prototype { padding-top: 28px; } .page-heading { align-items: start; } .page-heading h1 { font-size: 23px; } .page-heading p { max-width: 210px; } .detail { padding: 18px 14px; } .metadata { gap: 14px; } .section-toolbar { align-items: start; flex-wrap: wrap; } .modal-form { grid-template-columns: 1fr; } .page-footer { flex-direction: column; } }
</style>