<script>
    import Card from '$lib/Card.svelte';
    import RecipeForm from './[slug]/RecipeForm.svelte';

    /**
     * @typedef {Object} Props
     * @property {import('./$types').PageData} data
     */

    /** @type {Props} */
    let { data } = $props();
    let newRecipe = $state(false);

</script>
<svelte:head>
    <title>Yum! - Recipes</title>
</svelte:head>

<main class="py-6 px-4 sm:p-6 md:py-10 md:px-8">
{#if !newRecipe}
    <div class="container mx-auto flex flex-wrap items-stretch justify-around gap-4">
        {#if data}
            {#each data.recipeData as recipe}
                <Card obj={recipe} url="/recipes/{recipe.slug}"/>
            {/each}
        {:else}
            <p>Loading...</p>
        {/if}
    </div>
    <div class="container my-6">
        <button class="btn btn-base preset-filled-primary-500" onclick={() => {newRecipe = true;}} >New Recipe</button>
    </div>
{:else}
    <RecipeForm newRecipe={true} oncancel={() => {newRecipe = false}}/>
{/if}
</main>


<style>

</style>