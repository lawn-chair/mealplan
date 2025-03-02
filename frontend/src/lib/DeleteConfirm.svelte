<script lang="ts">
        interface Props {
        dialog:HTMLDialogElement;
        onselect: () => void;
        children?: any;
    };
    
    let { dialog = $bindable(), onselect, children }: Props = $props();

</script>
<div id="overlay" class="fixed hidden z-40 w-screen h-screen inset-0 bg-gray-900 bg-opacity-60"></div>
<dialog bind:this={dialog} aria-modal="true" class="card w-96 fixed z-50 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-md px-8 py-6 space-y-5 drop-shadow-lg">
    <div class="">
        {@render children?.()}
        <footer class="flex justify-end space-x-4 p-4">
            <button class="btn preset-tonal-error" type="submit" onclick={onselect}>Delete</button>
            <button class="btn preset-filled-secondary-500" type="button" onclick={function (event) {event.preventDefault(); dialog.close()}}>Cancel</button>
        </footer>
    </div>  
</dialog>

<style>
::backdrop {
  background-color: darkgray;
  opacity: 0.5;
}
</style>