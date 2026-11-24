<script lang="ts">
	import {
		AlertTriangle,
		CheckCircle2,
		Copy,
		Download,
		FileCode2,
		GitBranchPlus,
		Layers,
		Pencil,
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
	import StructuredText from '#lib/components/StructuredText.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';

	let {
		guide,
		relatedGuides = [],
		loadingRelated = false,
		categories = [],
		severities = [],
		onSelectGuide,
		onFilterErrorCode,
		onAddAlternate,
		onEdit,
		onExport,
		onDelete
	}: {
		guide: TroubleshootingGuideDetail;
		relatedGuides?: TroubleshootingGuideListItem[];
		loadingRelated?: boolean;
		categories?: TroubleshootingCategory[];
		severities?: RegistrySeverity[];
		onSelectGuide: (id: number) => void;
		onFilterErrorCode?: (errorCode: string) => void;
		onAddAlternate: (guide: TroubleshootingGuideDetail) => void;
		onEdit: (guide: TroubleshootingGuideDetail) => void;
		onExport: (guide: TroubleshootingGuideDetail) => void;
		onDelete: (guide: TroubleshootingGuideDetail) => void;
	} = $props();

	function severityOutlineClass(severity: TroubleshootingSeverity): string {
		const tone = severityTone(severity, severities);
		if (tone === 'destructive') return 'severity-badge border-destructive text-destructive';
		if (tone === 'warn') return 'severity-badge border-warn text-warn';
		if (tone === 'default') return 'severity-badge border-primary text-primary';
		return 'severity-badge border-border text-muted-foreground';
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
				<Badge variant="outline" class={severityOutlineClass(guide.severity)}>
					{severityLabel(guide.severity, severities)}
				</Badge>
				<Badge variant="outline">{categoryLabel(guide.category, categories)}</Badge>
				<Badge variant="default">{guide.device_name || 'Global'}</Badge>
				{#if guide.error_code}
					{#if onFilterErrorCode}
						<button
							type="button"
							class="error-code-pill border-primary/30 bg-primary/10 text-primary hover:bg-primary/15 focus-visible:ring-ring rounded-full border px-2.5 py-0.5 font-mono text-[11px] font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none"
							title="Filter Guides by {guide.error_code}"
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
					disabled={loadingRelated}
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
				<Button variant="secondary" size="sm" onclick={() => onExport(guide)}>
					<Download size={14} />
					<span>Export</span>
				</Button>
				<Button variant="destructive" size="sm" onclick={() => onDelete(guide)}>
					<Trash2 size={14} />
					<span>Delete</span>
				</Button>
			</div>
		</div>

		<h2 class="text-foreground mt-2.5 text-lg leading-snug font-semibold">{guide.title}</h2>
	</header>

	{#if loadingRelated}
		<section
			aria-label="Alternative Solutions"
			class="border-primary/25 bg-primary/5 rounded-lg border p-3.5"
		>
			<p class="text-muted-foreground text-xs" role="status">Loading alternative solutions…</p>
		</section>
	{:else if relatedGuides.length > 0}
		<section
			aria-label="Alternative Solutions"
			class="border-primary/25 bg-primary/5 rounded-lg border p-3.5"
		>
			<div class="flex flex-col gap-1">
				<div class="flex items-center gap-2">
					<Layers size={15} class="text-primary shrink-0" />
					<h3
						class="alternative-solutions-title text-foreground text-xs font-semibold tracking-wide uppercase"
					>
						Alternative Solutions ({relatedGuides.length})
					</h3>
				</div>
				<p class="text-muted-foreground text-xs">
					One error code or symptom can have multiple causes. Choose another solution below if
					needed.
				</p>
			</div>
			<div class="mt-2.5 grid gap-2 sm:grid-cols-2">
				{#each relatedGuides as sibling (sibling.id)}
					<button
						type="button"
						class="border-border bg-card hover:border-primary/50 hover:bg-accent/50 focus-visible:ring-ring flex flex-col items-start gap-1 rounded-md border p-2.5 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
						onclick={() => onSelectGuide(sibling.id)}
					>
						<div class="flex w-full items-center justify-between gap-2">
							<span class="text-foreground line-clamp-1 text-xs font-semibold">
								{sibling.title}
							</span>
							<Badge variant="outline" class={severityOutlineClass(sibling.severity)}>
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
							<span class="text-muted-foreground">
								· {sibling.step_count}
								{sibling.step_count === 1 ? 'step' : 'steps'}
							</span>
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
			<StructuredText text={guide.symptom} class="mt-2" />
		</section>

		<section class="border-border bg-card rounded-lg border p-4">
			<div class="flex items-center gap-2">
				<FileCode2 size={15} class="text-primary shrink-0" />
				<h3 class="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
					Root Cause
				</h3>
			</div>
			<StructuredText text={guide.root_cause || 'No root cause recorded yet.'} class="mt-2" />
		</section>
	</div>

	<section class="border-border bg-card rounded-lg border p-4" aria-labelledby="resolution-heading">
		<div class="flex flex-wrap items-center justify-between gap-2">
			<h3 id="resolution-heading" class="text-foreground text-sm font-semibold">
				Step-by-Step Resolution Guide
			</h3>
			<span class="text-muted-foreground text-xs">
				{guide.steps.length}
				{guide.steps.length === 1 ? 'step' : 'steps'}
			</span>
		</div>

		<ol role="list" class="mt-3.5 grid list-none gap-3 p-0">
			{#each guide.steps as step (step.position)}
				<li class="step-card border-border bg-background min-w-0 rounded-md border p-3.5">
					<div class="flex flex-wrap items-center gap-2">
						<span
							class="bg-primary/10 text-primary inline-flex items-center rounded px-1.5 py-0.5 font-mono text-xs font-semibold"
						>
							Step {step.position}
						</span>
						<h4 class="text-foreground text-sm font-semibold">
							{step.title}
						</h4>
					</div>

					<StructuredText text={step.instruction} class="mt-1.5" />

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
							<div class="min-w-0">
								<p class="font-semibold">Expected Verification:</p>
								<StructuredText
									text={step.verification}
									class="text-muted-foreground mt-0.5 gap-1 text-xs"
								/>
							</div>
						</div>
					{/if}
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
	.error-code-pill {
		cursor: pointer;
	}
</style>
