<script lang="ts">
	import { dev } from '$app/env';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';

	let { current, onChange }: { current: string; onChange: (key: string) => void } = $props();
	const variants = [
		{ key: 'A', name: 'Split catalog' },
		{ key: 'B', name: 'Device hub' },
		{ key: 'C', name: 'Function tree' },
		{ key: 'D', name: 'Inspection desk' },
		{ key: 'E', name: 'Device directory' }
	];
	let index = $derived(
		Math.max(
			0,
			variants.findIndex((variant) => variant.key === current)
		)
	);

	function cycle(direction: number) {
		onChange(variants[(index + direction + variants.length) % variants.length].key);
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
		if (
			event.target instanceof Element &&
			event.target.closest('input, textarea, select, [contenteditable], [role="tablist"]')
		)
			return;
		if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
			event.preventDefault();
			cycle(event.key === 'ArrowLeft' ? -1 : 1);
		}
	}
</script>

<svelte:window onkeydown={dev ? onKeydown : undefined} />

{#if dev}
	<nav class="switcher" aria-label="Prototype layouts">
		<button onclick={() => cycle(-1)} aria-label="Previous layout" title="Previous layout"
			><ChevronLeft size={18} /></button
		>
		<div aria-live="polite">
			<span>PROTOTYPE</span><strong>{variants[index].key} / {variants[index].name}</strong>
		</div>
		<button onclick={() => cycle(1)} aria-label="Next layout" title="Next layout"
			><ChevronRight size={18} /></button
		>
	</nav>
{/if}

<style>
	.switcher {
		position: fixed;
		bottom: 18px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 100;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px;
		width: 300px;
		max-width: calc(100vw - 24px);
		background: var(--foreground);
		color: var(--background);
		border-radius: 8px;
		box-shadow: 0 4px 20px #0003;
	}
	.switcher div {
		flex: 1;
		text-align: center;
	}
	.switcher span {
		display: block;
		font-size: 10px;
		opacity: 0.7;
	}
	.switcher strong {
		display: block;
		font-size: 13px;
	}
	.switcher button {
		display: grid;
		place-items: center;
		flex: 0 0 40px;
		height: 40px;
		border-radius: 4px;
		cursor: pointer;
	}
	.switcher button:hover {
		background: #ffffff20;
	}
	.switcher button:focus-visible {
		outline: 2px solid currentColor;
		outline-offset: 2px;
	}
</style>
