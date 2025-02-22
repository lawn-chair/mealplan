<script lang="ts">
	import { env } from '$env/dynamic/public';
  	import "../app.css";
	import type { Snippet } from 'svelte';
	import { ClerkProvider } from 'svelte-clerk';
	import Navbar from './Navbar.svelte';

	const { children }: { children: Snippet } = $props();
</script>

<svelte:head>
	<title>Yum!</title>
</svelte:head>
<!-- ... -->
<ClerkProvider publishableKey={env.PUBLIC_CLERK_PUBLISHABLE_KEY}>
	<Navbar />
	{@render children()}
</ClerkProvider> 


