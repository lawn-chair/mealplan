<script lang="ts">
        interface Props {
        dialog:HTMLDialogElement;
        onselect: () => void;
        children?: any;
    };
    
    let { dialog = $bindable(), onselect, children }: Props = $props();

</script>
<dialog bind:this={dialog} class="modal modal-bottom sm:modal-middle">
    <div class="modal-box p-4">
        {@render children?.()}
        <button class="btn btn-error" type="submit" onclick={onselect}>Delete</button>
        <button class="btn btn-secondary" type="button" onclick={function (event) {event.preventDefault(); dialog.close()}}>Cancel</button>
    </div>  
</dialog>