<script lang="ts">
    /** @type {import('./$types').PageData} */
    import Navbar from '../../Navbar.svelte';
    import RecipeForm from './RecipeForm.svelte';
    import type { RecipeData } from '$lib/types.js';
	import Hero from '$lib/Hero.svelte';

    interface Props {
        data: RecipeData;
        form: any;
    }

    let { data, form }: Props = $props();

    let editing = $state(false);
</script>

<Navbar />

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
    <Hero {...data} {editing} fallback='/recipe-blank.jpg' on:button-click={() => {editing = !editing}}/>
    {#if !editing}
    <div class="prose">
    <h3>Ingredients:</h3>
    <ul>
        {#if !data.ingredients}
            <li class="">No ingredients</li>
        {:else}
            {#each data.ingredients as ingredient}
                <li class="">{ingredient.amount} {ingredient.name}</li>
            {/each}
        {/if}
    </ul>
    <h3 class="">Instructions:</h3>
    <ul>
        {#if !data.steps}
            <li>No steps</li>
        {:else}
        {#each data.steps as step}
            <li class="">{step.text}</li>
        {/each}
        {/if}
    </ul>
    </div>
{:else}
    <RecipeForm {data} on:cancel={function () {editing = !editing}}/>
{/if}
{:else}
    <p>Loading...</p>
{/if}
</div>

<style>

</style>
