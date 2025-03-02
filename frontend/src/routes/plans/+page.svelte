<script lang="ts">
    import { useClerkContext } from 'svelte-clerk';
    import type { PlanData } from '$lib/types.js';
    import { formatDate } from '$lib/utils';

    interface Props {
        data: { planData: PlanData[] };
    }

    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';

    let { data } : Props = $props();
    $inspect(data);
</script>

<svelte:head>
    <title>Yum! - Plans</title>
</svelte:head>

<div class="container mx-auto">
    <h1 class="h3">Plans</h1>
    <div class="text-xs p-4 card background-secondary-100">
        <p class="text-xs">Plans are a way to organize your meals for the week. You can create a plan, add meals to it, and then generate a shopping list based on the meals in the plan.</p>
        <p class="text-xs">Plans can be shared with other users, so you can collaborate on meal planning with your friends and family.</p>
        <p class="text-xs">Click on a plan to view it, or click the "New Plan" button to create a new plan.</p>
    </div>
    <a href="/plans/new" class="btn preset-filled-primary-500">New Plan</a>
    <p class="h5 p-4">Upcoming Meal Plans:</p>    
    <ul class="p-4">
        {#await data}
            <p>Loading...</p>
        {:then data}
            {#if data.planData?.length === 0}
                <p>No plans found</p>
            {/if}
            {#each data.planData as plan}
            <li>
                <a class="text-xl p-4" href="/plans/{plan.id}">{formatDate(plan.start_date)} - {formatDate(plan.end_date)}</a>
            </li>
            {/each}
        {:catch error}
            <p>{error.message}</p>
        {/await}

    </ul>
</div>