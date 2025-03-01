<script lang="ts">
	import type { MealData } from "$lib/types";
	import MealCard from "../../meals/MealCard.svelte";

    interface Props {
        dialog:HTMLDialogElement;
        meals: MealData[];
        onselect: (meal: MealData) => void;
    };
    
    let { dialog = $bindable(), meals, onselect }: Props = $props();

</script>

<dialog bind:this={dialog} class="modal" id="modal">
    <div class="modal-box">
        <div class="modal-header">
            <h2 class="modal-title">Add Meal</h2>
        </div>
        <button class="btn btn-clear float-right" aria-label="Close" onclick={() => {dialog.close()}}>x</button>

        <div class="modal-body">
            {#each meals as meal}
                <MealCard {meal} selector={true} onselect={onselect} />
            {/each}
        </div>
    </div>
</dialog>