<script lang="ts">
	import { getDefinition, type DefinitionKind, type DefinitionRow } from '#lib/definitions.ts';
	import type { DeviceDefinitionRow } from '#lib/devices.ts';
	import DefinitionDetail from '#lib/components/DefinitionDetail.svelte';
	import DefinitionDialog from '#lib/components/DefinitionDialog.svelte';
	import DefinitionDeleteDialog from '#lib/components/DefinitionDeleteDialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import { RefreshCw } from '@lucide/svelte';
	let {
		definition,
		kind,
		device,
		ownershipLabel,
		onChanged
	}: {
		definition: DeviceDefinitionRow;
		kind: DefinitionKind;
		device: string;
		ownershipLabel?: string;
		onChanged: (row: DefinitionRow | null) => void | Promise<void>;
	} = $props();
	let retry = $state(0);
	const content = $derived.by(() => {
		void retry;
		return getDefinition(kind, definition.id);
	});
	let target = $state<DefinitionRow | null>(null);
	let dialogOpen = $state(false);
	let override = $state(false);
	let deleting = $state<DefinitionRow | null>(null);
	function edit(row: DefinitionRow) {
		target = row;
		override = false;
		dialogOpen = true;
	}
	function createOverride(row: DefinitionRow) {
		target = row;
		override = true;
		dialogOpen = true;
	}
</script>

{#await content}<p class="banner-empty" role="status">Loading definition details...</p>
{:then row}
	<DefinitionDetail
		{row}
		ownershipLabel={ownershipLabel ??
			(kind === 'parameter-definitions' ? 'Parameter definition' : 'Criteria definition')}
		onEdit={edit}
		onDelete={(selected) => {
			deleting = selected;
		}}
		onOverride={!row.device_name && device ? createOverride : undefined}
	/>
{:catch error}<div class="banner-error" role="alert">
		{error instanceof Error ? error.message : String(error)}<Button
			variant="outline"
			onclick={() => retry++}><RefreshCw size={16} /> Retry definition</Button
		>
	</div>{/await}
{#if dialogOpen && target}<DefinitionDialog
		bind:open={dialogOpen}
		{kind}
		title={kind === 'parameter-definitions' ? 'Parameter Definitions' : 'Criteria Definitions'}
		editRow={override ? null : target}
		createFrom={override ? target : null}
		prefillDevice={override ? device : target.device_name}
		lockDevice
		onsaved={async (row) => {
			await onChanged(row);
			retry++;
		}}
	/>{/if}
{#if deleting}<DefinitionDeleteDialog
		{kind}
		row={deleting}
		onClose={() => {
			deleting = null;
		}}
		onDeleted={() => onChanged(null)}
	/>{/if}
