<script lang="ts">
	import { copyText } from '#lib/clipboard.ts';
	import { resolve } from '$app/paths';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		getSequence,
		updateSequence,
		deleteSequence,
		emitSequence,
		type SequenceDetail,
		type SequenceInput,
		type EmitResult
	} from '#lib/sequences.ts';
	import { parseAssignments } from '#lib/sequence-assignments.ts';
	import { formatApiError } from '#lib/client.ts';
	import { toast } from 'svelte-sonner';
	import Breadcrumb from '#lib/components/ui/breadcrumb.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import { RefreshCw } from '@lucide/svelte';

	// The steps array is the source of truth; every mutation is a full-replace PUT.
	// Reorder stays button-driven and gives no toast; durable saves and deletes toast.
	let seqId = $derived(Number(page.params.id));

	let seq = $state<SequenceDetail | null>(null);
	let loading = $state(true);
	let actionError = $state('');

	// ---- emission ----------------------------------------------------------
	let emit = $state<EmitResult | null>(null);
	let emitLoading = $state(false);
	let emitFailed = $state(false);

	// ---- header edit -------------------------------------------------------
	let editMeta = $state(false);
	let nameDraft = $state('');
	let descDraft = $state('');

	// ---- delete confirmation -----------------------------------------------
	let confirmDelete = $state(false);

	// ---- step edit ---------------------------------------------------------
	let editingIndex = $state<number | null>(null);
	let editFnSource = $state('');
	let editFnInterface = $state('');
	let editFnName = $state('');
	let editParams = $state('');
	let editCriteria = $state('');
	let editTimeout = $state('');
	let editFlow = $state('');
	let editOnFail = $state('');
	let editDevice = $state('');
	let editNote = $state('');
	let stepError = $state('');
	$effect(() => {
		const id = seqId;
		load(id);
	});

	async function load(id: number) {
		loading = true;
		actionError = '';
		try {
			seq = await getSequence(id);
			await refreshEmit();
		} catch (e) {
			seq = null;
			actionError = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	async function refreshEmit() {
		if (!seq) return;
		emitLoading = true;
		try {
			emit = await emitSequence(seq.id);
			emitFailed = false;
		} catch (e) {
			emit = null;
			emitFailed = true;
			actionError = e instanceof Error ? e.message : String(e);
		} finally {
			emitLoading = false;
		}
	}

	// The stored step carries id/position/function_exists on top of the
	// SequenceStepInput shape; the server's step decoder ignores unknown
	// fields, so the spread is a sufficient copy for the full-replace PUT.
	function toInput(): SequenceInput {
		if (!seq) throw new Error('sequence not loaded');
		return {
			name: seq.name,
			description: seq.description,
			steps: seq.steps.map((s) => ({ ...s }))
		};
	}

	// Returns the failure message ('' on success) so each form shows its own error.
	async function put(body: SequenceInput): Promise<string> {
		if (!seq) return 'Sequence is not loaded.';
		try {
			seq = await updateSequence(seq.id, body);
			actionError = '';
			await refreshEmit();
			return '';
		} catch (e) {
			return formatApiError(e);
		}
	}

	async function saveMeta() {
		if (!seq || !nameDraft.trim()) return;
		const body = toInput();
		body.name = nameDraft.trim();
		body.description = descDraft;
		const error = await put(body);
		if (error) {
			actionError = error;
			return;
		}
		editMeta = false;
		toast.success('Sequence updated');
	}

	function startEdit(i: number) {
		if (!seq) return;
		const s = seq.steps[i];
		editingIndex = i;
		stepError = '';
		editFnSource = s.function_source;
		editFnInterface = s.function_interface;
		editFnName = s.function_name;
		editParams = Object.entries(s.parameters)
			.map(([k, v]) => `${k}=${v}`)
			.join('\n');
		editCriteria = Object.entries(s.criteria)
			.map(([k, v]) => `${k}=${v}`)
			.join('\n');
		editTimeout = s.timeout_ms;
		editFlow = s.flow;
		editOnFail = s.on_fail;
		editDevice = s.device_name;
		editNote = s.note;
	}

	async function saveStep() {
		if (!seq || editingIndex === null) return;
		const parameters = parseAssignments(editParams);
		if ('error' in parameters) {
			stepError = parameters.error;
			return;
		}
		const criteria = parseAssignments(editCriteria);
		if ('error' in criteria) {
			stepError = criteria.error;
			return;
		}
		const body = toInput();
		const target = body.steps[editingIndex];
		if (!target) return;
		target.function_source = editFnSource.trim();
		target.function_interface = editFnInterface.trim();
		target.function_name = editFnName.trim();
		target.parameters = parameters.values;
		target.criteria = criteria.values;
		target.timeout_ms = editTimeout;
		target.flow = editFlow;
		target.on_fail = editOnFail;
		target.device_name = editDevice;
		target.note = editNote;
		const error = await put(body);
		if (error) {
			stepError = error;
			return;
		}
		editingIndex = null;
		stepError = '';
		toast.success('Step saved');
	}

	function cancelStep() {
		editingIndex = null;
		stepError = '';
	}

	async function move(i: number, dir: -1 | 1) {
		if (!seq) return;
		const j = i + dir;
		if (j < 0 || j >= seq.steps.length) return;
		const body = toInput();
		const [step] = body.steps.splice(i, 1);
		body.steps.splice(j, 0, step);
		const error = await put(body);
		if (error) {
			actionError = error;
			return;
		}
		if (editingIndex === i) editingIndex = j;
		else if (editingIndex === j) editingIndex = i;
	}

	async function removeStep(i: number) {
		if (!seq || seq.steps.length <= 1) return;
		const body = toInput();
		body.steps.splice(i, 1);
		const error = await put(body);
		if (error) {
			actionError = error;
			return;
		}
		if (editingIndex === i) cancelStep();
		else if (editingIndex !== null && editingIndex > i) editingIndex -= 1;
		toast.success('Step removed');
	}

	async function removeSequence() {
		if (!seq) return;
		try {
			await deleteSequence(seq.id);
			toast.success('Sequence deleted');
			await goto(resolve('sequences'));
		} catch (e) {
			actionError = e instanceof Error ? e.message : String(e);
		}
	}

	async function copyAll() {
		if (!emit?.text) return;
		await copyText(emit.text);
	}
</script>

{#if loading}
	<p class="muted">Loading…</p>
{:else if !seq}
	<div class="banner-error" role="alert">
		{actionError || 'Unknown error.'}<Button variant="outline" onclick={() => load(seqId)}
			><RefreshCw size={16} /> Retry sequence</Button
		>
	</div>
	<a href={resolve('sequences')}>Back to sequences</a>
{:else}
	<div class="crumbs">
		<Breadcrumb
			items={[{ label: 'Sequences', href: resolve('sequences') }, { label: seq.name }]}
			class="mt-1 mb-3"
		/>
	</div>

	<header class="sequence-head">
		{#if editMeta}
			<form
				class="rename-form"
				onsubmit={(e) => {
					e.preventDefault();
					saveMeta();
				}}
			>
				<Input aria-label="Sequence name" class="min-w-48 flex-1" bind:value={nameDraft} required />

				<Textarea
					aria-label="Description"
					rows={3}
					class="min-w-48 flex-1"
					bind:value={descDraft}
					placeholder="Description"
				/>
				<Button type="submit" disabled={!nameDraft.trim()}>Save</Button>
				<Button variant="outline" onclick={() => (editMeta = false)}>Cancel</Button>
			</form>
		{:else}
			<div class="title-block">
				<p class="kicker">Sequence</p>
				<h1>{seq.name}</h1>
				{#if seq.description}<p class="muted">{seq.description}</p>{/if}
			</div>
			<div class="actions">
				<Button
					variant="outline"
					onclick={() => {
						nameDraft = seq?.name ?? '';
						descDraft = seq?.description ?? '';
						editMeta = true;
					}}>Rename</Button
				>
				<Button variant="destructive" onclick={() => (confirmDelete = true)}>Delete sequence</Button
				>
			</div>
		{/if}
	</header>

	<dl class="sequence-meta">
		<div>
			<dt>Steps</dt>
			<dd>{seq.steps.length}</dd>
		</div>

		<div>
			<dt>Created</dt>
			<dd>{seq.created_at.slice(0, 10)}</dd>
		</div>
		<div>
			<dt>Updated</dt>
			<dd>{seq.updated_at.slice(0, 10)}</dd>
		</div>
	</dl>

	{#if actionError}<div class="banner-error" role="alert">{actionError}</div>{/if}

	<section class="emitted" aria-labelledby="emitted-heading">
		<div class="section-heading">
			<h2 id="emitted-heading">Emitted text</h2>
			{#if emit?.text}
				<Button type="button" variant="outline" size="sm" onclick={copyAll}>Copy all</Button>
			{/if}
		</div>
		{#if emitLoading}
			<p class="muted">Validating steps…</p>
		{:else if emit?.text}
			<pre class="emitted-preview"><code>{emit.text}</code></pre>
			{#if emit.steps.some((s) => !s.resolved || Object.keys(s.errors).length > 0)}
				<p class="muted">
					Flagged steps were left out of the emitted text: fix or re-link them below.
				</p>
			{/if}
		{:else if emitFailed}
			<p class="muted">Emitted text could not be loaded.</p>
		{:else}
			<p class="muted">No step emits cleanly: fix the flagged steps below.</p>
		{/if}
	</section>

	<section class="steps-section" aria-labelledby="steps-heading">
		<div class="section-heading">
			<h2 id="steps-heading">Ordered steps</h2>
		</div>
		<ol class="steps">
			{#each seq.steps as step, i (step.id)}
				{@const stepEmit = emit?.steps.find((r) => r.step_id === step.id)}
				{@const stepErrors =
					stepEmit && Object.keys(stepEmit.errors).length > 0 ? stepEmit.errors : null}
				{@const flagged = !step.function_exists || stepErrors !== null}
				<li class="step" class:flagged>
					<div class="step-head">
						<strong>Step {step.position}</strong>
						<span class="step-identity">
							{step.function_source}{step.function_interface} · {step.function_name}
							{#if !step.function_exists}
								<Badge variant="warn" title="The referenced function is missing">orphan</Badge>
							{/if}
						</span>
						<Button
							type="button"
							size="icon"
							variant="outline"
							aria-label="Move step up"
							disabled={i === 0}
							onclick={() => move(i, -1)}
						>
							↑
						</Button>
						<Button
							type="button"
							size="icon"
							variant="outline"
							aria-label="Move step down"
							disabled={i === seq.steps.length - 1}
							onclick={() => move(i, 1)}>↓</Button
						>

						<Button type="button" variant="outline" onclick={() => startEdit(i)}>Edit</Button>

						<Button
							type="button"
							variant="destructive"
							disabled={seq.steps.length <= 1}
							onclick={() => removeStep(i)}
						>
							Remove
						</Button>
					</div>

					{#if editingIndex === i}
						<form
							class="step-form"
							onsubmit={(e) => {
								e.preventDefault();
								saveStep();
							}}
						>
							{#if stepError}<p class="banner-error" role="alert">{stepError}</p>{/if}
							<div class="grid gap-2 sm:grid-cols-3">
								<Input
									aria-label="Function source"
									bind:value={editFnSource}
									placeholder="source"
									required
								/>
								<DevicePicker
									device={editFnInterface}
									label="Function Device / Interface"
									blankLabel="Global (no Device / Interface)"
									fullWidth
									onValueChange={(value) => (editFnInterface = value)}
								/>
								<Input
									aria-label="Function name"
									bind:value={editFnName}
									placeholder="function name"
									required
								/>
							</div>
							<div class="grid gap-2 sm:grid-cols-2">
								<label class="field-label">
									<span>Parameters (one name=value per line)</span>
									<Textarea rows={3} bind:value={editParams} />
								</label>
								<label class="field-label">
									<span>Criteria (one name=value per line)</span>
									<Textarea rows={3} bind:value={editCriteria} />
								</label>
							</div>
							<div class="grid gap-2 sm:grid-cols-3">
								<label class="field-label">
									<span>Timeout-ms</span><Input bind:value={editTimeout} />
								</label>

								<label class="field-label"><span>Flow</span><Input bind:value={editFlow} /></label>

								<label class="field-label">
									<span>On-fail</span>
									<Input bind:value={editOnFail} />
								</label>
							</div>
							<div class="grid gap-2 sm:grid-cols-2">
								<DevicePicker
									device={editDevice}
									label="Device / Interface"
									blankLabel="Global (no Device / Interface)"
									fullWidth
									onValueChange={(value) => (editDevice = value)}
								/>
								<label class="field-label">
									<span>Note</span><Textarea rows={3} bind:value={editNote} />
								</label>
							</div>
							<div class="step-form-actions">
								<Button type="submit">Save step</Button>
								<Button type="button" variant="outline" onclick={cancelStep}>Cancel</Button>
							</div>
						</form>
					{:else}
						<div class="chip-row">
							{#each Object.entries(step.parameters) as [k, v] (k)}
								<span class="chip">{k}={v}</span>
							{:else}
								<span class="muted">no parameters</span>
							{/each}
						</div>
						<div class="chip-row">
							{#each Object.entries(step.criteria) as [k, v] (k)}
								<span class="chip chip-criteria">{k}={v}</span>
							{:else}
								<span class="muted">criteria: NULL</span>
							{/each}
						</div>
						<dl class="step-fields">
							<div>
								<dt>Timeout-ms</dt>
								<dd>{step.timeout_ms || 'Not set'}</dd>
							</div>
							<div>
								<dt>Flow</dt>
								<dd>{step.flow || 'Not set'}</dd>
							</div>
							<div>
								<dt>On-fail</dt>
								<dd>{step.on_fail || 'Not set'}</dd>
							</div>
							<div>
								<dt>Device</dt>
								<dd>{step.device_name || 'Not set'}</dd>
							</div>
							{#if step.note}
								<div class="step-note">
									<dt>Note</dt>
									<dd>{step.note}</dd>
								</div>
							{/if}
						</dl>
						{#if !step.function_exists}
							<p class="step-error">The referenced function is missing.</p>
						{:else if stepErrors}
							{#each Object.entries(stepErrors) as [k, msg] (k)}
								<p class="step-error">{k}: {msg}</p>
							{/each}
						{/if}
					{/if}
				</li>
			{/each}
		</ol>
	</section>
{/if}

<a class="mt-4 inline-block" href={resolve('sequences')}>Back to sequences</a>

<AlertDialog
	open={confirmDelete}
	onOpenChange={(o) => (confirmDelete = o)}
	title="Delete sequence"
	description={`Delete sequence ${seq?.name ?? ''}? This cannot be undone.`}
	confirmLabel="Delete"
	onConfirm={removeSequence}
/>

<style>
	.crumbs :global(span:last-child) {
		min-width: 0;
		overflow-wrap: anywhere;
	}
	.step-note {
		grid-column: 1 / -1;
	}
	.sequence-head {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px 16px;
		margin: 8px 0 16px;
	}
	.title-block {
		flex: 1 1 320px;
		min-width: 0;
	}
	.kicker,
	dt {
		color: var(--muted-foreground);
		font-size: 12px;
	}
	h1 {
		margin: 4px 0 8px;
		font: 600 22px/1.4 var(--mono);
		overflow-wrap: anywhere;
	}
	.actions,
	.step-form-actions,
	.chip-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
	}
	.rename-form {
		display: flex;
		flex: 1 1 360px;
		flex-wrap: wrap;
		gap: 8px;
		max-width: 640px;
		min-width: 0;
	}
	.sequence-meta,
	.step-fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 12px 20px;
		margin: 0;
	}
	.sequence-meta dd,
	.step-fields dd {
		margin: 0;
		font-size: 13px;
		overflow-wrap: anywhere;
	}
	.emitted,
	.steps-section {
		min-width: 0;
		margin-top: 20px;
	}
	.section-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 8px;
	}
	.section-heading h2 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
	}
	.emitted-preview {
		max-height: 320px;
		margin: 0;
		overflow: auto;
		padding: 12px;
		border-radius: 4px;
		background: var(--muted);
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		font-family: var(--mono);
		font-size: 12px;
		line-height: 1.8;
	}
	.steps {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.step {
		min-width: 0;
		padding: 14px 16px;
		border: 1px solid var(--border);
		border-radius: var(--radius);
		background: var(--card);
	}
	.step.flagged {
		border-color: var(--warn);
	}
	.step-head {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px 12px;
	}
	.step-identity {
		flex: 1 1 220px;
		min-width: 0;
		font: 600 13px/1.5 var(--mono);
		overflow-wrap: anywhere;
	}
	.chip {
		max-width: 100%;
		padding: 2px 6px;
		border-radius: 4px;
		background: var(--secondary);
		font: 12px/1.6 var(--mono);
		overflow-wrap: anywhere;
	}
	.chip-criteria {
		background: var(--crit-soft);
	}
	.step-fields {
		margin-top: 12px;
	}
	.step-error {
		margin: 8px 0 0;
		color: var(--destructive);
		font-size: 13px;
		overflow-wrap: anywhere;
	}
	.step-form {
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 0;
		margin-top: 12px;
	}
	.field-label {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
		font-size: 13px;
	}
	@media (max-width: 600px) {
		.rename-form {
			flex-basis: 100%;
		}
	}
</style>
