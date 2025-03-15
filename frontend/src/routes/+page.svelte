<script lang="ts">
    import { API } from '$lib/api.js';
    import { useClerkContext } from 'svelte-clerk';
    import { formatDate } from '$lib/utils.js';
    import type { PlanData } from '$lib/types.js';
	import { onMount } from 'svelte';
    
    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';

    let list : {
        plan: PlanData,
        ingredients: {name: string, amount: string}[]
    } = $state();
   
    async function getShoppingList() {
        console.log("here", userId)
        if(userId) {
            let token = await ctx.session?.getToken();
            console.log("again")
            fetch(API + `/shopping-list`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            })
            .then(response => response.json())
            .then(data => {
                console.log(data);
                list = data;
            });
        }
    }

    onMount(() => {
        getShoppingList();
    });

    
</script>

<svelte:head>
    <title>Yum! - Shopping List</title>
</svelte:head>

{#if list}
<div class="container mx-auto">
    <h1 class="h1">Shopping List</h1>
    <div class="card p-4 my-4">
        <h2 class="h2">{formatDate(list.plan.start_date)} - {formatDate(list.plan.end_date)}</h2>
        <ul class="p-4">
            {#each list.ingredients as ingredient}
            <li class="p-2"><input type="checkbox" /> {ingredient.amount} {ingredient.name}</li>
            {/each}
        </ul>
    </div>
</div>
{/if}

