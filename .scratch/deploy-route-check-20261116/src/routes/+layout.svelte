<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { Menu, Moon, Sun } from '@lucide/svelte';
	import './layout.css';
	import type { Path } from '$app/types';
	import { dev } from '$app/env';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { locales, localizeHref } from '#lib/paraglide/runtime.js';
	import appIcon from '#lib/assets/mac_10.ico?url';
	import Button from '#lib/components/ui/button.svelte';
	import * as DropdownMenu from '#lib/components/ui/dropdown-menu/index.js';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import '../app.css';

	let { children } = $props();

	let dark = $state(false);
	let manualTheme = false;

	onMount(() => {
		dark = document.documentElement.classList.contains('dark');
		try {
			manualTheme = localStorage.getItem('theme') !== null;
		} catch {
			manualTheme = false;
		}
		const mq = matchMedia('(prefers-color-scheme: dark)');
		const onSystemChange = (e: MediaQueryListEvent) => {
			if (!manualTheme) {
				dark = e.matches;
				document.documentElement.classList.toggle('dark', e.matches);
			}
		};
		mq.addEventListener('change', onSystemChange);
		return () => mq.removeEventListener('change', onSystemChange);
	});

	onMount(() => {
		const timers = new SvelteMap<HTMLElement, ReturnType<typeof setTimeout>>();
		const onScroll = (event: Event) => {
			const target = event.target === document ? document.scrollingElement : event.target;
			if (!(target instanceof HTMLElement)) return;
			const timer = timers.get(target);
			if (timer !== undefined) clearTimeout(timer);
			else target.dataset.scrollActive = 'true';
			timers.set(
				target,
				setTimeout(() => {
					target.removeAttribute('data-scroll-active');
					timers.delete(target);
				}, 600)
			);
		};
		document.addEventListener('scroll', onScroll, { capture: true, passive: true });
		return () => {
			document.removeEventListener('scroll', onScroll, true);
			for (const [target, timer] of timers) {
				clearTimeout(timer);
				target.removeAttribute('data-scroll-active');
			}
		};
	});

	function toggleTheme() {
		dark = !dark;
		manualTheme = true;
		document.documentElement.classList.toggle('dark', dark);
		try {
			localStorage.setItem('theme', dark ? 'dark' : 'light');
		} catch {
			return;
		}
	}

	const navItems = [
		{ href: '/devices', label: 'Devices' },
		{ href: '/functions', label: 'Functions' },
		{ href: '/parameter-definitions', label: 'Parameters' },
		{ href: '/criteria-definitions', label: 'Criteria' },
		{ href: '/examples', label: 'Examples' },
		{ href: '/sequences', label: 'Sequences' },
		{ href: '/orphans', label: 'Orphans' },
		{ href: '/import', label: 'Import' }
	] as const;

	function isActive(href: string): boolean {
		const pathname = page.url.pathname;
		return pathname === href || pathname.startsWith(href + '/');
	}
	function isOptionalRoute(routeId: string | null, optionalRoute: string): boolean {
		return routeId === optionalRoute;
	}
	let functionsPrototype = $derived(dev && isOptionalRoute(page.route.id, '/demo/functions-prototype'));
	let prototypePage = $derived(
		functionsPrototype || (dev && isOptionalRoute(page.route.id, '/demo/device-prototype'))
	);
	let referencePage = $derived(
		page.url.pathname === '/devices' ||
			page.url.pathname.startsWith('/devices/') ||
			page.url.pathname === '/functions' ||
			page.url.pathname.startsWith('/functions/')
	);
</script>

<svelte:head>
	<link rel="icon" type="image/x-icon" href={appIcon} />
	<title>Web Program Functions Database</title>
</svelte:head>

<a href="#main-content" class="skip-link">Skip to main content</a>

<header
	class="app-chrome border-border bg-card flex items-center gap-6 border-b px-5 py-2.5"
	class:reference-chrome={referencePage}
	inert={prototypePage}
	style:display={prototypePage ? 'none' : undefined}
>
	<a href={resolve('devices')} class="brand text-foreground font-semibold no-underline">
		<img src={appIcon} alt="" width="28" height="28" />
		{#if referencePage}<span class="brand-name"
				><strong>MasterControl</strong><small>Functions Database</small></span
			>{:else}MasterControl Functions Database{/if}
	</a>
	<nav class="flex items-center gap-4 text-sm" aria-label="Main navigation">
		{#each navItems as item (item.href)}
			<a
				href={resolve(item.href)}
				aria-current={isActive(item.href) ? 'page' : undefined}
				class={isActive(item.href)
					? 'text-foreground font-medium no-underline'
					: 'text-muted-foreground hover:text-foreground no-underline'}>{item.label}</a
			>
		{/each}
	</nav>
	<span class="spacer"></span>
	{#if referencePage}<div class="reference-menu">
			<DropdownMenu.Root
				><DropdownMenu.Trigger
					>{#snippet child({ props })}<Button
							{...props}
							variant="outline"
							size="icon"
							aria-label="Open navigation menu"
							title="Open navigation menu"><Menu data-icon="inline-start" /></Button
						>{/snippet}</DropdownMenu.Trigger
				><DropdownMenu.Content align="end"
					><DropdownMenu.Group aria-label="Main navigation"
						>{#each navItems as item (item.href)}<DropdownMenu.Item textValue={item.label}
								>{#snippet child({ props })}<a
										{...props}
										href={resolve(item.href)}
										aria-current={isActive(item.href) ? 'page' : undefined}>{item.label}</a
									>{/snippet}</DropdownMenu.Item
							>{/each}</DropdownMenu.Group
					></DropdownMenu.Content
				></DropdownMenu.Root
			>
		</div>{/if}
	<Button
		variant="outline"
		size="sm"
		class="theme-toggle"
		onclick={toggleTheme}
		aria-label="Toggle dark mode"
		aria-pressed={dark}
		title={dark ? 'Switch to light mode' : 'Switch to dark mode'}
	>
		{#if dark}<Moon size={16} />{:else}<Sun size={16} />{/if}
		<span class="theme-label">{dark ? 'Dark mode' : 'Light mode'}</span>
	</Button>
</header>

<main
	id="main-content"
	tabindex="-1"
	class="app-workspace"
	data-reference-theme={referencePage ? '' : undefined}
	style:padding={functionsPrototype ? '0' : undefined}
>
	{@render children()}
</main>

<!-- Theme tracks the dark-state below so toasts follow the manual override. -->
<Toaster theme={dark ? 'dark' : 'light'} />

<div style="display:none">
	{#each locales as locale (locale)}
		<a href={resolve(localizeHref(page.url.pathname, { locale }).slice(1) as Path)}>{locale}</a>
	{/each}
</div>

<div style="display:none">
	{#each locales as locale (locale)}
		<a href={resolve(localizeHref(page.url.pathname, { locale }).slice(1) as Path)}>{locale}</a>
	{/each}
</div>

<style>
	:global(html:has([data-reference-theme])) {
		--background: #f5f5f4;
		--foreground: #181a19;
		--card: #fff;
		--card-foreground: #181a19;
		--popover: #fff;
		--popover-foreground: #181a19;
		--primary-foreground: #fff;
		--secondary: #efefed;
		--secondary-foreground: #181a19;
		--muted: #efefed;
		--muted-foreground: #686e6a;
		--accent-foreground: #181a19;
		--border: #dedfdd;
		--input: #dedfdd;
		--device-accent: #181a19;
		--device-badge-foreground: #181a19;
		--device-soft: #efefed;
	}
	:global(html.dark:has([data-reference-theme])) {
		--background: #151716;
		--foreground: #f4f4f2;
		--card: #202221;
		--card-foreground: #f4f4f2;
		--popover: #202221;
		--popover-foreground: #f4f4f2;
		--primary-foreground: #202221;
		--secondary: #303330;
		--secondary-foreground: #f4f4f2;
		--muted: #303330;
		--muted-foreground: #b6bbb7;
		--accent-foreground: #f4f4f2;
		--border: #414542;
		--input: #414542;
		--device-accent: #f4f4f2;
		--device-badge-foreground: #f4f4f2;
		--device-soft: #303330;
	}
	.skip-link {
		position: fixed;
		top: 8px;
		left: 16px;
		transform: translateY(-200%);
		padding: 8px 16px;
		background: var(--background);
		border: 2px solid var(--ring);
		z-index: 200;
	}
	.skip-link:focus {
		transform: translateY(0);
	}
	main {
		min-width: 0;
	}
	.app-chrome,
	.app-workspace {
		--primary: var(--device-accent);
		--accent: var(--device-soft);
		--ring: var(--device-accent);
	}
	.app-chrome {
		min-height: 64px;
		padding: 8px 32px;
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 0;
		overflow-wrap: anywhere;
	}
	.brand img {
		object-fit: contain;
		flex-shrink: 0;
	}
	.app-chrome nav a[aria-current='page'] {
		color: var(--device-accent);
	}
	.app-workspace {
		width: 100%;
		max-width: none;
		margin: 0;
		padding: 28px 32px 64px;
	}
	.app-chrome :global(.theme-toggle) {
		width: 120px;
		min-height: 44px;
		flex-shrink: 0;
		gap: 8px;
	}
	@media (max-width: 1200px) {
		.app-chrome {
			flex-wrap: wrap;
			gap: 12px;
		}
		.brand {
			flex: 1;
		}
		.app-chrome > .spacer {
			display: none;
		}
		.app-chrome nav {
			order: 3;
			flex-basis: 100%;
			max-width: 100%;
			overflow-x: auto;
			padding: 4px 0;
		}
		.app-chrome nav a {
			flex-shrink: 0;
		}
	}
	@media (max-width: 600px) {
		.app-chrome {
			padding: 12px 16px;
		}
		.app-workspace {
			padding: 24px 16px 48px;
		}
		.app-chrome :global(.theme-toggle) {
			width: 44px;
			padding: 0;
		}
		.theme-label {
			display: none;
		}
	}
	@media (max-width: 1000px) {
		header {
			flex-wrap: wrap;
			gap: 12px;
		}
		nav {
			order: 3;
			flex-basis: 100%;
			max-width: 100%;
			overflow-x: auto;
			padding: 4px 0;
		}
		nav a {
			flex-shrink: 0;
		}
		header > :global(button) {
			flex-shrink: 0;
		}
	}
	.reference-chrome {
		min-height: 68px;
		gap: 20px;
	}
	.reference-chrome .brand {
		flex: 0 0 auto;
		gap: 12px;
	}
	.brand-name strong,
	.brand-name small {
		display: block;
	}
	.brand-name strong {
		font-size: 15px;
		line-height: 1.2;
	}
	.brand-name small {
		font-size: 11px;
		font-weight: 400;
		color: var(--muted-foreground);
	}
	.reference-chrome nav {
		flex: 1;
		gap: 4px;
	}
	.reference-chrome nav a {
		flex-shrink: 0;
		padding: 12px 14px;
		border-bottom: 2px solid transparent;
		font-size: 13px;
	}
	.reference-chrome nav a[aria-current='page'] {
		border-bottom-color: var(--foreground);
		font-weight: 700;
	}
	.reference-chrome :global(.theme-toggle) {
		width: 44px;
		min-height: 44px;
		padding: 0;
	}
	.reference-chrome .theme-label {
		display: none;
	}
	.reference-menu {
		display: none;
		margin-left: auto;
	}
	@media (max-width: 1200px) {
		.reference-chrome {
			flex-wrap: nowrap;
		}
		.reference-chrome nav {
			display: none;
		}
		.reference-menu {
			display: block;
		}
	}
	@media (max-width: 600px) {
		.reference-chrome .brand-name {
			display: none;
		}
	}
</style>
