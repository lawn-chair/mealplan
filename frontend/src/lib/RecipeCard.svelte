<script lang="ts">
     import { createEventDispatcher } from "svelte";
     import type { RecipeData } from '$lib/types.js';

    export let recipe : RecipeData;
    export let selector = false;
    export let compact = false;
    console.log(recipe)
    const dispatch = createEventDispatcher();
    
    function selectRecipe() {
        dispatch('select', recipe);
    }
</script>

<div class="card w-96 {selector || compact? 'card-side card-compact' : 'card-normal'} shadow-xl">
{#if recipe}
        <figure>
            <img src="{recipe.image && recipe.image.Valid ? recipe.image.String : "/recipe-blank.jpg"}" alt="{recipe.name}" class="w-full h-48 object-cover">
        </figure>
        <div class="card-body">

        <h1 class="prose card-title"><a href="/recipes/{recipe.slug}">{recipe.name}</a></h1>
        <p>{recipe.description}</p>
        {#if selector}
            <button class="btn btn-primary" on:click|preventDefault={selectRecipe}>Select</button>
        {/if}
    </div>
{:else}
    <p>Loading...</p>
{/if}
</div>

<style>

</style>
