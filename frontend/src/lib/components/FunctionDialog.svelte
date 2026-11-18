<script lang="ts">
	import { onMount } from 'svelte';
	import {
		createFunction,
		updateFunction,
		deleteFunction,
		ApiError,
		type FunctionInput,
		type FunctionRow
	} from '#lib/api.ts';
	import type { DeviceFunctionRow } from '#lib/devices.ts';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';
	import Label from '#lib/components/ui/label.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import { toast } from 'svelte-sonner';

	let {
		mode,
		fn,
		defaults = {},
		onClose,
		onSaved
	}: {
		mode: 'create' | 'edit' | 'delete';
		fn?: FunctionRow | DeviceFunctionRow;
		defaults?: Partial<FunctionInput>;
		onClose: () => void;
		onSaved: (row: FunctionRow | null) => void | Promise<void>;
	} = $props();
	let form = $state<FunctionInput>({
		source: 'DLL',
		interface: '',
		name: '',
		description: '',
		convention: null,
		category: '',
		execution_role: '',
		execution_phase: ''
	});
	let fieldErrors = $state<Record<string, string>>({});
	let submitError = $state('');
	let saving = $state(false);
	onMount(() => {
		form = fn
			? {
					source: fn.source,
					interface: fn.interface,
					name: fn.name,
					description: fn.description ?? '',
					convention: fn.convention,
					category: fn.category ?? '',
					execution_role: fn.execution_role ?? '',
					execution_phase: fn.execution_phase ?? ''
				}
			: { ...form, ...defaults };
	});
	function changeOpen(open: boolean) {
		if (!open && !saving) onClose();
	}
	async function save() {
		if (saving) return;
		saving = true;
		fieldErrors = {};
		submitError = '';
		try {
			let row: FunctionRow | null;
			if (mode === 'delete') {
				if (!fn) throw new Error('No Function selected.');
				await deleteFunction(fn.id);
				row = null;
			} else if (mode === 'edit') {
				if (!fn) throw new Error('No Function selected.');
				row = await updateFunction(fn.id, form);
			} else row = await createFunction(form);
			toast.success(
				mode === 'create'
					? 'Function created'
					: mode === 'edit'
						? 'Function updated'
						: 'Function deleted'
			);
			await onSaved(row);
			onClose();
		} catch (error) {
			if (error instanceof ApiError) fieldErrors = error.fields;
			submitError = error instanceof Error ? error.message : String(error);
		} finally {
			saving = false;
		}
	}
</script>

{#if mode === 'delete'}
	<AlertDialog
		open={true}
		onOpenChange={changeOpen}
		title="Delete function"
		description={fn
			? `${fn.interface ? `Device / Interface: ${fn.interface}` : 'Global - shared across Devices'}. Delete function ${fn.source}/${fn.interface}/${fn.name}? Links and examples referencing it become unresolved across all Devices. This cannot be undone.`
			: ''}
		confirmLabel={saving ? 'Deleting...' : 'Delete'}
		onConfirm={() => {
			void save();
		}}
	>
		{#if submitError}<div class="banner-error" role="alert">{submitError}</div>{/if}
	</AlertDialog>
{:else}
	<Dialog
		open={true}
		onOpenChange={changeOpen}
		title={mode === 'create' ? 'New function' : `Edit function #${fn?.id}`}
	>
		<p class="muted mb-4">
			{form.interface ? `Device: ${form.interface}` : 'Global - shared across Devices'}
		</p>
		{#if submitError}<div class="banner-error" role="alert">{submitError}</div>{/if}
		<form
			class="crud-form"
			onsubmit={(event) => {
				event.preventDefault();
				void save();
			}}
		>
			<div class="field">
				<Label for="f-source">Source</Label><Select
					id="f-source"
					ariaLabel="Source"
					bind:value={form.source}
					options={[
						{ value: 'BUILT-IN', label: 'BUILT-IN' },
						{ value: 'DLL', label: 'DLL' }
					]}
				/>{#if fieldErrors.source}<span class="field-error">{fieldErrors.source}</span>{/if}
			</div>
			<div class="field">
				<Label for="f-interface">Device / Interface</Label><DevicePicker
					id="f-interface"
					device={form.interface}
					label="Device / Interface"
					blankLabel="Global (no Device / Interface)"
					fullWidth
					disabled={saving}
					onValueChange={(value) => (form.interface = value)}
				/>{#if fieldErrors.interface}<span class="field-error">{fieldErrors.interface}</span>{/if}
			</div>
			<div class="field">
				<Label for="f-name">Name</Label><Input
					id="f-name"
					bind:value={form.name}
				/>{#if fieldErrors.name}<span class="field-error">{fieldErrors.name}</span>{/if}
			</div>
			<div class="field field-wide">
				<Label for="f-description">Description</Label><Textarea
					id="f-description"
					rows={4}
					bind:value={form.description}
				/>{#if fieldErrors.description}<span class="field-error">{fieldErrors.description}</span
					>{/if}
			</div>
			<div class="field">
				<Label for="f-convention">Convention</Label><Input
					id="f-convention"
					type="number"
					value={form.convention ?? ''}
					oninput={(event) => {
						const value = event.currentTarget.value;
						form.convention = value === '' ? null : Number(value);
					}}
				/>{#if fieldErrors.convention}<span class="field-error">{fieldErrors.convention}</span>{/if}
			</div>
			<div class="field">
				<Label for="f-category">Category</Label><Input
					id="f-category"
					bind:value={form.category}
				/>{#if fieldErrors.category}<span class="field-error">{fieldErrors.category}</span>{/if}
			</div>
			<div class="field">
				<Label for="f-role">Execution role</Label><Input
					id="f-role"
					bind:value={form.execution_role}
				/>{#if fieldErrors.execution_role}<span class="field-error"
						>{fieldErrors.execution_role}</span
					>{/if}
			</div>
			<div class="field">
				<Label for="f-phase">Execution phase</Label><Input
					id="f-phase"
					bind:value={form.execution_phase}
				/>{#if fieldErrors.execution_phase}<span class="field-error"
						>{fieldErrors.execution_phase}</span
					>{/if}
			</div>
			<div class="dialog-actions">
				<Button variant="outline" disabled={saving} onclick={onClose}>Cancel</Button><Button
					type="submit"
					disabled={saving}>{saving ? 'Saving...' : 'Save'}</Button
				>
			</div>
		</form>
	</Dialog>
{/if}
