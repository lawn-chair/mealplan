<script>
    /** @type {import('./$types').PageData} */
    export let data;
    import MealCard from './MealCard.svelte';
    import Navbar from '../Navbar.svelte'
</script>
<Navbar />

<main class="py-6 px-4 sm:p-6 md:py-10 md:px-8">
    <div class="container mx-auto flex flex-wrap items-center justify-around gap-4">
        {#if data}
            {#each data.mealData as meal}
                <MealCard {meal} />
            {/each}
        {:else}
            <p>Loading...</p>
        {/if}
    </div>
    <div class="container my-6">
        <a class="btn btn-primary" href="/meals/new">New Meal</a>
    </div>
</main>


<style>

</style>