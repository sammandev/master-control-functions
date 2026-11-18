<script lang="ts">
	import { Combobox } from 'bits-ui';
	import { cn } from '#lib/utils.ts';

	// Combobox wrapper (shadcn-svelte pattern, compact density): async
	// search-as-you-type picker (ticket 04). bits-ui 2.x applies NO
	// client-side filtering — the Root's `inputValue` is NOT bindable and
	// the rendered Items are entirely ours — so the caller hands in a
	// `search` function and the wrapper owns the 250 ms debounce, the
	// in-flight race guard, the loading state and the empty state. `value`/`onValueChange`
	// follow the dialog wrapper's controlled pattern (explicit arrows;
	// svelte-check mis-types bare on* shorthands). Selecting an item fills
	// the input with its label and closes the listbox but does NOT commit
	// any surrounding form — callers submit their own form (Enter inside
	// the input falls through bits-ui's early return and submits the form
	// implicitly).
	type Opt = { value: string; label: string; hint?: string; disabled?: boolean; key?: string };

	let {
		value = $bindable(''),
		search,
		ariaLabel,
		id,
		placeholder = 'Type to search…',
		emptyLabel = 'No matching definitions.',
		emptyActionLabel,
		onEmptyAction,
		debounceMs = 250,
		class: className,
		onValueChange
	}: {
		value?: string;
		search: (term: string) => Promise<Opt[]>;
		ariaLabel?: string;
		id?: string;
		placeholder?: string;
		emptyLabel?: string;
		emptyActionLabel?: string;
		onEmptyAction?: () => void;
		debounceMs?: number;
		class?: string;
		onValueChange?: (value: string) => void;
	} = $props();

	let typed = $state('');
	let open = $state(false);
	let results = $state<Opt[]>([]);
	let searching = $state(false);
	// Term whose search last completed; the empty state only renders for
	// it, so typing never flashes a premature "no matches".
	let settledTerm = $state<string | null>(null);
	let reqId = 0;

	async function runSearch(term: string) {
		const ticket = ++reqId;
		searching = true;
		try {
			const res = await search(term);
			if (ticket === reqId) {
				results = res;
				settledTerm = term;
			}
		} catch {
			if (ticket === reqId) {
				results = [];
				settledTerm = term;
			}
		} finally {
			if (ticket === reqId) searching = false;
		}
	}

	$effect(() => {
		const term = typed;
		if (!open) return;
		const t = setTimeout(() => runSearch(term), debounceMs);
		return () => clearTimeout(t);
	});

	const inputBase =
		'h-8 w-full rounded-[4px] border border-border bg-card px-2.5 text-sm text-foreground transition-colors placeholder:text-muted-foreground focus-visible:border-primary focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring/40 disabled:cursor-default disabled:opacity-50';
	const contentBase =
		'z-50 max-h-72 min-w-32 overflow-y-auto rounded-[4px] border border-border bg-card p-1 text-sm shadow-md';
	const itemBase =
		'relative flex w-full cursor-pointer select-none items-center justify-between gap-2 rounded-sm px-2 py-1.5 outline-none data-[highlighted]:bg-accent data-[selected]:font-semibold data-[disabled]:pointer-events-none data-[disabled]:opacity-50';
</script>

<Combobox.Root
	type="single"
	allowDeselect={false}
	{value}
	onValueChange={(v) => {
		value = v;
		// bits-ui rewrites the input's displayed text to the selected
		// label internally (Root.inputValue is not bindable); keep the
		// tracked term in step so the next open searches from it.
		typed = v;
		onValueChange?.(v);
	}}
	bind:open
>
	<Combobox.Input
		{id}
		aria-label={ariaLabel}
		{placeholder}
		autocomplete="off"
		class={cn(inputBase, className)}
		// Focus opens the listbox with the first page of results — the
		// picker is input-anchored, there is no separate trigger button.
		// bits-ui composes (not overrides) handlers passed to Input, so
		// this oninput tracks the term alongside its own.
		onfocus={() => (open = true)}
		oninput={(e) => (typed = e.currentTarget.value)}
	/>
	<Combobox.Portal>
		<Combobox.Content class={contentBase} sideOffset={4}>
			<Combobox.Viewport class="p-1">
				{#each results as option (option.key ?? option.value)}
					<Combobox.Item
						value={option.value}
						label={option.label}
						disabled={option.disabled}
						class={itemBase}
					>
						<span class="font-medium">{option.label}</span>
						{#if option.hint}<span class="text-muted-foreground text-xs">{option.hint}</span>{/if}
					</Combobox.Item>
				{/each}
				<div
					class={searching ? 'text-muted-foreground px-2 py-1.5 text-sm' : 'sr-only'}
					aria-live="polite"
				>
					{#if searching}Searching…{/if}
				</div>
				{#if !searching && results.length === 0 && settledTerm === typed}
					<div class="text-muted-foreground px-2 py-1.5 text-sm">{emptyLabel}</div>
					{#if emptyActionLabel}
						<button
							type="button"
							class="text-primary hover:bg-accent focus-visible:outline-ring mt-1 flex w-full items-center rounded-sm px-2 py-1.5 text-sm font-medium outline-none focus-visible:outline-2 focus-visible:outline-offset-1"
							onclick={() => {
								open = false;
								onEmptyAction?.();
							}}
						>
							{emptyActionLabel}
						</button>
					{/if}
				{/if}
			</Combobox.Viewport>
		</Combobox.Content>
	</Combobox.Portal>
</Combobox.Root>
