<script lang="ts">
	import {
		createExample,
		updateExample,
		ApiError,
		type ExampleRow,
		type ExampleInput
	} from '#lib/examples.ts';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Label from '#lib/components/ui/label.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import { toast } from 'svelte-sonner';

	// The example create/edit form, extracted from ExamplesPanel so the
	// /examples workspace, the Device Examples tab and the per-function
	// panels open the same modal. fixedFunction pins the function reference
	// (function-detail tabs); prefillDevice seeds new rows with the current
	// Device context; lockDevice disables the Device picker so a contextual
	// mutation cannot silently retarget the row (a Global example edited
	// from a Device workspace stays Global).
	let {
		open = $bindable(false),
		editRow = null,
		fixedFunction = null,
		prefillDevice = '',
		lockDevice = false,
		onsaved
	}: {
		open?: boolean;
		editRow?: ExampleRow | null;
		fixedFunction?: { source: string; name: string } | null;
		prefillDevice?: string;
		lockDevice?: boolean;
		onsaved?: (row: ExampleRow) => void | Promise<void>;
	} = $props();

	let form = $state<ExampleInput>(emptyForm());
	let fieldErrors = $state<Record<string, string>>({});
	let submitError = $state('');
	let saving = $state(false);

	function emptyForm(): ExampleInput {
		return {
			function_source: fixedFunction?.source ?? 'DLL',
			function_name: fixedFunction?.name ?? '',
			device_name: '',
			script: '',
			meaning: ''
		};
	}

	// Every open resets the form: edit mode from the row's current values
	// (the Function reference is immutable and the Device keeps its own
	// scope, never the caller's context), create mode from the prefill.
	$effect(() => {
		if (open) {
			form = editRow
				? {
						function_source: editRow.function_source,
						function_name: editRow.function_name,
						device_name: editRow.device_name,
						script: editRow.script,
						meaning: editRow.meaning
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
			let saved: ExampleRow;
			if (editRow) {
				const payload = {
					device_name: lockDevice ? editRow.device_name : form.device_name,
					script: form.script,
					meaning: form.meaning
				};
				saved = await updateExample(editRow.id, payload);
				toast.success('Example updated');
			} else {
				saved = await createExample(form);
				toast.success('Example created');
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
	title={editRow ? 'Edit example' : 'New example'}
>
	<p class="text-muted-foreground mb-3 text-xs">
		{#if editRow || fixedFunction}
			Function: {editRow?.function_source ?? fixedFunction?.source} /
			{editRow?.function_name ?? fixedFunction?.name}
		{/if}
		{form.device_name ? `Device: ${form.device_name}` : 'Global - shared across Devices'}
	</p>
	{#if submitError}
		<div class="banner-error" role="alert">{submitError}</div>
	{/if}
	<form
		onsubmit={(e) => {
			e.preventDefault();
			save();
		}}
	>
		{#if !editRow && !fixedFunction}
			<div class="field">
				<Label for="example-source">Function source</Label>
				<Select
					id="example-source"
					ariaLabel="Function source"
					bind:value={form.function_source}
					options={[
						{ value: 'BUILT-IN', label: 'BUILT-IN' },
						{ value: 'DLL', label: 'DLL' }
					]}
				/>
				{#if fieldErrors.function_source}
					<span class="field-error">{fieldErrors.function_source}</span>
				{/if}
			</div>
			<div class="field">
				<Label for="example-fn-name">Function name</Label>
				<Input id="example-fn-name" bind:value={form.function_name} />
				{#if fieldErrors.function_name}
					<span class="field-error">{fieldErrors.function_name}</span>
				{/if}
			</div>
		{/if}
		<div class="field">
			<Label for="example-device">Device / Interface</Label>
			<!-- Device (row attribute, ADR-0009): blank = Global. Locked when
			     opened from a Device workspace in context so a Global example
			     cannot be silently retargeted to the current Device. -->
			<DevicePicker
				id="example-device"
				device={form.device_name}
				label="Device / Interface"
				blankLabel="Global (no Device / Interface)"
				fullWidth
				disabled={lockDevice || saving}
				onValueChange={(value) => (form.device_name = value)}
			/>
			{#if fieldErrors.device_name}
				<span class="field-error">{fieldErrors.device_name}</span>
			{/if}
		</div>
		<div class="field">
			<Label for="example-script">Script</Label>
			<Textarea id="example-script" bind:value={form.script} rows={3} class="font-mono" />
			{#if fieldErrors.script}
				<span class="field-error">{fieldErrors.script}</span>
			{/if}
		</div>
		<div class="field">
			<Label for="example-meaning">Meaning</Label>
			<Textarea id="example-meaning" bind:value={form.meaning} rows={2} />
		</div>
		<div class="mt-4 flex justify-end gap-2">
			<Button variant="outline" disabled={saving} onclick={() => (open = false)}>Cancel</Button>
			<Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</Button>
		</div>
	</form>
</Dialog>
