<script lang="ts">
	import { untrack } from 'svelte';
	import { ArrowDown, ArrowUp, LockKeyhole, Plus, Trash2 } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import AlertDialog from './ui/alert-dialog.svelte';
	import Badge from './ui/badge.svelte';
	import Button from './ui/button.svelte';
	import Dialog from './ui/dialog.svelte';
	import Input from './ui/input.svelte';
	import Select from './ui/select.svelte';
	import Textarea from './ui/textarea.svelte';
	import * as Tabs from './ui/tabs';
	import { formatApiError } from '../client';
	import {
		createTroubleshootingCategory,
		createTroubleshootingSeverity,
		deleteTroubleshootingCategory,
		deleteTroubleshootingSeverity,
		reorderTroubleshootingCategory,
		reorderTroubleshootingSeverity,
		updateTroubleshootingCategory,
		updateTroubleshootingSeverity,
		type TroubleshootingCategory,
		type TroubleshootingSeverity
	} from '../troubleshooting-registry';

	type Kind = 'category' | 'severity';
	const labelKinds = ['category', 'severity'] as const;

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
		persistedLabel = '';
		persistedDescription = '';
		persistedTone = '';

		constructor(item: TroubleshootingCategory | TroubleshootingSeverity) {
			this.key = item.key;
			this.label = item.label;
			this.description = item.description;
			this.sort_order = item.sort_order;
			this.built_in = item.built_in;
			this.guide_count = item.guide_count;
			this.displayName = item.label;
			this.tone = 'tone' in item ? item.tone : '';
			this.persistedLabel = item.label;
			this.persistedDescription = item.description;
			this.persistedTone = this.tone;
		}

		reconcile(item: TroubleshootingCategory | TroubleshootingSeverity) {
			const labelIsDirty = this.label !== this.persistedLabel;
			const descriptionIsDirty = this.description !== this.persistedDescription;
			const toneIsDirty = this.tone !== this.persistedTone;
			if (!labelIsDirty) this.label = item.label;
			if (!descriptionIsDirty) this.description = item.description;
			if (!toneIsDirty && 'tone' in item) this.tone = item.tone;
			this.sort_order = item.sort_order;
			this.built_in = item.built_in;
			this.guide_count = item.guide_count;
			this.displayName = item.label;
			this.persistedLabel = item.label;
			this.persistedDescription = item.description;
			this.persistedTone = 'tone' in item ? item.tone : '';
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
	let categoryRows = $state(untrack(() => categories.map((item) => new EditableRow(item))));
	let severityRows = $state(untrack(() => severities.map((item) => new EditableRow(item))));
	let reordering = $state<Kind | null>(null);
	let deleteTarget = $state<DeleteTarget | null>(null);
	let replacement = $state('');
	let deleting = $state(false);
	let deleteError = $state('');
	let newLabels = $state({ category: '', severity: '' });
	let addErrors = $state({ category: '', severity: '' });
	let adding = $state<Kind | null>(null);
	let blockDeleteClose = false;

	function reconcileRows<T extends TroubleshootingCategory | TroubleshootingSeverity>(
		current: EditableRow[],
		items: T[]
	) {
		const byKey = new Map(current.map((row) => [row.key, row]));
		return items.map((item) => {
			const row = byKey.get(item.key) ?? new EditableRow(item);
			row.reconcile(item);
			return row;
		});
	}

	$effect(() => {
		const nextCategories = categories;
		untrack(() => {
			categoryRows = reconcileRows(categoryRows, nextCategories);
		});
	});

	$effect(() => {
		const nextSeverities = severities;
		untrack(() => {
			severityRows = reconcileRows(severityRows, nextSeverities);
		});
	});

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

	async function refreshAfterChange() {
		try {
			await onChanged();
		} catch (failure) {
			toast.error(formatApiError(failure));
		}
	}

	async function addLabel(kind: Kind) {
		const label = newLabels[kind].trim();
		if (adding || !label) return;
		adding = kind;
		addErrors[kind] = '';
		try {
			if (kind === 'category') {
				const created = await createTroubleshootingCategory({ label });
				categoryRows = [...categoryRows, new EditableRow(created)];
			} else {
				const created = await createTroubleshootingSeverity({ label, tone: 'outline' });
				severityRows = [...severityRows, new EditableRow(created)];
			}
			newLabels[kind] = '';
			await refreshAfterChange();
		} catch (failure) {
			addErrors[kind] = formatApiError(failure);
		} finally {
			adding = null;
		}
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
		let saved = false;
		try {
			const body = saveBody(row, kind);
			if (kind === 'category') {
				await updateTroubleshootingCategory(row.key, body);
			} else {
				await updateTroubleshootingSeverity(row.key, body);
			}
			row.persistedLabel = body.label;
			row.persistedDescription = body.description ?? '';
			row.persistedTone = kind === 'severity' ? (row as SeverityRow).tone : '';
			row.displayName = row.label.trim();
			toast.success(`${kindLabel(kind)} saved`);
			saved = true;
		} catch (failure) {
			row.error = formatApiError(failure);
			toast.error(row.error);
		} finally {
			row.busy = false;
		}
		if (saved) await refreshAfterChange();
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
		let moved = false;
		try {
			const result =
				kind === 'category'
					? await reorderTroubleshootingCategory(row.key, neighbor.key)
					: await reorderTroubleshootingSeverity(row.key, neighbor.key);
			for (const item of result.items) {
				const editable = rows.find((candidate) => candidate.key === item.key);
				editable?.reconcile(item);
			}
			if (kind === 'category') {
				categoryRows = categoryRows.toSorted((a, b) => a.sort_order - b.sort_order);
			} else {
				severityRows = severityRows.toSorted((a, b) => a.sort_order - b.sort_order);
			}
			moved = true;
		} catch (failure) {
			row.error = formatApiError(failure);
			toast.error(row.error);
		} finally {
			row.busy = false;
			neighbor.busy = false;
			reordering = null;
		}
		if (moved) await refreshAfterChange();
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
		const target = deleteTarget;
		try {
			if (target.kind === 'category') {
				await deleteTroubleshootingCategory(target.row.key, replacement);
			} else {
				await deleteTroubleshootingSeverity(target.row.key, replacement);
			}
			toast.success(`${kindLabel(target.kind)} “${target.row.label}” deleted`);
			if (target.kind === 'category') {
				categoryRows = categoryRows.filter((item) => item.key !== target.row.key);
			} else {
				severityRows = severityRows.filter((item) => item.key !== target.row.key);
			}
			closeDelete(true);
		} catch (failure) {
			deleteError = formatApiError(failure);
			toast.error(deleteError);
			return;
		} finally {
			deleting = false;
		}
		await refreshAfterChange();
	}
</script>

<Dialog
	{open}
	{onOpenChange}
	title="Manage Labels"
	description="Add and edit the categories and severities used on troubleshooting guides."
	class="max-w-[880px]"
>
	<Tabs.Root bind:value={tab} activationMode="automatic">
		<Tabs.List
			aria-label="Troubleshooting labels"
			class="bg-card sticky top-0 z-10 mb-4 grid w-full grid-cols-2 py-1"
		>
			<Tabs.Trigger value="category">
				Categories <span aria-hidden="true">({categoryRows.length})</span>
			</Tabs.Trigger>
			<Tabs.Trigger value="severity">
				Severities <span aria-hidden="true">({severityRows.length})</span>
			</Tabs.Trigger>
		</Tabs.List>

		{#each labelKinds as kind (kind)}
			<Tabs.Content value={kind} class="outline-none">
				<div class="flex flex-col gap-3">
					<form
						class="border-border bg-muted/30 grid gap-2 rounded-md border p-3"
						onsubmit={(event) => {
							event.preventDefault();
							void addLabel(kind);
						}}
					>
						<label for={`new-${kind}-label`} class="text-sm font-semibold">New {kind} label</label>
						<div class="flex flex-wrap items-center gap-2">
							<Input
								id={`new-${kind}-label`}
								class="min-w-40 flex-1"
								bind:value={newLabels[kind]}
								disabled={adding === kind}
								placeholder={kind === 'category' ? 'e.g. Network setup' : 'e.g. Needs review'}
							/>
							<Button type="submit" disabled={!!adding || !newLabels[kind].trim()}>
								<Plus size={14} />{adding === kind ? 'Adding…' : `Add ${kind}`}
							</Button>
						</div>
						{#if addErrors[kind]}<p role="alert" class="text-destructive text-sm">
								{addErrors[kind]}
							</p>{/if}
					</form>
					{#each rowsFor(kind) as row, index (row.key)}
						{@const rows = rowsFor(kind)}
						<section
							class="border-border bg-background overflow-hidden rounded-md border"
							aria-label={`${row.displayName} settings`}
						>
							<div
								class="border-border bg-muted/20 flex flex-wrap items-center justify-between gap-2 border-b px-3 py-2"
							>
								<div class="flex min-w-0 items-center gap-2">
									<span class="truncate text-sm font-semibold">{row.displayName}</span>
									<Badge variant={row.built_in ? 'outline' : 'default'}>
										{#if row.built_in}<LockKeyhole size={11} class="mr-1" />{/if}
										{row.built_in ? 'Built-in' : 'Custom'}
									</Badge>
								</div>
								<div class="flex items-center gap-2">
									<code class="text-muted-foreground max-w-48 truncate text-[11px]">{row.key}</code>
									<Badge variant="outline">
										{row.guide_count}
										{row.guide_count === 1 ? 'guide' : 'guides'}
									</Badge>
								</div>
							</div>
							<div class="p-3">
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
											onclick={() => move(row, kind, -1)}
										>
											<ArrowUp size={15} />
										</Button>
										<Button
											variant="outline"
											size="icon"
											aria-label={`Move ${row.displayName} down`}
											disabled={row.busy || reordering !== null || index === rows.length - 1}
											onclick={() => move(row, kind, 1)}
										>
											<ArrowDown size={15} />
										</Button>
									</div>
								</div>

								<div class="mt-3 flex flex-wrap items-end justify-between gap-3">
									<div class="flex flex-wrap items-end gap-3">
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
										<Button
											variant="outline"
											class="text-destructive"
											aria-label={`Delete ${row.displayName}`}
											title={row.built_in
												? 'Built-in values are migration-owned and cannot be deleted'
												: `Delete ${row.displayName}`}
											disabled={row.busy || row.built_in}
											onclick={() => beginDelete(kind, row)}
										>
											<Trash2 size={14} />Delete
										</Button>
										<Button
											disabled={row.busy || !row.label.trim()}
											aria-label={`Save ${row.displayName}`}
											onclick={() => save(row, kind)}
										>
											{row.busy ? 'Saving…' : 'Save'}
										</Button>
									</div>
								</div>
								{#if row.error}<p role="alert" class="text-destructive mt-2 text-sm">
										{row.error}
									</p>{/if}
							</div>
						</section>
					{:else}
						<p class="text-muted-foreground p-3 text-sm">
							No {kind === 'category' ? 'categories' : 'severities'} yet. Add a label to get started.
						</p>
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
		description={`${deleteTarget.row.guide_count} ${deleteTarget.row.guide_count === 1 ? 'guide uses' : 'guides use'} this ${deleteTarget.kind}. Select a replacement to use before permanently deleting the label.`}
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
