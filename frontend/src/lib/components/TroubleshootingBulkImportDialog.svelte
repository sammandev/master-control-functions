<script lang="ts">
	import { formatApiError } from '#lib/client.ts';
	import {
		commitTroubleshootingImport,
		getTroubleshootingPreview,
		troubleshootingTransferFileTypes,
		type TroubleshootingImportResult,
		type TroubleshootingPreview
	} from '#lib/troubleshooting-transfer.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';

	let {
		onClose,
		onImported
	}: {
		onClose: () => void;
		onImported: (result: TroubleshootingImportResult) => void;
	} = $props();

	let file = $state<File>();
	let preview = $state<TroubleshootingPreview>();
	let busy = $state<'preview' | 'commit'>();
	let error = $state('');
	const actionLabels = { create: 'Create', update: 'Update', error: 'Error' };

	async function previewFile() {
		if (!file || busy) return;
		busy = 'preview';
		preview = undefined;
		error = '';
		try {
			preview = await getTroubleshootingPreview(file);
		} catch (err) {
			error = formatApiError(err);
		} finally {
			busy = undefined;
		}
	}

	async function confirmImport() {
		if (!file || !preview?.valid || busy) return;
		busy = 'commit';
		error = '';
		try {
			onImported(await commitTroubleshootingImport(file));
			onClose();
		} catch (err) {
			preview = undefined;
			error = formatApiError(err);
		} finally {
			busy = undefined;
		}
	}
</script>

<Dialog
	open
	title="Bulk Import Guides"
	description="Upload a JSON or XLSX file and review what will be created or updated. Nothing is saved until you confirm."
	onOpenChange={(open) => {
		if (!open && !busy) onClose();
	}}
>
	<div class="grid gap-4">
		<div class="border-border grid gap-2 rounded-md border p-3">
			<label for="bulk-guide-file" class="text-sm font-semibold">Guide file</label>
			<input
				id="bulk-guide-file"
				type="file"
				accept={troubleshootingTransferFileTypes}
				disabled={!!busy}
				class="file:bg-muted file:text-foreground block max-w-full text-sm file:mr-2 file:rounded file:border-0 file:px-2 file:py-1"
				onchange={(event) => {
					file = event.currentTarget.files?.[0];
					preview = undefined;
					error = '';
				}}
			/>
			<Button
				variant="secondary"
				class="justify-self-start"
				disabled={!file || !!busy}
				onclick={previewFile}
			>
				{busy === 'preview' ? 'Previewing…' : 'Preview import'}
			</Button>
		</div>
		{#if error}<p class="text-destructive text-sm" role="alert">{error}</p>{/if}
		{#if preview}
			<section aria-label="Import preview" class="grid gap-2">
				<h3 class="text-sm font-semibold">Import preview</h3>
				<p class="text-muted-foreground text-xs" role="status">
					{#if preview.valid}
						{preview.rows.length} {preview.rows.length === 1 ? 'guide' : 'guides'} ready to import.
					{:else}
						Fix the errors in the file and preview it again before importing.
					{/if}
				</p>
				{#each preview.errors ?? [] as issue (issue)}
					<p class="text-destructive text-sm">{issue}</p>
				{/each}
				<ul class="border-border divide-border divide-y rounded-md border">
					{#each preview.rows as row (row.index)}
						<li class="p-3 text-sm">
							<div class="flex flex-wrap items-center gap-2">
								<Badge
									variant={row.action === 'error'
										? 'destructive'
										: row.action === 'update'
											? 'default'
											: 'outline'}
								>
									{actionLabels[row.action]}
								</Badge>
								<span class="font-medium">{row.title || `Guide ${row.index}`}</span>
								{#if row.id !== undefined}<span class="text-muted-foreground text-xs">
										ID {row.id}
									</span>{/if}
							</div>
							{#if row.reasons?.length}
								<ul class="text-destructive mt-1 list-disc pl-5 text-xs">
									{#each row.reasons as reason (reason)}
										<li>{reason}</li>
									{/each}
								</ul>
							{/if}
						</li>
					{/each}
				</ul>
			</section>
		{/if}
		<div class="dialog-actions">
			<Button variant="outline" disabled={!!busy} onclick={onClose}>Cancel</Button>
			<Button disabled={!!busy || !preview?.valid} onclick={confirmImport}>
				{busy === 'commit' ? 'Importing…' : 'Confirm import'}
			</Button>
		</div>
	</div>
</Dialog>
