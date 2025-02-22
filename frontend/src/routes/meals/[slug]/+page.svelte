<script lang="ts">
    import MealForm from '../MealForm.svelte';
    import RecipeCard from '$lib/RecipeCard.svelte'
    import Hero from '$lib/Hero.svelte';
    import type { RecipeData, MealData } from '$lib/types.js';
    import { API } from '$lib/api.js';
    import { useClerkContext } from 'svelte-clerk';

    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';

    interface Props {
        data: MealData;
        form: any;
    }

    let { data, form }: Props = $props();
    let mealData = $derived.by(() => {
        let mealData = $state(data);
        return mealData;
    });

    console.log(form);
    let editing = $state(false);

    let recipes : RecipeData[] = $state([]);
    function getRecipes() {
        fetch(API + `/recipes`)
            .then(response => response.json())
            .then(data => {
                recipes = data;
            });
    }
    getRecipes();
</script>

<svelte:head>
    <title>Yum! - Meals - {data.name}</title>
</svelte:head>


{#if form?.message}
    <div class="toast toast-top toast-end">
        <span role="alert" class="alert alert-error">Update Failed: {JSON.parse(form?.message).message}</span>
    </div>
{:else if form?.success}
    <div class="toast toast-top toast-end">
        <span role="alert" class="alert alert-success">Meal Updated</span>
    </div>
{/if}

<div class="container mx-auto">
{#if mealData}

    <Hero {...mealData} {editing} fallback='/meal-blank.jpg' {userId} onclick={() => {editing = !editing}}/>

    {#if !editing}
    <div class="prose">
        <h2>Ingredients:</h2>
        <ul>
        {#each mealData.ingredients as ingredient}
            <li>{ingredient.amount} {ingredient.name}</li>
        {/each}
        </ul>
    </div>
    <div class="prose">
        <h2>Steps:</h2>
        <ol>
        {#each mealData.steps as step}
            <li>{step.text}</li>
        {/each}
        </ol>
    </div>
    <div>
        <span class="prose"><h2>Recipes:</h2></span>
        {#await recipes}
            <p>Loading...</p>
        {:then recipes}
            {#if !mealData.recipes || mealData.recipes.length == 0}
                <p>No recipes</p>
            {:else}
                <div class="flex flex-wrap gap-4">
                {#each mealData.recipes as recipe}
                    {#if recipes.find((x) => x.id === recipe.recipe_id)}
                        <RecipeCard recipe={recipes.find((x) => x.id === recipe.recipe_id) || { id: 0, name: '', description: '', slug: '', ingredients: [], steps: [] }} 
                            compact={true} />
                    {/if}
                {/each}
                </div>
            {/if}
        {/await}
    </div>

    {:else}
    <MealForm data={mealData} oncancel={() => {editing = !editing}} onsave={() => {editing = !editing}}/>
    {/if}
{:else}
    <p>Loading...</p>
{/if}
</div>

<style>

</style>
