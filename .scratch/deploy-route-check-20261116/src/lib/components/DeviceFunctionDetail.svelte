<script lang="ts">
	import { copyText } from '#lib/clipboard.ts';
	import { resolve } from '$app/paths';
	import {
		Copy,
		Check,
		ExternalLink,
		RefreshCw,
		SlidersHorizontal,
		ListChecks,
		Code,
		Pencil,
		Trash2,
		Plus
	} from '@lucide/svelte';
	import { listLinks, type LinkRow } from '#lib/links.ts';
	import {
		listDeviceDefinitions,
		type DeviceFunctionRow,
		type DeviceDefinitionRow,
		type DeviceDefinitionKind,
		type DeviceDefinitionsResult
	} from '#lib/devices.ts';
	import {
		deleteExample,
		listExamples,
		type ExampleListResult,
		type ExampleRow
	} from '#lib/examples.ts';
	import Button from '#lib/components/ui/button.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import ExampleFieldUsage from '#lib/components/ExampleFieldUsage.svelte';
	import ExampleDialog from '#lib/components/ExampleDialog.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import FunctionDialog from '#lib/components/FunctionDialog.svelte';
	import ComposeDialog from '#lib/components/ComposeDialog.svelte';
	import { toast } from 'svelte-sonner';

	let {
		fn,
		device,
		onChanged
	}: {
		fn: DeviceFunctionRow;
		device: string;
		onChanged?: (deletedId?: number) => void | Promise<void>;
	} = $props();
	let dialogMode = $state<'edit' | 'delete' | null>(null);
	let composeOpen = $state(false);
	let examplePage = $state(1);
	let globalPage = $state(1);
	let examplePageSize = $state(20);
	let globalPageSize = $state(20);
	let retry = $state(0);
	let copied = $state<number | null>(null);
	let copyStatus = $state('');
	let exampleDialogOpen = $state(false);
	let exampleEditRow = $state<ExampleRow | null>(null);
	let exampleConfirmRow = $state<ExampleRow | null>(null);
	let deletingExample = $state(false);
	type DetailField = { link: LinkRow; definition: DeviceDefinitionRow | undefined };
	type ExamplePage = { page: number; page_size: number };

	async function definitions(
		kind: DeviceDefinitionKind,
		scope: string
	): Promise<DeviceDefinitionsResult> {
		const result = await listDeviceDefinitions(kind, scope, 1, 100);
		const items = [...result.items];
		for (let next = 2; items.length < result.total; next++) {
			const more = await listDeviceDefinitions(kind, scope, next, 100);
			if (!more.items.length) throw new Error('The definition list changed. Please retry.');
			items.push(...more.items);
		}
		return { ...result, items };
	}

	async function fields(
		kind: DeviceDefinitionKind,
		id: number,
		scope: string
	): Promise<{ scoped: DetailField[]; global: DetailField[] }> {
		const links = await listLinks(id, kind === 'parameters' ? 'parameter-links' : 'criteria-links');
		const scopedLinks = links.filter((link) => link.device_name === scope);
		const scopedNames = new Set(scopedLinks.map((link) => link.name));
		const relevant = scope
			? [
					...scopedLinks,
					...links.filter((link) => link.device_name === '' && !scopedNames.has(link.name))
				]
			: scopedLinks;
		if (!relevant.length) return { scoped: [], global: [] };
		const catalogs = await Promise.all([
			definitions(kind, scope),
			scope && relevant.some((link) => link.device_name === '')
				? definitions(kind, '')
				: Promise.resolve(null)
		]);
		const byId = new Map(
			catalogs
				.flatMap((catalog) => (catalog ? [...catalog.items, ...catalog.global_items] : []))
				.map((definition) => [definition.id, definition])
		);
		const resolved = relevant.map((link) => ({
			link,
			definition:
				link.resolved && link.definition_id !== null ? byId.get(link.definition_id) : undefined
		}));
		const isShared = (field: DetailField) =>
			(field.definition?.device_name ?? field.link.device_name) === '';
		return {
			scoped: resolved.filter((field) => !isShared(field)),
			global: resolved.filter(isShared)
		};
	}

	async function load({
		fn,
		scope,
		examples: deviceExamples,
		sharedExamples,
		attempt
	}: {
		fn: DeviceFunctionRow;
		scope: string;
		examples: ExamplePage;
		sharedExamples: ExamplePage;
		attempt: number;
	}) {
		void attempt;
		const filters = { function_source: fn.source, function_name: fn.name };
		const [parameters, criteria, examples, globalExamples] = await Promise.all([
			fields('parameters', fn.id, scope),
			fields('criteria', fn.id, scope),
			listExamples({ ...filters, device: scope, ...deviceExamples }),
			scope ? listExamples({ ...filters, device: '', ...sharedExamples }) : Promise.resolve(null)
		]);
		return { parameters, criteria, examples, globalExamples };
	}

	let content = $derived(
		load({
			fn,
			scope: device,
			examples: { page: examplePage, page_size: examplePageSize },
			sharedExamples: { page: globalPage, page_size: globalPageSize },
			attempt: retry
		})
	);

	async function copy(script: string, id: number) {
		const success = await copyText(script);
		copied = success ? id : null;
		copyStatus = success ? 'Script copied.' : 'Copy failed. The script text is selectable.';
	}

	function openCreateExample() {
		exampleEditRow = null;
		exampleDialogOpen = true;
	}

	function openEditExample(example: ExampleRow) {
		exampleEditRow = example;
		exampleDialogOpen = true;
	}

	async function deleteSelectedExample() {
		if (!exampleConfirmRow || deletingExample) return;
		deletingExample = true;
		try {
			await deleteExample(exampleConfirmRow.id);
			toast.success('Example deleted');
			exampleConfirmRow = null;
			retry++;
			await onChanged?.();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : String(error));
		} finally {
			deletingExample = false;
		}
	}
</script>

<section class="function-detail" aria-label="Function details">
	<header>
		<h2>{fn.name}</h2>
		{#if fn.description}<p>{fn.description}</p>{/if}
		<div class="function-meta">
			<span>{fn.category || 'Uncategorized'}</span><a
				href={resolve('/functions/[id]', { id: String(fn.id) })}
				><ExternalLink size={14} /> Manage function</a
			>
		</div>
		<div class="function-actions">
			<Button variant="outline" onclick={() => (dialogMode = 'edit')}
				><Pencil size={16} /> Edit function</Button
			>
			<Button variant="outline" onclick={() => (dialogMode = 'delete')}
				><Trash2 size={16} /> Delete function</Button
			>
			<Button onclick={() => (composeOpen = true)}><Code size={16} /> Compose script</Button>
		</div>
	</header>
	{#await content}
		<p role="status" class="banner-empty">Loading function details...</p>
	{:then data}
		{@const parameters = data.parameters}
		{@const criteria = data.criteria}
		{@const parameterNames = [...parameters.scoped, ...parameters.global].map(
			(field) => field.link.name
		)}
		{@const criteriaNames = [...criteria.scoped, ...criteria.global].map(
			(field) => field.link.name
		)}
		<section class="detail-section" aria-label="Parameters">
			<h3><SlidersHorizontal size={16} /> Parameters</h3>
			{#if parameters.scoped.length}{@render fieldsTable(
					`${device} parameters`,
					parameters.scoped
				)}{/if}
			{#if parameters.global.length}{@render fieldsTable(
					'Global parameters',
					parameters.global,
					true
				)}{/if}
			{#if !parameters.scoped.length && !parameters.global.length}<p class="muted">
					No parameter links for this device.
				</p>{/if}
		</section>
		<section class="detail-section" aria-label="Criteria">
			<h3><ListChecks size={16} /> Criteria</h3>
			{#if criteria.scoped.length}{@render fieldsTable(`${device} criteria`, criteria.scoped)}{/if}
			{#if criteria.global.length}{@render fieldsTable(
					'Global criteria',
					criteria.global,
					true
				)}{/if}
			{#if !criteria.scoped.length && !criteria.global.length}<p class="muted">
					No criteria links for this device.
				</p>{/if}
		</section>
		<section class="detail-section" aria-label="Function examples">
			<div class="examples-section-heading">
				<h3><Code size={16} /> Examples</h3>
				<Button size="sm" variant="outline" onclick={openCreateExample}
					><Plus size={16} /> New example</Button
				>
			</div>
			{@render exampleGroup(device || 'Global', data.examples, parameterNames, criteriaNames)}
			{#if data.examples.total > examplePageSize}<Pagination
					bind:page={examplePage}
					total={data.examples.total}
					bind:pageSize={examplePageSize}
				/>{/if}
			{#if data.globalExamples}
				{@render exampleGroup(
					'Global references',
					data.globalExamples,
					parameterNames,
					criteriaNames
				)}
				{#if data.globalExamples.total > globalPageSize}<Pagination
						bind:page={globalPage}
						total={data.globalExamples.total}
						bind:pageSize={globalPageSize}
					/>{/if}
			{/if}
		</section>
	{:catch error}
		<div class="banner-error" role="alert">
			{error instanceof Error ? error.message : String(error)}<Button
				variant="outline"
				size="sm"
				onclick={() => retry++}><RefreshCw size={16} /> Retry details</Button
			>
		</div>
	{/await}
	<p role="status" class="copy-status">{copyStatus}</p>
</section>

{#if dialogMode}<FunctionDialog
		mode={dialogMode}
		{fn}
		onClose={() => (dialogMode = null)}
		onSaved={async (row) => {
			await onChanged?.(row === null ? fn.id : undefined);
			retry++;
		}}
	/>{/if}
{#if composeOpen}<ComposeDialog fnId={fn.id} onClose={() => (composeOpen = false)} />{/if}
<ExampleDialog
	bind:open={exampleDialogOpen}
	editRow={exampleEditRow}
	fixedFunction={{ source: fn.source, name: fn.name }}
	prefillDevice={exampleEditRow?.device_name ?? device}
	lockDevice={exampleEditRow?.device_name === ''}
	onsaved={async () => {
		retry++;
		await onChanged?.();
	}}
/>
<AlertDialog
	open={exampleConfirmRow !== null}
	onOpenChange={(open) => {
		if (!open) exampleConfirmRow = null;
	}}
	title="Delete example"
	description={exampleConfirmRow
		? `Delete this example of ${exampleConfirmRow.function_name} from ${exampleConfirmRow.device_name || 'Global'}? This cannot be undone.`
		: ''}
	confirmLabel="Delete"
	onConfirm={deleteSelectedExample}
/>

{#snippet fieldsTable(title: string, fields: DetailField[], shared = false)}
	<section class="field-group" aria-label={title}>
		<h4><span class="scope-badge" class:shared>{title}</span> <span>{fields.length}</span></h4>
		<div class="table-scroll">
			<table>
				<thead
					><tr
						><th>Name</th><th>Type</th><th>Default</th><th>Allowed values</th><th>Required</th></tr
					></thead
				><tbody>
					{#each fields as field (field.link.id)}<tr
							><td
								><code>{field.link.name}</code>{#if !field.definition}<span class="unresolved"
										>Unresolved</span
									>{/if}</td
							><td>{field.definition?.type || '-'} {field.definition?.unit || ''}</td><td
								><code>{field.definition?.default_value ?? '-'}</code></td
							><td>{field.definition?.allowed_values || '-'}</td><td
								>{field.link.required ? 'Yes' : 'No'}</td
							></tr
						>{/each}
				</tbody>
			</table>
		</div>
	</section>
{/snippet}

{#snippet exampleGroup(
	title: string,
	result: ExampleListResult,
	parameterNames: string[],
	criteriaNames: string[]
)}
	<section class="example-group" aria-label={`${title} examples`}>
		<h4>{title} <span>{result.total}</span></h4>
		{#each result.items as example (example.id)}<article class="example">
				<div class="example-heading">
					<p>{example.meaning || `${example.function_source} / ${example.function_name}`}</p>
					<div class="example-actions">
						<Button
							variant="ghost"
							size="icon"
							aria-label={`Copy example ${example.id}`}
							title="Copy script"
							onclick={() => copy(example.script, example.id)}
							>{#if copied === example.id}<Check size={16} />{:else}<Copy size={16} />{/if}</Button
						>
						<Button
							variant="ghost"
							size="icon"
							aria-label={`${example.device_name ? 'Edit' : 'Edit Global'} example ${example.id}`}
							title={example.device_name ? 'Edit example' : 'Edit Global example'}
							onclick={() => openEditExample(example)}><Pencil size={16} /></Button
						>
						<Button
							variant="ghost"
							size="icon"
							aria-label={`${example.device_name ? 'Delete' : 'Delete Global'} example ${example.id}`}
							title={example.device_name ? 'Delete example' : 'Delete Global example'}
							disabled={deletingExample}
							onclick={() => (exampleConfirmRow = example)}><Trash2 size={16} /></Button
						>
					</div>
				</div>
				<ExampleFieldUsage
					script={example.script}
					parameters={parameterNames}
					criteria={criteriaNames}
				/>
				<pre><code>{example.script}</code></pre>
			</article>{:else}<p class="muted">No examples in this group.</p>{/each}
	</section>
{/snippet}

<style>
	.function-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: 16px;
	}
	.function-detail {
		min-width: 0;
		padding: 26px;
	}
	h2 {
		font-family: var(--mono);
		font-size: 19px;
		font-weight: 650;
		overflow-wrap: anywhere;
		margin: 8px 0;
	}
	h3 {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 650;
		margin-bottom: 12px;
	}
	h4 {
		font-size: 12px;
		font-weight: 500;
		margin-bottom: 8px;
	}
	h4 span {
		color: var(--muted-foreground);
		font-size: 13px;
	}
	.scope-badge {
		display: inline-block;
		padding: 3px 7px;
		border-radius: 4px;
		background: var(--device-soft);
		color: var(--device-badge-foreground);
		font-size: 11px;
		font-weight: 600;
	}
	.scope-badge.shared {
		background: var(--secondary);
		color: var(--muted-foreground);
	}
	header > p {
		max-width: 100%;
		font-size: 13px;
		line-height: 1.55;
	}
	.function-meta {
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
		justify-content: space-between;
		margin-top: 16px;
		font-size: 13px;
	}
	.function-meta > span {
		color: var(--device-accent);
	}
	.function-meta a {
		display: flex;
		gap: 6px;
		align-items: center;
		color: var(--primary);
	}
	a:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 4px;
	}
	.detail-section {
		margin-top: 28px;
	}
	.examples-section-heading {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
	}
	.examples-section-heading h3 {
		min-width: 0;
	}
	.field-group,
	.example-group {
		margin: 16px 0;
		min-width: 0;
	}
	.table-scroll {
		overflow-x: auto;
	}
	table {
		width: 100%;
		min-width: 560px;
		border-collapse: collapse;
		font-size: 12px;
		background: var(--card);
	}
	th:first-child,
	td:first-child {
		min-width: 140px;
	}
	th,
	td {
		position: relative;
		text-align: left;
		padding: 10px;
		border-bottom: 1px solid var(--border);
		vertical-align: top;
	}
	th {
		padding: 9px 10px;
		font-size: 11px;
		background: var(--muted);
		color: var(--muted-foreground);
		font-weight: 500;
		white-space: nowrap;
	}
	code {
		font-family: var(--mono);
		overflow-wrap: anywhere;
	}
	.unresolved {
		display: block;
		font-size: 12px;
		color: var(--destructive);
	}
	.example {
		margin: 12px 0;
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 12px 14px;
		background: var(--card);
	}
	.example-heading {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		align-items: center;
		margin-bottom: 12px;
		font-size: 13px;
		font-weight: 600;
	}
	.example-heading p {
		overflow-wrap: anywhere;
		min-width: 0;
	}
	.example-heading :global(button) {
		flex-shrink: 0;
	}
	.example-actions {
		display: flex;
		align-items: center;
		gap: 2px;
		flex-shrink: 0;
	}
	pre {
		padding: 12px;
		background: var(--muted);
		border-radius: 4px;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		font-size: 12px;
		line-height: 1.8;
	}
	.copy-status {
		min-height: 20px;
		font-size: 13px;
		color: var(--muted-foreground);
	}
	@media (max-width: 700px) {
		.function-detail {
			padding: 16px 0;
		}
	}
</style>
