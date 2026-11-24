<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { FunctionRow } from '#lib/api.ts';
	import { resolve } from '$app/paths';
	import { Pencil, Trash2, Code, ExternalLink } from '@lucide/svelte';
	import LinksPanel from '#lib/components/LinksPanel.svelte';
	import ExamplesPanel from '#lib/components/ExamplesPanel.svelte';
	import McpReadinessPanel from '#lib/components/McpReadinessPanel.svelte';
	import ComposeDialog from '#lib/components/ComposeDialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import * as Tabs from '#lib/components/ui/tabs/index.ts';

	let {
		fn,
		onEdit,
		onDelete,
		standalone = false,
		heading
	}: {
		fn: FunctionRow;
		onEdit?: (fn: FunctionRow) => void;
		onDelete?: (fn: FunctionRow) => void;
		standalone?: boolean;
		heading?: Snippet;
	} = $props();
	let tab = $state('parameters');
	let composeOpen = $state(false);
	let parameterCount = $state<number | null>(null);
	let criteriaCount = $state<number | null>(null);
</script>

<section class="function-detail" class:standalone aria-label="Function details">
	<header>
		<div class="detail-top">
			<div>
				{#if !standalone}<p class="ownership">
						Function reference / {fn.interface || 'GLOBAL'}
					</p>{/if}
				{#if heading}{@render heading()}{:else}<svelte:element this={standalone ? 'h1' : 'h2'}
						>{fn.name}</svelte:element
					>{/if}
				<div class="identity">
					<Badge>{fn.source}</Badge><Badge class="max-w-full break-all whitespace-normal"
						>{fn.interface || 'GLOBAL'}</Badge
					>
				</div>
			</div>
			<div class="actions">
				{#if onEdit}<Button
						variant="outline"
						aria-label="Edit function"
						onclick={() => onEdit?.(fn)}><Pencil data-icon="inline-start" /> Edit</Button
					>{/if}
				{#if onDelete}<Button
						variant="outline"
						size="icon"
						aria-label="Delete function"
						title="Delete function"
						onclick={() => onDelete?.(fn)}><Trash2 size={16} /></Button
					>{/if}
				<Button onclick={() => (composeOpen = true)}><Code size={16} /> Compose script</Button>
			</div>
		</div>
		{#if fn.description}<p class="description">{fn.description}</p>{/if}
		<dl>
			<div>
				<dt>Category</dt>
				<dd>{fn.category || '-'}</dd>
			</div>
			<div>
				<dt>Execution role</dt>
				<dd>{fn.execution_role || '-'}</dd>
			</div>
			<div>
				<dt>Execution phase</dt>
				<dd>{fn.execution_phase || '-'}</dd>
			</div>
			<div>
				<dt>Convention</dt>
				<dd>{fn.convention ?? '-'}</dd>
			</div>
			<div>
				<dt>Updated</dt>
				<dd>{fn.updated_at.slice(0, 10)}</dd>
			</div>
		</dl>
	</header>
	<Tabs.Root bind:value={tab}>
		<Tabs.List class="function-sections" aria-label="Function sections">
			<Tabs.Trigger class="function-section" value="parameters" aria-label="Parameters"
				>Parameters{#if parameterCount !== null}<span class="tab-count" aria-hidden="true"
						>{parameterCount}</span
					>{/if}</Tabs.Trigger
			>
			<Tabs.Trigger class="function-section" value="criteria" aria-label="Criteria"
				>Criteria{#if criteriaCount !== null}<span class="tab-count" aria-hidden="true"
						>{criteriaCount}</span
					>{/if}</Tabs.Trigger
			>
			<Tabs.Trigger class="function-section" value="examples">Examples</Tabs.Trigger>
			<Tabs.Trigger class="function-section" value="mcp">MCP readiness</Tabs.Trigger>
		</Tabs.List>
		<Tabs.Content value="parameters" class="panel-scroll"
			><LinksPanel
				fnId={fn.id}
				defaultDevice={fn.interface}
				kind="parameter-links"
				defKind="parameter-definitions"
				noun="parameter"
				onCountChange={(count) => (parameterCount = count)}
			/></Tabs.Content
		>
		<Tabs.Content value="criteria" class="panel-scroll"
			><LinksPanel
				fnId={fn.id}
				defaultDevice={fn.interface}
				kind="criteria-links"
				defKind="criteria-definitions"
				noun="criteria"
				onCountChange={(count) => (criteriaCount = count)}
			/></Tabs.Content
		>
		<Tabs.Content value="examples" class="panel-scroll"
			><ExamplesPanel
				fixed={{ source: fn.source, name: fn.name }}
				defaultDevice={fn.interface}
				functionId={fn.id}
			/></Tabs.Content
		>
		<Tabs.Content value="mcp" class="panel-scroll"
			><McpReadinessPanel fnId={fn.id} defaultDevice={fn.interface} /></Tabs.Content
		>
	</Tabs.Root>
	{#if !standalone}<div class="detail-footer">
			<span>Function reference</span><a href={resolve('/functions/[id]', { id: String(fn.id) })}
				>Manage function <ExternalLink size={16} /></a
			>
		</div>{/if}
</section>

{#if composeOpen}<ComposeDialog fnId={fn.id} onClose={() => (composeOpen = false)} />{/if}

<style>
	.function-detail {
		min-width: 0;
		padding: 24px;
	}
	.function-detail.standalone {
		padding: 0;
		margin-top: 24px;
	}
	.standalone :global(.function-heading) {
		font: 650 28px/1.3 var(--mono);
		margin: 0 0 10px;
	}
	.function-detail :global(.panel-scroll) {
		max-width: 100%;
		overflow-x: auto;
	}
	.function-detail :global(.function-sections) {
		width: 100%;
		justify-content: flex-start;
		gap: 24px;
		overflow-x: auto;
		padding: 0;
		border-bottom: 1px solid var(--border);
		border-radius: 0;
		background: transparent;
	}
	.function-detail :global(.function-section) {
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
	.function-detail :global(.function-section[data-state='active']) {
		color: var(--device-accent);
		border-bottom-color: var(--device-accent);
	}
	.tab-count {
		color: var(--muted-foreground);
		font-size: 10px;
		margin-left: 4px;
	}
	header {
		margin-bottom: 24px;
	}
	.detail-top {
		display: flex;
		justify-content: space-between;
		align-items: start;
		flex-wrap: wrap;
		gap: 16px;
	}
	.detail-top > div:first-child {
		min-width: 0;
	}
	header :global(h1),
	header :global(h2) {
		font-family: var(--mono);
		font-size: 20px;
		font-weight: 700;
		overflow-wrap: anywhere;
		margin: 7px 0 10px;
	}
	.ownership {
		color: var(--muted-foreground);
		font-size: 10px;
		font-weight: 700;
		text-transform: uppercase;
		overflow-wrap: anywhere;
	}
	.identity,
	.actions {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
	}
	.description {
		font-size: 13px;
		line-height: 1.6;
		margin: 17px 0;
		max-width: 780px;
		overflow-wrap: anywhere;
	}
	dl {
		display: flex;
		flex-wrap: wrap;
		gap: 16px 30px;
		padding: 15px 0 22px;
	}
	dt {
		font-size: 11px;
		color: var(--muted-foreground);
		margin-bottom: 4px;
	}
	dd {
		font-size: 12px;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.detail-footer {
		margin-top: 30px;
		border-top: 1px solid var(--border);
		padding-top: 14px;
		display: flex;
		justify-content: space-between;
		gap: 12px;
		color: var(--muted-foreground);
		font-size: 11px;
	}
	.detail-footer a {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		color: var(--primary);
		min-height: 32px;
	}
	.detail-footer a:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 3px;
	}
	@media (max-width: 600px) {
		.function-detail {
			padding: 20px 14px;
		}
	}
</style>
