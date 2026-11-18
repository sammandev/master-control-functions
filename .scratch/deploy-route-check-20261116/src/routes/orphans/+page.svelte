<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import {
		listOrphans,
		ApiError,
		type OrphanReport,
		type OrphanKind,
		type OrphanLinkRow,
		type OrphanExampleRow
	} from '#lib/orphans.ts';
	import { createFunction, type FunctionInput } from '#lib/api.ts';
	import { createDefinition, type DefinitionKind, type DefinitionInput } from '#lib/definitions.ts';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Label from '#lib/components/ui/label.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir, type SortValue } from '#lib/table-sort.ts';
	import { SvelteSet } from 'svelte/reactivity';
	import { RefreshCw } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import {
		Table,
		TableHeader,
		TableBody,
		TableRow,
		TableHead,
		TableCell
	} from '#lib/components/ui/table/index.ts';

	// Orphan report page (ticket 05, ADR-0003): every unresolved reference
	// grouped by kind. Nothing is dropped: each row opens a create dialog
	// prefilled from that row, in the row's own Device or Global scope. A
	// successful repair reloads the report, because one repair can leave
	// related rows unresolved.
	type MissingTarget = 'function' | 'definition';

	let report = $state<OrphanReport | null>(null);
	let loading = $state(true);
	let loadError = $state('');
	let loadRevision = 0;

	// 'all' shows every group; a kind filters to that group.
	let filter = $state<'all' | OrphanKind>('all');

	// ---- repair dialog state -------------------------------------------------
	let resolveKind = $state<'function' | 'parameter' | 'criteria' | null>(null);
	// The affected usage the open dialog repairs, shown as context.
	let repairContext = $state<{ usage: string; identity: string; scope: string } | null>(null);
	// The create dialog prefills from the row being resolved.
	let functionForm = $state<FunctionInput | null>(null);
	let definitionForm = $state<DefinitionInput | null>(null);
	let fieldErrors = $state<Record<string, string>>({});
	let submitError = $state('');
	let saving = $state(false);

	const shownGroups = $derived(
		report ? report.groups.filter((g) => filter === 'all' || g.kind === filter) : []
	);

	// Distinct missing targets across every group, not just the filtered view.
	const missingSummary = $derived.by(() => {
		const functions = new SvelteSet<string>();
		const parameters = new SvelteSet<string>();
		const criteria = new SvelteSet<string>();
		for (const group of report?.groups ?? []) {
			for (const row of group.items) {
				for (const target of missingTargets(group.kind, row)) {
					if (target === 'function') {
						functions.add(`${row.function_source}/${row.function_name}`);
					} else if (isLink(row)) {
						const identity = `${row.name}@${row.device_name || 'Global'}`;
						(group.kind === 'criteria-links' ? criteria : parameters).add(identity);
					}
				}
			}
		}
		return { functions: functions.size, parameters: parameters.size, criteria: criteria.size };
	});

	// ---- sorting (ticket 09) ------------------------------------------------
	// The orphan report returns full groups (no pagination, no sort
	// params), so the shared sortRows helper orders the loaded rows. One
	// sort state serves every group table: a column that a group does not
	// render cannot be activated from it, and rows missing that value
	// (examples in a link column) keep their original order.
	const ORPHAN_SORT_KEYS: Record<string, (row: OrphanLinkRow | OrphanExampleRow) => SortValue> = {
		function: (r) => `${r.function_source}/${r.function_name}`,
		definition: (r) => (isLink(r) ? r.name : ''),
		device: (r) => r.device_name || '',
		script: (r) => (isLink(r) ? '' : (r as OrphanExampleRow).script)
	};

	let sortKey = $state('');
	let sortDir = $state<SortDir>('asc');

	function applySort(col: string) {
		({ key: sortKey, dir: sortDir } = nextSort(sortKey, sortDir, col));
	}

	function groupView(items: (OrphanLinkRow | OrphanExampleRow)[]) {
		return sortKey === '' ? items : sortRows(items, ORPHAN_SORT_KEYS[sortKey]!, sortDir);
	}

	onMount(() => {
		load();
	});

	async function load() {
		const revision = ++loadRevision;
		loading = true;
		loadError = '';
		try {
			const next = await listOrphans();
			if (revision === loadRevision) report = next;
		} catch (e) {
			// Keep the last loaded report under the banner: the banner says it
			// may be out of date, so stale rows are never presented as current.
			if (revision === loadRevision) loadError = e instanceof Error ? e.message : String(e);
		} finally {
			if (revision === loadRevision) loading = false;
		}
	}

	function kindLabel(kind: OrphanKind): string {
		switch (kind) {
			case 'parameter-links':
				return 'Parameter links';
			case 'criteria-links':
				return 'Criteria links';
			case 'examples':
				return 'Examples';
		}
	}

	function isLink(row: OrphanLinkRow | OrphanExampleRow): row is OrphanLinkRow {
		return (row as OrphanLinkRow).definition_exists !== undefined;
	}

	function missingLabel(groupKind: OrphanKind, target: MissingTarget): string {
		if (target === 'function') return 'Function';
		return groupKind === 'criteria-links' ? 'Criteria definition' : 'Parameter definition';
	}

	// The first missing target is the one the repair action creates next.
	function primaryTarget(
		groupKind: OrphanKind,
		row: OrphanLinkRow | OrphanExampleRow
	): MissingTarget {
		return missingTargets(groupKind, row)[0] ?? 'definition';
	}

	function repairLabel(groupKind: OrphanKind, row: OrphanLinkRow | OrphanExampleRow): string {
		return missingLabel(groupKind, primaryTarget(groupKind, row));
	}

	function groupDescription(kind: OrphanKind): string {
		switch (kind) {
			case 'parameter-links':
				return 'Parameter links whose Function or Parameter definition is missing.';
			case 'criteria-links':
				return 'Criteria links whose Function or Criteria definition is missing.';
			case 'examples':
				return 'Examples whose Function is missing.';
		}
	}

	// Every missing target a row waits on, in the order it must be created.
	// Examples only reference Functions; links can miss either side.
	function missingTargets(
		groupKind: OrphanKind,
		row: OrphanLinkRow | OrphanExampleRow
	): MissingTarget[] {
		if (groupKind === 'examples' || !isLink(row)) return ['function'];
		const targets: MissingTarget[] = [];
		if (!row.function_exists) targets.push('function');
		if (!row.definition_exists) targets.push('definition');
		return targets;
	}

	function usageLabel(groupKind: OrphanKind, row: OrphanLinkRow | OrphanExampleRow): string {
		if (!isLink(row)) return 'Example';
		return `${groupKind === 'criteria-links' ? 'Criteria' : 'Parameter'} link ${row.name}`;
	}

	// Open the prefilled create dialog for the row's first missing target. The
	// definition keeps the link's own Device or Global scope, so the repair
	// resolves that usage without widening a Device-scoped link to Global.
	function openResolve(groupKind: OrphanKind, row: OrphanLinkRow | OrphanExampleRow) {
		fieldErrors = {};
		submitError = '';
		repairContext = {
			usage: usageLabel(groupKind, row),
			identity: `${row.function_source}/${row.function_name}`,
			scope: row.device_name || 'Global'
		};
		if (primaryTarget(groupKind, row) === 'function') {
			// The missing target is the Function: source + name prefilled.
			resolveKind = 'function';
			definitionForm = null;
			functionForm = {
				source: row.function_source,
				interface: '',
				name: row.function_name,
				description: '',
				convention: null,
				category: '',
				execution_role: '',
				execution_phase: ''
			};
			return;
		}
		if (!isLink(row)) return;
		// The Function exists; the definition is what's missing.
		resolveKind = groupKind === 'criteria-links' ? 'criteria' : 'parameter';
		functionForm = null;
		definitionForm = {
			name: row.name,
			device_name: row.device_name,
			description: '',
			type: 'string',
			unit: '',
			default_value: '',
			allowed_values: '',
			cardinality: 'SINGLE',
			value_pattern: '',
			required: false
		};
	}

	function closeRepair() {
		resolveKind = null;
		repairContext = null;
	}

	async function resolveOrphans() {
		if (saving || (!functionForm && !definitionForm)) return;
		saving = true;
		fieldErrors = {};
		submitError = '';
		try {
			if (resolveKind === 'function' && functionForm) {
				await createFunction(functionForm);
				toast.success('Function created');
			} else if (definitionForm) {
				const kind: DefinitionKind =
					resolveKind === 'criteria' ? 'criteria-definitions' : 'parameter-definitions';
				await createDefinition(kind, definitionForm);
				toast.success(`${definitionForm.name} created`);
			}
		} catch (e) {
			// The rows stay listed and the dialog stays open for correction.
			fieldErrors = e instanceof ApiError ? e.fields : {};
			submitError = e instanceof Error ? e.message : String(e);
			return;
		} finally {
			saving = false;
		}
		closeRepair();
		await load();
	}
</script>

<h1 class="mt-2 mb-4 text-xl font-semibold">Orphan report</h1>

<p class="muted">
	Unresolved references imported from the workbook (or left behind by deletions). Nothing is
	dropped: create the missing target from a row to resolve it in place.
</p>

{#if loadError}
	<div class="banner-error" role="alert">
		{loadError}{report ? ' Showing the last loaded report, which may be out of date.' : ''}
		<Button variant="outline" onclick={() => load()}
			><RefreshCw size={16} /> Retry orphan report</Button
		>
	</div>
{/if}

{#if loading && !report}
	<p class="muted">Loading…</p>
{:else if report && report.total === 0}
	<div class="banner-empty">No unresolved references. The knowledge base is fully linked.</div>
{:else if report}
	<div class="filters" role="group" aria-label="Filter orphans by usage">
		{#each [{ kind: 'all' as const, count: report.total }, ...report.groups.map( (g) => ({ kind: g.kind, count: g.count }) )] as f (f.kind)}
			<Button
				variant={filter === f.kind ? 'default' : 'outline'}
				size="sm"
				aria-pressed={filter === f.kind}
				onclick={() => (filter = f.kind)}
			>
				{f.kind === 'all' ? 'All' : kindLabel(f.kind)} ({f.count})
			</Button>
		{/each}
	</div>
	<div class="catalog-toolbar">
		<span class="muted">
			{report.total} unresolved {report.total === 1 ? 'reference' : 'references'}
		</span>
		<span class="muted">Missing Functions: {missingSummary.functions}</span>
		<span class="muted">Missing Parameter definitions: {missingSummary.parameters}</span>
		<span class="muted">Missing Criteria definitions: {missingSummary.criteria}</span>
		{#if loading}<span class="muted" aria-live="polite">Refreshing…</span>{/if}
	</div>

	{#each shownGroups as group (group.kind)}
		{#if group.count > 0}
			<section class="orphan-group">
				<h2 class="group-title">
					{kindLabel(group.kind)} <span class="muted">({group.count})</span>
				</h2>
				<p class="muted">{groupDescription(group.kind)}</p>
				<div class="orphan-table">
					<Table class="border-0">
						<TableHeader>
							<TableRow>
								<TableHead>Missing</TableHead>
								<SortHead
									label="Function"
									col="function"
									active={sortKey}
									dir={sortDir}
									onclick={() => applySort('function')}
								/>
								{#if group.kind !== 'examples'}
									<SortHead
										label="Definition"
										col="definition"
										active={sortKey}
										dir={sortDir}
										onclick={() => applySort('definition')}
									/>
								{:else}
									<SortHead
										label="Script"
										col="script"
										active={sortKey}
										dir={sortDir}
										onclick={() => applySort('script')}
									/>
								{/if}
								<SortHead
									label="Device / Global"
									col="device"
									active={sortKey}
									dir={sortDir}
									onclick={() => applySort('device')}
								/>
								<TableHead>Repair</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each groupView(group.items) as row (row.id)}
								<TableRow>
									<TableCell>
										<div class="missing-list">
											{#each missingTargets(group.kind, row) as target (target)}
												<Badge variant="warn" class="max-w-full whitespace-normal">
													{missingLabel(group.kind, target)} missing
												</Badge>
											{/each}
										</div>
									</TableCell>
									<TableCell>
										<span class="identifier">
											{#if isLink(row) && row.function_id !== null}
												<!-- jump-to-row: straight to the function detail -->
												<a
													class="text-primary hover:underline"
													href={resolve('/functions/[id]', { id: String(row.function_id) })}
												>
													{row.function_source}/{row.function_name}
												</a>
											{:else}
												{row.function_source}/{row.function_name}
											{/if}
										</span>
									</TableCell>
									{#if isLink(row)}
										<TableCell><span class="identifier">{row.name}</span></TableCell>
									{:else}
										<TableCell><span class="script-scroll">{row.script}</span></TableCell>
									{/if}
									<TableCell>
										{#if row.device_name}
											<span class="device-name">{row.device_name}</span>
										{:else}
											<span class="muted" title="Shared across Devices">Global</span>
										{/if}
									</TableCell>
									<TableCell>
										<div class="repair-cell">
											<Button size="sm" onclick={() => openResolve(group.kind, row)}>
												Create missing {repairLabel(group.kind, row)}
											</Button>
											{#if missingTargets(group.kind, row).length > 1}
												<span class="muted"
													>Also needs a {missingLabel(group.kind, 'definition')}</span
												>
											{/if}
										</div>
									</TableCell>
								</TableRow>
							{/each}
						</TableBody>
					</Table>
				</div>
			</section>
		{/if}
	{/each}
{/if}

<Dialog
	open={resolveKind !== null}
	onOpenChange={(o) => {
		if (!o && !saving) closeRepair();
	}}
	title={resolveKind === 'function'
		? `Create function ${functionForm?.source}/${functionForm?.name}`
		: `Create ${resolveKind === 'criteria' ? 'criteria' : 'parameter'} definition "${definitionForm?.name}"`}
	description="Creating the missing target resolves the reference in place. Affected rows stay listed until the refreshed report confirms they are resolved."
>
	{#if repairContext}
		<dl class="repair-context">
			<div>
				<dt>Affected usage</dt>
				<dd>{repairContext.usage}</dd>
			</div>
			<div>
				<dt>Function</dt>
				<dd class="identifier">{repairContext.identity}</dd>
			</div>
			<div>
				<dt>Device / Global</dt>
				<dd class={repairContext.scope === 'Global' ? '' : 'identifier'}>{repairContext.scope}</dd>
			</div>
		</dl>
	{/if}
	{#if submitError}
		<div class="banner-error" role="alert">
			<p>{submitError}</p>
			<p class="muted">
				The affected rows stay unresolved. Correct the form and save again, or cancel.
			</p>
		</div>
	{/if}
	<form
		onsubmit={(e) => {
			e.preventDefault();
			resolveOrphans();
		}}
	>
		{#if resolveKind === 'function' && functionForm}
			<div class="field">
				<Label for="orphan-source">Source</Label>
				<Select
					id="orphan-source"
					ariaLabel="Source"
					bind:value={functionForm.source}
					options={[
						{ value: 'BUILT-IN', label: 'BUILT-IN' },
						{ value: 'DLL', label: 'DLL' }
					]}
				/>
				{#if fieldErrors.source}<span class="field-error">{fieldErrors.source}</span>{/if}
			</div>
			<div class="field">
				<Label for="orphan-interface">Device / Interface</Label>
				<DevicePicker
					id="orphan-interface"
					device={functionForm.interface}
					label="Device / Interface"
					blankLabel="Global (no Device / Interface)"
					fullWidth
					onValueChange={(value) => {
						if (functionForm) functionForm.interface = value;
					}}
				/>
				{#if fieldErrors.interface}<span class="field-error">{fieldErrors.interface}</span>{/if}
			</div>
			<div class="field">
				<Label for="orphan-fn-name">Name</Label>
				<Input id="orphan-fn-name" type="text" bind:value={functionForm.name} />
				{#if fieldErrors.name}<span class="field-error">{fieldErrors.name}</span>{/if}
			</div>
		{:else if definitionForm}
			<div class="field">
				<Label for="orphan-def-device">Device / Interface</Label>
				<DevicePicker
					id="orphan-def-device"
					device={definitionForm.device_name ?? ''}
					label="Device / Interface"
					blankLabel="Global definitions"
					disabled
					fullWidth
					onValueChange={() => undefined}
				/>
			</div>
			<div class="field">
				<Label for="orphan-def-name">Name</Label>
				<Input id="orphan-def-name" type="text" bind:value={definitionForm.name} />
				{#if fieldErrors.name}<span class="field-error">{fieldErrors.name}</span>{/if}
			</div>
			<div class="field">
				<Label for="orphan-def-type">Type</Label>
				<Select
					id="orphan-def-type"
					ariaLabel="Type"
					bind:value={definitionForm.type}
					options={[
						{ value: 'int', label: 'int' },
						{ value: 'double', label: 'double' },
						{ value: 'string', label: 'string' },
						{ value: 'list[int]', label: 'list[int]' }
					]}
				/>
				{#if fieldErrors.type}<span class="field-error">{fieldErrors.type}</span>{/if}
			</div>
			<div class="field">
				<Label for="orphan-def-unit">Unit</Label>
				<Input id="orphan-def-unit" type="text" bind:value={definitionForm.unit} />
				{#if fieldErrors.unit}<span class="field-error">{fieldErrors.unit}</span>{/if}
			</div>
		{/if}
		<div class="mt-4 flex justify-end gap-2">
			<Button variant="outline" type="button" disabled={saving} onclick={closeRepair}>Cancel</Button
			>
			<Button type="submit" disabled={saving}>
				{saving ? 'Creating…' : 'Create and resolve'}
			</Button>
		</div>
	</form>
</Dialog>

<style>
	.filters,
	.catalog-toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		margin-bottom: 16px;
	}
	.group-title {
		margin: 20px 0 4px;
		font-size: 15px;
		font-weight: 600;
	}
	.orphan-table {
		min-width: 0;
		margin-top: 8px;
		overflow-x: auto;
		border: 1px solid var(--border);
		border-radius: var(--radius);
	}
	.orphan-table :global(table) {
		min-width: 720px;
	}
	.identifier {
		font-family: var(--mono);
		overflow-wrap: anywhere;
	}
	.device-name {
		font-family: var(--mono);
		white-space: nowrap;
	}
	.missing-list {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}
	.repair-cell {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 4px;
	}
	.script-scroll {
		display: block;
		max-width: 360px;
		overflow-x: auto;
		white-space: nowrap;
		font-family: var(--mono);
		font-size: 12px;
	}
	.repair-context {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 12px;
		margin: 0 0 12px;
		font-size: 13px;
	}
	.repair-context dt {
		color: var(--muted-foreground);
		font-size: 12px;
	}
	.repair-context dd {
		margin: 0;
		overflow-wrap: anywhere;
	}
</style>
