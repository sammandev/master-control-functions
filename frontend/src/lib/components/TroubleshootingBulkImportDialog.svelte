<script lang="ts">
	import { Download, FileUp, Upload } from '@lucide/svelte';
	import { formatApiError } from '#lib/client.ts';
	import {
		commitTroubleshootingImport,
		downloadTroubleshootingTemplate,
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
	let templateError = $state('');
	let templateNotice = $state('');
	let downloading = $state<'json' | 'xlsx'>();
	const actionLabels = { create: 'Create', update: 'Update', error: 'Error' };
	const templateFormats = ['json', 'xlsx'] as const;
	const counts = $derived.by(() => {
		const summary = { create: 0, update: 0, error: preview?.errors?.length ?? 0 };
		for (const row of preview?.rows ?? []) summary[row.action]++;
		return summary;
	});

	function selectFile(next?: File) {
		if (busy) return;
		preview = undefined;
		error = '';
		file = undefined;
		if (next && !/\.(json|xlsx)$/i.test(next.name)) {
			error = 'Choose a JSON or XLSX guide file.';
			return;
		}
		file = next;
	}

	function dropFile(event: DragEvent) {
		event.preventDefault();
		if (busy) return;
		const files = event.dataTransfer?.files;
		if (files && files.length > 1) {
			selectFile();
			error = 'Choose one guide file at a time.';
			return;
		}
		if (files?.length) selectFile(files[0]);
	}

	async function downloadTemplate(format: 'json' | 'xlsx') {
		if (downloading) return;
		downloading = format;
		templateError = '';
		templateNotice = '';
		try {
			await downloadTroubleshootingTemplate(format);
			templateNotice = `${format.toUpperCase()} template download started.`;
		} catch (err) {
			templateError = formatApiError(err);
		} finally {
			downloading = undefined;
		}
	}

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
	title="Import Guides"
	description="Upload a JSON or XLSX file and review what will be created or updated. Nothing is saved until you confirm."
	onOpenChange={(open) => {
		if (!open && !busy) onClose();
	}}
>
	<div class="grid gap-4">
		<section
			aria-label="Guide templates"
			class="border-border bg-muted/30 grid gap-2 rounded-md border p-3"
		>
			<h3 class="text-sm font-semibold">Start with a template</h3>
			<p class="text-muted-foreground text-xs">
				Use the template to format your guides and resolution steps.
			</p>
			<div class="flex flex-wrap gap-2">
				{#each templateFormats as format (format)}
					<Button
						variant="outline"
						size="sm"
						disabled={!!downloading}
						onclick={() => downloadTemplate(format)}
					>
						<Download size={14} />
						{downloading === format ? 'Downloading…' : `Download ${format.toUpperCase()} template`}
					</Button>
				{/each}
			</div>
			{#if templateError}<p class="text-destructive text-sm" role="alert">{templateError}</p>{/if}
			{#if templateNotice}<p class="text-muted-foreground text-xs" role="status">
					{templateNotice}
				</p>{/if}
		</section>
		<div
			class="border-border grid gap-3 rounded-md border border-dashed p-4"
			role="region"
			aria-label="Upload a guide file"
			ondragover={(event) => event.preventDefault()}
			ondrop={dropFile}
		>
			<div class="flex items-start gap-2">
				<Upload size={18} class="text-muted-foreground shrink-0" />
				<div>
					<h3 class="text-sm font-semibold">Guide file</h3>
					<p class="text-muted-foreground text-xs">
						Drag and drop one JSON or XLSX file, or browse below.
					</p>
				</div>
			</div>
			<input
				id="bulk-guide-file"
				type="file"
				accept={troubleshootingTransferFileTypes}
				disabled={!!busy}
				aria-label="Guide file"
				class="peer sr-only"
				onchange={(event) => {
					selectFile(event.currentTarget.files?.[0]);
					event.currentTarget.value = '';
				}}
			/>
			<label
				for="bulk-guide-file"
				class="border-border bg-card hover:bg-accent peer-focus-visible:ring-ring cursor-pointer justify-self-start rounded-md border px-3 py-1.5 text-sm font-medium peer-focus-visible:ring-2 peer-disabled:pointer-events-none peer-disabled:opacity-50"
			>
				Browse files
			</label>
			{#if file}
				<div class="bg-muted/40 flex items-center gap-2 rounded px-2 py-1 text-sm">
					<FileUp size={15} class="shrink-0" />
					<span class="min-w-0 grow truncate" title={file.name}>{file.name}</span>
					<span class="text-muted-foreground text-xs whitespace-nowrap">
						{file.size < 1024 ? `${file.size} B` : `${(file.size / 1024).toFixed(1)} KB`}
					</span>
					<Button variant="ghost" size="sm" disabled={!!busy} onclick={() => selectFile()}>
						Remove
					</Button>
				</div>
			{/if}
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
				<div class="flex flex-wrap gap-2">
					<Badge variant="outline">Create: {counts.create}</Badge>
					<Badge variant="default">Update: {counts.update}</Badge>
					<Badge variant={counts.error ? 'destructive' : 'outline'}>Error: {counts.error}</Badge>
				</div>
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
				<ul class="border-border divide-border max-h-64 divide-y overflow-y-auto rounded-md border">
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
