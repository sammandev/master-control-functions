<script lang="ts">
	import type { DefinitionRow } from '#lib/definitions.ts';
	import { resolve } from '$app/paths';
	import { devicePath, deviceBadgeLinkClass } from '#lib/devices.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import { Pencil, Trash2, CopyPlus } from '@lucide/svelte';
	let {
		row,
		onEdit,
		onDelete,
		onOverride,
		ownershipLabel
	}: {
		row: DefinitionRow;
		onEdit: (row: DefinitionRow) => void;
		onDelete: (row: DefinitionRow) => void;
		onOverride?: (row: DefinitionRow) => void;
		ownershipLabel?: string;
	} = $props();
</script>

<section aria-label="Definition details" class="definition-detail">
	<div class="detail-top">
		<div>
			<p class="ownership">
				{row.device_name
					? (ownershipLabel ?? 'Device definition')
					: 'Global - shared across Devices'}
			</p>
			<h2>{row.name}</h2>
			{#if row.device_name}<a
					href={resolve('/devices/[name]', { name: devicePath(row.device_name) })}
					class={deviceBadgeLinkClass}
					><Badge class="max-w-full break-all whitespace-normal">{row.device_name}</Badge></a
				>{/if}
		</div>
		{@render actions()}
	</div>
	{#if row.description}<p class="description">{row.description}</p>{/if}
	<dl>
		<div>
			<dt>Type</dt>
			<dd>{row.type}</dd>
		</div>
		<div>
			<dt>Unit</dt>
			<dd>{row.unit || 'None'}</dd>
		</div>
		<div>
			<dt>Default value</dt>
			<dd class="identifier">{row.default_value ?? 'None'}</dd>
		</div>
		<div>
			<dt>Allowed values</dt>
			<dd class="identifier">{row.allowed_values || 'None'}</dd>
		</div>
		<div>
			<dt>Cardinality</dt>
			<dd>{row.cardinality}</dd>
		</div>
		<div>
			<dt>Value pattern</dt>
			<dd class="identifier">{row.value_pattern || 'None'}</dd>
		</div>
		<div>
			<dt>Required</dt>
			<dd>{row.required ? 'Yes' : 'No'}</dd>
		</div>
		<div>
			<dt>Used by</dt>
			<dd>{row.usage_count} Function link{row.usage_count === 1 ? '' : 's'}</dd>
		</div>
		<div>
			<dt>Orphan usage</dt>
			<dd>{row.orphan_usage_count} orphan link{row.orphan_usage_count === 1 ? '' : 's'}</dd>
		</div>
		<div>
			<dt>Updated</dt>
			<dd>{row.updated_at.slice(0, 10)}</dd>
		</div>
	</dl>
</section>

{#snippet actions()}
	<div class="actions">
		{#if onOverride}<Button onclick={() => onOverride?.(row)}
				><CopyPlus size={16} /> Create Device override</Button
			>{/if}
		<Button variant="outline" onclick={() => onEdit(row)}
			><Pencil size={16} /> {row.device_name ? 'Edit definition' : 'Edit Global definition'}</Button
		>
		<Button variant="outline" onclick={() => onDelete(row)}
			><Trash2 size={16} />
			{row.device_name ? 'Delete definition' : 'Delete Global definition'}</Button
		>
	</div>
{/snippet}

<style>
	.definition-detail {
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
	.description {
		max-width: 780px;
		margin: 16px 0;
		overflow-wrap: anywhere;
	}
	dl {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(180px, 100%), 1fr));
		gap: 20px;
		margin: 24px 0;
	}
	dt {
		margin-bottom: 4px;
	}
	dd {
		margin: 0;
		font-size: 13px;
		overflow-wrap: anywhere;
	}
	.identifier {
		font-family: var(--mono);
		white-space: pre-wrap;
	}
	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	@media (max-width: 600px) {
		.definition-detail {
			padding: 20px 14px;
		}
		dl {
			gap: 16px;
		}
	}
</style>
