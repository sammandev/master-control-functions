<script lang="ts">
	import Badge from '#lib/components/ui/badge.svelte';

	let {
		script,
		parameters = [],
		criteria = [],
		loading = false,
		unavailable = false
	}: {
		script: string;
		parameters?: string[];
		criteria?: string[];
		loading?: boolean;
		unavailable?: boolean;
	} = $props();

	function assignmentKeys(segment: string): string[] {
		const keys: string[] = [];
		for (const part of segment.split(/[|&]/)) {
			const separator = part.indexOf('=');
			if (separator <= 0) continue;
			const key = part.slice(0, separator).trim().toUpperCase();
			if (key && !keys.includes(key)) keys.push(key);
		}
		return keys;
	}

	function matchingFields(fields: string[], keys: string[]): string[] {
		const matched: string[] = [];
		return fields.filter((field) => {
			const normalized = field.trim().toUpperCase();
			if (!normalized || !keys.includes(normalized) || matched.includes(normalized)) return false;
			matched.push(normalized);
			return true;
		});
	}

	const usage = $derived.by(() => {
		if (loading || unavailable) return null;
		const segments = script.split('|');
		const canonical = segments.length >= 4;
		const allKeys = assignmentKeys(script);
		const parameterKeys = canonical ? assignmentKeys(segments[2] ?? '') : allKeys;
		const criteriaKeys = canonical ? assignmentKeys(segments[3] ?? '') : allKeys;
		const usedParameters = matchingFields(parameters, parameterKeys);
		const usedCriteria = matchingFields(criteria, criteriaKeys);
		const usedKeys = [...usedParameters, ...usedCriteria].map((field) => field.toUpperCase());

		return {
			parameters: usedParameters,
			criteria: usedCriteria,
			other: allKeys.filter((key) => !usedKeys.includes(key))
		};
	});
</script>

<div class="example-usage" role="group" aria-label="Fields used in script">
	{#if loading}
		<span class="usage-state">Loading field usage...</span>
	{:else if unavailable}
		<span class="usage-state">Field usage unavailable</span>
	{:else if usage}
		{#if usage.parameters.length}
			<div class="usage-group">
				<span class="usage-label">Parameters</span>
				{#each usage.parameters as field (field)}<Badge variant="outline">{field}</Badge>{/each}
			</div>
		{/if}
		{#if usage.criteria.length}
			<div class="usage-group">
				<span class="usage-label">Criteria</span>
				{#each usage.criteria as field (field)}<Badge variant="outline">{field}</Badge>{/each}
			</div>
		{/if}
		{#if usage.other.length}
			<div class="usage-group">
				<span class="usage-label">Other script fields</span>
				{#each usage.other as field (field)}<Badge variant="outline">{field}</Badge>{/each}
			</div>
		{:else if !usage.parameters.length && !usage.criteria.length}
			<span class="usage-state">No named fields detected</span>
		{/if}
	{/if}
</div>

<style>
	.example-usage {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 6px 12px;
		margin-top: 8px;
		font-family: var(--default-font-family);
		white-space: normal;
	}
	.usage-group {
		display: inline-flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 4px;
	}
	.usage-label,
	.usage-state {
		color: var(--muted-foreground);
		font-size: 11px;
		font-weight: 600;
	}
</style>
