<script lang="ts">
    import type { MealData, RecipeData } from '$lib/types.js';
    import { Trash2 } from 'lucide-svelte';
    
    interface Props {
        obj?: MealData | RecipeData;
        url?: string;
        compact?: boolean;
        onselect?: (id: MealData|RecipeData) => void;
        ondelete?: (id: MealData|RecipeData) => void;
    }

    let { obj, url, compact = false, onselect, ondelete }: Props = $props();
    
    function selectMeal(event: MouseEvent) {
        event.preventDefault();
        if(obj) {
            onselect?.(obj);
        }
    }

    function deleteMeal(event: MouseEvent) {
        event.preventDefault();
        if(obj) {
            ondelete?.(obj);
        }
    }

    $inspect(obj);
</script>

{#if obj}
<div class="card card-hover relative preset-filled-surface-100-900 border-surface-200-800 w-96 shadow-xl divide-surface-200-800 divide-y overflow-hidden">
    <header class="static block">
        <img src={obj.image && obj.image.Valid ? obj.image.String : "/meal-blank.jpg"} alt={obj.name} class="w-full h-48 object-cover">
        {#if ondelete}
            <button class="btn btn-ghost absolute right-0 bottom-0" type="button" onclick={deleteMeal}><Trash2 /></button>
        {/if}
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