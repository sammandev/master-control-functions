<script lang="ts">
	import { Checkbox as CheckboxPrimitive } from 'bits-ui';
	import { Check, Minus } from '@lucide/svelte';
	import { cn } from '#lib/utils.ts';

	let {
		ref = $bindable(null),
		checked = $bindable(false),
		indeterminate = $bindable(false),
		class: className,
		...rest
	}: Omit<CheckboxPrimitive.RootProps, 'children' | 'child'> = $props();
</script>

<CheckboxPrimitive.Root
	bind:ref
	bind:checked
	bind:indeterminate
	class={cn(
		'border-muted-foreground bg-card text-primary-foreground hover:border-primary focus-visible:outline-ring data-[state=checked]:border-primary data-[state=checked]:bg-primary data-[state=indeterminate]:border-primary data-[state=indeterminate]:bg-primary inline-flex size-5 shrink-0 cursor-pointer items-center justify-center rounded-[4px] border-2 transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	{...rest}
>
	{#snippet children({ checked: isChecked, indeterminate: isIndeterminate })}
		{#if isIndeterminate}
			<Minus size={14} strokeWidth={3} />
		{:else if isChecked}
			<Check size={14} strokeWidth={3} />
		{/if}
	{/snippet}
</CheckboxPrimitive.Root>
