<script>
    /** @type {import('./$types').PageData} */
    export let data;
	import RecipeCard from '$lib/RecipeCard.svelte';
    import RecipeForm from './[slug]/RecipeForm.svelte';
    import Navbar from '../Navbar.svelte'
    let newRecipe = false;

</script>
<Navbar />

<main class="py-6 px-4 sm:p-6 md:py-10 md:px-8">
{#if !newRecipe}
    <div class="container mx-auto flex flex-wrap items-center justify-around gap-4">
        {#if data}
            {#each data.recipeData as recipe}
                <RecipeCard {recipe} />
            {/each}
        {:else}
            <p>Loading...</p>
        {/if}
    </div>
    <div class="container my-6">
        <button class="btn btn-primary" on:click={() => {newRecipe = true;}} >New Recipe</button>
    </div>
{:else}
    <RecipeForm newRecipe={true} on:cancel={() => {newRecipe = false}}/>
{/if}
</main>


<style>

</style>