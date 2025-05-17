<script lang="ts">
	import { API } from '$lib/api.js';
	import { useClerkContext } from 'svelte-clerk';
	import { Trash2 } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { toaster } from '$lib/toaster-svelte';
	import { getContext } from 'svelte';

	const ctx = useClerkContext();
	const userId = ctx.auth.userId || '';
	let pantry: { items: string[] } = $state({ items: [] });

	async function getPantry() {
		if (userId) {
			let token = await ctx.session?.getToken();
			fetch(API + `/pantry`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			})
				.then((response) => response.json())
				.then((data) => {
					console.log(data);
					pantry = data;
				});
		}
	}

	async function updatePantry() {
		if (userId) {
			let token = await ctx.session?.getToken();
			fetch(API + `/pantry`, {
				method: 'PUT',
				body: JSON.stringify(pantry),
				headers: {
					Authorization: `Bearer ${token}`
				}
			})
				.then((response) => {
					if (response.ok) {
						console.log('Pantry updated successfully');
						toaster.create({
							title: 'Saved',
							description: 'Pantry was successfully updated',
							type: 'success'
						});
					} else {
						console.error('Failed to update pantry');
						toaster.create({
							title: 'Error',
							description: 'Failed to update pantry',
							type: 'error'
						});
					}
					return response.json();
				})
				.then((data) => {
					console.log(data);
					pantry = data;
				});
		}
	}

	onMount(() => {
		getPantry();
	});
</script>

<svelte:head>
	<title>Yum! - Pantry</title>
</svelte:head>

<main class="px-4 py-6 sm:p-6 md:px-8 md:py-10">
	<div class="mx-auto gap-4">
		<h1 class="text-3xl">Pantry</h1>
		<p class="text-xs">
			The Pantry represents items you always have on hand and will be excluded from your shopping
			list.
		</p>
	</div>
	<div class="container mx-auto gap-4">
		{#if pantry}
			{#each pantry.items as item, i}
				<div class="flex">
					<p class="text-2xl capitalize">{item}</p>
					<button
						class="btn btn-ghost"
						type="button"
						aria-label="remove"
						onclick={() => {
							pantry.items.splice(i, 1);
							updatePantry();
						}}><Trash2 /></button
					>
				</div>
			{/each}
			<form
				class="flex flex-col gap-4"
				onsubmit={() => {
					updatePantry();
				}}
			>
				<input
					type="text"
					class="input input-bordered w-full"
					name="item"
					placeholder="Add item..."
					onchange={(e: Event) => {
						if (!(e.target instanceof HTMLInputElement) || !e.target.value) {
							return;
						}
						pantry.items.push(e.target?.value);
						e.target.value = '';
						e.target.focus();
					}}
				/>
				<button class="btn btn-base preset-filled-primary-500" type="submit">Add</button>
			</form>
		{:else}
			<p>Loading...</p>
		{/if}
	</div>
</main>
