<script lang="ts">
  	import "../app.css";
	import type { Snippet } from 'svelte';
	import { ClerkProvider } from 'svelte-clerk';
	import Navbar from './Navbar.svelte';

	const { children }: { children: Snippet } = $props();
</script>

<!-- ... -->
<ClerkProvider publishableKey="pk_test_Y2hpZWYtZmx5LTYzLmNsZXJrLmFjY291bnRzLmRldiQ">
	<Navbar />
	{@render children()}
</ClerkProvider> 


