<script lang="ts">
	import { resolve } from '$app/paths';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import {
		listLinks,
		attachLink,
		updateLink,
		detachLink,
		ApiError,
		type LinkRow,
		type LinkKind
	} from '#lib/links.ts';
	import { deviceLinkClass, devicePath } from '#lib/devices.ts';
	import {
		listDefinitions,
		deleteDefinition,
		createDefinition,
		type DefinitionKind,
		type DefinitionInput
	} from '#lib/definitions.ts';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import Combobox from '#lib/components/ui/combobox.svelte';
	import Tooltip from '#lib/components/ui/tooltip.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';
	import Label from '#lib/components/ui/label.svelte';
	import Checkbox from '#lib/components/ui/checkbox.svelte';
	import SortHead from '#lib/components/SortHead.svelte';
	import { nextSort, sortRows, type SortDir } from '#lib/table-sort.ts';
	import RadioGroup from '#lib/components/ui/radio-group.svelte';
	import RadioGroupItem from '#lib/components/ui/radio-group-item.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import * as DropdownMenu from '#lib/components/ui/dropdown-menu/index.ts';
	import { EllipsisVertical, Plus, Search } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import {
		Table,
		TableHeader,
		TableBody,
		TableRow,
		TableHead,
		TableCell
	} from '#lib/components/ui/table/index.ts';

	// One tab of the function detail page (ticket 04): the link rows of one
	// dictionary for one function. kind picks the link endpoint, defKind the
	// dictionary the searchable dropdown and inline creation write to.
	// Fully on the wrapper layer as of the polish step (ticket 08): the
	// scoped style block is gone.
	let {
		fnId,
		kind,
		defKind,
		noun,
		defaultDevice = '',
		onCountChange
	}: {
		fnId: number;
		kind: LinkKind;
		defKind: DefinitionKind;
		noun: string;
		defaultDevice?: string;
		onCountChange?: (id: number, count: number) => void;
	} = $props();

	// ---- list state -------------------------------------------------------
	let links = $state<LinkRow[]>([]);
	let sortKey = $state('');
	let sortDir = $state<SortDir>('asc');
	let linkSearch = $state('');
	const matching = $derived(
		links.filter((row) =>
			`${row.name} ${row.device_name}`.toLowerCase().includes(linkSearch.trim().toLowerCase())
		)
	);
	const requiredCount = $derived(links.filter((row) => row.required).length);
	const unresolvedCount = $derived(links.filter((row) => !row.resolved).length);
	const view = $derived(
		sortKey === ''
			? matching
			: sortRows(
					matching,
					(row) => {
						if (sortKey === 'name') return row.name;
						if (sortKey === 'device') return row.device_name;
						if (sortKey === 'required') return Number(row.required);
						return row.position;
					},
					sortDir
				)
	);
	function applySort(column: string) {
		({ key: sortKey, dir: sortDir } = nextSort(sortKey, sortDir, column));
	}
	let loading = $state(false);
	let loadError = $state('');

	// ---- row edit state (in place, orphan links included) ------------------
	let editingId = $state<number | null>(null);
	let editForm = $state<{ device_name: string; position: string; required: boolean }>({
		device_name: '',
		position: '',
		required: false
	});
	let editError = $state('');
	let savingRow = $state(false);

	// ---- add-link dialog state ---------------------------------------------
	let addOpen = $state(false);
	// mode 'pick': searchable dropdown over existing definitions.
	// mode 'inline': create a new definition and link it in one flow.
	// mode 'resolve' (ticket 05): create the missing definition behind an
	// unresolved link row: the link itself already exists.
	let addMode = $state<'pick' | 'inline' | 'resolve'>('pick');
	let pickName = $state('');
	let inlineForm = $state<DefinitionInput>(emptyInlineForm());
	let addDevice = $state('');
	let addPosition = $state('');
	let addRequired = $state(false);
	let fieldErrors = $state<Record<string, string>>({});
	let submitError = $state('');
	let attaching = $state(false);

	function emptyInlineForm(): DefinitionInput {
		return {
			name: '',
			description: '',
			type: 'string',
			unit: '',
			default_value: '',
			allowed_values: '',
			cardinality: 'SINGLE',
			value_pattern: '',
			required: false
		};
	}

	async function load(id: number) {
		loading = true;
		loadError = '';
		try {
			links = await listLinks(id, kind);
			onCountChange?.(id, links.length);
		} catch (e) {
			links = [];
			loadError = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		// fnId is read synchronously so the effect re-runs when the route
		// param changes under a reused component instance.
		const id = fnId;
		load(id);
	});

	// Server-side search for the pick-mode Combobox (ui/combobox.svelte
	// owns the debounce): the picker only needs display fields, the
	// attach call goes by name.
	async function searchDefinitions(term: string) {
		const res = await listDefinitions(defKind, { search: term || undefined, page_size: 20 });
		return res.items.map((d) => ({
			value: d.name,
			// Device-scoped definitions share the name (ADR-0009); the label
			// disambiguates the scope and the key keeps the each block unique.
			label: d.device_name ? `${d.name} · ${d.device_name}` : d.name,
			hint: `${d.type}${d.unit ? ` · ${d.unit}` : ''}`,
			key: `${d.name}::${d.device_name}`
		}));
	}

	function openAdd() {
		addMode = 'pick';
		pickName = '';
		inlineForm = emptyInlineForm();
		addDevice = defaultDevice;
		addPosition = '';
		addRequired = false;
		fieldErrors = {};
		submitError = '';
		addOpen = true;
	}

	// One-action resolution from an orphan row (ticket 05, ADR-0003):
	// open the dialog in resolve mode with the missing definition's name
	// prefilled; submitting creates only the definition, the existing
	// (unresolved) link then resolves by name.
	function resolveOrphan(row: LinkRow) {
		addMode = 'resolve';
		inlineForm = emptyInlineForm();
		inlineForm.name = row.name;
		addDevice = row.device_name;
		addPosition = row.position === null ? '' : String(row.position);
		addRequired = row.required;
		fieldErrors = {};
		submitError = '';
		addOpen = true;
	}

	// Attach by picking (name) or inline (definition object): the server
	// validates and, for inline, creates + links atomically. In resolve
	// mode only the definition is created.
	async function attach() {
		attaching = true;
		fieldErrors = {};
		submitError = '';
		const position = addPosition.trim() === '' ? null : Number(addPosition);
		try {
			if (addMode === 'resolve') {
				await createDefinition(defKind, { ...inlineForm, device_name: addDevice });
			} else if (addMode === 'pick') {
				await attachLink(fnId, kind, {
					name: pickName,
					device_name: addDevice,
					position,
					required: addRequired
				});
			} else {
				await attachLink(fnId, kind, {
					definition: { ...inlineForm, device_name: addDevice },
					device_name: addDevice,
					position,
					required: addRequired
				});
			}
			addOpen = false;
			await load(fnId);
			if (addMode === 'resolve') {
				toast.success('Definition created');
			} else {
				toast.success(`${noun} link attached`);
			}
		} catch (e) {
			if (e instanceof ApiError) {
				fieldErrors = e.fields;
				submitError = e.message;
			} else {
				submitError = e instanceof Error ? e.message : String(e);
			}
		} finally {
			attaching = false;
		}
	}

	function startEdit(row: LinkRow) {
		editingId = row.id;
		editForm = {
			device_name: row.device_name,
			position: row.position === null ? '' : String(row.position),
			required: row.required
		};
		editError = '';
	}

	async function saveEdit(row: LinkRow) {
		savingRow = true;
		editError = '';
		const raw = editForm.position.trim();
		try {
			const updated = await updateLink(fnId, kind, row.id, {
				device_name: editForm.device_name,
				position: raw === '' ? null : Number(raw),
				required: editForm.required
			});
			links = links.map((l) => (l.id === updated.id ? updated : l));
			editingId = null;
			toast.success('Link updated');
		} catch (e) {
			editError = e instanceof Error ? e.message : String(e);
		} finally {
			savingRow = false;
		}
	}

	// ---- confirm dialog state (detach / force delete) -----------------------
	// The row awaiting confirmation; kept after the dialog closes so the
	// description stays stable until the next confirm opens.
	let detachRow = $state<LinkRow | null>(null);
	let detachOpen = $state(false);
	let forceRow = $state<LinkRow | null>(null);
	let forceOpen = $state(false);

	function requestDetach(row: LinkRow) {
		detachRow = row;
		detachOpen = true;
	}

	// Detach removes only the link row; the definition always survives.
	async function confirmDetach() {
		if (!detachRow) return;
		try {
			await detachLink(fnId, kind, detachRow.id);
			await load(fnId);
			toast.success('Link detached');
		} catch (e) {
			loadError = e instanceof Error ? e.message : String(e);
		}
	}

	// Delete-with-warning: a definition still referenced by links is
	// protected by the server (409 + usage count); forcing turns those
	// links into orphans, which the user must confirm through the
	// AlertDialog (opened from the 409 response).
	async function removeDefinition(row: LinkRow) {
		if (row.definition_id === null) return;
		try {
			await deleteDefinition(defKind, row.definition_id);
			await load(fnId);
			toast.success('Definition deleted');
		} catch (e) {
			if (e instanceof ApiError && e.status === 409) {
				forceRow = row;
				forceOpen = true;
			} else {
				loadError = e instanceof Error ? e.message : String(e);
			}
		}
	}

	async function confirmForceRemove() {
		if (!forceRow || forceRow.definition_id === null) return;
		try {
			await deleteDefinition(defKind, forceRow.definition_id, true);
			await load(fnId);
			toast.success('Definition deleted');
		} catch (e) {
			loadError = e instanceof Error ? e.message : String(e);
		}
	}
</script>

<!-- Mounts flush inside the detail page's tabs Card: the toolbar and
     table carry the card's horizontal padding, the table itself is
     borderless so the Card frames it. -->
<div>
	<div class="link-toolbar">
		<div class="link-title">
			<strong>{noun === 'parameter' ? 'Parameter links' : 'Criteria links'}</strong>
			<p>{loading ? 'Loading…' : `${links.length} linked definitions`}</p>
			{#if !loading && links.length}
				<div class="link-stats">
					<Badge variant="outline">{requiredCount} required</Badge>
					<Badge variant="outline">{links.length - requiredCount} optional</Badge>
					{#if unresolvedCount}<Badge variant="warn">{unresolvedCount} unresolved</Badge>{/if}
				</div>
			{/if}
		</div>
		<div class="link-actions">
			<div class="link-search">
				<Search size={16} /><Input
					bind:value={linkSearch}
					aria-label={`Search ${noun} links`}
					placeholder="Search name or device"
				/>
			</div>
			<Button onclick={openAdd}><Plus data-icon="inline-start" /> Attach {noun}</Button>
		</div>
	</div>

	{#if loadError}
		<div class="px-4"><div class="banner-error" role="alert">{loadError}</div></div>
	{/if}

	<Table class="border-0">
		<TableHeader>
			<TableRow>
				<SortHead
					label="Position"
					col="position"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('position')}
				/>
				<SortHead
					label="Name"
					col="name"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('name')}
				/>
				<SortHead
					label="Device"
					col="device"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('device')}
				/>
				<SortHead
					label="Required"
					col="required"
					active={sortKey}
					dir={sortDir}
					onclick={() => applySort('required')}
				/>
				<TableHead></TableHead>
			</TableRow>
		</TableHeader>
		<TableBody>
			{#each view as row (row.id)}
				{#if editingId === row.id}
					<TableRow class="bg-popover">
						<TableCell>
							<Input type="number" min="1" bind:value={editForm.position} aria-label="Position" />
						</TableCell>
						<TableCell>{row.name}</TableCell>
						<TableCell>
							<DevicePicker
								device={editForm.device_name}
								blankLabel="Global (no Device / Interface)"
								fullWidth
								disabled={savingRow}
								onValueChange={(value) => (editForm.device_name = value)}
							/>
						</TableCell>
						<TableCell>
							<Checkbox
								bind:checked={editForm.required}
								aria-label="Toggle required for {row.name}"
							/>
						</TableCell>
						<TableCell class="text-right whitespace-nowrap">
							<Button size="sm" disabled={savingRow} onclick={() => saveEdit(row)}>
								{savingRow ? 'Saving…' : 'Save'}
							</Button>
							<Button size="sm" variant="outline" onclick={() => (editingId = null)}>Cancel</Button>
						</TableCell>
					</TableRow>
					{#if editError}
						<TableRow>
							<TableCell colspan={5} class="banner-error" aria-live="assertive"
								>{editError}</TableCell
							>
						</TableRow>
					{/if}
				{:else}
					<TableRow>
						<TableCell class="text-muted-foreground text-xs">{row.position ?? 'Not set'}</TableCell>
						<TableCell>
							{row.name}
							{#if !row.resolved}
								<Tooltip content="No {noun} definition with this name exists">
									{#snippet children(tp)}
										<!-- tabindex makes the hint keyboard-reachable (tooltip opens on focus); the badge itself is not an action target. -->
										<Badge variant="warn" {...tp} tabindex={0}>unresolved</Badge>
									{/snippet}
								</Tooltip>
							{/if}
						</TableCell>
						<TableCell>
							{#if row.device_name}
								<!-- Function→device back-path (ADR-0009 view). -->
								<a
									class={deviceLinkClass}
									href={resolve('/devices/[name]', { name: devicePath(row.device_name) })}
									title={`Browse device ${row.device_name}`}
								>
									{row.device_name}
								</a>
							{:else}
								Not set
							{/if}
						</TableCell>
						<TableCell>
							<Checkbox
								checked={row.required}
								aria-label="Toggle required for {row.name}"
								onCheckedChange={(checked) => {
									// Required toggle edits in place, orphan links included.
									editForm = {
										device_name: row.device_name,
										position: row.position === null ? '' : String(row.position),
										required: checked
									};
									saveEdit(row);
								}}
							/>
						</TableCell>
						<TableCell class="text-right">
							<DropdownMenu.Root>
								<DropdownMenu.Trigger>
									{#snippet child({ props })}
										<Button
											{...props}
											variant="ghost"
											size="icon"
											aria-label={`Actions for ${row.name}`}
										>
											<EllipsisVertical class="size-4" />
										</Button>
									{/snippet}
								</DropdownMenu.Trigger>
								<DropdownMenu.Content align="end" class="w-44">
									{#if !row.resolved}
										<DropdownMenu.Item onclick={() => resolveOrphan(row)}>
											Create definition
										</DropdownMenu.Item>
									{/if}
									<DropdownMenu.Item onclick={() => startEdit(row)}>Edit</DropdownMenu.Item>
									<DropdownMenu.Item
										disabled={row.definition_id === null}
										onclick={() => removeDefinition(row)}
									>
										Delete definition
									</DropdownMenu.Item>
									<DropdownMenu.Separator />
									<DropdownMenu.Item variant="destructive" onclick={() => requestDetach(row)}>
										Detach
									</DropdownMenu.Item>
								</DropdownMenu.Content>
							</DropdownMenu.Root>
						</TableCell>
					</TableRow>
				{/if}
			{:else}
				<TableRow>
					<TableCell colspan={5} class="banner-empty text-center" aria-live="polite">
						{loading
							? 'Loading…'
							: linkSearch && links.length
								? `No ${noun} links match your search.`
								: `No ${noun} links yet. Attach an existing definition or create one inline.`}
					</TableCell>
				</TableRow>
			{/each}
		</TableBody>
	</Table>
</div>

<Dialog
	open={addOpen}
	onOpenChange={(o) => (addOpen = o)}
	title={addMode === 'resolve' ? `Create missing ${noun} definition` : `Attach ${noun}`}
	description={addMode === 'resolve'
		? 'Creating this definition resolves the unresolved link in place: the link row keeps its device and position.'
		: ''}
>
	{#if addMode !== 'resolve'}
		<RadioGroup
			value={addMode}
			onValueChange={(v) => (addMode = v as 'pick' | 'inline')}
			class="mb-3 flex flex-row gap-4"
		>
			<div class="flex items-center gap-1.5">
				<RadioGroupItem id="add-mode-pick" value="pick" />
				<Label for="add-mode-pick">Pick existing definition</Label>
			</div>
			<div class="flex items-center gap-1.5">
				<RadioGroupItem id="add-mode-inline" value="inline" />
				<Label for="add-mode-inline">Create new definition</Label>
			</div>
		</RadioGroup>
	{/if}

	{#if submitError}
		<div class="banner-error" role="alert">{submitError}</div>
	{/if}

	<form
		onsubmit={(e) => {
			e.preventDefault();
			attach();
		}}
	>
		{#if addMode === 'pick'}
			<div class="field">
				<Label for="pick-name">Definition</Label>
				<Combobox
					id="pick-name"
					bind:value={pickName}
					search={searchDefinitions}
					placeholder="Type to search {noun} definitions…"
					emptyActionLabel="Create new definition"
					onEmptyAction={() => (addMode = 'inline')}
				/>
				{#if fieldErrors.name}<span class="field-error">{fieldErrors.name}</span>{/if}
			</div>
		{:else}
			<div class="field">
				<Label for="inline-name">Name</Label>
				<Input id="inline-name" type="text" bind:value={inlineForm.name} />
				{#if addMode === 'resolve'}
					{#if fieldErrors.name}<span class="field-error">{fieldErrors.name}</span>{/if}
				{:else if fieldErrors['definition.name']}
					<span class="field-error">{fieldErrors['definition.name']}</span>
				{/if}
			</div>
			<div class="field">
				<Label for="inline-type">Type</Label>
				<Select
					id="inline-type"
					ariaLabel="Type"
					bind:value={inlineForm.type}
					options={[
						{ value: 'int', label: 'int' },
						{ value: 'double', label: 'double' },
						{ value: 'string', label: 'string' },
						{ value: 'list[int]', label: 'list[int]' }
					]}
				/>
				{#if addMode === 'resolve'}
					{#if fieldErrors.type}<span class="field-error">{fieldErrors.type}</span>{/if}
				{:else if fieldErrors['definition.type']}
					<span class="field-error">{fieldErrors['definition.type']}</span>
				{/if}
			</div>
			<div class="field">
				<Label for="inline-unit">Unit</Label>
				<Input id="inline-unit" type="text" bind:value={inlineForm.unit} />
			</div>
			<div class="field">
				<Label for="inline-desc">Description</Label>
				<Textarea id="inline-desc" rows={3} bind:value={inlineForm.description} />
			</div>
		{/if}

		<div class="field">
			<Label for="add-device">Device / Interface</Label>
			<DevicePicker
				id="add-device"
				device={addDevice}
				label="Device / Interface"
				blankLabel="Global (no Device / Interface)"
				fullWidth
				disabled={addMode === 'resolve' || attaching}
				onValueChange={(value) => (addDevice = value)}
			/>
			{#if fieldErrors.device_name}<span class="field-error">{fieldErrors.device_name}</span>{/if}
		</div>
		{#if addMode !== 'resolve'}
			<div class="field">
				<Label for="add-position">Position</Label>
				<Input id="add-position" type="number" min="1" bind:value={addPosition} />
				{#if fieldErrors.position}<span class="field-error">{fieldErrors.position}</span>{/if}
			</div>
			<div class="field">
				<div class="flex min-h-10 items-center gap-3">
					<Checkbox id="add-required" bind:checked={addRequired} />
					<Label for="add-required" class="cursor-pointer py-2">Required</Label>
				</div>
			</div>
		{/if}

		<div class="mt-4 flex justify-end gap-2">
			<Button variant="outline" type="button" onclick={() => (addOpen = false)}>Cancel</Button>
			<Button type="submit" disabled={attaching}>
				{attaching ? 'Attaching…' : addMode === 'resolve' ? 'Create' : 'Attach'}
			</Button>
		</div>
	</form>
</Dialog>

<AlertDialog
	open={detachOpen}
	onOpenChange={(o) => (detachOpen = o)}
	title="Detach link"
	description={detachRow
		? `Detach ${detachRow.name} from this function? The ${noun} definition itself is kept.`
		: ''}
	confirmLabel="Detach"
	onConfirm={confirmDetach}
/>

<AlertDialog
	open={forceOpen}
	onOpenChange={(o) => (forceOpen = o)}
	title="Force delete definition"
	description={forceRow
		? `${forceRow.name} is referenced by other function link(s). Force delete? The references will remain but become unresolved.`
		: ''}
	confirmLabel="Force delete"
	onConfirm={confirmForceRemove}
/>

<style>
	.link-toolbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 12px 20px;
		padding: 20px 0 14px;
	}
	.link-title {
		min-width: 0;
	}
	.link-stats,
	.link-actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
	}
	.link-stats {
		margin-top: 8px;
	}
	.link-actions {
		gap: 8px;
		flex-wrap: nowrap;
		margin-left: auto;
		max-width: 100%;
	}
	.link-actions :global(button) {
		flex-shrink: 0;
	}
	.link-search {
		position: relative;
		flex: 1 1 240px;
		min-width: 160px;
		width: 240px;
	}
	.link-search > :global(svg) {
		position: absolute;
		top: 8px;
		left: 10px;
		color: var(--muted-foreground);
		pointer-events: none;
	}
	.link-search :global(input) {
		padding-left: 32px;
	}
	.link-toolbar strong {
		font-size: 13px;
	}
	.link-toolbar p {
		color: var(--muted-foreground);
		font-size: 13px;
		margin: 3px 0 0;
	}
	@media (max-width: 600px) {
		.link-toolbar {
			align-items: start;
			flex-wrap: wrap;
		}
		.link-actions {
			width: 100%;
			flex-wrap: wrap;
		}
		.link-search {
			min-width: 0;
			flex-basis: 180px;
		}
	}
</style>
