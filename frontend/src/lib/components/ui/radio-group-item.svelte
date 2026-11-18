<script lang="ts">
	import { RadioGroup as RadioGroupPrimitive } from 'bits-ui';
	import { cn } from '#lib/utils.ts';

	// RadioGroup item wrapper over the bits-ui RadioGroup.Item primitive.
	// Playwright treats it as role=radio like the native control it
	// replaces (ticket 01).
	let {
		ref = $bindable(null),
		class: className,
		...rest
	}: RadioGroupPrimitive.ItemProps = $props();
</script>

<RadioGroupPrimitive.Item
	bind:ref
	class={cn(
		'border-input bg-card focus-visible:outline-ring data-[state=checked]:border-primary aspect-square size-3.5 shrink-0 cursor-pointer rounded-full border transition-colors focus-visible:outline-2 focus-visible:outline-offset-1 disabled:cursor-default disabled:opacity-50',
		className
	)}
	{...rest}
>
	{#snippet children({ checked })}
		{#if checked}
			<span class="flex items-center justify-center">
				<span class="bg-primary size-1.5 rounded-full"></span>
			</span>
		{/if}
	{/snippet}
</RadioGroupPrimitive.Item>
