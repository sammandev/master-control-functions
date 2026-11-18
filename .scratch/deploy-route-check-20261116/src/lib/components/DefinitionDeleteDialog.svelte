<script lang="ts">
	import {
		deleteDefinition,
		ApiError,
		type DefinitionRow,
		type DefinitionKind
	} from '#lib/definitions.ts';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import { toast } from 'svelte-sonner';
	let {
		kind,
		row,
		onClose,
		onDeleted
	}: {
		kind: DefinitionKind;
		row: DefinitionRow;
		onClose: () => void;
		onDeleted: () => void | Promise<void>;
	} = $props();
	let force = $derived(row.usage_count > 0);
	const impact = $derived(
		row.device_name
			? 'Links are retained and may resolve to a Global fallback. Without a fallback, they become unresolved.'
			: 'Links using this Global definition are retained but become unresolved.'
	);
	let deleting = $state(false);
	let error = $state('');
	async function remove() {
		if (deleting) return;
		deleting = true;
		error = '';
		try {
			await deleteDefinition(kind, row.id, force);
			toast.success('Definition deleted');
			await onDeleted();
			onClose();
		} catch (failure) {
			error = failure instanceof Error ? failure.message : String(failure);
			if (failure instanceof ApiError && failure.status === 409) force = true;
		} finally {
			deleting = false;
		}
	}
</script>

<AlertDialog
	open={true}
	onOpenChange={(open) => {
		if (!open && !deleting) onClose();
	}}
	title={`${force ? 'Force delete' : 'Delete'} ${row.name}?`}
	description={`${row.device_name ? `Device: ${row.device_name}` : 'Global - shared across Devices'}. ${force ? `${row.name} is referenced by Function links. ${impact}` : 'This permanently removes the definition. This cannot be undone.'}`}
	confirmLabel="Delete"
	onConfirm={remove}
>
	{#if error}<p class="banner-error" role="alert">{error}</p>{/if}
</AlertDialog>
