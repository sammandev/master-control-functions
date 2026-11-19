<script lang="ts">
	import {
		AlertTriangle,
		CheckCircle2,
		Copy,
		FileCode2,
		GitBranchPlus,
		Layers,
		Pencil,
		RotateCcw,
		Sparkles,
		Trash2
	} from '@lucide/svelte';
	import { copyText } from '#lib/clipboard.ts';
	import {
		categoryLabel,
		formatTroubleshootingAiContext,
		severityLabel,
		severityTone,
		type TroubleshootingGuideDetail,
		type TroubleshootingGuideListItem,
		type TroubleshootingSeverity
	} from '#lib/troubleshooting.ts';
	import type {
		TroubleshootingCategory,
		TroubleshootingSeverity as RegistrySeverity
	} from '#lib/troubleshooting-registry.ts';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Checkbox from '#lib/components/ui/checkbox.svelte';

	let {
		guide,
		relatedGuides = [],
		categories = [],
		severities = [],
		completedPositions = [],
		onChecklistChange,
		onResetChecklist,
		onSelectGuide,
		onFilterErrorCode,
		onAddAlternate,
		onEdit,
		onDelete
	}: {
		guide: TroubleshootingGuideDetail;
		relatedGuides?: TroubleshootingGuideListItem[];
		categories?: TroubleshootingCategory[];
		severities?: RegistrySeverity[];
		completedPositions?: number[];
		onChecklistChange: (position: number, checked: boolean) => void;
		onResetChecklist: () => void;
		onSelectGuide: (id: number) => void;
		onFilterErrorCode?: (errorCode: string) => void;
		onAddAlternate: (guide: TroubleshootingGuideDetail) => void;
		onEdit: (guide: TroubleshootingGuideDetail) => void;
		onDelete: (guide: TroubleshootingGuideDetail) => void;
	} = $props();

	function isStepChecked(position: number): boolean {
		return completedPositions.includes(position);
	}

	const completedCount = $derived(
		guide.steps.filter((step) => completedPositions.includes(step.position)).length
	);

	function severityVariant(
		severity: TroubleshootingSeverity
	): 'destructive' | 'warn' | 'default' | 'outline' {
		const tone = severityTone(severity, severities);
		return tone === 'destructive' || tone === 'warn' || tone === 'default' ? tone : 'outline';
	}

	async function copyAiContext() {
		const text = formatTroubleshootingAiContext(guide, relatedGuides);
		await copyText(text);
	}
</script>

<div class="detail-stack">
	<header class="detail-head border-border bg-card rounded-lg border p-4">
		<div class="detail-head-top">
			<div class="badge-row">
				<Badge variant={severityVariant(guide.severity)}
					>{severityLabel(guide.severity, severities)}</Badge
				>
				<Badge variant="outline">{categoryLabel(guide.category, categories)}</Badge>
				<Badge variant="default">{guide.device_name || 'Global'}</Badge>
				{#if guide.error_code}
					{#if onFilterErrorCode}
						<button
							type="button"
							class="error-code-pill border-primary/30 bg-primary/10 text-primary hover:bg-primary/15 focus-visible:ring-ring rounded-full border px-2.5 py-0.5 font-mono text-[11px] font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none"
							title="Filter playbook by {guide.error_code}"
							onclick={() => onFilterErrorCode(guide.error_code)}
						>
							{guide.error_code}
						</button>
					{:else}
						<Badge variant="default"><span class="font-mono">{guide.error_code}</span></Badge>
					{/if}
				{/if}
				{#if guide.function_name}
					<Badge variant="outline"><span class="font-mono">{guide.function_name}</span></Badge>
				{/if}
			</div>

			<div class="action-row">
				<Button
					variant="secondary"
					size="sm"
					title="Copy structured Markdown diagnostic context for AI assistants"
					onclick={copyAiContext}
				>
					<Sparkles size={14} />
					<span>Copy AI Context</span>
				</Button>
				<Button
					variant="secondary"
					size="sm"
					title="Document another root cause or solution for this same problem / error code"
					onclick={() => onAddAlternate(guide)}
				>
					<GitBranchPlus size={14} />
					<span>Add Alternate Solution</span>
				</Button>
				<Button variant="secondary" size="sm" onclick={() => onEdit(guide)}>
					<Pencil size={14} />
					<span>Edit</span>
				</Button>
				<Button variant="destructive" size="sm" onclick={() => onDelete(guide)}>
					<Trash2 size={14} />
					<span>Delete</span>
				</Button>
			</div>
		</div>

		<h2 class="text-foreground mt-2.5 text-lg leading-snug font-semibold">{guide.title}</h2>
	</header>

	{#if relatedGuides.length > 0}
		<section
			aria-label="Alternative solutions for this problem or error code"
			class="border-primary/25 bg-primary/5 rounded-lg border p-3.5"
		>
			<div class="flex flex-wrap items-center justify-between gap-2">
				<div class="flex items-center gap-2">
					<Layers size={15} class="text-primary shrink-0" />
					<h3 class="text-foreground text-xs font-semibold tracking-wide uppercase">
						Alternative Solutions for This Problem / Error Code ({relatedGuides.length})
					</h3>
				</div>
				<span class="text-muted-foreground text-xs">
					One error code or symptom can have multiple causes — switch solution below if needed:
				</span>
			</div>
			<div class="mt-2.5 grid gap-2 sm:grid-cols-2">
				{#each relatedGuides as sibling (sibling.id)}
					<button
						type="button"
						class="border-border bg-card hover:border-primary/50 hover:bg-accent/50 focus-visible:ring-ring flex flex-col items-start gap-1 rounded-md border p-2.5 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
						onclick={() => onSelectGuide(sibling.id)}
					>
						<div class="flex w-full items-center justify-between gap-2">
							<span class="text-foreground line-clamp-1 text-xs font-semibold">{sibling.title}</span
							>
							<Badge variant={severityVariant(sibling.severity)}>
								{severityLabel(sibling.severity, severities)}
							</Badge>
						</div>
						<p class="text-muted-foreground line-clamp-1 text-xs">{sibling.symptom}</p>
						<div class="mt-0.5 flex flex-wrap items-center gap-1.5 text-[11px]">
							{#if sibling.error_code}
								<span class="text-primary font-mono font-semibold">{sibling.error_code}</span>
							{/if}
							{#if sibling.function_name}
								<span class="text-muted-foreground font-mono">{sibling.function_name}</span>
							{/if}
							<span class="text-muted-foreground"
								>· {sibling.step_count} {sibling.step_count === 1 ? 'step' : 'steps'}</span
							>
						</div>
					</button>
				{/each}
			</div>
		</section>
	{/if}

	<div class="diagnostic-grid">
		<section class="border-border bg-card rounded-lg border p-4">
			<div class="flex items-center gap-2">
				<AlertTriangle size={15} class="text-warn shrink-0" />
				<h3 class="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
					Observed Symptom / Log Signature
				</h3>
			</div>
			<p class="text-foreground mt-2 text-sm leading-relaxed">{guide.symptom}</p>
		</section>

		<section class="border-border bg-card rounded-lg border p-4">
			<div class="flex items-center gap-2">
				<FileCode2 size={15} class="text-primary shrink-0" />
				<h3 class="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
					Root Cause
				</h3>
			</div>
			<p class="text-foreground mt-2 text-sm leading-relaxed">
				{guide.root_cause || 'No root cause recorded yet.'}
			</p>
		</section>
	</div>

	<section class="border-border bg-card rounded-lg border p-4">
		<div class="checklist-header">
			<div>
				<h3 class="text-foreground text-sm font-semibold">Step-by-Step Resolution Checklist</h3>
				<p class="text-muted-foreground text-xs">
					Check off each diagnostic step as you verify it on the DUT or station.
				</p>
			</div>
			<div class="flex items-center gap-2">
				<Badge variant={completedCount === guide.steps.length ? 'default' : 'outline'}>
					<CheckCircle2 size={12} class="mr-1" />
					{completedCount} / {guide.steps.length} completed
				</Badge>
				{#if completedCount > 0}
					<Button variant="ghost" size="sm" onclick={onResetChecklist}>
						<RotateCcw size={13} />
						<span>Reset progress</span>
					</Button>
				{/if}
			</div>
		</div>

		<ol class="mt-3.5 grid list-none gap-3 p-0">
			{#each guide.steps as step (step.position)}
				{@const checked = isStepChecked(step.position)}
				<li
					class={[
						'step-card border-border rounded-md border p-3.5 transition-colors',
						checked ? 'bg-accent/35 border-primary/35' : 'bg-background'
					]}
				>
					<div class="flex items-start gap-3">
						<div class="pt-0.5">
							<Checkbox
								{checked}
								aria-label="Mark step {step.position}: {step.title} as completed"
								onCheckedChange={(v) => onChecklistChange(step.position, v === true)}
							/>
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex flex-wrap items-center gap-2">
								<span
									class="bg-primary/10 text-primary inline-flex items-center rounded px-1.5 py-0.5 font-mono text-xs font-semibold"
								>
									Step {step.position}
								</span>
								<h4
									class={[
										'text-foreground text-sm font-semibold',
										checked && 'text-muted-foreground line-through'
									]}
								>
									{step.title}
								</h4>
							</div>

							<p class="text-foreground mt-1.5 text-sm leading-relaxed whitespace-pre-line">
								{step.instruction}
							</p>

							{#if step.script_snippet}
								<div class="border-border bg-card mt-2.5 overflow-hidden rounded border">
									<div
										class="border-border bg-muted/40 flex items-center justify-between border-b px-2.5 py-1"
									>
										<span class="text-muted-foreground font-mono text-[11px]">
											Script / INI Reference
										</span>
										<Button
											variant="ghost"
											size="sm"
											class="h-6 px-2 text-xs"
											onclick={() => copyText(step.script_snippet)}
										>
											<Copy size={12} />
											<span>Copy</span>
										</Button>
									</div>
									<pre
										class="text-foreground m-0 overflow-x-auto p-2.5 font-mono text-xs leading-relaxed">{step.script_snippet}</pre>
								</div>
							{/if}

							{#if step.verification}
								<div
									class="border-border bg-accent/30 text-foreground mt-2.5 flex items-start gap-2 rounded border px-2.5 py-1.5 text-xs"
								>
									<CheckCircle2 size={14} class="text-primary mt-0.5 shrink-0" />
									<div>
										<span class="font-semibold">Expected Verification:</span>
										<span class="text-muted-foreground ml-1">{step.verification}</span>
									</div>
								</div>
							{/if}
						</div>
					</div>
				</li>
			{/each}
		</ol>
	</section>
</div>

<style>
	.detail-stack {
		display: grid;
		gap: 14px;
		align-content: start;
	}
	.detail-head-top {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
	}
	.badge-row,
	.action-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
	}
	.diagnostic-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 12px;
	}
	.checklist-header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
	}
	.error-code-pill {
		cursor: pointer;
	}
</style>
