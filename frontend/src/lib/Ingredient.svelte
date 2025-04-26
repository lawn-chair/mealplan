<script lang="ts">
    import { Trash2 } from 'lucide-svelte';

    interface Props {
        ingredient: any;
        i: number;
        onremove: () => void;
    }

    let { ingredient, i, onremove }: Props = $props();


</script>

<div class="flex">
    <div class="w-full">
        <input type="text" class="input join-item input-bordered w-24" name="ingredient.{i}.amount" bind:value={ingredient.amount}>
        <input type="text" class="input join-item input-bordered" name="ingredient.{i}.name" bind:value={ingredient.name}>
    </div>
    <button aria-label="remove" class="btn btn-ghost" type="button" onclick={onremove}>
        <Trash2 />
    </button>
    <!--<input type="hidden" name="ingredient.{i}.calories" value={ingredient.calories}>-->
</div>
