<script lang="ts">
	import { untrack } from 'svelte';
	import { ArrowDown, ArrowUp, GitBranchPlus, Plus, Trash2 } from '@lucide/svelte';
	import {
		ApiError,
		createTroubleshootingGuide,
		moveDraftStep,
		updateTroubleshootingGuide,
		type TroubleshootingCategory,
		type TroubleshootingGuideDetail,
		type TroubleshootingGuideInput,
		type TroubleshootingSeverity,
		type TroubleshootingDraftStep
	} from '#lib/troubleshooting.ts';
	import {
		createTroubleshootingCategory,
		createTroubleshootingSeverity,
		type TroubleshootingCategory as RegistryCategory,
		type TroubleshootingSeverity as RegistrySeverity
	} from '#lib/troubleshooting-registry.ts';
	import type { DeviceSummary } from '#lib/devices.ts';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import TroubleshootingRegistryPicker from '#lib/components/TroubleshootingRegistryPicker.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';

	export type TroubleshootingDialogMode = 'create' | 'edit' | 'alternate';

	let {
		open = false,
		mode = 'create',
		initialGuide = null,
		deviceSummaries = [],
		categories = [],
		severities = [],
		onOpenChange,
		onSaved,
		onRegistryChanged
	}: {
		open: boolean;
		mode: TroubleshootingDialogMode;
		initialGuide: TroubleshootingGuideDetail | null;
		deviceSummaries?: DeviceSummary[];
		categories?: RegistryCategory[];
		severities?: RegistrySeverity[];
		onOpenChange: (open: boolean) => void;
		onSaved: (saved: TroubleshootingGuideDetail) => void;
		onRegistryChanged: () => void | Promise<void>;
	} = $props();

	let draftSequence = 0;
	function makeDraftStep(
		step?: Omit<TroubleshootingDraftStep, 'client_id'>
	): TroubleshootingDraftStep {
		return {
			client_id: `draft-${++draftSequence}`,
			title: step?.title ?? '',
			instruction: step?.instruction ?? '',
			script_snippet: step?.script_snippet ?? '',
			verification: step?.verification ?? ''
		};
	}

	let title = $state(untrack(() => (mode === 'edit' ? (initialGuide?.title ?? '') : '')));
	let category = $state<TroubleshootingCategory>(
		untrack(() => initialGuide?.category ?? 'connection')
	);
	let severity = $state<TroubleshootingSeverity>(
		untrack(() => initialGuide?.severity ?? 'critical')
	);
	let deviceName = $state(untrack(() => initialGuide?.device_name ?? ''));
	let functionName = $state(untrack(() => initialGuide?.function_name ?? ''));
	let errorCode = $state(untrack(() => initialGuide?.error_code ?? ''));
	let symptom = $state(untrack(() => (mode === 'edit' ? (initialGuide?.symptom ?? '') : '')));
	let rootCause = $state(untrack(() => (mode === 'edit' ? (initialGuide?.root_cause ?? '') : '')));
	let steps = $state<TroubleshootingDraftStep[]>(
		untrack(() =>
			mode === 'edit' && initialGuide?.steps.length
				? initialGuide.steps.map((step) => makeDraftStep(step))
				: [makeDraftStep()]
		)
	);

	let busy = $state(false);
	let generalError = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	const dialogTitle = $derived(
		mode === 'edit'
			? 'Edit Troubleshooting Guide'
			: mode === 'alternate'
				? 'Add Alternate Solution for Existing Problem / Error Code'
				: 'New Troubleshooting Guide'
	);

	const dialogDescription = $derived(
		mode === 'alternate'
			? 'A single error code or function can fail in multiple ways. The problem tags below are pre-filled so this guide is automatically linked as an alternative solution.'
			: 'Document the symptom, root cause, and step-by-step verification guides.'
	);

	function addStep() {
		steps = [...steps, makeDraftStep()];
	}

	function removeStep(index: number) {
		if (steps.length <= 1) return;
		steps = steps.filter((_, i) => i !== index);
	}

	function moveStep(index: number, delta: -1 | 1) {
		steps = moveDraftStep(steps, index, delta);
	}

	function validateClient(): Record<string, string> {
		const errs: Record<string, string> = {};
		if (!title.trim()) errs.title = 'Title is required';
		if (!symptom.trim()) errs.symptom = 'Observed symptom is required';
		if (steps.length === 0) errs.steps = 'At least one resolution step is required';
		steps.forEach((s, idx) => {
			if (!s.title.trim() && !errs['steps.title']) {
				errs['steps.title'] = `Step ${idx + 1}: step title is required`;
			}
			if (!s.instruction.trim() && !errs['steps.instruction']) {
				errs['steps.instruction'] = `Step ${idx + 1}: instruction is required`;
			}
		});
		return errs;
	}

	async function createCategory(label: string) {
		const created = await createTroubleshootingCategory({ label });
		await onRegistryChanged();
		category = created.key;
		return created;
	}

	async function createSeverity(label: string) {
		const created = await createTroubleshootingSeverity({ label, tone: 'outline' });
		await onRegistryChanged();
		severity = created.key;
		return created;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		generalError = '';
		const errs = validateClient();
		if (Object.keys(errs).length > 0) {
			fieldErrors = errs;
			return;
		}
		fieldErrors = {};
		busy = true;
		const payload: TroubleshootingGuideInput = {
			title: title.trim(),
			category,
			severity,
			device_name: deviceName,
			function_name: functionName.trim(),
			error_code: errorCode.trim().toUpperCase(),
			symptom: symptom.trim(),
			root_cause: rootCause.trim(),
			steps: steps.map((s) => ({
				title: s.title.trim(),
				instruction: s.instruction.trim(),
				script_snippet: s.script_snippet.trim(),
				verification: s.verification.trim()
			}))
		};
		try {
			const saved =
				mode === 'edit' && initialGuide
					? await updateTroubleshootingGuide(initialGuide.id, payload)
					: await createTroubleshootingGuide(payload);
			onSaved(saved);
			onOpenChange(false);
		} catch (err) {
			if (err instanceof ApiError) {
				generalError = err.message;
				if (err.fields) fieldErrors = err.fields;
			} else {
				generalError = err instanceof Error ? err.message : String(err);
			}
		} finally {
			busy = false;
		}
	}
</script>

<Dialog
	{open}
	{onOpenChange}
	title={dialogTitle}
	description={dialogDescription}
	class="max-w-[780px]"
>
	<form class="guide-form" onsubmit={handleSubmit}>
		{#if mode === 'alternate' && initialGuide}
			<div
				class="border-primary/30 bg-primary/5 text-foreground flex items-start gap-2.5 rounded-md border p-3 text-xs"
			>
				<GitBranchPlus size={15} class="text-primary mt-0.5 shrink-0" />
				<div>
					<span class="font-semibold">Linked to existing problem:</span>
					<span class="text-muted-foreground ml-1">
						Sharing Error Code
						{#if errorCode}<Badge variant="default" class="mx-0.5 font-mono">{errorCode}</Badge
							>{/if}
						or Function
						{#if functionName}<Badge variant="outline" class="mx-0.5 font-mono"
								>{functionName}</Badge
							>{/if}
						groups this guide alongside <strong>"{initialGuide.title}"</strong> as an alternative solution.
					</span>
				</div>
			</div>
		{/if}

		{#if generalError}
			<p class="error" role="alert">{generalError}</p>
		{/if}

		<section class="form-section border-border bg-background/50 rounded-md border p-3.5">
			<h3 class="text-foreground text-xs font-semibold tracking-wider uppercase">
				1. Problem & Context Tags
			</h3>
			<div class="mt-2.5 grid gap-3 sm:grid-cols-2">
				<div class="grid gap-1.5 sm:col-span-2">
					<label for="guide-title" class="text-foreground text-xs font-semibold">
						Solution / Manifestation Title <span class="text-destructive">*</span>
					</label>
					<Input
						id="guide-title"
						bind:value={title}
						placeholder="e.g. TX_MULTI_VERIFICATION Power Low Due to Uncalibrated RF Cable Loss"
						aria-invalid={fieldErrors.title ? 'true' : undefined}
						aria-describedby={fieldErrors.title ? 'guide-title-error' : undefined}
						required
					/>
					{#if fieldErrors.title}
						<span id="guide-title-error" class="text-destructive text-xs">{fieldErrors.title}</span>
					{/if}
				</div>

				<div class="grid gap-1.5 sm:col-span-2">
					<label for="guide-error-code" class="text-foreground text-xs font-semibold">
						Error Code / Problem Tag
					</label>
					<Input
						id="guide-error-code"
						bind:value={errorCode}
						class="font-mono"
						placeholder="e.g. E00WI005638 or IoT_TX1_FIXTURE_OR_DUT_PROBLEM_POW"
					/>
					<span class="text-muted-foreground text-[11px]">
						Multiple guides can share the same Error Code to document different causes.
					</span>
				</div>

				<div class="grid gap-1.5">
					<label for="guide-device" class="text-foreground text-xs font-semibold">
						Device / Interface Scope
					</label>
					<DevicePicker
						id="guide-device"
						label="Device / Interface Scope"
						blankLabel="Global (All Devices)"
						device={deviceName}
						summaries={deviceSummaries}
						invalid={Boolean(fieldErrors.device_name)}
						describedBy={fieldErrors.device_name ? 'guide-device-error' : undefined}
						fullWidth
						onValueChange={(d) => {
							deviceName = d;
						}}
					/>
					{#if fieldErrors.device_name}
						<span id="guide-device-error" class="text-destructive text-xs"
							>{fieldErrors.device_name}</span
						>
					{/if}
				</div>

				<div class="grid gap-1.5">
					<label for="guide-function-name" class="text-foreground text-xs font-semibold">
						Related Function Name
					</label>
					<Input
						id="guide-function-name"
						bind:value={functionName}
						class="font-mono"
						placeholder="e.g. TX_MULTI_VERIFICATION, SENDSTR_AND_WAITSTR, MAIN_FLOW"
					/>
				</div>

				<div class="grid gap-1.5">
					<label for="guide-category" class="text-foreground text-xs font-semibold">
						Testing Stage / Category
					</label>
					<div class="registry-field">
						<TroubleshootingRegistryPicker
							id="guide-category"
							items={categories}
							kind="category"
							bind:value={category}
							invalid={Boolean(fieldErrors.category)}
							describedBy={fieldErrors.category ? 'guide-category-error' : undefined}
							onCreate={createCategory}
						/>
					</div>
					{#if fieldErrors.category}
						<span id="guide-category-error" class="text-destructive text-xs"
							>{fieldErrors.category}</span
						>
					{/if}
				</div>

				<div class="grid gap-1.5">
					<label for="guide-severity" class="text-foreground text-xs font-semibold">Severity</label>
					<div class="registry-field">
						<TroubleshootingRegistryPicker
							id="guide-severity"
							items={severities}
							kind="severity"
							bind:value={severity}
							invalid={Boolean(fieldErrors.severity)}
							describedBy={fieldErrors.severity ? 'guide-severity-error' : undefined}
							onCreate={createSeverity}
						/>
					</div>
					{#if fieldErrors.severity}
						<span id="guide-severity-error" class="text-destructive text-xs"
							>{fieldErrors.severity}</span
						>
					{/if}
				</div>
			</div>
		</section>

		<section class="form-section border-border bg-background/50 rounded-md border p-3.5">
			<h3 class="text-foreground text-xs font-semibold tracking-wider uppercase">
				2. Diagnostic Summary
			</h3>
			<div class="mt-2.5 grid gap-3">
				<div class="grid gap-1.5">
					<label for="guide-symptom" class="text-foreground text-xs font-semibold">
						Observed Symptom / Log Signature <span class="text-destructive">*</span>
					</label>
					<Textarea
						id="guide-symptom"
						rows={2}
						bind:value={symptom}
						placeholder="What does the test engineer see in the runner log, console, or measurement report?"
						aria-invalid={fieldErrors.symptom ? 'true' : undefined}
						aria-describedby={fieldErrors.symptom ? 'guide-symptom-error' : undefined}
						required
					/>
					{#if fieldErrors.symptom}
						<span id="guide-symptom-error" class="text-destructive text-xs"
							>{fieldErrors.symptom}</span
						>
					{/if}
				</div>

				<div class="grid gap-1.5">
					<label for="guide-root-cause" class="text-foreground text-xs font-semibold">
						Root Cause Explanation
					</label>
					<Textarea
						id="guide-root-cause"
						rows={2}
						bind:value={rootCause}
						placeholder="Why does this happen in the MCV2 script, INI configuration, DUT firmware, or RF station?"
					/>
				</div>
			</div>
		</section>

		<section class="form-section border-border bg-background/50 rounded-md border p-3.5">
			<div class="flex flex-wrap items-center justify-between gap-2">
				<div>
					<h3 class="text-foreground text-xs font-semibold tracking-wider uppercase">
						3. Step-by-Step Resolution Guide ({steps.length})
					</h3>
					<p class="text-muted-foreground text-[11px]">
						Break the fix into clear, verifiable steps. Include optional INI / script examples.
					</p>
				</div>
				<Button type="button" variant="secondary" size="sm" onclick={addStep}>
					<Plus size={14} />
					<span>Add Step</span>
				</Button>
			</div>

			{#if fieldErrors.steps || fieldErrors['steps.title'] || fieldErrors['steps.instruction']}
				<p class="text-destructive mt-2 text-xs" role="alert">
					{fieldErrors.steps || fieldErrors['steps.title'] || fieldErrors['steps.instruction']}
				</p>
			{/if}

			<div class="mt-3 grid gap-3">
				{#each steps as step, idx (step.client_id)}
					<div class="border-border bg-card rounded-md border p-3">
						<div class="flex items-center justify-between gap-2">
							<Badge variant="default" class="font-mono">Step {idx + 1}</Badge>
							<div class="flex items-center gap-1">
								<Button
									type="button"
									variant="ghost"
									size="icon"
									aria-label="Move step {idx + 1} up"
									title="Move step up"
									disabled={idx === 0}
									onclick={() => moveStep(idx, -1)}
								>
									<ArrowUp size={14} />
								</Button>
								<Button
									type="button"
									variant="ghost"
									size="icon"
									aria-label="Move step {idx + 1} down"
									title="Move step down"
									disabled={idx === steps.length - 1}
									onclick={() => moveStep(idx, 1)}
								>
									<ArrowDown size={14} />
								</Button>
								<Button
									type="button"
									variant="ghost"
									size="icon"
									aria-label="Remove step {idx + 1}"
									title="Remove step"
									disabled={steps.length <= 1}
									onclick={() => removeStep(idx)}
								>
									<Trash2 size={14} />
								</Button>
							</div>
						</div>

						<div class="mt-2.5 grid gap-2.5">
							<div class="grid gap-1">
								<label for="step-title-{idx}" class="text-foreground text-xs font-semibold">
									Action Headline <span class="text-destructive">*</span>
								</label>
								<Input
									id="step-title-{idx}"
									bind:value={step.title}
									placeholder="e.g. Check TX_TARGET_POWER and POWER tolerance in TEST_SPEC.ini"
									aria-invalid={fieldErrors['steps.title'] ? 'true' : undefined}
									aria-describedby={fieldErrors['steps.title']
										? `step-title-error-${idx}`
										: undefined}
									required
								/>
								{#if fieldErrors['steps.title']}
									<span id="step-title-error-{idx}" class="text-destructive text-xs">
										{fieldErrors['steps.title']}
									</span>
								{/if}
							</div>

							<div class="grid gap-1">
								<label for="step-instruction-{idx}" class="text-foreground text-xs font-semibold">
									Detailed Instruction <span class="text-destructive">*</span>
								</label>
								<Textarea
									id="step-instruction-{idx}"
									rows={2}
									bind:value={step.instruction}
									placeholder="Explain exactly what file, parameter, or fixture component to inspect or modify."
									aria-invalid={fieldErrors['steps.instruction'] ? 'true' : undefined}
									aria-describedby={fieldErrors['steps.instruction']
										? `step-instruction-error-${idx}`
										: undefined}
									required
								/>
								{#if fieldErrors['steps.instruction']}
									<span id="step-instruction-error-{idx}" class="text-destructive text-xs">
										{fieldErrors['steps.instruction']}
									</span>
								{/if}
							</div>

							<div class="grid gap-2.5 sm:grid-cols-2">
								<div class="grid gap-1">
									<label for="step-snippet-{idx}" class="text-foreground text-xs font-semibold">
										Script Line / INI Snippet (optional)
									</label>
									<Textarea
										id="step-snippet-{idx}"
										rows={2}
										class="font-mono text-xs"
										bind:value={step.script_snippet}
										placeholder="DLL:WIFI_TEST,TX_MULTI_VERIFICATION,..."
									/>
								</div>

								<div class="grid gap-1">
									<label
										for="step-verification-{idx}"
										class="text-foreground text-xs font-semibold"
									>
										Expected Verification Result (optional)
									</label>
									<Textarea
										id="step-verification-{idx}"
										rows={2}
										bind:value={step.verification}
										placeholder="How to confirm this step solved the problem."
									/>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</section>

		<div class="dialog-actions">
			<Button type="button" variant="secondary" onclick={() => onOpenChange(false)}>Cancel</Button>
			<Button type="submit" disabled={busy}>
				{busy
					? 'Saving…'
					: mode === 'edit'
						? 'Save Changes'
						: mode === 'alternate'
							? 'Create Alternate Solution'
							: 'Create Guide'}
			</Button>
		</div>
	</form>
</Dialog>

<style>
	.guide-form {
		display: grid;
		gap: 14px;
	}
</style>
