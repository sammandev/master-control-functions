<script lang="ts">
	import { Combobox, Popover } from 'bits-ui';
	import { Check, ChevronDown, Plus, Search } from '@lucide/svelte';

	type RegistryItem = {
		key: string;
		label: string;
		description?: string;
		sort_order: number;
		built_in: boolean;
		guide_count: number;
		tone?: string;
	};

	let {
		items,
		value = $bindable(''),
		kind,
		onCreate
	}: {
		items: RegistryItem[];
		value?: string;
		kind: 'category' | 'severity';
		onCreate: (label: string) => Promise<RegistryItem>;
	} = $props();

	let open = $state(false);
	let term = $state('');
	let error = $state('');
	let creating = $state(false);
	let createdItem = $state<RegistryItem | null>(null);
	let input = $state<HTMLInputElement | null>(null);

	const kindLabel = $derived(kind === 'category' ? 'Category' : 'Severity');
	const selectedLabel = $derived(
		items.find((item) => item.key === value)?.label ??
			(createdItem?.key === value ? createdItem.label : undefined) ??
			(value || `Select ${kind}`)
	);
	const normalizedTerm = $derived(term.trim().toLocaleLowerCase());
	const filteredItems = $derived(
		[
			...items,
			...(createdItem && !items.some((item) => item.key === createdItem?.key) ? [createdItem] : [])
		]
			.filter((item) => {
				if (!normalizedTerm) return true;
				return [item.label, item.key, item.description]
					.filter(Boolean)
					.some((field) => field?.toLocaleLowerCase().includes(normalizedTerm));
			})
			.toSorted((left, right) => left.sort_order - right.sort_order)
	);
	const hasExactMatch = $derived(
		[...items, ...(createdItem ? [createdItem] : [])].some(
			(item) =>
				item.label.trim().toLocaleLowerCase() === normalizedTerm ||
				item.key.trim().toLocaleLowerCase() === normalizedTerm
		)
	);
	const canCreate = $derived(normalizedTerm.length > 0 && !hasExactMatch);

	async function createItem() {
		const label = term.trim();
		if (!canCreate || creating || !label) return;

		creating = true;
		error = '';
		try {
			const item = await onCreate(label);
			createdItem = item;
			value = item.key;
			term = '';
			open = false;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : String(cause);
		} finally {
			creating = false;
		}
	}
</script>

<Popover.Root
	bind:open
	onOpenChange={(next) => {
		if (next) {
			term = '';
			error = '';
		}
	}}
>
	<Popover.Trigger
		aria-label={`${kindLabel}: ${selectedLabel}`}
		class="border-border bg-card text-foreground focus-visible:outline-ring flex h-10 w-full min-w-0 items-center justify-between gap-2 rounded-[4px] border px-3 py-2 text-sm focus-visible:outline-2 focus-visible:outline-offset-2"
	>
		<span class="min-w-0 truncate text-left">{selectedLabel}</span>
		<ChevronDown data-icon="inline-end" />
	</Popover.Trigger>
	<Popover.Portal>
		<Popover.Content
			role="dialog"
			aria-label={`Choose ${kind}`}
			sideOffset={6}
			align="start"
			class="border-border bg-card text-foreground z-50 w-[360px] max-w-[calc(100vw-32px)] rounded-[4px] border p-2 shadow-md"
			onOpenAutoFocus={(event) => {
				event.preventDefault();
				input?.focus({ preventScroll: true });
			}}
		>
			<Combobox.Root
				type="single"
				{value}
				inputValue={term}
				open={true}
				allowDeselect={false}
				onValueChange={(next) => {
					if (!next) return;
					value = next;
					term = '';
					open = false;
				}}
			>
				<div class="border-border relative mb-2 border-b pb-2">
					<Search class="text-muted-foreground absolute top-2.5 left-2.5" />
					<Combobox.Input
						bind:ref={input}
						aria-label={`Search ${kind === 'category' ? 'categories' : 'severities'}`}
						placeholder={`Search ${kind === 'category' ? 'categories' : 'severities'}`}
						autocomplete="off"
						class="border-border bg-background placeholder:text-muted-foreground focus-visible:outline-ring h-10 w-full rounded-[4px] border pr-3 pl-9 text-sm focus-visible:outline-2"
						oninput={(event) => {
							term = event.currentTarget.value;
							error = '';
						}}
					/>
				</div>

				<Combobox.ContentStatic>
					<Combobox.Viewport class="max-h-64 overflow-y-auto overscroll-contain">
						{#each filteredItems as item (item.key)}
							<Combobox.Item
								value={item.key}
								label={item.label}
								class="data-[highlighted]:bg-accent relative flex min-h-10 w-full cursor-pointer items-center justify-between gap-2 rounded-sm px-2 py-2 text-sm outline-none data-[selected]:font-semibold"
							>
								{#snippet children({ selected })}
									<span class="flex min-w-0 flex-col gap-0.5">
										<span class="truncate">{item.label}</span>
										{#if item.description}
											<span class="text-muted-foreground truncate text-xs">{item.description}</span>
										{/if}
									</span>
									{#if selected}<Check />{/if}
								{/snippet}
							</Combobox.Item>
						{/each}
					</Combobox.Viewport>
				</Combobox.ContentStatic>

				{#if canCreate}
					<button
						type="button"
						disabled={creating}
						class="border-border hover:bg-accent focus-visible:outline-ring mt-2 flex min-h-10 w-full items-center gap-2 rounded-sm border px-2 py-2 text-left text-sm font-medium focus-visible:outline-2 focus-visible:outline-offset-1 disabled:cursor-wait disabled:opacity-60"
						onclick={createItem}
					>
						<Plus data-icon="inline-start" />
						{creating ? 'Creating…' : `Create “${term.trim()}”`}
					</button>
				{:else if filteredItems.length === 0}
					<p role="status" class="text-muted-foreground px-2 py-3 text-sm">
						No matching {kind === 'category' ? 'categories' : 'severities'}.
					</p>
				{/if}

				{#if error}
					<p role="alert" class="text-destructive px-2 pt-2 text-sm">{error}</p>
				{/if}
			</Combobox.Root>
		</Popover.Content>
	</Popover.Portal>
</Popover.Root>
