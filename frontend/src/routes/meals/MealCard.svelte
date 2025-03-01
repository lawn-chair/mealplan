<script lang="ts">
    import type { MealData } from '$lib/types.js';
    
    interface Props {
        meal: MealData;
        selector?: boolean;
        compact?: boolean;
        onselect?: (meal: MealData) => void;

    }

    let { meal, selector = false, compact = false, onselect = () => {} }: Props = $props();
    
    function selectMeal(event: MouseEvent) {
        event.preventDefault();
        onselect(meal);
    }
</script>

<div class="card w-96 {selector || compact? 'card-side card-compact' : 'card-normal'} shadow-xl">
{#if meal}
    <figure>
        <img src="{meal.image && meal.image.Valid ? meal.image.String : "/meal-blank.jpg"}" alt="{meal.name}" class="w-full h-48 object-cover">
    </figure>
    <div class="card-body prose">
        <h1 class="card-title"><a href="/meals/{meal.slug}">{meal.name}</a></h1>
        <p>{meal.description}</p>
        {#if selector}
            <button class="btn btn-primary" onclick={selectMeal}>Select</button>
        {/if}
    </div>
{:else}
    <p>Loading...</p>
{/if}
</div>

<style>

</style>