<script lang="ts">
    import { useClerkContext } from 'svelte-clerk';
    import { goto } from '$app/navigation';
    import type { PlanData } from '$lib/types';
    import { formatDate } from '$lib/utils';
    import { API } from '$lib/api';

    interface Props {
        data: {planData: PlanData};
        form: any;
    }

    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';

    let { data, form }: Props = $props();
    let planData = $derived.by(() => {
        let planData = $state(data.planData);
        return planData;
    });

    let editing = $state(false);

    async function deletePlan() {
        let token = await ctx.session?.getToken();

        await fetch(API + `/plans/${planData.id}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
        });
        goto('/plans/');
    }

    $inspect(planData)
</script>

<svelte:head>
    <title>Yum! - Plans - {planData.start_date} - {planData.end_date}</title>
</svelte:head>

{#if form?.message}
    <div class="toast toast-top toast-end">
        <span role="alert" class="alert alert-error">Update Failed: {JSON.parse(form?.message).message}</span>
    </div>
{:else if form?.success}
    <div class="toast toast-top toast-end">
        <span role="alert" class="alert alert-success">Plan Updated</span>
    </div>
{/if}

<div class="container mx-auto">
{#await planData}
    <p>Loading...</p>
{:then planData}
    <h1>{formatDate(planData.start_date)} - {formatDate(planData.end_date)}</h1>
    <button class="btn btn-primary" onclick={deletePlan}>Delete Plan</button>
    <button class="btn btn-primary" onclick={() => {editing = !editing}}>Edit Plan</button>
    <div class="prose">
        <h2>Meals:</h2>
        <ul>
            {#if !planData.meals}
                <li>No meals</li>
            {:else}
                {#each planData.meals as meal}
                    <li>{meal}</li>
                {/each}
            {/if}
        </ul>
    </div>
{/await}
</div>