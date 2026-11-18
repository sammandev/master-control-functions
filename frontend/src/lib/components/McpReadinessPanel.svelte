<script lang="ts">
	import { formatApiError } from '#lib/client.ts';
	import {
		deleteGrammar,
		deleteScriptKey,
		getReadiness,
		saveGrammar,
		saveScriptKey,
		statusLabel,
		type GrammarScope,
		type Readiness,
		type ReadinessField,
		type ScriptIssue,
		type ScriptKey
	} from '#lib/mcp.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import { toast } from 'svelte-sonner';
	import { CircleAlert, CircleCheck, RefreshCw } from '@lucide/svelte';
	import { untrack } from 'svelte';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir, type SortValue } from '#lib/table-sort.ts';

	let { fnId, defaultDevice = '' }: { fnId: number; defaultDevice?: string } = $props();

	let report = $state<Readiness | null>(null);
	let error = $state('');
	let loading = $state(false);
	let device = $state('');
	let auditor = $state('');
	let encoding = $state<'raw' | 'named'>('named');
	let scope = $state<GrammarScope>('bundle');
	let note = $state('');
	let fieldSort = $state<Record<ScriptKey['kind'], { key: string; dir: SortDir }>>({
		parameter: { key: '', dir: 'asc' },
		criterion: { key: '', dir: 'asc' }
	});
	const fieldKeys: Record<string, (field: ReadinessField) => SortValue> = {
		name: (field) => field.name,
		definition: (field) => Number(field.has_definition),
		audit: (field) => field.script_key_scope
	};
	function sortFields(kind: ScriptKey['kind'], column: string) {
		fieldSort[kind] = nextSort(fieldSort[kind].key, fieldSort[kind].dir, column);
	}

	async function load() {
		loading = true;
		error = '';
		try {
			report = await getReadiness(fnId, device.trim());
			if (report.grammar) {
				encoding = report.grammar.encoding;
				scope = report.grammar.scope;
			}
		} catch (e) {
			error = formatApiError(e);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const id = fnId;
		const scope = defaultDevice;
		untrack(() => {
			void id;
			device = scope;
			load();
		});
	});

	function requireAuditor(): boolean {
		if (auditor.trim()) return true;
		toast.error('Enter who audited this before saving.');
		return false;
	}

	async function run(action: () => Promise<unknown>, success: string) {
		try {
			await action();
			toast.success(success);
			await load();
		} catch (e) {
			toast.error(formatApiError(e));
		}
	}

	function submitGrammar(event: SubmitEvent) {
		event.preventDefault();
		if (!requireAuditor()) return;
		run(
			() => saveGrammar(fnId, { encoding, scope, audited_by: auditor.trim(), note }),
			'Grammar saved'
		);
	}

	function auditKey(kind: ScriptKey['kind'], field: ReadinessField, keyScope: GrammarScope) {
		if (!requireAuditor()) return;
		run(
			() =>
				saveScriptKey(fnId, {
					kind,
					name: field.name,
					scope: keyScope,
					audited_by: auditor.trim(),
					note: ''
				}),
			`${field.name} audited for ${keyScope}`
		);
	}

	const statusVariant = (status: Readiness['status']) =>
		status === 'ready' ? 'default' : status === 'bundle_ready' ? 'outline' : 'warn';
</script>

{#snippet issues(title: string, list: ScriptIssue[])}
	{#if list.length}
		<div class="issue-group">
			<h4><CircleAlert size={14} aria-hidden="true" /> {title} <span>{list.length}</span></h4>
			<ul class="issues">
				{#each list as issue, i (i)}
					<li><code>{issue.field}</code> <span class="code">{issue.code}</span> {issue.message}</li>
				{/each}
			</ul>
		</div>
	{/if}
{/snippet}

{#snippet readiness(label: string, ready: boolean, count: number)}
	<div class="readiness" class:ready>
		{#if ready}<CircleCheck size={16} aria-hidden="true" />{:else}<CircleAlert
				size={16}
				aria-hidden="true"
			/>{/if}
		<span><strong>{label}</strong>{ready ? 'composable' : `${count} issue(s)`}</span>
	</div>
{/snippet}

{#snippet fieldRows(kind: ScriptKey['kind'], label: string, fields: ReadinessField[])}
	{@const sorting = fieldSort[kind]}
	{@const rows = sorting.key ? sortRows(fields, fieldKeys[sorting.key], sorting.dir) : fields}
	{#if fields.length}
		<div class="block">
			<h4>{label} <span>{fields.length}</span></h4>
			<div class="table-scroll">
				<table>
					<thead>
						<tr
							><SortHead
								label="Key"
								col="name"
								active={sorting.key}
								dir={sorting.dir}
								onclick={() => sortFields(kind, 'name')}
							/><SortHead
								label="Definition"
								col="definition"
								active={sorting.key}
								dir={sorting.dir}
								onclick={() => sortFields(kind, 'definition')}
							/><SortHead
								label="Script key audit"
								col="audit"
								active={sorting.key}
								dir={sorting.dir}
								onclick={() => sortFields(kind, 'audit')}
							/><th>Actions</th></tr
						>
					</thead>
					<tbody>
						{#each rows as field (field.name)}
							<tr>
								<td
									><code>{field.name}</code>{#if field.device_name}<span class="scope"
											>{field.device_name}</span
										>{/if}</td
								>
								<td
									>{#if field.has_definition}<Badge variant="outline">yes</Badge>{:else}<Badge
											variant="warn">missing</Badge
										>{/if}</td
								>
								<td
									>{#if field.script_key_scope}<Badge>{field.script_key_scope}</Badge>{:else}<span
											class="muted">not audited</span
										>{/if}</td
								>
								<td>
									{#if report?.grammar?.encoding === 'named'}
										<div class="row-actions">
											<Button
												size="sm"
												variant="outline"
												onclick={() => auditKey(kind, field, 'line')}>Line</Button
											>
											<Button
												size="sm"
												variant="outline"
												onclick={() => auditKey(kind, field, 'bundle')}>Bundle</Button
											>
											{#if field.script_key_scope}
												<Button
													size="sm"
													variant="ghost"
													onclick={() =>
														run(
															() => deleteScriptKey(fnId, kind, field.name),
															`${field.name} audit removed`
														)}>Remove</Button
												>
											{/if}
										</div>
									{:else}<span class="muted">Named grammar required</span>{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
{/snippet}

<section class="mcp" aria-label="MCP readiness">
	<div class="toolbar">
		<label>
			Device scope
			<Input
				bind:value={device}
				placeholder={defaultDevice || 'Global'}
				aria-label="Device scope"
			/>
		</label>
		<label>
			Audited by
			<Input bind:value={auditor} placeholder="Your name" aria-label="Audited by" />
		</label>
		<Button variant="outline" onclick={load} disabled={loading}
			><RefreshCw data-icon="inline-start" class={loading ? 'animate-spin' : ''} />
			{loading ? 'Checking…' : 'Check'}</Button
		>
	</div>

	{#if error}
		<div class="banner-error" role="alert">
			{error}<Button variant="outline" size="sm" onclick={load}
				><RefreshCw data-icon="inline-start" /> Retry</Button
			>
		</div>
	{:else if report}
		<div class="status" aria-live="polite">
			<div class="status-head">
				<span>Overall</span>
				<Badge variant={statusVariant(report.status)}>{statusLabel[report.status]}</Badge>
			</div>
			{@render readiness('Single line', report.line.ready, report.line.issues.length)}
			{@render readiness('INI bundle', report.bundle.ready, report.bundle.issues.length)}
		</div>

		<form class="grammar block" onsubmit={submitGrammar}>
			<h4>Script grammar</h4>
			<p class="hint">
				{#if report.grammar}
					Audited by {report.grammar.audited_by} on {report.grammar.audited_at.slice(0, 10)}.
				{:else}
					Not audited — script-audit-mcp will not compose this function.
				{/if}
				Raw needs convention 1; named needs convention 2.
			</p>
			<div class="grammar-fields">
				<label
					>Encoding
					<Select
						ariaLabel="Encoding"
						value={encoding}
						options={[
							{ value: 'named', label: 'named (KEY=value)' },
							{ value: 'raw', label: 'raw (single payload)' }
						]}
						onchange={(value) => (encoding = value as 'raw' | 'named')}
					/>
				</label>
				<label
					>Scope
					<Select
						ariaLabel="Scope"
						value={scope}
						options={[
							{ value: 'line', label: 'line and bundle' },
							{ value: 'bundle', label: 'INI bundle only' }
						]}
						onchange={(value) => (scope = value as GrammarScope)}
					/>
				</label>
				<label class="note">Note <Input bind:value={note} aria-label="Grammar note" /></label>
			</div>
			<div class="grammar-actions">
				{#if report.grammar}
					<Button
						variant="outline"
						onclick={() => run(() => deleteGrammar(fnId), 'Grammar removed')}>Remove grammar</Button
					>
				{/if}
				<Button type="submit">Save grammar</Button>
			</div>
		</form>

		{@render fieldRows('parameter', 'Parameters', report.parameters)}
		{@render fieldRows('criterion', 'Criteria', report.criteria)}
		{@render issues('Line issues', report.line.issues)}
		{@render issues('Bundle issues', report.bundle.issues)}
		{@render issues('Warnings', report.warnings)}
	{:else}
		<p class="hint" role="status">Checking…</p>
	{/if}
</section>

<style>
	.mcp {
		display: grid;
		gap: 16px;
		padding: 20px 0 12px;
	}
	.toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: 12px;
	}
	.toolbar label {
		flex: 1 1 180px;
		max-width: 260px;
	}
	label {
		display: grid;
		gap: 5px;
		font-size: 11px;
		font-weight: 600;
		color: var(--muted-foreground);
	}
	.status {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 1px;
		overflow: hidden;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--border);
	}
	.status > * {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		background: var(--card);
		font-size: 12px;
	}
	.status-head > span {
		color: var(--muted-foreground);
	}
	.readiness {
		color: var(--muted-foreground);
	}
	.readiness.ready {
		color: var(--foreground);
	}
	.readiness strong {
		display: block;
		color: var(--foreground);
		font-weight: 600;
	}
	.block {
		display: grid;
		gap: 10px;
		padding: 16px;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--card);
	}
	.grammar-fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 12px;
	}
	.grammar-actions {
		display: flex;
		justify-content: flex-end;
		flex-wrap: wrap;
		gap: 8px;
	}
	h4 {
		display: flex;
		align-items: center;
		gap: 8px;
		font-weight: 600;
		font-size: 13px;
		margin: 0;
	}
	h4 > span {
		min-width: 20px;
		padding: 1px 6px;
		border: 1px solid var(--border);
		border-radius: 999px;
		color: var(--muted-foreground);
		font-size: 10px;
		font-weight: 500;
		text-align: center;
	}
	.hint,
	.muted {
		color: var(--muted-foreground);
		font-size: 12px;
		margin: 0;
	}
	.table-scroll {
		overflow-x: auto;
	}
	table {
		width: 100%;
		min-width: 520px;
		border-collapse: collapse;
		font-size: 13px;
	}
	th {
		color: var(--muted-foreground);
		font-size: 11px;
		font-weight: 600;
	}
	th,
	td {
		text-align: left;
		vertical-align: middle;
		padding: 8px;
		border-bottom: 1px solid var(--border);
	}
	tbody tr:last-child td {
		border-bottom: 0;
	}
	.scope {
		margin-left: 6px;
		color: var(--muted-foreground);
		font-size: 11px;
	}
	.row-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.issue-group {
		display: grid;
		gap: 8px;
		padding: 12px 14px;
		border: 1px solid var(--border);
		border-left: 3px solid var(--destructive);
		border-radius: 6px;
		background: var(--card);
	}
	.issues {
		margin: 0;
		padding-left: 18px;
		font-size: 13px;
		display: grid;
		gap: 4px;
	}
	.code {
		color: var(--muted-foreground);
		font-family: var(--mono);
		font-size: 12px;
	}
</style>
