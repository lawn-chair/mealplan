<script lang="ts">
	import type { MealData, RecipeData } from "$lib/types";
    import Card from "$lib/Card.svelte";

    interface Props {
        dialog:HTMLDialogElement;
        meals: MealData[];
        // Including RecipeData is dumb but it clears a vs code error and shouldn't hurt
        onselect: (id: MealData | RecipeData) => void; 
    };
    
    let { dialog = $bindable(), meals, onselect }: Props = $props();

</script>

<dialog bind:this={dialog} class="card fixed z-50 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-md drop-shadow-lg">
    <header class="indent-0 container items-center bg-primary-contrast-800-200">
        <h2 class="h4 float px-4">Add Meal</h2>
        <button class="btn h5 btn-sm btn-circle btn-ghost absolute right-2 top-2" aria-label="Close" onclick={() => {dialog.close()}}>X</button>
    </header>

    <article>
        {#each meals as meal}
            <div class="p-4">
                <Card obj={meal} onselect={onselect} />
            </div>
        {/each}
    </article>
</dialog>