<script lang="ts">
  	import "../app.css";
	import type { Snippet } from 'svelte';
	import { ClerkProvider } from 'svelte-clerk';

	const { children }: { children: Snippet } = $props();
</script>

<!-- ... -->

<ClerkProvider publishableKey="pk_test_Y2hpZWYtZmx5LTYzLmNsZXJrLmFjY291bnRzLmRldiQ">
	{@render children()}
</ClerkProvider>