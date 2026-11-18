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
	import { toast } from 'svelte-sonner';
	import { RefreshCw } from '@lucide/svelte';
	import { untrack } from 'svelte';

	let { fnId, defaultDevice = '' }: { fnId: number; defaultDevice?: string } = $props();

	let report = $state<Readiness | null>(null);
	let error = $state('');
	let loading = $state(false);
	let device = $state('');
	let auditor = $state('');
	let encoding = $state<'raw' | 'named'>('named');
	let scope = $state<GrammarScope>('bundle');
	let note = $state('');

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
		<h4>{title}</h4>
		<ul class="issues">
			{#each list as issue, i (i)}
				<li><code>{issue.field}</code> <span class="code">{issue.code}</span> {issue.message}</li>
			{/each}
		</ul>
	{/if}
{/snippet}

{#snippet fieldRows(kind: ScriptKey['kind'], label: string, fields: ReadinessField[])}
	{#if fields.length}
		<h4>{label}</h4>
		<table>
			<thead>
				<tr><th>Key</th><th>Definition</th><th>Script key audit</th><th>Actions</th></tr>
			</thead>
			<tbody>
				{#each fields as field (field.name)}
					<tr>
						<td><code>{field.name}</code>{field.device_name ? ` (${field.device_name})` : ''}</td>
						<td>{field.has_definition ? 'yes' : 'missing'}</td>
						<td>{field.script_key_scope || 'not audited'}</td>
						<td class="row-actions">
							{#if report?.grammar?.encoding === 'named'}
								<Button size="sm" variant="outline" onclick={() => auditKey(kind, field, 'line')}
									>Line</Button
								>
								<Button size="sm" variant="outline" onclick={() => auditKey(kind, field, 'bundle')}
									>Bundle</Button
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
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
{/snippet}

<section class="mcp" aria-label="MCP readiness">
	<div class="toolbar">
		<label>
			Device scope
			<Input bind:value={device} placeholder={defaultDevice} aria-label="Device scope" />
		</label>
		<Button variant="outline" onclick={load} disabled={loading}
			><RefreshCw size={16} /> Check</Button
		>
		<label>
			Audited by
			<Input bind:value={auditor} placeholder="your name" aria-label="Audited by" />
		</label>
	</div>

	{#if error}
		<div class="banner-error" role="alert">{error}</div>
	{:else if report}
		<p class="status" aria-live="polite">
			<Badge variant={statusVariant(report.status)}>{statusLabel[report.status]}</Badge>
			<span
				>Single line: {report.line.ready ? 'composable' : `${report.line.issues.length} issue(s)`} · INI
				bundle: {report.bundle.ready
					? 'composable'
					: `${report.bundle.issues.length} issue(s)`}</span
			>
		</p>

		<form class="grammar" onsubmit={submitGrammar}>
			<h4>Script grammar</h4>
			<p class="hint">
				{#if report.grammar}
					Audited by {report.grammar.audited_by} on {report.grammar.audited_at.slice(0, 10)}.
				{:else}
					Not audited — script-audit-mcp will not compose this function.
				{/if}
				Raw needs convention 1; named needs convention 2.
			</p>
			<label
				>Encoding
				<select bind:value={encoding}>
					<option value="named">named (KEY=value)</option>
					<option value="raw">raw (single payload)</option>
				</select>
			</label>
			<label
				>Scope
				<select bind:value={scope}>
					<option value="line">line and bundle</option>
					<option value="bundle">INI bundle only</option>
				</select>
			</label>
			<label>Note <Input bind:value={note} aria-label="Grammar note" /></label>
			<Button type="submit">Save grammar</Button>
			{#if report.grammar}
				<Button variant="outline" onclick={() => run(() => deleteGrammar(fnId), 'Grammar removed')}
					>Remove grammar</Button
				>
			{/if}
		</form>

		{@render fieldRows('parameter', 'Parameters', report.parameters)}
		{@render fieldRows('criterion', 'Criteria', report.criteria)}
		{@render issues('Line issues', report.line.issues)}
		{@render issues('Bundle issues', report.bundle.issues)}
		{@render issues('Warnings', report.warnings)}
	{:else}
		<p class="hint">Checking…</p>
	{/if}
</section>

<style>
	.mcp {
		display: grid;
		gap: 16px;
		padding: 12px 0;
	}
	.toolbar,
	.grammar,
	.status {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: 12px;
	}
	.status {
		align-items: center;
		font-size: 13px;
	}
	label {
		display: grid;
		gap: 4px;
		font-size: 12px;
		color: var(--muted-foreground);
	}
	select {
		height: 32px;
		border: 1px solid var(--border);
		border-radius: 4px;
		background: var(--card);
		color: var(--foreground);
		padding: 0 8px;
	}
	h4 {
		width: 100%;
		font-weight: 600;
		font-size: 13px;
		margin: 0;
	}
	.hint {
		width: 100%;
		color: var(--muted-foreground);
		font-size: 12px;
		margin: 0;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}
	th,
	td {
		text-align: left;
		padding: 6px 8px;
		border-bottom: 1px solid var(--border);
	}
	.row-actions {
		display: flex;
		gap: 6px;
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
