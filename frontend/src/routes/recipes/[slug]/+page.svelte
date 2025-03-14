<script lang="ts">
    /** @type {import('./$types').PageData} */
    import RecipeForm from './RecipeForm.svelte';
    import type { RecipeData } from '$lib/types.js';
	import Hero from '$lib/Hero.svelte';
	import { useClerkContext } from 'svelte-clerk';

    interface Props {
        data: RecipeData;
        form: any;
    }
    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';
    
    let { data, form }: Props = $props();
    let recipeData = $derived.by(() => {
        let recipeData = $state(data);
        return recipeData;
    });

    let editing = $state(false);
</script>

<svelte:head>
    <title>Yum! - Recipes - {data.name}</title>
</svelte:head>

{#if form?.message}
    <div class="toast toast-top toast-end">
        <span role="alert" class="alert alert-error">Update Failed: {JSON.parse(form?.message).message}</span>
    </div>
{:else if form?.success}
    <div class="toast toast-top toast-end">
        <span role="alert" class="alert alert-success">Recipe Updated</span>
    </div>
{/if}

<div class="container mx-auto">
{#if data}
    <Hero {...recipeData} {editing} fallback='/recipe-blank.jpg' {userId} onclick={() => {editing = !editing}}/>
    {#if !editing}
    <div class="container flex flex-col md:flex-row gap-4">
        <div class="p-4 card backdrop-brightness-200 h-full md:w-1/3">
            <h3 class="h4">Ingredients:</h3>
            <ul class="list-disc p-4">
                {#if !recipeData.ingredients}
                    <li class="">No ingredients</li>
                {:else}
                    {#each recipeData.ingredients as ingredient}
                        <li class="">{ingredient.amount} {ingredient.name}</li>
                    {/each}
                {/if}
            </ul>
        </div>

        <div class="p-4 md:w-2/3">
            <h3 class="h4">Instructions:</h3>
            <ul class="list-decimal p-4">
                {#if !recipeData.steps}
                    <li>No steps</li>
                {:else}
                {#each recipeData.steps as step}
                    <li class="">{step.text}</li>
                {/each}
                {/if}
            </ul>
        </div>
    </div>
{:else}
    <RecipeForm data={recipeData} oncancel={function () {editing = !editing}}/>
{/if}
{:else}
    <p>Loading...</p>
{/if}
</div>

<style>

</style>
