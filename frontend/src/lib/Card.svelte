<script lang="ts">
    import type { MealData, RecipeData } from '$lib/types.js';
    
    interface Props {
        obj?: MealData | RecipeData;
        url?: string;
        compact?: boolean;
        onselect?: (id: MealData|RecipeData|undefined) => void;
    }

    let { obj, url, compact = false, onselect }: Props = $props();
    
    function selectMeal(event: MouseEvent) {
        event.preventDefault();
        onselect?.(obj);
    }

    $inspect(obj);
</script>

{#if obj}
<div class="card card-hover preset-filled-surface-100-900 border-surface-200-800 w-96 shadow-xl divide-surface-200-800 block divide-y overflow-hidden">
    <header>
        <img src={obj.image && obj.image.Valid ? obj.image.String : "/meal-blank.jpg"} alt={obj.name} class="w-full h-48 object-cover">
    </header>
    <article class="p-4">
        <h1 class="text-2xl font-bold">
            {#if url}
                <a href={url}>{obj.name}</a>
            {:else}
                {obj.name}
            {/if}
        </h1>
        {#if !compact}
        <p>{obj.description}</p>
        {/if}
    </article>
    {#if onselect}
    <footer>
            <button class="btn preset-filled-primary-500" onclick={selectMeal}>Select</button>
    </footer>
    {/if}
</div>
{/if}

<style>

</style>