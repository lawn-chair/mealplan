<script lang="ts">
    import MealForm from '../MealForm.svelte';
    import Card from '$lib/Card.svelte'
    import Hero from '$lib/Hero.svelte';
    import type { RecipeData, MealData } from '$lib/types.js';
    import { API } from '$lib/api.js';
    import { useClerkContext } from 'svelte-clerk';

    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';

    interface Props {
        data: MealData;
    }

    let { data }: Props = $props();
    let mealData = $derived.by(() => {
        let mealData = $state(data);
        return mealData;
    });

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

<div class="container mx-auto">
{#if mealData}

    <Hero {...mealData} {editing} fallback='/meal-blank.jpg' {userId} onclick={() => {editing = !editing}}/>

    {#if !editing}
    <div class="container flex flex-col md:flex-row gap-4">
        <div class="p-4 card backdrop-brightness-200 h-full md:w-1/3">
            <h2 class="h4">Ingredients:</h2>
            <ul class="list-disc p-4">
            {#each mealData.ingredients as ingredient}
                <li>{ingredient.amount} {ingredient.name}</li>
            {/each}
            </ul>
        </div>

        <div class="p-4 md:w-2/3">
            <h2 class="h4">Steps:</h2>
            <ol class="list-decimal p-4">
            {#each mealData.steps as step}
                <li>{step.text}</li>
            {/each}
            </ol>
        </div>
    </div>
    <div class="p-4 md:w-2/3">
        <span><h2 class="h4">Recipes:</h2></span>
        {#await recipes}
            <p>Loading...</p>
        {:then recipes}
            {#if !mealData.recipes || mealData.recipes.length == 0}
                <p>No recipes</p>
            {:else}
                <div class="flex flex-wrap gap-4">
                {#each mealData.recipes as recipe}
                    {#if recipes.find((x) => x.id === recipe.recipe_id)}
                        <Card obj={recipes.find((x) => x.id === recipe.recipe_id) || { id: 0, name: '', description: '', slug: '', ingredients: [], steps: [] }} 
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
