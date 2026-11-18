<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	onMount(() => {
		void goto(resolve('devices'), { replace: true });
	});
</script>

<p class="text-muted-foreground">Redirecting…</p>
