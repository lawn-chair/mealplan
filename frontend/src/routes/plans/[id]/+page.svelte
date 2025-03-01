<script lang="ts">
    import { useClerkContext } from 'svelte-clerk';
    import { goto, invalidateAll } from '$app/navigation';
    import type { PlanData, MealData } from '$lib/types';
    import { formatDate } from '$lib/utils';
    import { API } from '$lib/api';
	import MealChooser from './MealChooser.svelte';
    import DeleteConfirm from '$lib/DeleteConfirm.svelte';
	import MealCard from '../../meals/MealCard.svelte';
	import { applyAction, deserialize } from '$app/forms';

    interface Props {
        data: {
            planData: PlanData,
            meals: MealData[]
        };
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

    async function handleSave() {
        let token = await ctx.session?.getToken();

        let response = await fetch(API + `/plans/${planData.id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify(planData),
        });

        /** @type {import('@sveltejs/kit').ActionResult} */
		const result = deserialize(await response.text());

        if (result.type === 'success') {
            // rerun all `load` functions, following the successful update
            await invalidateAll();
            console.log(result);
            applyAction(result);
            editing = false;
        }

    }

    $inspect(planData);

    let dialog :HTMLDialogElement
    let deleteDialog :HTMLDialogElement
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
    <button class="btn btn-primary" onclick={() => {deleteDialog.showModal()}}>Delete Plan</button>
    <button class="btn btn-primary" onclick={() => {editing = !editing}}>Edit Plan</button>
    <h2>Meals:</h2>
    <div class="flex flex-wrap gap-4">
        {#if !planData.meals}
            <li>No meals</li>
        {:else}
            {#each planData.meals as meal}
                {#if data.meals.find(m => m.id === meal)}
                    <MealCard compact={true} meal={data.meals.find(m => m.id === meal )}/>
                {/if}
            {/each}
        {/if}
    </div>
    {#if editing}
    <button class="btn btn-primary" onclick={() => dialog.showModal()}>Add Meal</button>
    <button class="btn btn-primary" onclick={() => {handleSave()}}>Save</button>
    {/if}
{/await}
</div>

<DeleteConfirm bind:dialog={deleteDialog} onselect={deletePlan}>
    <h2>Are you sure you want to delete this plan?</h2>
</DeleteConfirm>    

<MealChooser bind:dialog={dialog} meals={data.meals} onselect={function (meal) {
    if (!planData.meals) {
        planData.meals = [];
    }
    if(meal.id) {
        planData.meals.push(meal.id);
    }
}}/>