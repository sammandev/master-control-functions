<script lang="ts" generics="V extends string">
	import { Combobox, Popover } from 'bits-ui';
	import { Check, ChevronDown, Search } from '@lucide/svelte';
	import { cn } from '#lib/utils.ts';
	type Option = { value: V; label: string; disabled?: boolean };

	let {
		value = $bindable(),
		options,
		id,
		ariaLabel,
		placeholder = 'Select…',
		disabled = false,
		required = false,
		name,
		class: className,
		onchange
	}: {
		value: V;
		options: Option[];
		id?: string;
		ariaLabel?: string;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		name?: string;
		class?: string;
		onchange?: (value: V) => void;
	} = $props();

	let open = $state(false);
	let term = $state('');
	let input = $state<HTMLInputElement | null>(null);
	const selectedLabel = $derived(
		options.find((option) => option.value === value)?.label ?? placeholder
	);
	const filtered = $derived(
		options.filter((option) => option.label.toLowerCase().includes(term.trim().toLowerCase()))
	);
	const triggerBase =
		'flex h-8 w-full items-center justify-between gap-1.5 rounded-[4px] border border-border bg-card px-2.5 text-sm text-foreground transition-colors outline-none focus-visible:border-primary focus-visible:ring-1 focus-visible:ring-ring/40 disabled:cursor-default disabled:opacity-50';
	const contentBase =
		'z-50 w-[max(240px,var(--bits-popover-anchor-width))] max-w-[calc(100vw-24px)] rounded-[4px] border border-border bg-card p-1 text-sm shadow-md';
	const itemBase =
		'relative flex w-full cursor-pointer select-none items-center justify-between gap-2 rounded-sm px-2 py-1.5 outline-none data-[highlighted]:bg-accent data-[selected]:font-semibold data-[disabled]:pointer-events-none data-[disabled]:opacity-50';
</script>

{#if name}<input type="hidden" {name} {value} {disabled} {required} />{/if}
<Popover.Root
	bind:open
	onOpenChange={() => {
		term = '';
	}}
>
	<Popover.Trigger
		{id}
		{disabled}
		aria-label={ariaLabel}
		aria-required={required}
		class={cn(triggerBase, className)}
	>
		<span class="min-w-0 truncate">{selectedLabel}</span>
		<ChevronDown size={14} class="shrink-0 opacity-50" />
	</Popover.Trigger>
	<Popover.Portal>
		<Popover.Content
			role="dialog"
			aria-label={ariaLabel ? `Choose ${ariaLabel}` : 'Choose an option'}
			class={contentBase}
			sideOffset={4}
			align="start"
			onOpenAutoFocus={(event) => {
				event.preventDefault();
				input?.focus();
			}}
		>
			<Combobox.Root
				type="single"
				{value}
				inputValue={term}
				open={true}
				allowDeselect={false}
				onOpenChange={(next) => {
					if (!next) open = false;
				}}
				onValueChange={(next) => {
					const option = options.find((option) => option.value === next && !option.disabled);
					if (!option) return;
					value = option.value;
					onchange?.(option.value);
					open = false;
				}}
			>
				<div class="border-border relative mb-1 border-b pb-1">
					<Search
						size={14}
						class="text-muted-foreground pointer-events-none absolute top-2 left-2"
					/>
					<Combobox.Input
						bind:ref={input}
						aria-label="Search options"
						placeholder="Search options"
						autocomplete="off"
						class="border-border bg-background focus-visible:outline-ring h-8 w-full rounded-sm border pr-2 pl-7 text-sm focus-visible:outline-2"
						oninput={(event) => {
							term = event.currentTarget.value;
						}}
					/>
				</div>
				<Combobox.ContentStatic>
					<Combobox.Viewport class="max-h-60 overflow-y-auto overscroll-contain">
						{#each filtered as option (option.value)}
							<Combobox.Item
								value={option.value}
								label={option.label}
								disabled={option.disabled}
								class={itemBase}
							>
								<span class="min-w-0 break-words">{option.label}</span>
								{#if option.value === value}<Check size={14} class="shrink-0" />{/if}
							</Combobox.Item>
						{:else}<p role="status" class="text-muted-foreground px-2 py-3">
								No options match your search.
							</p>{/each}
					</Combobox.Viewport>
				</Combobox.ContentStatic>
			</Combobox.Root>
		</Popover.Content>
	</Popover.Portal>
</Popover.Root>
