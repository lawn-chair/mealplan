<script lang="ts">
    import Card from '$lib/Card.svelte';
	import type { MealData } from '$lib/types';

    /**
     * @typedef {Object} Props
     * @property {import('./$types').PageData} data
     */

    interface Props {
        data: {
            mealData: MealData[]
        }
    }
    
    let {data} : Props = $props();
</script>

<svelte:head>
    <title>Yum! - Meals</title>
</svelte:head>

<main class="py-6 px-4 sm:p-6 md:py-10 md:px-8">
    <div class="container mx-auto flex flex-wrap items-center justify-around gap-4">
        {#if data}
            {#each data.mealData as meal}
                <Card obj={meal}  url="/meals/{meal.slug}" />
            {/each}
        {:else}
            <p>Loading...</p>
        {/if}
    </div>
    <div class="container my-6">
        <a class="btn btn-base preset-filled-primary-500" href="/meals/new">New Meal</a>
    </div>
</main>


<style>

</style>