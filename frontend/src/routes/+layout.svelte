<script lang="ts">
  import "../app.css";
	import type { Snippet } from 'svelte';
	import { ClerkProvider } from 'svelte-clerk';

	const { children }: { children: Snippet } = $props();
</script>

<!-- ... -->

<ClerkProvider>
	{@render children()}
</ClerkProvider>