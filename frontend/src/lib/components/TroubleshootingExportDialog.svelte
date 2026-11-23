<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { Search } from '@lucide/svelte';
	import { formatApiError } from '#lib/client.ts';
	import {
		listTroubleshootingGuides,
		severityLabel,
		severityTone,
		type TroubleshootingFilters,
		type TroubleshootingGuideListItem
	} from '#lib/troubleshooting.ts';
	import type { TroubleshootingSeverity } from '#lib/troubleshooting-registry.ts';
	import {
		downloadTroubleshootingExport,
		requestTroubleshootingExport,
		type TroubleshootingExportFormat,
		type TroubleshootingExportSelection
	} from '#lib/troubleshooting-transfer.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Input from '#lib/components/ui/input.svelte';

	let {
		initialGuideId,
		severities,
		filters,
		onClose
	}: {
		initialGuideId?: number;
		severities: TroubleshootingSeverity[];
		filters: TroubleshootingFilters;
		onClose: () => void;
	} = $props();

	let ids = $state(untrack(() => (initialGuideId === undefined ? [] : [initialGuideId])));
	let scope = $state<TroubleshootingExportSelection['scope']>('selected');
	let search = $state('');
	let guides = $state<TroubleshootingGuideListItem[]>([]);
	let loading = $state(true);
	let listError = $state('');
	let format = $state<TroubleshootingExportFormat>('json');
	let busy = $state(false);
	let error = $state('');
	const formats = [
		{ value: 'json', label: 'JSON' },
		{ value: 'xlsx', label: 'XLSX' },
		{ value: 'both', label: 'Both (ZIP)' }
	] as const;

	const visible = $derived.by(() => {
		const query = search.trim().toLowerCase();
		return guides.filter((guide) =>
			[guide.title, guide.error_code, severityLabel(guide.severity, severities)].some((value) =>
				value.toLowerCase().includes(query)
			)
		);
	});
	const canSelectVisible = $derived(visible.some((guide) => !ids.includes(guide.id)));
	const disabled = $derived(busy || (scope === 'selected' && !ids.length));
	let mounted = false;

	async function loadGuides() {
		listError = '';
		loading = true;
		const byId = new SvelteMap<number, TroubleshootingGuideListItem>();
		try {
			const pageSize = 100;
			const first = await listTroubleshootingGuides({ page: 1, page_size: pageSize });
			for (const guide of first.items) byId.set(guide.id, guide);
			for (let page = 2; mounted && page <= Math.ceil(first.total / pageSize); page++) {
				const result = await listTroubleshootingGuides({ page, page_size: pageSize });
				for (const guide of result.items) byId.set(guide.id, guide);
			}
			guides = [...byId.values()];
		} catch (err) {
			listError = formatApiError(err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		mounted = true;
		void loadGuides();
		return () => {
			mounted = false;
		};
	});

	function toggleGuide(id: number, checked: boolean) {
		ids = checked ? [...ids, id] : ids.filter((value) => value !== id);
	}

	async function download() {
		if (disabled) return;
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
	description="Choose guides and a file format. Exports keep guide IDs so you can import them as updates."
	onOpenChange={(open) => {
		if (!open && !busy) onClose();
	}}
>
	<div class="grid gap-4">
		<fieldset class="border-border grid gap-2 rounded-md border p-3 text-sm" disabled={busy}>
			<legend class="px-1 font-semibold">Guides</legend>
			<label class="flex items-center gap-2">
				<input type="radio" name="export-scope" value="selected" bind:group={scope} />
				Choose guides
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
		{#if scope === 'selected'}
			<section
				class="border-border grid gap-2 rounded-md border p-3"
				aria-label="Choose guides to export"
			>
				<div class="relative">
					<Search
						size={16}
						class="text-muted-foreground pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2"
					/>
					<Input
						type="search"
						class="pl-8"
						bind:value={search}
						disabled={busy}
						aria-label="Search guides to export"
						placeholder="Search by title, error code, or severity"
					/>
				</div>
				<div class="flex flex-wrap items-center justify-between gap-2">
					<p class="text-muted-foreground text-sm" role="status">
						{ids.length} selected {ids.length === 1 ? 'guide' : 'guides'}
					</p>
					<div class="flex flex-wrap gap-1">
						<Button
							variant="ghost"
							size="sm"
							disabled={busy || !canSelectVisible}
							onclick={() => (ids = [...new Set([...ids, ...visible.map((guide) => guide.id)])])}
						>
							Select all shown
						</Button>
						<Button
							variant="ghost"
							size="sm"
							disabled={busy || !ids.length}
							onclick={() => (ids = [])}
						>
							Clear selection
						</Button>
					</div>
				</div>
				{#if loading}
					<p class="text-muted-foreground p-3 text-sm" role="status">Loading guides…</p>
				{:else if listError}
					<div class="flex flex-wrap items-center justify-between gap-2">
						<p class="text-destructive text-sm" role="alert">{listError}</p>
						<Button variant="outline" size="sm" onclick={loadGuides}>Retry</Button>
					</div>
				{:else if !visible.length}
					<p class="text-muted-foreground p-3 text-sm">
						{search.trim() ? 'No guides match your search.' : 'No guides are available to export.'}
					</p>
				{:else}
					<ul
						class="border-border divide-border max-h-56 divide-y overflow-y-auto rounded-md border"
					>
						{#each visible as guide (guide.id)}
							{@const tone = severityTone(guide.severity, severities)}
							<li>
								<label class="hover:bg-muted/40 flex cursor-pointer items-start gap-3 px-3 py-2">
									<input
										type="checkbox"
										class="mt-0.5"
										checked={ids.includes(guide.id)}
										disabled={busy}
										onchange={(event) => toggleGuide(guide.id, event.currentTarget.checked)}
									/>
									<span class="grid min-w-0 gap-1">
										<span class="text-sm font-medium break-words">{guide.title}</span>
										<span class="flex flex-wrap gap-1">
											{#if guide.error_code}<Badge variant="outline">{guide.error_code}</Badge>{/if}
											<Badge
												variant={tone === 'destructive' || tone === 'warn' || tone === 'default'
													? tone
													: 'outline'}
											>
												{severityLabel(guide.severity, severities)}
											</Badge>
										</span>
									</span>
								</label>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		{/if}
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
			<Button onclick={download} {disabled}>
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
