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
    <div class="prose">
    <h3>Ingredients:</h3>
    <ul>
        {#if !recipeData.ingredients}
            <li class="">No ingredients</li>
        {:else}
            {#each recipeData.ingredients as ingredient}
                <li class="">{ingredient.amount} {ingredient.name}</li>
            {/each}
        {/if}
    </ul>
    <h3 class="">Instructions:</h3>
    <ul>
        {#if !recipeData.steps}
            <li>No steps</li>
        {:else}
        {#each recipeData.steps as step}
            <li class="">{step.text}</li>
        {/each}
        {/if}
    </ul>
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
