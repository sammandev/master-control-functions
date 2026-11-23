<script lang="ts">
	import { copyText } from '#lib/clipboard.ts';
	import { dev } from '$app/env';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		CircuitBoard,
		Radio,
		Terminal,
		Cable,
		ChevronRight,
		ChevronDown,
		ArrowLeft,
		Copy,
		Check,
		Search,
		Layers,
		FileCode
	} from '@lucide/svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import DevicePrototypeSwitcher from '#lib/components/DevicePrototypeSwitcher.svelte';
	import DeviceDirectoryPrototype from '#lib/components/DeviceDirectoryPrototype.svelte';

	type Definition = {
		name: string;
		type: string;
		value: string;
		constraint: string;
		scope: 'device' | 'global';
		required: boolean;
	};
	type Example = { title: string; meaning: string; script: string; scope: 'device' | 'global' };
	type MockFunction = {
		id: string;
		name: string;
		source: string;
		channel: string;
		category: string;
		description: string;
		parameters: Definition[];
		criteria: Definition[];
		examples: Example[];
	};
	type Section = 'functions' | 'criteria' | 'examples';
	const functions: MockFunction[] = [
		{
			id: 'tx',
			name: 'TX_MULTI_VERIFICATION',
			source: 'DLL',
			channel: 'WIFI_TEST',
			category: 'Transmit verification',
			description:
				'Measure transmit power and EVM for the selected antenna, frequency, and wireless standard.',
			parameters: [
				{
					name: 'ANT',
					type: 'integer',
					value: '1',
					constraint: '1, 2, 4, 8',
					scope: 'device',
					required: true
				},
				{
					name: 'STANDARD',
					type: 'string',
					value: '11AX',
					constraint: '11AC, 11AX',
					scope: 'device',
					required: true
				},
				{
					name: 'FREQ',
					type: 'integer / MHz',
					value: '5975',
					constraint: '5955 - 7115',
					scope: 'device',
					required: true
				},
				{
					name: 'BW',
					type: 'integer / MHz',
					value: '20',
					constraint: '20, 40, 80, 160',
					scope: 'device',
					required: true
				},
				{
					name: 'DATA_RATE',
					type: 'string',
					value: 'MCS9',
					constraint: 'MCS0 - MCS11',
					scope: 'device',
					required: true
				},
				{
					name: 'DUT_TX_SETTLE_TIME_MS',
					type: 'integer / ms',
					value: '200',
					constraint: '0 - 5000',
					scope: 'global',
					required: false
				}
			],
			criteria: [
				{
					name: 'POWER',
					type: 'range / dB',
					value: '-3~5dB',
					constraint: 'Lower ~ upper',
					scope: 'device',
					required: true
				},
				{
					name: 'EVM',
					type: 'number / dB',
					value: '-35',
					constraint: 'Maximum threshold',
					scope: 'device',
					required: true
				}
			],
			examples: [
				{
					title: '6 GHz / 11AX / MCS9',
					meaning: 'Verify transmit power and EVM with antenna 1 at 5975 MHz.',
					scope: 'device',
					script:
						'DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=1&STANDARD=11AX&FREQ=5975&BW=20&DATA_RATE=MCS9|POWER=-3~5dB&EVM=-35dB|400000|NEXT|FAIL'
				},
				{
					title: '80 MHz transmit verification',
					meaning: 'A wider-bandwidth verification using antenna 2.',
					scope: 'device',
					script:
						'DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=2&STANDARD=11AX&FREQ=5985&BW=80&DATA_RATE=MCS9|POWER=-3~5dB&EVM=-35dB|400000|NEXT|FAIL'
				},
				{
					title: 'Shared transmit reference',
					meaning: 'A shared function example with explicit power and EVM criteria.',
					scope: 'global',
					script:
						'DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=1&STANDARD=11AC&FREQ=5180&BW=20&DATA_RATE=MCS0|POWER=-3~5dB&EVM=-30dB|400000|NEXT|FAIL'
				}
			]
		},
		{
			id: 'rx',
			name: 'RX_PER',
			source: 'DLL',
			channel: 'WIFI_TEST',
			category: 'Receive verification',
			description: 'Measure packet error rate at the configured receive level.',
			parameters: [
				{
					name: 'FREQ',
					type: 'integer / MHz',
					value: '5975',
					constraint: '5955 - 7115',
					scope: 'device',
					required: true
				},
				{
					name: 'RX_LEVEL',
					type: 'number / dBm',
					value: '-65',
					constraint: '-95 - -20',
					scope: 'device',
					required: true
				}
			],
			criteria: [
				{
					name: 'PER',
					type: 'number / %',
					value: '10',
					constraint: 'Maximum 10%',
					scope: 'device',
					required: true
				}
			],
			examples: [
				{
					title: 'Receive sensitivity',
					meaning: 'Check packet error rate at -65 dBm.',
					scope: 'device',
					script: 'DLLWIFI_TEST|RX_PER|FREQ=5975&RX_LEVEL=-65|PER=10%|400000|NEXT|FAIL'
				}
			]
		},
		{
			id: 'serial',
			name: 'GET_SERIAL_NUMBER',
			source: 'DLL',
			channel: 'TELNET',
			category: 'Device identity',
			description: 'Read the unit serial number through the device terminal.',
			parameters: [
				{
					name: 'COMMAND',
					type: 'string',
					value: 'get_serial',
					constraint: 'Terminal command',
					scope: 'device',
					required: true
				}
			],
			criteria: [
				{
					name: 'SERIAL_LENGTH',
					type: 'integer',
					value: '12',
					constraint: 'Expected length',
					scope: 'device',
					required: true
				}
			],
			examples: [
				{
					title: 'Read unit identity',
					meaning: 'Read and check the serial number before testing.',
					scope: 'device',
					script: 'DLLTELNET|GET_SERIAL_NUMBER|COMMAND=get_serial|SERIAL_LENGTH=12|10000|NEXT|FAIL'
				}
			]
		},
		{
			id: 'report',
			name: 'UPLOAD_RESULT',
			source: 'DLL',
			channel: 'IPLAS',
			category: 'Result reporting',
			description: 'Submit the completed test result to the reporting interface.',
			parameters: [
				{
					name: 'RESULT',
					type: 'string',
					value: 'PASS',
					constraint: 'PASS, FAIL',
					scope: 'device',
					required: true
				}
			],
			criteria: [],
			examples: [
				{
					title: 'Submit passing result',
					meaning: 'Upload the passing result after verification.',
					scope: 'device',
					script: 'DLLIPLAS|UPLOAD_RESULT|RESULT=PASS|NULL|10000|NEXT|FAIL'
				}
			]
		}
	];
	const devices = [
		{
			name: 'WIFI_TEST',
			description: 'Wireless measurement and verification',
			category: 'Wireless',
			ids: ['tx', 'rx']
		},
		{
			name: 'TELNET',
			description: 'Terminal access and device identity',
			category: 'Control',
			ids: ['serial']
		},
		{
			name: 'IPLAS',
			description: 'Test results and production reporting',
			category: 'Reporting',
			ids: ['report']
		},
		{
			name: 'GOLDENKV',
			description: 'Reference measurements and golden-unit checks',
			category: 'Reference',
			ids: ['tx', 'rx']
		},
		{ name: 'GPIB', description: 'Instrument control', category: 'Control', ids: [] }
	];
	let variant = $derived(
		['A', 'B', 'C', 'D', 'E'].includes(page.url.searchParams.get('variant') ?? '')
			? page.url.searchParams.get('variant')!
			: 'A'
	);
	let selectedDevice = $state('');
	let selectedFunction = $state('');
	let deviceSearch = $state('');
	let functionSearch = $state('');
	let section = $state<Section>('functions');
	let functionFocused = $state(false);
	let copied = $state('');
	let copyStatus = $state('');
	let activeDevice = $derived(devices.find((device) => device.name === selectedDevice));
	let deviceFunctions = $derived(functions.filter((fn) => activeDevice?.ids.includes(fn.id)));
	let visibleDevices = $derived(
		devices.filter((device) =>
			`${device.name} ${device.description}`.toLowerCase().includes(deviceSearch.toLowerCase())
		)
	);
	let visibleFunctions = $derived(
		deviceFunctions.filter((fn) =>
			`${fn.name} ${fn.description} ${fn.category}`
				.toLowerCase()
				.includes(functionSearch.toLowerCase())
		)
	);
	let activeFunction = $derived(deviceFunctions.find((fn) => fn.id === selectedFunction));

	function changeVariant(key: string) {
		const url = new URL(page.url.href);
		url.searchParams.set('variant', key);
		goto(resolve(`demo/device-prototype?${url.searchParams.toString()}${url.hash}`), {
			replace: true,
			reset: false
		});
	}

	function selectDevice(name: string) {
		selectedDevice = name;
		selectedFunction = devices.find((device) => device.name === name)?.ids[0] ?? '';
		functionSearch = '';
		section = 'functions';
		functionFocused = false;
		copied = '';
		copyStatus = '';
	}

	function selectFunction(id: string) {
		selectedFunction = id;
		functionFocused = true;
		copied = '';
		copyStatus = '';
	}

	function deviceStats(ids: string[]) {
		const related = functions.filter((fn) => ids.includes(fn.id));
		return {
			criteria: new Set(related.flatMap((fn) => fn.criteria.map((definition) => definition.name)))
				.size,
			examples: related.reduce(
				(count, fn) => count + fn.examples.filter((example) => example.scope === 'device').length,
				0
			)
		};
	}

	function parametersFor(fn: MockFunction) {
		return fn.parameters.map((definition) =>
			selectedDevice === 'GOLDENKV' && definition.name === 'ANT'
				? { ...definition, value: '2' }
				: definition
		);
	}

	async function copyExample(script: string) {
		const success = await copyText(script);
		copied = success ? script : '';
		copyStatus = success ? 'Script copied.' : 'Copy failed. The script text is selectable.';
	}
</script>

<svelte:head><title>Device UX Prototype | MasterControl Functions Database</title></svelte:head>

{#snippet scopeBadge(scope: 'device' | 'global')}
	<span class:shared={scope === 'global'} class="scope-badge"
		>{scope === 'global' ? 'Global' : selectedDevice}</span
	>
{/snippet}

{#snippet definitions(rows: Definition[])}
	<div class="table-scroll">
		<table class="definition-table">
			<thead
				><tr><th>Name</th><th>Type</th><th>Default</th><th>Constraint</th><th>Required</th></tr
				></thead
			>
			<tbody>
				{#each rows as definition (definition.name)}
					<tr
						><td><code>{definition.name}</code></td><td>{definition.type}</td><td
							><code>{definition.value}</code></td
						><td>{definition.constraint}</td><td>{definition.required ? 'Yes' : 'No'}</td></tr
					>
				{/each}
			</tbody>
		</table>
	</div>
{/snippet}

{#snippet exampleList(fn: MockFunction)}
	{#each ['device', 'global'] as scope (scope)}
		{@const examples = fn.examples.filter((example) => example.scope === scope)}
		{#if examples.length}
			<h4 class="group-heading">
				{scope === 'global' ? 'Global' : selectedDevice} <span>{examples.length}</span>
			</h4>
			<div class="examples-list">
				{#each examples as example (example.title)}
					<article class="example">
						<div class="example-title">
							<strong>{example.title}</strong><Button
								variant="ghost"
								size="icon"
								aria-label={`Copy ${example.title}`}
								title={`Copy ${example.title}`}
								onclick={() => copyExample(example.script)}
							>
								{#if copied === example.script}
									<Check size={16} />
								{:else}
									<Copy size={16} />
								{/if}
							</Button>
						</div>
						<p>{example.meaning}</p>
						<pre><code>{example.script}</code></pre>
					</article>
				{/each}
			</div>
		{/if}
	{/each}
{/snippet}

{#snippet functionContent(fn: MockFunction)}
	{@const scoped = parametersFor(fn).filter((definition) => definition.scope === 'device')}
	{@const global = parametersFor(fn).filter((definition) => definition.scope === 'global')}
	<div class="function-detail">
		<header class="function-heading">
			<div class="eyebrow">{fn.source} / {fn.channel}</div>
			<h2>{fn.name}</h2>
			<p>{fn.description}</p>
			<span class="category-label">{fn.category}</span>
		</header>
		<section class="detail-section" aria-label="Function parameters">
			<h3><Layers size={16} /> Parameters <span>{fn.parameters.length}</span></h3>
			{#if scoped.length}<div class="group-heading">{@render scopeBadge('device')}</div>
				{@render definitions(scoped)}{/if}
			{#if global.length}<div class="group-heading">{@render scopeBadge('global')}</div>
				{@render definitions(global)}{/if}
		</section>
		<section class="detail-section" aria-label="Function criteria">
			<h3><Check size={16} /> Criteria <span>{fn.criteria.length}</span></h3>
			{#if fn.criteria.length}{@render definitions(fn.criteria)}{:else}<p class="empty">
					No criteria defined.
				</p>{/if}
		</section>
		<section class="detail-section" aria-label="Function examples">
			<h3><FileCode size={16} /> Examples <span>{fn.examples.length}</span></h3>
			{@render exampleList(fn)}
		</section>
	</div>
{/snippet}

{#snippet deviceRail()}
	<aside class="device-rail" aria-label="Devices">
		<h2>Devices <span>{devices.length}</span></h2>
		{#each devices as device (device.name)}
			<button
				class:chosen={selectedDevice === device.name}
				class="rail-device"
				onclick={() => selectDevice(device.name)}
				aria-pressed={selectedDevice === device.name}
			>
				<CircuitBoard size={16} />
				<span>{device.name}</span>
				<small>{device.ids.length}</small>
			</button>
		{/each}
	</aside>
{/snippet}

{#snippet functionSearchBox()}
	<div class="search-box">
		<Search size={16} /><Input
			bind:value={functionSearch}
			placeholder="Search functions"
			aria-label="Search functions"
		/>
	</div>
{/snippet}

{#snippet deviceSections()}
	<div class="section-nav" role="group" aria-label="Device sections">
		{#each ['functions', 'criteria', 'examples'] as key (key)}
			<button
				class:active={section === key}
				aria-pressed={section === key}
				onclick={() => {
					section = key as Section;
					functionFocused = false;
				}}
			>
				{key === 'functions'
					? 'Functions'
					: key === 'criteria' ? 'Criteria' : 'Examples'}
			</button>
		{/each}
	</div>
{/snippet}

{#snippet aggregateSection()}
	<div class="aggregate">
		{#each deviceFunctions as fn (fn.id)}
			{#if section === 'examples' || fn.criteria.length}
				<section class="aggregate-function">
					<div class="aggregate-heading">
						<h2>{fn.name}</h2>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => {
								section = 'functions';
								selectFunction(fn.id);
							}}>Open function <ChevronRight size={14} /></Button
						>
					</div>
					{#if section === 'criteria'}{@render definitions(fn.criteria)}
						<h3>Function examples</h3>{/if}
					{@render exampleList(fn)}
				</section>
			{/if}
		{/each}
		{#if section === 'criteria' && !deviceFunctions.some((fn) => fn.criteria.length)}<p
				class="empty"
			>
				No criteria defined for this device.
			</p>{/if}
	</div>
{/snippet}

{#if dev}
	<div class="prototype" class:variant-b={variant === 'B'} class:variant-c={variant === 'C'} class:variant-d={variant === 'D'}>
		{#if variant === 'E'}
			<DeviceDirectoryPrototype />
		{:else}
		<header class="app-header">
			<button class="brand" onclick={() => selectDevice('')}
				><CircuitBoard size={24} /><span>MasterControl Functions Database</span></button
			><span class="mock-label">Prototype / Mock data</span>
		</header>
		{#if variant === 'D'}
			<div class="desk">
				<div class="desk-heading">
					<div><span class="desk-kicker">MASTERCONTROL / REFERENCE CATALOG</span><h1>Inspection desk</h1><p>Devices and their function definitions</p></div>
					<span class="desk-count">{devices.length} DEVICES</span>
				</div>
				<div class="desk-device-bar" aria-label="Select a device">
					{#each devices as device (device.name)}
						<button class:active={selectedDevice === device.name} aria-pressed={selectedDevice === device.name} onclick={() => selectDevice(device.name)}>
							<span>{device.name}</span><small>{device.ids.length}</small>
						</button>
					{/each}
				</div>
				{#if activeDevice}
					<div class="desk-context"><div><span class="desk-kicker">SELECTED DEVICE / {activeDevice.category}</span><h2>{activeDevice.name}</h2><p>{activeDevice.description}</p></div><div class="desk-metrics"><span><strong>{deviceFunctions.length}</strong> Functions</span><span><strong>{deviceStats(activeDevice.ids).criteria}</strong> Criteria</span><span><strong>{deviceStats(activeDevice.ids).examples}</strong> Examples</span></div></div>
					<section class="desk-section" aria-label="Functions">
						<div class="desk-section-heading"><div><span class="desk-kicker">01 / CATALOG</span><h2>Functions</h2></div><div class="desk-search"><Search size={16} /><input bind:value={functionSearch} placeholder="Search functions" aria-label="Search functions" /></div></div>
						<div class="table-scroll"><table class="desk-table"><thead><tr><th>Function</th><th>Source / Interface</th><th>Parameters</th><th>Criteria</th><th>Examples</th></tr></thead><tbody>
							{#each visibleFunctions as fn (fn.id)}<tr class:selected={selectedFunction === fn.id}><td><button onclick={() => selectFunction(fn.id)} aria-pressed={selectedFunction === fn.id}><strong>{fn.name}</strong><small>{fn.category}</small></button></td><td>{fn.source} / {fn.channel}</td><td>{fn.parameters.length}</td><td>{fn.criteria.length}</td><td>{fn.examples.length}</td></tr>{:else}<tr><td colspan="5">No matching functions.</td></tr>{/each}
						</tbody></table></div>
					</section>
					{#if activeFunction}
						<section class="desk-section desk-inspector" aria-label="Function inspector">
							<div class="desk-section-heading"><div><span class="desk-kicker">02 / INSPECTOR</span><h2>{activeFunction.name}</h2><p>{activeFunction.description}</p></div><span class="desk-badge">{activeFunction.source} / {activeFunction.channel}</span></div>
							<div class="desk-tabs" role="group" aria-label="Inspect function">
								{#each ['functions', 'criteria', 'examples'] as key (key)}<button class:active={section === key} aria-pressed={section === key} onclick={() => (section = key as Section)}>{key === 'functions' ? `Parameters (${activeFunction.parameters.length})` : key === 'criteria' ? `Criteria (${activeFunction.criteria.length})` : `Examples (${activeFunction.examples.length})`}</button>{/each}
							</div>
							{#if section === 'functions'}{@render definitions(parametersFor(activeFunction))}{:else if section === 'criteria'}{#if activeFunction.criteria.length}{@render definitions(activeFunction.criteria)}{:else}<p class="empty">No criteria defined.</p>{/if}{:else}{@render exampleList(activeFunction)}{/if}
						</section>
					{/if}
				{:else}
					<section class="desk-welcome"><span class="desk-kicker">DEVICE REGISTRY</span><h2>Select a device to inspect</h2><div class="desk-registry">{#each visibleDevices as device (device.name)}<button onclick={() => selectDevice(device.name)}><span><strong>{device.name}</strong><small>{device.description}</small></span><span>{device.ids.length} functions <ChevronRight size={16} /></span></button>{/each}</div></section>
				{/if}
				<p class="copy-status" role="status">{copyStatus}</p>
				<output class="context-state" aria-label="Prototype state">Layout D | Device: {selectedDevice || 'All'} | Function: {activeFunction?.name ?? 'None'} | Section: {section === 'functions' ? 'parameters' : section}</output>
			</div>
		{:else}
		<div class="workspace">
			{#if !activeDevice}
				<section class="directory">
					<div class="directory-heading">
						<div>
							<div class="eyebrow">REFERENCE CATALOG</div>
							<h1>Devices <span>{devices.length}</span></h1>
						</div>
						<div class="search-box">
							<Search size={16} /><Input
								bind:value={deviceSearch}
								placeholder="Search devices"
								aria-label="Search devices"
							/>
						</div>
					</div>
					{#if variant === 'A'}
						<div class="table-scroll">
							<table class="device-table">
								<thead
									><tr
										><th>Device</th><th>Functions</th><th>Criteria</th><th>Examples</th><th
											><span class="sr-only">Open</span></th
										></tr
									></thead
								><tbody>
									{#each visibleDevices as device (device.name)}{@const stats = deviceStats(
											device.ids
										)}<tr
											><td
												><button class="device-name" onclick={() => selectDevice(device.name)}
													><CircuitBoard size={20} /><span
														><strong>{device.name}</strong><small>{device.description}</small></span
													></button
												></td
											><td>{device.ids.length}</td><td>{stats.criteria}</td><td>{stats.examples}</td
											><td
												><Button
													variant="ghost"
													size="icon"
													aria-label={`Open ${device.name}`}
													title={`Open ${device.name}`}
													onclick={() => selectDevice(device.name)}
													><ChevronRight size={18} /></Button
												></td
											></tr
										>{/each}
								</tbody>
							</table>
						</div>
					{:else if variant === 'B'}
						<div class="device-grid">
							{#each visibleDevices as device (device.name)}{@const stats = deviceStats(
									device.ids
								)}<button class="device-tile" onclick={() => selectDevice(device.name)}
									><div class="tile-icon">
										{#if device.category === 'Wireless' || device.category === 'Reference'}<Radio
												size={28}
											/>{:else if device.category === 'Control'}<Terminal size={28} />{:else}<Cable
												size={28}
											/>{/if}<ChevronRight size={18} />
									</div>
									<span class="eyebrow">{device.category}</span>
									<h2>{device.name}</h2>
									<p>{device.description}</p>
									<div class="tile-stats">
										<span><strong>{device.ids.length}</strong> Functions</span><span
											><strong>{stats.criteria}</strong> Criteria</span
										><span><strong>{stats.examples}</strong> Examples</span>
									</div></button
								>{/each}
						</div>
					{:else}
						<div class="directory-list">
							{#each visibleDevices as device (device.name)}<button
									class="directory-item"
									onclick={() => selectDevice(device.name)}
									><CircuitBoard size={28} />
									<div>
										<span class="eyebrow">{device.category}</span>
										<h2>{device.name}</h2>
										<p>{device.description}</p>
										<div class="function-tags">
											{#each functions.filter( (fn) => device.ids.includes(fn.id) ) as fn (fn.id)}<span
													>{fn.name}</span
												>{:else}<span>No linked functions</span>{/each}
										</div>
									</div>
									<ChevronRight size={20} /></button
								>{/each}
						</div>
					{/if}
					{#if !visibleDevices.length}<div class="empty" role="status">
							No devices match "{deviceSearch}".
						</div>{/if}
				</section>
			{:else}
				<nav class="breadcrumbs" aria-label="Breadcrumb">
					<button onclick={() => selectDevice('')}>Devices</button><ChevronRight size={14} /><span
						>{activeDevice.name}</span
					>{#if functionFocused && activeFunction}<ChevronRight size={14} /><span
							>{activeFunction.name}</span
						>{/if}
				</nav>
				<header class="device-heading">
					<div class="device-identity">
						<div class="device-symbol"><CircuitBoard size={28} /></div>
						<div>
							<div class="eyebrow">{activeDevice.category}</div>
							<h1>{activeDevice.name}</h1>
							<p>{activeDevice.description}</p>
						</div>
					</div>
					<div class="device-select">Device<Select ariaLabel="Device" value={selectedDevice} onchange={selectDevice} options={devices.map((device) => ({ value: device.name, label: device.name }))} /></div>
				</header>
				{#if !deviceFunctions.length}<section class="empty-device">
						<CircuitBoard size={36} />
						<h2>No linked functions</h2>
						<p>{activeDevice.name}</p>
						<Button variant="outline" onclick={() => selectDevice('')}
							><ArrowLeft size={16} /> All devices</Button
						>
					</section>
				{:else if variant === 'A'}
					<div class="split-layout">
						{@render deviceRail()}
						<aside class="function-rail" aria-label="Device functions">
							<div class="rail-heading">
								<h2>Functions</h2>
								<span>{deviceFunctions.length}</span>
							</div>
							{@render functionSearchBox()}{#each visibleFunctions as fn (fn.id)}<button
									class="function-choice"
									class:chosen={selectedFunction === fn.id}
									aria-pressed={selectedFunction === fn.id}
									onclick={() => selectFunction(fn.id)}
									><span class="eyebrow">{fn.source} / {fn.channel}</span><strong>{fn.name}</strong
									><small>{fn.category}</small>
									<div>
										<span>{fn.parameters.length} parameters</span><span
											>{fn.examples.length} examples</span
										>
									</div></button
								>{:else}<p class="empty" role="status">No matching functions.</p>{/each}
						</aside>
						<div class="detail-pane">
							{#if activeFunction}{@render functionContent(activeFunction)}{/if}
						</div>
					</div>
				{:else if variant === 'B'}
					<div class="hub-stats">
						<span><strong>{deviceFunctions.length}</strong> Functions</span><span
							><strong>{deviceStats(activeDevice.ids).criteria}</strong> Criteria</span
						><span><strong>{deviceStats(activeDevice.ids).examples}</strong> Device examples</span>
					</div>
					{@render deviceSections()}
					{#if section !== 'functions'}{@render aggregateSection()}
					{:else if functionFocused && activeFunction}<div class="hub-detail">
							<Button
								variant="ghost"
								size="sm"
								onclick={() => {
									functionFocused = false;
								}}><ArrowLeft size={16} /> All functions</Button
							>{@render functionContent(activeFunction)}
						</div>
					{:else}<div class="hub-list">
							<div class="list-toolbar">
								<h2>Functions</h2>
								{@render functionSearchBox()}
							</div>
							<div class="table-scroll">
								<table class="function-table">
									<thead
										><tr
											><th>Function</th><th>Interface</th><th>Parameters</th><th>Criteria</th><th
												>Examples</th
											></tr
										></thead
									><tbody
										>{#each visibleFunctions as fn (fn.id)}<tr
												><td
													><button class="function-link" onclick={() => selectFunction(fn.id)}
														><strong>{fn.name}</strong><small>{fn.description}</small></button
													></td
												><td>{fn.channel}</td><td>{fn.parameters.length}</td><td
													>{fn.criteria.length}</td
												><td>{fn.examples.length}</td></tr
											>{/each}</tbody
									>
								</table>
							</div>
							{#if !visibleFunctions.length}<p class="empty" role="status">
									No matching functions.
								</p>{/if}
						</div>{/if}
				{:else}
					<div class="tree-layout">
						{@render deviceRail()}
						<div class="tree-content">
							<div class="list-toolbar">
								<h2>Functions <span>{deviceFunctions.length}</span></h2>
								{@render functionSearchBox()}
							</div>
							{#each visibleFunctions as fn (fn.id)}<section class="function-branch">
									<button
										class="branch-toggle"
										aria-expanded={selectedFunction === fn.id}
										onclick={() => {
											selectedFunction = selectedFunction === fn.id ? '' : fn.id;
											functionFocused = selectedFunction !== '';
										}}
										><ChevronDown
											size={18}
											class={selectedFunction === fn.id ? '' : 'collapsed-chevron'}
										/>
										<div><strong>{fn.name}</strong><small>{fn.category}</small></div>
										<span>{fn.examples.length} examples</span></button
									>{#if selectedFunction === fn.id}{@render functionContent(fn)}{/if}
								</section>{:else}<p class="empty" role="status">No matching functions.</p>{/each}
						</div>
					</div>
				{/if}
			{/if}
			<p class="copy-status" role="status">{copyStatus}</p>
			<output class="context-state" aria-label="Prototype state"
				>Layout {variant} | Device: {selectedDevice || 'All'} | Function: {activeFunction?.name ??
					'None'} | Section: {variant === 'B' ? section : 'functions'}</output
			>
		</div>
		{/if}
		{/if}
		<DevicePrototypeSwitcher current={variant} onChange={changeVariant} />
	</div>
{:else}
	<p>This prototype is available only in development.</p>
{/if}

<style>
	.variant-d .app-header { background: #191919; color: #fff; border-color: #191919; }
	.variant-d .brand :global(svg) { color: #fff; }
	.variant-d .mock-label { background: #383838; color: #ededed; }
	.desk { max-width: 1440px; margin: auto; padding: 40px 40px 110px; color: #181818; }
	.desk h1, .desk h2, .desk p { margin: 0; }
	.desk-heading, .desk-context, .desk-section-heading { display: flex; justify-content: space-between; align-items: end; gap: 24px; }
	.desk-heading h1 { margin-top: 10px; font-size: 32px; }
	.desk-heading p, .desk-context p, .desk-section-heading p { margin-top: 6px; color: #666; font-size: 14px; }
	.desk-kicker { color: #666; font-size: 11px; font-weight: 700; }
	.desk-count { font-size: 11px; color: #666; white-space: nowrap; }
	.desk-device-bar { display: flex; gap: 0; margin-top: 32px; overflow-x: auto; border-block: 1px solid #d6d6d6; }
	.desk-device-bar button { display: flex; align-items: center; gap: 16px; flex: 0 0 auto; padding: 16px 18px; border-bottom: 3px solid transparent; font-size: 13px; font-weight: 700; }
	.desk-device-bar button.active { border-bottom-color: #171717; background: #f2f2f2; }
	.desk-device-bar small { color: #777; font-weight: 400; }
	.desk-context { padding: 30px 0; border-bottom: 1px solid #dedede; }
	.desk-context h2 { margin-top: 8px; font-size: 24px; }
	.desk-metrics { display: flex; gap: 30px; color: #666; font-size: 12px; }
	.desk-metrics strong { display: block; color: #181818; font-size: 20px; }
	.desk-section { padding-top: 32px; }
	.desk-section-heading { margin-bottom: 20px; }
	.desk-section-heading h2 { margin-top: 6px; font-size: 20px; }
	.desk-search { display: flex; align-items: center; gap: 8px; width: 240px; padding: 9px 10px; border: 1px solid #ccc; }
	.desk-search input { width: 100%; min-width: 0; outline: none; font-size: 13px; }
	.desk-table { width: 100%; min-width: 650px; border-collapse: collapse; text-align: left; font-size: 13px; }
	.desk-table th { padding: 12px; border-bottom: 1px solid #aaa; color: #666; font-size: 11px; font-weight: 600; }
	.desk-table td { padding: 14px 12px; border-bottom: 1px solid #e6e6e6; }
	.desk-table tr.selected { background: #f1f1f1; }
	.desk-table button { text-align: left; }
	.desk-table button strong, .desk-table button small { display: block; }
	.desk-table button small { color: #666; margin-top: 4px; }
	.desk-inspector { margin-top: 36px; border-top: 3px solid #202020; }
	.desk-badge { padding: 5px 8px; background: #e7f2eb; color: #27623d; font-size: 11px; white-space: nowrap; }
	.desk-tabs { display: flex; gap: 24px; overflow-x: auto; border-bottom: 1px solid #ddd; margin-bottom: 14px; }
	.desk-tabs button { padding: 12px 0; border-bottom: 2px solid transparent; white-space: nowrap; color: #666; font-size: 13px; }
	.desk-tabs button.active { border-bottom-color: #181818; color: #181818; font-weight: 700; }
	.desk-welcome { padding: 48px 0; }
	.desk-welcome h2 { margin-top: 8px; font-size: 22px; }
	.desk-registry { margin-top: 28px; border-top: 1px solid #ccc; }
	.desk-registry button { display: flex; align-items: center; justify-content: space-between; gap: 16px; width: 100%; padding: 16px 4px; border-bottom: 1px solid #e3e3e3; text-align: left; }
	.desk-registry button > span:last-child { display: flex; align-items: center; gap: 12px; color: #666; font-size: 12px; white-space: nowrap; }
	.desk-registry strong, .desk-registry small { display: block; }
	.desk-registry small { margin-top: 4px; color: #666; font-size: 12px; }
	@media (max-width: 700px) { .desk { padding: 28px 16px 110px; } .desk-heading, .desk-context, .desk-section-heading { align-items: flex-start; flex-direction: column; } .desk-search { width: 100%; } .desk-metrics { gap: 20px; } .desk-heading h1 { font-size: 26px; } }
	.prototype {
		position: fixed;
		inset: 0;
		z-index: 40;
		overflow: auto;
		background: var(--background);
		color: var(--foreground);
		font-family: 'Segoe UI', sans-serif;
		--device-accent: #087f70;
		--device-soft: #e8f5ef;
	}
	.prototype :global(*) {
		box-sizing: border-box;
	}
	.prototype :global(button) {
		cursor: pointer;
	}
	.prototype :global(button:focus-visible),
	.prototype :global(select:focus-visible) {
		outline: 2px solid var(--ring);
		outline-offset: 3px;
	}
	.prototype :global(.collapsed-chevron) {
		transform: rotate(-90deg);
	}
	:global(.dark) .prototype {
		--device-accent: #77d9bc;
		--device-soft: #183b31;
	}
	.prototype.variant-d {
		--background: #fff; --foreground: #181818; --card: #fff; --border: #dedede;
		--muted-foreground: #606060; --secondary: #f0f0f0; --ring: #181818;
		--device-accent: #181818; --device-soft: #ededed;
	}
	.app-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		min-height: 64px;
		padding: 12px 32px;
		border-bottom: 1px solid var(--border);
		background: var(--card);
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 12px;
		font-weight: 650;
		text-align: left;
	}
	.brand :global(svg) {
		color: var(--device-accent);
		flex-shrink: 0;
	}
	.mock-label {
		flex-shrink: 0;
		padding: 4px 8px;
		font-size: 12px;
		background: var(--secondary);
		color: var(--muted-foreground);
		border-radius: 4px;
	}
	.workspace {
		max-width: 1560px;
		padding: 28px 32px 112px;
		margin: 0 auto;
	}
	h1,
	h2,
	h3,
	h4,
	p {
		margin: 0;
	}
	h1 {
		font-size: 28px;
		font-weight: 650;
		line-height: 1.3;
		overflow-wrap: anywhere;
	}
	h1 span,
	h2 span,
	h3 span {
		font-size: 13px;
		font-weight: 400;
		color: var(--muted-foreground);
	}
	h2 {
		font-size: 17px;
		font-weight: 650;
		overflow-wrap: anywhere;
	}
	h3 {
		font-size: 14px;
		font-weight: 650;
	}
	p {
		color: var(--muted-foreground);
		line-height: 1.55;
	}
	.eyebrow {
		font-size: 11px;
		font-weight: 600;
		color: var(--muted-foreground);
		margin-bottom: 6px;
	}
	.directory-heading {
		display: flex;
		justify-content: space-between;
		align-items: end;
		gap: 24px;
		margin: 8px 0 28px;
	}
	.search-box {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
		width: 290px;
		max-width: 100%;
	}
	.search-box :global(svg) {
		flex: 0 0 16px;
		color: var(--muted-foreground);
	}
	.search-box :global(input) {
		min-width: 0;
		background: var(--card);
	}
	.table-scroll {
		overflow-x: auto;
		max-width: 100%;
	}
	table {
		width: 100%;
		text-align: left;
		border-collapse: collapse;
		background: var(--card);
	}
	th {
		position: relative;
		padding: 12px 16px;
		font-size: 12px;
		color: var(--muted-foreground);
		font-weight: 500;
		background: var(--secondary);
		border-block: 1px solid var(--border);
		white-space: nowrap;
	}
	td {
		padding: 16px;
		border-bottom: 1px solid var(--border);
		font-size: 13px;
	}
	.device-table td:not(:first-child) {
		font-variant-numeric: tabular-nums;
	}
	.device-name {
		display: flex;
		align-items: center;
		gap: 16px;
		text-align: left;
	}
	.device-name :global(svg) {
		color: var(--device-accent);
		flex-shrink: 0;
	}
	.device-name strong,
	.device-name small {
		display: block;
	}
	.device-name strong {
		font-size: 15px;
		font-family: var(--mono);
	}
	.device-name small {
		color: var(--muted-foreground);
		font-size: 13px;
		margin-top: 5px;
	}
	.device-name:hover strong,
	.function-link:hover strong {
		color: var(--device-accent);
		text-decoration: underline;
	}
	.device-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 20px;
	}
	.device-tile {
		display: flex;
		flex-direction: column;
		min-width: 0;
		padding: 24px;
		min-height: 245px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--card);
		text-align: left;
	}
	.device-tile:hover {
		border-color: var(--device-accent);
		background: var(--device-soft);
	}
	.tile-icon {
		display: flex;
		align-items: center;
		justify-content: space-between;
		color: var(--device-accent);
		width: 100%;
		margin-bottom: 22px;
	}
	.device-tile h2 {
		font-family: var(--mono);
		font-size: 20px;
		margin-bottom: 8px;
	}
	.device-tile p {
		font-size: 13px;
		margin-bottom: 24px;
	}
	.tile-stats {
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
		margin-top: auto;
		font-size: 11px;
		color: var(--muted-foreground);
	}
	.tile-stats strong {
		display: block;
		font-size: 20px;
		font-weight: 500;
		font-variant-numeric: tabular-nums;
		color: var(--foreground);
	}
	.directory-list {
		border-top: 1px solid var(--border);
	}
	.directory-item {
		display: flex;
		gap: 24px;
		align-items: center;
		width: 100%;
		padding: 24px 16px;
		border-bottom: 1px solid var(--border);
		text-align: left;
	}
	.directory-item > div {
		flex: 1;
		min-width: 0;
	}
	.directory-item :global(svg) {
		flex-shrink: 0;
		color: var(--device-accent);
	}
	.directory-item:hover {
		background: var(--card);
	}
	.directory-item h2 {
		font-family: var(--mono);
		font-size: 22px;
	}
	.directory-item p {
		font-size: 13px;
		margin-top: 5px;
	}
	.function-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: 12px;
	}
	.function-tags span {
		font-family: var(--mono);
		font-size: 11px;
		padding: 4px 6px;
		background: var(--secondary);
		overflow-wrap: anywhere;
	}
	.breadcrumbs {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
		font-size: 12px;
		color: var(--muted-foreground);
		margin-bottom: 24px;
		overflow-wrap: anywhere;
	}
	.breadcrumbs button {
		color: var(--device-accent);
	}
	.breadcrumbs button:hover {
		text-decoration: underline;
	}
	.breadcrumbs :global(svg) {
		flex-shrink: 0;
	}
	.device-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		margin-bottom: 28px;
	}
	.device-identity {
		display: flex;
		gap: 16px;
		align-items: center;
		min-width: 0;
	}
	.device-identity > div:last-child {
		min-width: 0;
	}
	.device-symbol {
		display: grid;
		place-items: center;
		width: 60px;
		height: 60px;
		flex-shrink: 0;
		background: var(--device-soft);
		color: var(--device-accent);
		border-radius: 6px;
	}
	.device-heading h1 {
		font-family: var(--mono);
	}
	.device-heading p {
		font-size: 13px;
		margin-top: 5px;
	}
	.device-select {
		font-size: 11px;
		color: var(--muted-foreground);
		display: grid;
		gap: 6px;
	}
	.split-layout {
		display: grid;
		grid-template-columns: 180px 260px minmax(0, 1fr);
		background: var(--card);
		border-block: 1px solid var(--border);
		min-height: 600px;
	}
	.device-rail {
		padding: 18px 12px;
		border-right: 1px solid var(--border);
		min-width: 0;
	}
	.device-rail h2 {
		font-size: 12px;
		color: var(--muted-foreground);
		margin: 0 8px 18px;
	}
	.rail-device {
		display: flex;
		gap: 10px;
		align-items: center;
		width: 100%;
		padding: 12px 8px;
		border-radius: 4px;
		text-align: left;
		margin-bottom: 4px;
	}
	.rail-device span {
		flex: 1;
		min-width: 0;
		font-size: 12px;
		font-family: var(--mono);
		overflow-wrap: anywhere;
	}
	.rail-device :global(svg) {
		flex-shrink: 0;
	}
	.rail-device small {
		font-size: 11px;
		color: var(--muted-foreground);
	}
	.rail-device.chosen {
		color: var(--device-accent);
		background: var(--device-soft);
	}
	.rail-device:hover {
		background: var(--secondary);
	}
	.function-rail {
		border-right: 1px solid var(--border);
		min-width: 0;
	}
	.device-rail,
	.function-rail {
		position: sticky;
		top: 16px;
		align-self: start;
		max-height: min(640px, calc(100dvh - 32px));
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
	}
	.rail-heading {
		display: flex;
		justify-content: space-between;
		padding: 20px 16px 12px;
	}
	.rail-heading h2 {
		font-size: 14px;
	}
	.rail-heading span {
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.function-rail .search-box {
		padding: 0 12px 16px;
		width: 100%;
	}
	.function-choice {
		display: block;
		width: 100%;
		text-align: left;
		padding: 18px 16px;
		border-top: 1px solid var(--border);
		border-left: 3px solid transparent;
	}
	.function-choice.chosen {
		background: var(--device-soft);
		border-left-color: var(--device-accent);
	}
	.function-choice:hover {
		background: var(--secondary);
	}
	.function-choice strong {
		display: block;
		font-family: var(--mono);
		font-size: 12px;
		overflow-wrap: anywhere;
		line-height: 1.6;
	}
	.function-choice small {
		display: block;
		color: var(--muted-foreground);
		font-size: 12px;
		margin-top: 4px;
	}
	.function-choice > div {
		display: flex;
		gap: 14px;
		margin-top: 12px;
		font-size: 11px;
		color: var(--muted-foreground);
	}
	.detail-pane {
		min-width: 0;
		padding: 26px;
	}
	.function-heading h2 {
		font-family: var(--mono);
		font-size: 19px;
	}
	.function-heading p {
		margin-top: 12px;
		font-size: 13px;
		max-width: 75ch;
	}
	.category-label {
		display: inline-block;
		margin-top: 12px;
		color: var(--device-accent);
		font-size: 11px;
	}
	.detail-section {
		margin-top: 28px;
	}
	.detail-section > h3 {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 12px;
	}
	.detail-section > h3 span {
		margin-left: auto;
	}
	.group-heading {
		margin: 14px 0 8px;
		font-size: 12px;
		font-weight: 500;
	}
	.group-heading > span:not(.scope-badge) {
		color: var(--muted-foreground);
		margin-left: 6px;
	}
	.scope-badge {
		display: inline-block;
		padding: 3px 7px;
		font-size: 10px;
		font-weight: 600;
		border-radius: 4px;
		background: var(--device-soft);
		color: var(--device-accent);
	}
	.scope-badge.shared {
		background: var(--secondary);
		color: var(--muted-foreground);
	}
	.definition-table th {
		padding: 9px 10px;
		font-size: 11px;
	}
	.definition-table td {
		padding: 10px;
		font-size: 12px;
		vertical-align: top;
	}
	.definition-table td:first-child {
		min-width: 130px;
	}
	.definition-table code {
		overflow-wrap: anywhere;
	}
	.examples-list {
		display: grid;
		gap: 12px;
	}
	.example {
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 12px 14px;
		background: var(--card);
		min-width: 0;
	}
	.example-title {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		align-items: center;
	}
	.example-title strong {
		font-size: 13px;
	}
	.example p {
		font-size: 12px;
		margin: 4px 0 12px;
	}
	.example pre {
		font-family: var(--mono);
		font-size: 11px;
		line-height: 1.8;
		margin: 0;
		background: var(--secondary);
		padding: 12px;
		border-radius: 4px;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		word-break: break-word;
		user-select: text;
	}
	.hub-stats {
		display: flex;
		gap: 36px;
		margin: 0 0 28px;
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.hub-stats strong {
		font-size: 20px;
		color: var(--foreground);
		font-weight: 500;
		margin-right: 6px;
		font-variant-numeric: tabular-nums;
	}
	.section-nav {
		display: flex;
		gap: 24px;
		border-bottom: 1px solid var(--border);
	}
	.section-nav button {
		padding: 12px 0;
		font-size: 14px;
		color: var(--muted-foreground);
		border-bottom: 2px solid transparent;
	}
	.section-nav button.active {
		color: var(--device-accent);
		border-color: var(--device-accent);
	}
	.list-toolbar {
		display: flex;
		justify-content: space-between;
		gap: 20px;
		align-items: center;
		margin-bottom: 20px;
	}
	.hub-list,
	.hub-detail,
	.aggregate {
		padding-top: 24px;
	}
	.function-link {
		display: block;
		text-align: left;
	}
	.function-link strong {
		display: block;
		font-family: var(--mono);
		font-size: 13px;
		overflow-wrap: anywhere;
	}
	.function-link small {
		display: block;
		font-size: 12px;
		color: var(--muted-foreground);
		margin-top: 6px;
		max-width: 55ch;
	}
	.hub-detail {
		max-width: 1050px;
	}
	.hub-detail .function-detail {
		margin-top: 20px;
	}
	.aggregate-function {
		padding-bottom: 32px;
		margin-bottom: 28px;
		border-bottom: 1px solid var(--border);
	}
	.aggregate-heading {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		margin-bottom: 16px;
	}
	.aggregate-heading h2 {
		font-family: var(--mono);
		font-size: 15px;
	}
	.aggregate-function h3 {
		margin: 20px 0 8px;
	}
	.tree-layout {
		display: grid;
		grid-template-columns: 210px minmax(0, 1fr);
		border-top: 1px solid var(--border);
	}
	.tree-content {
		padding: 24px 0 24px 28px;
		min-width: 0;
	}
	.function-branch {
		border-block-end: 1px solid var(--border);
		padding: 8px 0;
	}
	.branch-toggle {
		display: flex;
		align-items: center;
		gap: 14px;
		width: 100%;
		text-align: left;
		padding: 16px 4px;
	}
	.branch-toggle > div {
		flex: 1;
		min-width: 0;
	}
	.branch-toggle strong {
		font-family: var(--mono);
		font-size: 14px;
		overflow-wrap: anywhere;
	}
	.branch-toggle small {
		display: block;
		font-size: 12px;
		color: var(--muted-foreground);
		margin-top: 5px;
	}
	.branch-toggle > span {
		font-size: 12px;
		color: var(--muted-foreground);
		white-space: nowrap;
	}
	.branch-toggle :global(svg) {
		flex-shrink: 0;
		color: var(--device-accent);
	}
	.function-branch .function-detail {
		padding: 12px 12px 24px 36px;
	}
	.function-branch .function-heading {
		display: none;
	}
	.empty {
		padding: 24px 12px;
		color: var(--muted-foreground);
		font-size: 13px;
	}
	.empty-device {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
		padding: 64px 12px;
		border-block: 1px solid var(--border);
	}
	.empty-device :global(svg) {
		color: var(--muted-foreground);
	}
	.copy-status {
		margin-top: 12px;
		min-height: 22px;
		font-size: 12px;
	}
	.context-state {
		display: block;
		padding-top: 16px;
		margin-top: 16px;
		font-size: 11px;
		color: var(--muted-foreground);
		border-top: 1px solid var(--border);
		overflow-wrap: anywhere;
	}
	@media (min-width: 1100px) {
		.variant-c .directory {
			max-width: 1100px;
			margin: 0 auto;
		}
	}
	@media (max-width: 1100px) {
		.split-layout {
			grid-template-columns: 220px minmax(0, 1fr);
		}
		.split-layout > .device-rail {
			display: none;
		}
		.device-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
	@media (max-width: 700px) {
		.app-header {
			padding: 12px 16px;
			flex-wrap: wrap;
		}
		.brand {
			font-size: 13px;
			gap: 8px;
		}
		.mock-label {
			font-size: 10px;
		}
		.workspace {
			padding: 20px 16px 110px;
		}
		h1 {
			font-size: 24px;
		}
		.directory-heading {
			flex-direction: column;
			align-items: stretch;
			gap: 18px;
		}
		.search-box {
			width: 100%;
		}
		.device-grid {
			grid-template-columns: minmax(0, 1fr);
		}
		.device-tile {
			min-height: 215px;
		}
		.device-name {
			gap: 10px;
		}
		.device-name small {
			min-width: 150px;
		}
		.device-table th,
		.device-table td {
			padding: 12px 10px;
		}
		.directory-item {
			gap: 14px;
			padding: 22px 4px;
		}
		.directory-item h2 {
			font-size: 19px;
		}
		.device-heading {
			flex-direction: column;
			align-items: stretch;
			gap: 18px;
		}
		.device-symbol {
			width: 48px;
			height: 48px;
		}
		.device-select {
			grid-template-columns: auto minmax(0, 1fr);
			align-items: center;
			gap: 12px;
		}
		.split-layout {
			display: block;
		}
		.function-rail {
			position: static;
			max-height: 360px;
			border-right: 0;
			border-bottom: 1px solid var(--border);
		}
		.function-choice {
			padding: 12px 16px;
		}
		.function-choice > div {
			display: none;
		}
		.detail-pane {
			padding: 20px 12px;
		}
		.function-heading h2 {
			font-size: 17px;
		}
		.hub-stats {
			gap: 14px;
			flex-wrap: wrap;
		}
		.hub-stats strong {
			font-size: 18px;
		}
		.list-toolbar {
			flex-direction: column;
			align-items: stretch;
			gap: 14px;
		}
		.tree-layout {
			display: block;
		}
		.tree-layout > .device-rail {
			display: none;
		}
		.tree-content {
			padding: 20px 0;
		}
		.branch-toggle {
			gap: 10px;
			flex-wrap: wrap;
		}
		.branch-toggle > span {
			margin-left: 28px;
		}
		.function-branch .function-detail {
			padding: 4px 0 24px;
		}
		.aggregate-heading {
			align-items: start;
			flex-wrap: wrap;
		}
	}
</style>
