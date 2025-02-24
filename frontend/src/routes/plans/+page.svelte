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
    <h1>Plans</h1>
    <p>Plans are a way to organize your meals for the week. You can create a plan, add meals to it, and then generate a shopping list based on the meals in the plan.</p>
    <p>Plans can be shared with other users, so you can collaborate on meal planning with your friends and family.</p>
    <p>Click on a plan to view it, or click the "New Plan" button to create a new plan.</p>
    <a href="/plans/new" class="btn btn-primary">New Plan</a>
    <ul>
        {#await data}
            <p>Loading...</p>
        {:then data}
            {#if data.planData?.length === 0}
                <p>No plans found</p>
            {/if}
            {#each data.planData as plan}
            <li>
                <a href="/plans/{plan.id}">{formatDate(plan.start_date)} - {formatDate(plan.end_date)}</a>
            </li>
            {/each}
        {:catch error}
            <p>{error.message}</p>
        {/await}

    </ul>
</div>