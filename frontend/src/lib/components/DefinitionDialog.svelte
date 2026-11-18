<script lang="ts">
	import {
		createDefinition,
		updateDefinition,
		ApiError,
		type DefinitionRow,
		type DefinitionInput,
		type DefinitionKind
	} from '#lib/definitions.ts';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Checkbox from '#lib/components/ui/checkbox.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';
	import Label from '#lib/components/ui/label.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import { toast } from 'svelte-sonner';

	// The dictionary create/edit form, extracted from DictionaryPage so the
	// device page can open the same modal with the device prefilled and
	// visibly locked (ticket 05). kind picks the endpoint; everything else
	// is identical.
	let {
		kind,
		title,
		open = $bindable(false),
		editRow = null,
		createFrom = null,
		prefillDevice = '',
		lockDevice = false,
		onsaved
	}: {
		kind: DefinitionKind;
		title: string;
		open?: boolean;
		editRow?: DefinitionRow | null;
		createFrom?: DefinitionRow | null;
		prefillDevice?: string;
		lockDevice?: boolean;
		onsaved?: (row: DefinitionRow) => void | Promise<void>;
	} = $props();

	// Form shape for the create/edit modal. Scope (device_name, ADR-0009)
	// is part of a definition's identity, so it is always a definite string
	// here and always sent: an edit that omitted it would silently
	// downgrade a scoped definition to Global (the API is a full-replace
	// PUT — see TestParameterDefinitionUpdateScopeSemantics).
	type DefinitionForm = DefinitionInput & { device_name: string };

	let form = $state<DefinitionForm>(emptyForm());
	let fieldErrors = $state<Record<string, string>>({});
	let submitError = $state('');
	let saving = $state(false);

	function emptyForm(): DefinitionForm {
		return {
			name: '',
			device_name: '',
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

	// Every open resets the form: edit mode from the row's current values,
	// create mode from the prefill device (blank = Global).
	$effect(() => {
		if (open) {
			const source = editRow ?? createFrom;
			form = source
				? {
						name: source.name,
						device_name: editRow ? source.device_name : prefillDevice,
						description: source.description,
						type: source.type,
						unit: source.unit,
						default_value: source.default_value,
						allowed_values: source.allowed_values,
						cardinality: source.cardinality,
						value_pattern: source.value_pattern,
						required: source.required
					}
				: { ...emptyForm(), device_name: prefillDevice };
			fieldErrors = {};
			submitError = '';
		}
	});

	async function save() {
		if (saving) return;
		saving = true;
		fieldErrors = {};
		submitError = '';
		try {
			const input = lockDevice
				? { ...form, device_name: editRow?.device_name ?? prefillDevice }
				: form;
			let saved: DefinitionRow;
			if (editRow) {
				saved = await updateDefinition(kind, editRow.id, input);
				toast.success(`${editRow.name} updated`);
			} else {
				saved = await createDefinition(kind, input);
				toast.success(`${form.name} created`);
			}
			await onsaved?.(saved);
			open = false;
		} catch (e) {
			if (e instanceof ApiError) {
				fieldErrors = e.fields;
				submitError = e.message;
			} else {
				submitError = e instanceof Error ? e.message : String(e);
			}
		} finally {
			saving = false;
		}
	}
</script>

<Dialog
	{open}
	onOpenChange={(next) => {
		if (!saving) open = next;
	}}
	title={editRow ? `Edit ${editRow.name}` : `New ${title.toLowerCase()} entry`}
>
	<p class="text-muted-foreground mb-3 text-xs">
		{form.device_name ? `Device: ${form.device_name}` : 'Global - shared across Devices'}
	</p>
	{#if submitError}
		<div class="banner-error" role="alert">{submitError}</div>
	{/if}
	<form
		class="crud-form"
		onsubmit={(e) => {
			e.preventDefault();
			save();
		}}
	>
		<div class="field">
			<Label for="d-name">Name</Label>
			<Input id="d-name" bind:value={form.name} />
			{#if fieldErrors.name}<span class="field-error">{fieldErrors.name}</span>{/if}
		</div>
		<div class="field">
			<Label for="d-scope">Device / Interface</Label>
			<!-- Device (row attribute, ADR-0009): blank = Global; a device
			     overrides the global definition for that device's links.
			     Locked when opened from a device page in context. -->
			<DevicePicker
				id="d-scope"
				device={form.device_name}
				label="Device / Interface"
				blankLabel="Global (no Device / Interface)"
				fullWidth
				disabled={lockDevice || saving}
				onValueChange={(value) => (form.device_name = value)}
			/>
			{#if fieldErrors.device_name}<span class="field-error">{fieldErrors.device_name}</span>{/if}
		</div>
		<div class="field field-wide">
			<Label for="d-description">Description</Label>
			<Textarea id="d-description" rows={3} bind:value={form.description} />
			{#if fieldErrors.description}<span class="field-error">{fieldErrors.description}</span>{/if}
		</div>
		<div class="field">
			<Label for="d-type">Type</Label>
			<Select
				id="d-type"
				ariaLabel="Type"
				bind:value={form.type}
				options={[
					{ value: 'int', label: 'int' },
					{ value: 'double', label: 'double' },
					{ value: 'string', label: 'string' },
					{ value: 'list[int]', label: 'list[int]' }
				]}
				onchange={() => {
					// A type change may invalidate default/allowed members;
					// clear them rather than leave stale values in place.
					form.default_value = '';
					form.allowed_values = '';
				}}
			/>
			{#if fieldErrors.type}<span class="field-error">{fieldErrors.type}</span>{/if}
		</div>
		<div class="field">
			<Label for="d-unit">Unit</Label>
			<Input id="d-unit" bind:value={form.unit} />
			{#if fieldErrors.unit}<span class="field-error">{fieldErrors.unit}</span>{/if}
		</div>
		<div class="field">
			<Label for="d-allowed">Allowed values (&#123;v1,v2,&#125;…)</Label>
			<Textarea
				id="d-allowed"
				rows={2}
				bind:value={form.allowed_values}
				placeholder="&#123;20,40,80&#125;"
			/>
			{#if fieldErrors.allowed_values}<span class="field-error">{fieldErrors.allowed_values}</span
				>{/if}
		</div>
		<div class="field">
			<Label for="d-default">Default value</Label>
			<Textarea id="d-default" rows={2} bind:value={form.default_value} />
			{#if fieldErrors.default_value}<span class="field-error">{fieldErrors.default_value}</span
				>{/if}
		</div>
		<div class="field">
			<Label for="d-cardinality">Cardinality</Label>
			<Select
				id="d-cardinality"
				ariaLabel="Cardinality"
				bind:value={form.cardinality}
				options={[
					{ value: 'SINGLE', label: 'SINGLE' },
					{ value: 'MULTIPLE', label: 'MULTIPLE' }
				]}
			/>
			{#if fieldErrors.cardinality}<span class="field-error">{fieldErrors.cardinality}</span>{/if}
		</div>
		<div class="field field-wide">
			<Label for="d-pattern">Value pattern</Label>
			<Textarea id="d-pattern" rows={3} bind:value={form.value_pattern} />
			{#if fieldErrors.value_pattern}<span class="field-error">{fieldErrors.value_pattern}</span
				>{/if}
		</div>
		<div class="field">
			<div class="flex min-h-10 items-center gap-3">
				<Checkbox id="d-required" bind:checked={form.required} />
				<Label for="d-required" class="cursor-pointer py-2">Required</Label>
			</div>
			{#if fieldErrors.required}<span class="field-error">{fieldErrors.required}</span>{/if}
		</div>
		<div class="dialog-actions">
			<Button variant="outline" disabled={saving} onclick={() => (open = false)}>Cancel</Button>
			<Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</Button>
		</div>
	</form>
</Dialog>
