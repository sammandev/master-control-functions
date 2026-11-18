<script lang="ts">
	import type { ExampleRow } from '#lib/examples.ts';
	import { copyText } from '#lib/clipboard.ts';
	import { resolve } from '$app/paths';
	import { devicePath, deviceBadgeLinkClass } from '#lib/devices.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Tooltip from '#lib/components/ui/tooltip.svelte';
	import { Copy, Check, Pencil, Trash2 } from '@lucide/svelte';

	let {
		row,
		onEdit,
		onDelete
	}: {
		row: ExampleRow;
		onEdit: (row: ExampleRow) => void;
		onDelete: (row: ExampleRow) => void;
	} = $props();

	// ---- full-script copy ---------------------------------------------------
	// The detail pane always carries the complete script; copy success and
	// failure are announced through a live status line, and the script text
	// stays selectable as the clipboard fallback.
	let copied = $state(false);
	let copyStatus = $state('');

	async function copyScript() {
		const success = await copyText(row.script);
		copied = success;
		copyStatus = success ? 'Script copied.' : 'Copy failed. The script text is selectable.';
	}
</script>

<section aria-label="Example details" class="example-detail">
	<div class="detail-top">
		<div>
			<p class="ownership">
				{row.device_name ? 'Device example' : 'Global - shared across Devices'}
			</p>
			<h2>{row.function_source} / {row.function_name}</h2>
			<div class="identity">
				<Badge>{row.function_source}</Badge>
				{#if row.device_name}<a
						href={resolve('/devices/[name]', { name: devicePath(row.device_name) })}
						class={deviceBadgeLinkClass}
						><Badge class="max-w-full break-all whitespace-normal">{row.device_name}</Badge></a
					>{:else}<Badge class="max-w-full break-all whitespace-normal">Global</Badge>{/if}
				{#if !row.function_exists}
					<Tooltip content="No function with this source and name exists">
						{#snippet children(tp)}
							<Badge variant="warn" {...tp} tabindex={0}>orphan</Badge>
						{/snippet}
					</Tooltip>
				{/if}
			</div>
		</div>
		{@render actions()}
	</div>
	{#if row.meaning}<p class="meaning">{row.meaning}</p>{/if}
	<div class="script-block">
		<div class="script-heading">
			<span>Script</span>
			<Button variant="outline" size="sm" onclick={copyScript} aria-label={`Copy example ${row.id}`}
				>{#if copied}<Check size={16} />{:else}<Copy size={16} />{/if}
				Copy script</Button
			>
		</div>
		<pre class="script-scroll"><code>{row.script}</code></pre>
		<p role="status" class="copy-status">{copyStatus}</p>
	</div>
	<dl>
		<div>
			<dt>Ownership</dt>
			<dd>{row.device_name || 'Global'}</dd>
		</div>
		<div>
			<dt>Created</dt>
			<dd>{row.created_at.slice(0, 10)}</dd>
		</div>
		<div>
			<dt>Updated</dt>
			<dd>{row.updated_at.slice(0, 10)}</dd>
		</div>
	</dl>
</section>

{#snippet actions()}
	<div class="actions">
		<Button variant="outline" onclick={() => onEdit(row)}>
			<Pencil size={16} /> {row.device_name ? 'Edit example' : 'Edit Global example'}</Button
		>
		<Button variant="outline" onclick={() => onDelete(row)}>
			<Trash2 size={16} />
			{row.device_name ? 'Delete example' : 'Delete Global example'}</Button
		>
	</div>
{/snippet}

<style>
	.example-detail {
		min-width: 0;
		padding: 24px;
	}
	.detail-top {
		display: flex;
		align-items: start;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 16px;
	}
	.detail-top > div:first-child {
		min-width: 0;
	}
	.ownership,
	dt {
		color: var(--muted-foreground);
		font-size: 12px;
	}
	h2 {
		margin: 8px 0 12px;
		font: 700 20px/1.5 var(--mono);
		overflow-wrap: anywhere;
	}
	.identity {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
	}
	.identity a:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 3px;
	}
	.meaning {
		margin: 16px 0;
		max-width: 100%;
		overflow-wrap: anywhere;
	}
	.script-block {
		margin: 20px 0;
		min-width: 0;
	}
	.script-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 8px;
	}
	.script-heading span {
		color: var(--muted-foreground);
		font-size: 12px;
	}
	/* The script scrolls inside its own surface; the document never gains
	   horizontal overflow from long script lines. */
	.script-scroll {
		margin: 0;
		max-height: 320px;
		overflow: auto;
		padding: 12px;
		background: var(--muted);
		border-radius: 4px;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		font-family: var(--mono);
		font-size: 12px;
		line-height: 1.8;
	}
	.script-scroll:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 2px;
	}
	.copy-status {
		min-height: 18px;
		margin-top: 6px;
		font-size: 12px;
		color: var(--muted-foreground);
	}
	dl {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 20px;
		margin: 20px 0;
	}
	dt {
		margin-bottom: 4px;
	}
	dd {
		margin: 0;
		font-size: 13px;
		overflow-wrap: anywhere;
		white-space: pre-wrap;
	}
	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	@media (max-width: 600px) {
		.example-detail {
			padding: 20px 14px;
		}
		dl {
			grid-template-columns: minmax(0, 1fr);
			gap: 16px;
		}
	}
</style>
