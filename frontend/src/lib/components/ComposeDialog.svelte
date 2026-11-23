<script lang="ts">
	import { resolve } from '$app/paths';
	import { copyText } from '#lib/clipboard.ts';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import {
		getComposeFields,
		composeScript,
		type ComposeField,
		type ComposeFields,
		type ComposeRequest
	} from '#lib/compose.ts';
	import { ApiError, formatApiError } from '#lib/client.ts';
	import { createExample } from '#lib/examples.ts';
	import { createLatestRequestGuard } from '#lib/request-guard.ts';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Select from '#lib/components/ui/select.svelte';
	import Tooltip from '#lib/components/ui/tooltip.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Label from '#lib/components/ui/label.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Collapsible from '#lib/components/ui/collapsible.svelte';
	import { toast } from 'svelte-sonner';
	import {
		listSequences,
		getSequence,
		createSequence,
		updateSequenceSteps,
		type SequenceListItem
	} from '#lib/sequences.ts';

	// Script composer dialog (ticket 07): form generated from the
	// function's parameter links in position order, live preview of the
	// emitted script string, copy-to-clipboard, and save-as-example with
	// editable function + device scope + script + meaning. The parent
	// mounts/unmounts the dialog; Escape and backdrop clicks reach the
	// shared Dialog wrapper (ticket 02), which reports the close request
	// through the open state -> onClose effect. The save-as-example and
	// add-to-sequence detail groups are shared Collapsible sections
	// (ticket 07); all form chrome uses the wrapper layer and the scoped
	// style block is gone.
	let {
		fnId,
		onClose
	}: {
		fnId: number;
		onClose: () => void;
	} = $props();

	// Dialog open state: true while mounted; bits-ui flips it on
	// Escape/outside-interaction, the effect forwards that to the parent.
	let open = $state(true);

	$effect(() => {
		if (!open) onClose();
	});

	// ---- fields ------------------------------------------------------------
	let meta = $state<ComposeFields | null>(null);
	let loadError = $state('');
	let loading = $state(true);

	// ---- form state --------------------------------------------------------
	let paramValues = $state<Record<string, string>>({});
	let criteriaValues = $state<Record<string, string>>({});
	let timeoutMs = $state('10000');
	let flow = $state('NEXT');
	let onFail = $state('FAIL');

	// ---- preview -----------------------------------------------------------
	const previewRequestGuard = createLatestRequestGuard();
	let preview = $state('');
	let fieldErrors = $state<Record<string, string>>({});
	let composeTimer: ReturnType<typeof setTimeout> | undefined;

	// ---- save-as-example ---------------------------------------------------
	let saveDevice = $state('');
	let saveMeaning = $state('');
	let saveScript = $state('');
	let saveScriptEdited = false;
	let saving = $state(false);
	let savedNote = $state('');
	let saveError = $state('');

	$effect(() => {
		// fnId read synchronously so a route-param change reloads.
		const id = fnId;
		loadMeta(id);
	});

	async function loadMeta(id: number) {
		clearTimeout(composeTimer);
		previewRequestGuard.begin();
		loading = true;
		loadError = '';
		meta = null;
		try {
			const m = await getComposeFields(id);
			meta = m;
			// Prefill from definition defaults; MULTIPLE inputs hold raw
			// comma text, same as the emitted form.
			paramValues = {};
			criteriaValues = {};
			for (const f of m.parameters) {
				paramValues[f.name] = f.default_value ?? '';
			}
			for (const f of m.criteria) {
				criteriaValues[f.name] = f.default_value ?? '';
			}
			// Save-as-example starts from the links' device scope (first
			// non-empty scope wins); fully editable before saving.
			saveDevice =
				m.function.interface ||
				([m.parameters, m.criteria]
					.flat()
					.map((f) => f.device_name.trim())
					.find((d) => d !== '') ??
					'');
			seqDevice = saveDevice;
			previewScript();
		} catch (e) {
			loadError = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	function currentRequest(): ComposeRequest {
		return {
			parameters: { ...paramValues },
			criteria: { ...criteriaValues },
			timeout_ms: timeoutMs,
			flow,
			on_fail: onFail
		};
	}

	// Live preview: debounced compose call; the server is the single
	// source of truth for the grammar, so the preview shows either the
	// emitted script or the field-level rejections.
	function previewScript(request = previewRequestGuard.begin()) {
		if (!meta) return;
		composeScript(fnId, currentRequest())
			.then((res) => {
				if (!previewRequestGuard.isCurrent(request)) return;
				preview = res.script;
				if (!saveScriptEdited) saveScript = res.script;
				fieldErrors = {};
			})
			.catch((e) => {
				if (!previewRequestGuard.isCurrent(request)) return;
				preview = '';
				if (e instanceof ApiError) {
					fieldErrors = e.fields;
				} else {
					fieldErrors = { _request: e instanceof Error ? e.message : String(e) };
				}
			});
	}

	function schedulePreview() {
		clearTimeout(composeTimer);
		const request = previewRequestGuard.begin();
		composeTimer = setTimeout(() => previewScript(request), 300);
	}

	async function copyScript() {
		if (!preview) return;
		await copyText(preview);
	}

	async function saveAsExample() {
		saving = true;
		savedNote = '';
		saveError = '';
		try {
			if (!meta) throw new Error('not composed yet');
			await createExample({
				function_source: meta.function.source,
				function_name: meta.function.name,
				device_name: saveDevice,
				script: saveScript,
				meaning: saveMeaning
			});
			savedNote = 'Saved as an example of this function.';
			toast.success(savedNote);
		} catch (e) {
			saveError = formatApiError(e);
		} finally {
			saving = false;
		}
	}

	// ---- add-to-sequence (ticket 10) ---------------------------------------
	// The composed step (name-keyed values + fixed segments + device/note)
	// is stored as a SequenceStep; validation happens live at emission, so
	// an invalid step can be stored and is flagged later (ADR-0005).
	let seqs = $state<SequenceListItem[]>([]);
	let seqPick = $state<string>('new');
	let seqNewName = $state('');
	let seqNewDesc = $state('');
	let seqDevice = $state('');
	let seqNote = $state('');
	let seqSaving = $state(false);
	let seqNoteSaved = $state('');
	let seqError = $state('');
	const sequenceListGuard = createLatestRequestGuard();

	// Reload on every open: a sequence may have been created since the
	// dialog was last shown.
	async function reloadSequences() {
		const request = sequenceListGuard.begin();
		try {
			const pageSize = 100;
			const first = await listSequences({ page: 1, page_size: pageSize });
			const items = first.items;
			const pageCount = Math.ceil(first.total / first.page_size);
			for (let page = 2; page <= pageCount; page++) {
				const result = await listSequences({ page, page_size: pageSize });
				items.push(...result.items);
			}
			if (sequenceListGuard.isCurrent(request)) seqs = items;
		} catch {
			if (sequenceListGuard.isCurrent(request)) seqs = [];
		}
	}

	// stepFromForm builds the step payload from the current composer form.
	function stepFromForm() {
		if (!meta) throw new Error('not composed yet');
		return {
			function_source: meta.function.source,
			function_interface: meta.function.interface,
			function_name: meta.function.name,
			parameters: { ...paramValues },
			criteria: { ...criteriaValues },
			timeout_ms: timeoutMs,
			flow,
			on_fail: onFail,
			device_name: seqDevice,
			note: seqNote
		};
	}

	async function addToSequence() {
		seqSaving = true;
		seqNoteSaved = '';
		seqError = '';
		try {
			const step = stepFromForm();
			if (seqPick === 'new') {
				const created = await createSequence({
					name: seqNewName.trim(),
					description: seqNewDesc,
					steps: [step]
				});
				await reloadSequences();
				seqNoteSaved = `Created sequence '${created.name}' with this step.`;
			} else {
				const id = Number(seqPick);
				const current = await getSequence(id);
				await updateSequenceSteps(id, [...current.steps, step]);
				seqNoteSaved = 'Added as a new step.';
			}
			toast.success(seqNoteSaved);
		} catch (e) {
			seqError = formatApiError(e);
		} finally {
			seqSaving = false;
		}
	}

	function fieldLabel(f: ComposeField): string {
		const bits = [f.name];
		if (f.type) bits.push(f.type);
		if (f.cardinality === 'MULTIPLE') bits.push('MULTIPLE');
		if (f.required) bits.push('required');
		return bits.join(' · ');
	}
</script>

<Dialog {open} onOpenChange={(o) => (open = o)} title="Script composer" class="max-w-[640px]">
	{#if loading}
		<p class="muted">Loading fields…</p>
	{:else if loadError || !meta}
		<div class="banner-error" role="alert">{loadError || 'Unknown error.'}</div>
	{:else}
		<p class="muted">
			{meta.function.source}{meta.function.interface} · {meta.function.name}
		</p>

		<div class="mt-2 flex flex-col gap-2">
			{#each meta.parameters as f (f.name)}
				<div class="field mb-0">
					<Label for={`compose-param-${f.name}`}>
						{fieldLabel(f)}
						{#if !f.resolved}
							<Tooltip content="The linked definition is missing">
								{#snippet children(tp)}
									<!-- tabindex makes the hint keyboard-reachable (tooltip opens on focus); the badge itself is not an action target. -->
									<Badge variant="warn" {...tp} tabindex={0}>orphan</Badge>
								{/snippet}
							</Tooltip>
						{/if}
					</Label>
					<Input
						id={`compose-param-${f.name}`}
						type="text"
						bind:value={paramValues[f.name]}
						placeholder={f.cardinality === 'MULTIPLE' ? 'value1,value2,…' : ''}
						oninput={schedulePreview}
					/>
					{#if fieldErrors[f.name]}
						<span class="field-error">{fieldErrors[f.name]}</span>
					{/if}
				</div>
			{:else}
				<p class="muted">This function has no parameter links.</p>
			{/each}

			{#if meta.criteria.length > 0}
				<h4 class="mt-1 mb-1 text-[13px] font-semibold">Criteria</h4>
				{#each meta.criteria as f (f.name)}
					<div class="field mb-0">
						<Label for={`compose-criteria-${f.name}`}>{fieldLabel(f)}</Label>
						<Input
							id={`compose-criteria-${f.name}`}
							type="text"
							bind:value={criteriaValues[f.name]}
							oninput={schedulePreview}
						/>
						{#if fieldErrors[`criteria.${f.name}`]}
							<span class="field-error">{fieldErrors[`criteria.${f.name}`]}</span>
						{/if}
					</div>
				{/each}
			{/if}

			<div class="grid grid-cols-3 gap-2">
				<div class="field mb-0">
					<Label for="compose-timeout">Timeout-ms</Label>
					<Input
						id="compose-timeout"
						type="text"
						bind:value={timeoutMs}
						oninput={schedulePreview}
					/>
					{#if fieldErrors.timeout_ms}<span class="field-error">{fieldErrors.timeout_ms}</span>{/if}
				</div>
				<div class="field mb-0">
					<Label for="compose-flow">Flow</Label>
					<Input id="compose-flow" type="text" bind:value={flow} oninput={schedulePreview} />
					{#if fieldErrors.flow}<span class="field-error">{fieldErrors.flow}</span>{/if}
				</div>
				<div class="field mb-0">
					<Label for="compose-onfail">On-fail</Label>
					<Input id="compose-onfail" type="text" bind:value={onFail} oninput={schedulePreview} />
					{#if fieldErrors.on_fail}<span class="field-error">{fieldErrors.on_fail}</span>{/if}
				</div>
			</div>
			{#if fieldErrors._request}
				<span class="field-error">{fieldErrors._request}</span>
			{/if}
		</div>

		<div class="border-border mt-3 border-t pt-2">
			<div class="mb-1 flex items-center justify-between">
				<strong>Script</strong>
				{#if preview}
					<Button type="button" size="sm" variant="outline" onclick={copyScript}>Copy</Button>
				{/if}
			</div>
			{#if preview}
				<!-- Composer output: server-emitted script string, always mono
				     and visually distinct from the form fields. -->
				<code class="bg-secondary block rounded p-2 text-xs break-all whitespace-pre-wrap">
					{preview}
				</code>
			{:else}
				<p class="muted">
					{Object.keys(fieldErrors).length > 0
						? 'Fix the highlighted fields to preview the script.'
						: 'Type values to preview the script.'}
				</p>
			{/if}
		</div>

		<!-- The two Collapsible groups share the new-sequence field pair. -->
		{#snippet seqNewFields()}
			<div class="field">
				<Label for="compose-seq-new-name">New sequence name</Label>
				<Input id="compose-seq-new-name" type="text" bind:value={seqNewName} />
			</div>
			<div class="field">
				<Label for="compose-seq-new-desc">Description</Label>
				<Textarea id="compose-seq-new-desc" rows={3} bind:value={seqNewDesc} />
			</div>
		{/snippet}

		<Collapsible class="mt-3">
			{#snippet title()}
				<span>Save as example</span>
			{/snippet}
			<form
				class="mt-2"
				onsubmit={(e) => {
					e.preventDefault();
					saveAsExample();
				}}
			>
				<div class="field">
					<Label for="compose-save-script">Script (editable)</Label>

					<Textarea
						id="compose-save-script"
						rows={2}
						bind:value={saveScript}
						oninput={() => (saveScriptEdited = true)}
					/>
				</div>
				<div class="grid grid-cols-2 gap-2">
					<div class="field mb-0">
						<Label for="compose-save-device">Device / Interface</Label>
						<DevicePicker
							id="compose-save-device"
							device={saveDevice}
							label="Device / Interface"
							blankLabel="Global (no Device / Interface)"
							fullWidth
							disabled={saving}
							onValueChange={(value) => (saveDevice = value)}
						/>
					</div>
					<div class="field mb-0">
						<Label for="compose-save-meaning">Meaning</Label>
						<Textarea id="compose-save-meaning" rows={3} bind:value={saveMeaning} />
					</div>
				</div>
				<div class="mt-2 flex items-center gap-3">
					<Button type="submit" disabled={saving || !saveScript.trim()}>
						{saving ? 'Saving…' : 'Save example'}
					</Button>
					{#if savedNote}<span class="text-ok text-[13px]">{savedNote}</span>{/if}
					{#if saveError}<span class="field-error">{saveError}</span>{/if}
				</div>
			</form>
		</Collapsible>

		<Collapsible
			class="mt-2"
			onOpenChange={(o) => {
				if (o) reloadSequences();
			}}
		>
			{#snippet title()}
				<span>Add to sequence</span>
			{/snippet}
			<div class="mt-2">
				<p class="muted">
					Stored as a sequence step; values are re-validated live when the sequence is emitted.
					Manage sequences in the <a
						class="text-primary hover:underline"
						href={resolve('sequences')}
						target="_blank">sequences</a
					>

					list.
				</p>
				{#if seqs.length === 0}
					{@render seqNewFields()}
				{:else}
					<div class="field">
						<Label for="compose-seq-target">Target sequence</Label>
						<Select
							id="compose-seq-target"
							ariaLabel="Target sequence"
							bind:value={seqPick}
							options={[
								{ value: 'new', label: 'New sequence…' },
								...seqs.map((s) => ({
									value: String(s.id),
									label: `${s.name} (${s.step_count} steps)`
								}))
							]}
						/>
					</div>
					{#if seqPick === 'new'}
						<div class="grid grid-cols-2 gap-2">
							{@render seqNewFields()}
						</div>
					{/if}
				{/if}
				<div class="grid grid-cols-2 gap-2">
					<div class="field mb-0">
						<Label for="compose-seq-device">Device / Interface</Label>
						<DevicePicker
							id="compose-seq-device"
							device={seqDevice}
							label="Device / Interface"
							blankLabel="Global (no Device / Interface)"
							fullWidth
							disabled={seqSaving}
							onValueChange={(value) => (seqDevice = value)}
						/>
					</div>
					<div class="field mb-0">
						<Label for="compose-seq-note">Step note</Label>
						<Textarea id="compose-seq-note" rows={3} bind:value={seqNote} />
					</div>
				</div>
				<div class="mt-2 flex items-center gap-3">
					<Button
						type="button"
						disabled={seqSaving || (seqPick === 'new' && !seqNewName.trim())}
						onclick={addToSequence}
					>
						{seqSaving
							? 'Saving…'
							: seqPick === 'new'
								? 'Create sequence with this step'
								: 'Add step'}
					</Button>
					{#if seqNoteSaved}<span class="text-ok text-[13px]">{seqNoteSaved}</span>{/if}
					{#if seqError}<span class="field-error">{seqError}</span>{/if}
				</div>
			</div>
		</Collapsible>

		<div class="mt-4 flex justify-end">
			<Button variant="outline" type="button" onclick={() => (open = false)}>Close</Button>
		</div>
	{/if}
</Dialog>
