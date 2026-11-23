<script lang="ts">
	import { untrack } from 'svelte';
	import { formatApiError } from '#lib/client.ts';
	import type { TroubleshootingFilters } from '#lib/troubleshooting.ts';
	import {
		downloadTroubleshootingExport,
		requestTroubleshootingExport,
		type TroubleshootingExportFormat,
		type TroubleshootingExportSelection
	} from '#lib/troubleshooting-transfer.ts';
	import Button from '#lib/components/ui/button.svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';

	let {
		ids,
		initialScope,
		filters,
		onClose
	}: {
		ids: number[];
		initialScope: TroubleshootingExportSelection['scope'];
		filters: TroubleshootingFilters;
		onClose: () => void;
	} = $props();

	let scope = $state(untrack(() => initialScope));
	let format = $state<TroubleshootingExportFormat>('json');
	let busy = $state(false);
	let error = $state('');
	const formats = [
		{ value: 'json', label: 'JSON' },
		{ value: 'xlsx', label: 'XLSX' },
		{ value: 'both', label: 'Both (ZIP)' }
	] as const;

	async function download() {
		if (busy || (scope === 'selected' && !ids.length)) return;
		busy = true;
		error = '';
		try {
			const selection = scope === 'selected' ? ({ scope, ids } as const) : { scope };
			const blob = await requestTroubleshootingExport(format, selection, filters);
			downloadTroubleshootingExport(blob, format);
			onClose();
		} catch (err) {
			error = formatApiError(err);
		} finally {
			busy = false;
		}
	}
</script>

<Dialog
	open
	title="Export Guides"
	description="Choose which guides to export and the file format. Exported files include IDs so they can be imported as updates."
	onOpenChange={(open) => {
		if (!open && !busy) onClose();
	}}
>
	<div class="grid gap-4">
		<fieldset class="border-border grid gap-2 rounded-md border p-3 text-sm" disabled={busy}>
			<legend class="px-1 font-semibold">Guides</legend>
			<label class="flex items-center gap-2">
				<input
					type="radio"
					name="export-scope"
					value="selected"
					bind:group={scope}
					disabled={!ids.length}
				/>
				{ids.length} selected {ids.length === 1 ? 'guide' : 'guides'}
			</label>
			<label class="flex items-center gap-2">
				<input type="radio" name="export-scope" value="matching" bind:group={scope} />
				All matching current filters
			</label>
			<label class="flex items-center gap-2">
				<input type="radio" name="export-scope" value="all" bind:group={scope} />
				All guides
			</label>
		</fieldset>
		<fieldset
			class="border-border flex flex-wrap gap-4 rounded-md border p-3 text-sm"
			disabled={busy}
		>
			<legend class="px-1 font-semibold">File format</legend>
			{#each formats as option (option.value)}
				<label class="flex items-center gap-2">
					<input type="radio" name="export-format" value={option.value} bind:group={format} />
					{option.label}
				</label>
			{/each}
		</fieldset>
		{#if error}<p class="text-destructive text-sm" role="alert">{error}</p>{/if}
		<div class="dialog-actions">
			<Button variant="outline" onclick={onClose} disabled={busy}>Cancel</Button>
			<Button onclick={download} disabled={busy || (scope === 'selected' && !ids.length)}>
				{busy ? 'Preparing…' : 'Download export'}
			</Button>
		</div>
	</div>
</Dialog>

<style>
	input[type='radio'] {
		border-radius: 50%;
	}
</style>
