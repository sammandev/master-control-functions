<script lang="ts">
	import { ArrowDown, ArrowUp, Trash2 } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import AlertDialog from './ui/alert-dialog.svelte';
	import Button from './ui/button.svelte';
	import Dialog from './ui/dialog.svelte';
	import Input from './ui/input.svelte';
	import Select from './ui/select.svelte';
	import Textarea from './ui/textarea.svelte';
	import * as Tabs from './ui/tabs';
	import { formatApiError } from '../client';
	import {
		deleteTroubleshootingCategory,
		deleteTroubleshootingSeverity,
		updateTroubleshootingCategory,
		updateTroubleshootingSeverity,
		type TroubleshootingCategory,
		type TroubleshootingSeverity
	} from '../troubleshooting-registry';

	type Kind = 'category' | 'severity';

	class EditableRow {
		key: string;
		label = $state('');
		description = $state('');
		sort_order = $state(0);
		built_in = $state(false);
		guide_count = $state(0);
		busy = $state(false);
		error = $state('');
		displayName = $state('');
		tone = $state('');

		constructor(item: TroubleshootingCategory | TroubleshootingSeverity) {
			this.key = item.key;
			this.label = item.label;
			this.description = item.description;
			this.sort_order = item.sort_order;
			this.built_in = item.built_in;
			this.guide_count = item.guide_count;
			this.displayName = item.label;
			this.tone = 'tone' in item ? item.tone : '';
		}
	}

	type CategoryRow = InstanceType<typeof EditableRow>;
	type SeverityRow = CategoryRow;
	type DeleteTarget = { kind: Kind; row: CategoryRow };

	const tones = ['destructive', 'warn', 'default', 'outline'].map((tone) => ({
		value: tone,
		label: tone
	}));

	let {
		open,
		categories,
		severities,
		onOpenChange,
		onChanged
	}: {
		open: boolean;
		categories: TroubleshootingCategory[];
		severities: TroubleshootingSeverity[];
		onOpenChange: (open: boolean) => void;
		onChanged: () => void | Promise<void>;
	} = $props();

	let tab = $state<Kind>('category');
	let categoryRows = $derived(categories.map((item) => new EditableRow(item)));
	let severityRows = $derived(severities.map((item) => new EditableRow(item)));
	let reordering = $state<Kind | null>(null);
	let deleteTarget = $state<DeleteTarget | null>(null);
	let replacement = $state('');
	let deleting = $state(false);
	let deleteError = $state('');
	let blockDeleteClose = false;

	function kindLabel(kind: Kind) {
		return kind === 'category' ? 'Category' : 'Severity';
	}

	function rowsFor(kind: Kind): (CategoryRow | SeverityRow)[] {
		return kind === 'category' ? categoryRows : severityRows;
	}

	function replacementOptions(target: DeleteTarget) {
		return rowsFor(target.kind)
			.filter((item) => item.key !== target.row.key)
			.map((item) => ({ value: item.key, label: item.label }));
	}

	function saveBody(row: CategoryRow | SeverityRow, kind: Kind) {
		const body = {
			label: row.label.trim(),
			description: row.description?.trim() ?? '',
			sort_order: row.sort_order
		};
		return kind === 'severity' ? { ...body, tone: (row as SeverityRow).tone } : body;
	}

	async function save(row: CategoryRow | SeverityRow, kind: Kind) {
		if (row.busy) return;
		row.busy = true;
		row.error = '';
		try {
			const body = saveBody(row, kind);
			if (kind === 'category') {
				await updateTroubleshootingCategory(row.key, body);
			} else {
				await updateTroubleshootingSeverity(row.key, body);
			}
			row.displayName = row.label.trim();
			toast.success(`${kindLabel(kind)} saved`);
			await onChanged();
		} catch (failure) {
			row.error = formatApiError(failure);
			toast.error(row.error);
		} finally {
			row.busy = false;
		}
	}

	async function move(row: CategoryRow | SeverityRow, kind: Kind, direction: -1 | 1) {
		if (reordering) return;
		const rows = rowsFor(kind).toSorted((a, b) => a.sort_order - b.sort_order);
		const index = rows.findIndex((item) => item.key === row.key);
		const neighbor = rows[index + direction];
		if (!neighbor) return;

		reordering = kind;
		row.busy = true;
		neighbor.busy = true;
		row.error = '';
		neighbor.error = '';
		const rowOrder = row.sort_order;
		const neighborOrder = neighbor.sort_order;
		try {
			if (kind === 'category') {
				await updateTroubleshootingCategory(row.key, {
					label: row.label.trim(),
					description: row.description?.trim() ?? '',
					sort_order: neighborOrder
				});
				await updateTroubleshootingCategory(neighbor.key, {
					label: neighbor.label.trim(),
					description: neighbor.description?.trim() ?? '',
					sort_order: rowOrder
				});
			} else {
				await updateTroubleshootingSeverity(row.key, {
					label: row.label.trim(),
					description: row.description?.trim() ?? '',
					sort_order: neighborOrder,
					tone: (row as SeverityRow).tone
				});
				await updateTroubleshootingSeverity(neighbor.key, {
					label: neighbor.label.trim(),
					description: neighbor.description?.trim() ?? '',
					sort_order: rowOrder,
					tone: (neighbor as SeverityRow).tone
				});
			}
			row.sort_order = neighborOrder;
			neighbor.sort_order = rowOrder;
			if (kind === 'category') {
				categoryRows = categoryRows.toSorted((a, b) => a.sort_order - b.sort_order);
			} else {
				severityRows = severityRows.toSorted((a, b) => a.sort_order - b.sort_order);
			}
			await onChanged();
		} catch (failure) {
			row.error = formatApiError(failure);
			toast.error(row.error);
		} finally {
			row.busy = false;
			neighbor.busy = false;
			reordering = null;
		}
	}

	function beginDelete(kind: Kind, row: CategoryRow | SeverityRow) {
		deleteTarget = { kind, row };
		replacement = '';
		deleteError = '';
	}

	function closeDelete(force = false) {
		if (deleting && !force) return;
		deleteTarget = null;
		replacement = '';
		deleteError = '';
	}

	async function confirmDelete() {
		if (deleting || !deleteTarget) return;
		if (!replacement || replacement === deleteTarget.row.key) {
			deleteError = `Choose a different ${deleteTarget.kind}`;
			blockDeleteClose = true;
			return;
		}
		deleting = true;
		deleteError = '';
		try {
			if (deleteTarget.kind === 'category') {
				await deleteTroubleshootingCategory(deleteTarget.row.key, replacement);
			} else {
				await deleteTroubleshootingSeverity(deleteTarget.row.key, replacement);
			}
			const label = deleteTarget.row.label;
			toast.success(`${kindLabel(deleteTarget.kind)} “${label}” deleted`);
			await onChanged();
			closeDelete(true);
		} catch (failure) {
			deleteError = formatApiError(failure);
			toast.error(deleteError);
		} finally {
			deleting = false;
		}
	}
</script>

<Dialog
	{open}
	{onOpenChange}
	title="Manage troubleshooting lists"
	description="Edit labels, descriptions, display order, and severity tones."
	class="max-w-[880px]"
>
	<Tabs.Root bind:value={tab} activationMode="automatic">
		<Tabs.List aria-label="Troubleshooting lists" class="mb-4 grid w-full grid-cols-2">
			<Tabs.Trigger value="category">Categories</Tabs.Trigger>
			<Tabs.Trigger value="severity">Severities</Tabs.Trigger>
		</Tabs.List>

		{#each ['category', 'severity'] as currentKind (currentKind)}
			<Tabs.Content value={currentKind} class="outline-none">
				<div class="space-y-3">
					{#each rowsFor(currentKind as Kind) as row, index (row.key)}
						{@const kind = currentKind as Kind}
						{@const rows = rowsFor(kind)}
						<section
							class="border-border bg-background rounded-md border p-3"
							aria-label={`${row.displayName} settings`}
						>
							<div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1.25fr)_auto]">
								<label class="grid gap-1 text-xs font-medium">
									<span>Label</span>
									<Input
										bind:value={row.label}
										aria-label={`${row.displayName} label`}
										disabled={row.busy}
									/>
								</label>
								<label class="grid gap-1 text-xs font-medium md:col-start-2 md:row-start-1">
									<span>Description</span>
									<Textarea
										bind:value={row.description}
										aria-label={`${row.displayName} description`}
										disabled={row.busy}
										class="min-h-8 resize-y py-1.5"
									/>
								</label>
								<div class="flex items-start gap-1 md:col-start-3 md:row-start-1">
									<Button
										variant="outline"
										size="icon"
										aria-label={`Move ${row.displayName} up`}
										disabled={row.busy || reordering !== null || index === 0}
										onclick={() => move(row, kind, -1)}><ArrowUp size={15} /></Button
									>
									<Button
										variant="outline"
										size="icon"
										aria-label={`Move ${row.displayName} down`}
										disabled={row.busy || reordering !== null || index === rows.length - 1}
										onclick={() => move(row, kind, 1)}><ArrowDown size={15} /></Button
									>
								</div>
							</div>

							<div class="mt-3 flex flex-wrap items-end justify-between gap-3">
								<div class="flex flex-wrap items-end gap-3">
									<span class="text-muted-foreground py-2 text-xs"
										>{row.guide_count} {row.guide_count === 1 ? 'guide' : 'guides'}</span
									>
									{#if kind === 'severity'}
										<label class="grid min-w-36 gap-1 text-xs font-medium">
											<span>Tone</span>
											<Select
												bind:value={(row as SeverityRow).tone}
												options={tones}
												ariaLabel={`Tone: ${(row as SeverityRow).tone}`}
												disabled={row.busy}
											/>
										</label>
									{/if}
								</div>
								<div class="flex items-center gap-2">
									{#if !row.built_in}
										<Button
											variant="outline"
											class="text-destructive"
											aria-label={`Delete ${row.displayName}`}
											disabled={row.busy}
											onclick={() => beginDelete(kind, row)}><Trash2 size={14} />Delete</Button
										>
									{/if}
									<Button
										disabled={row.busy || !row.label.trim()}
										aria-label={`Save ${row.displayName}`}
										onclick={() => save(row, kind)}>{row.busy ? 'Saving…' : 'Save'}</Button
									>
								</div>
							</div>
							{#if row.error}<p role="alert" class="text-destructive mt-2 text-sm">
									{row.error}
								</p>{/if}
						</section>
					{/each}
				</div>
			</Tabs.Content>
		{/each}
	</Tabs.Root>
</Dialog>

{#if deleteTarget}
	<AlertDialog
		open={true}
		onOpenChange={(next) => {
			if (next) return;
			if (blockDeleteClose) {
				blockDeleteClose = false;
				return;
			}
			closeDelete();
		}}
		title={`Delete ${deleteTarget.row.label}?`}
		description={`${deleteTarget.row.guide_count} ${deleteTarget.row.guide_count === 1 ? 'guide uses' : 'guides use'} this ${deleteTarget.kind}. Choose a remaining ${deleteTarget.kind} before deletion.`}
		confirmLabel={`Delete ${deleteTarget.kind}`}
		onConfirm={confirmDelete}
	>
		<div class="mt-4 grid gap-1 text-sm">
			<span class="font-medium">Reassign guides to</span>
			<Select
				bind:value={replacement}
				options={replacementOptions(deleteTarget)}
				ariaLabel={`Choose replacement ${deleteTarget.kind}`}
				required
				placeholder="Select replacement…"
			/>
		</div>
		{#if deleteError}<p role="alert" class="text-destructive mt-3 text-sm">{deleteError}</p>{/if}
	</AlertDialog>
{/if}
