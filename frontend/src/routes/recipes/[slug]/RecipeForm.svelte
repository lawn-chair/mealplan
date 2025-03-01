<script lang="ts">
    import { onMount } from "svelte";
    import { API } from '$lib/api';
	import { invalidateAll, goto } from '$app/navigation';
	import { applyAction, deserialize } from '$app/forms';

    import type { RecipeData } from '$lib/types.js';
	import Ingredient from '$lib/Ingredient.svelte';
	import SortableRow from '$lib/SortableRow.svelte';
    import { updatePosition } from '$lib/utils';
	import { useClerkContext } from "svelte-clerk";

    interface Props {
        newRecipe?: boolean;
        data?: RecipeData;
        oncancel: () => void;
    }
    // Do not destructure context or you'll lose reactivity!
    const ctx = useClerkContext();
    const userId = $derived(ctx.auth.userId);

    let { newRecipe = false, data = $bindable({
        slug: '',
        id: 0,
        name: '',
        description: '',
        ingredients: [{amount: '', name: ''}],
        steps: [{text: '', order: 0}],
        image: {String: '', Valid: false}
    }), oncancel }: Props = $props();

	onMount(async function() {
        if(!ctx.auth.userId) {
            oncancel();
        }
    });

    async function uploadFile() {
        const file = (document.getElementById('image_file') as HTMLInputElement).files?.[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('file', file);
        console.log(formData);

        if (!ctx.session) {
            console.error('Session is not available');
            return;
        }

        const token = await ctx.session.getToken();
        try {
            const response = await fetch(API + '/images', {
                method: 'POST',
                headers: {
                    Authorization: `Bearer ${token}`
                },
                body: formData
            });

            if(response.ok) {
                const result = await response.json();
                data.image = {String: result.url, Valid: true};
            }
        } catch (error) {
            console.error('Error uploading file:', error);
        }
    }
   
	/** @param {SubmitEvent & { currentTarget: EventTarget & HTMLFormElement}} event */
    async function handleSubmit(event : SubmitEvent) {
		event.preventDefault();
        if (!ctx.session) {
            console.error('Session is not available');
            return;
        }
        const token = await ctx.session.getToken();
        console.log("Sending to: ", API + (newRecipe ? '/recipes' : `/recipes/${data.id}`));
        
        const deleteMeal = event.submitter && 
            (event.submitter as HTMLButtonElement).formAction && 
            (event.submitter as HTMLButtonElement).formAction.includes("?/delete");

		if (deleteMeal) {
            const response = await fetch(API + `/recipes/${data.id}`, {
                method: 'DELETE',
                headers: {
                    Authorization: `Bearer ${token}`
                }
            });
            goto('/recipes');
        }
        
        const response = await fetch(API + (newRecipe ? '/recipes' : `/recipes/${data.id}`), {
			method: newRecipe ? 'POST' : 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
			body: JSON.stringify(data),
		});

		if (response.ok) {
            /** @type {import('@sveltejs/kit').ActionResult} */
            const result = deserialize(await response.text());

            if (result.type === 'success') {
                // rerun all `load` functions, following the successful update
                await invalidateAll();
            }

            applyAction(result);
        }
	}

    let delete_warning : HTMLDialogElement;
</script>
<form method="POST" enctype="multipart/form-data" action={newRecipe ? "/recipes" : "?/update"} onsubmit={handleSubmit}>
    <div>
    <input type="hidden" id="slug" name="slug" value={data.slug} />
    <input type="hidden" id="id" name="id" value={data.id} />

    <label class="" for="name">
        <div class="label"><span class="label-text">Name:</span></div>
        <input class="input input-bordered" type="text" id="name" name="name" bind:value={data.name}>
    </label>

    <label class="form-control" for="description">
       <span class="label label-text">Description:</span>
        <textarea class="textarea textarea-bordered" id="description" name="description" bind:value={data.description}></textarea>
    </label>

    <label class="form-control" for="image">
        <span class="label label-text">Image:</span>
        <div role="tablist" class="tabs tabs-lifted">
            <input type="radio" name="my_tabs_2" role="tab" class="tab" aria-label="URL" checked={true} />
            <div role="tabpanel" class="tab-content bg-base-100 border-base-300 rounded-box p-6">
                <input class="input input-bordered" type="text" id="image" name="image" value={data.image?.String}>
            </div>
          
            <input type="radio" name="my_tabs_2" role="tab" class="tab" aria-label="Upload" />
            <div role="tabpanel" class="tab-content bg-base-100 border-base-300 rounded-box p-6">
              <input type="file" id="image_file" name="image_file" accept="image/*" onchange={uploadFile}>
            </div>
          </div>
        
    </label>

</div>
    <label class="form-control" for="ingredients">
        <span class="label label-text">Ingredients:</span>

        {#each data.ingredients as ingredient, i}
            <Ingredient {ingredient} {i} 
                onremove={function () {data.ingredients.splice(i, 1); data.ingredients = data.ingredients}}/> 
        {/each}
    </label>
    <div class="my-5">
    <button class="btn" type="button" onclick={function () {data.ingredients = [...data.ingredients, {amount: '', name: '', calories: 0}]}}>Add Ingredient</button>
    </div>

    <label class="form-control" for="steps">
        <span class="label label-text">Steps:</span>
        {#each data.steps as step, i}
            <SortableRow class="flex" data-id={i} 
                    handleClass="handle" 
                    this="div"
                    drag={(from, to) => updatePosition(data.steps, from, to)}>
                <input type="text" class="input input-bordered w-full" name="step.{step.order}.text" bind:value={step.text}>
                <input class="ghost drag" type="hidden" name="step.{step.order}.order" bind:value={step.order}>
                <button class="btn btn-ghost" aria-label="delete" type="button" onclick={function () {data.steps.splice(i, 1); data.steps = data.steps;}}>
                    <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 36 36"><path fill="currentColor" d="M27.14 34H8.86A2.93 2.93 0 0 1 6 31V11.23h2V31a.93.93 0 0 0 .86 1h18.28a.93.93 0 0 0 .86-1V11.23h2V31a2.93 2.93 0 0 1-2.86 3" class="clr-i-outline clr-i-outline-path-1"/><path fill="currentColor" d="M30.78 9H5a1 1 0 0 1 0-2h25.78a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-2"/><path fill="currentColor" d="M21 13h2v15h-2z" class="clr-i-outline clr-i-outline-path-3"/><path fill="currentColor" d="M13 13h2v15h-2z" class="clr-i-outline clr-i-outline-path-4"/><path fill="currentColor" d="M23 5.86h-1.9V4h-6.2v1.86H13V4a2 2 0 0 1 1.9-2h6.2A2 2 0 0 1 23 4Z" class="clr-i-outline clr-i-outline-path-5"/><path fill="none" d="M0 0h36v36H0z"/></svg>
                </button>
                <span class="handle"><svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 36 36"><path fill="currentColor" d="M32 29H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-1"/><path fill="currentColor" d="M32 19H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-2"/><path fill="currentColor" d="M32 9H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-3"/><path fill="none" d="M0 0h36v36H0z"/></svg></span>
            </SortableRow>
        {/each}
    </label>
    <div class="my-5">
        <button class="btn" type="button" onclick={function () {data.steps = [...data.steps, {text: '', order: data.steps.length}]}}>Add Step</button>
    </div>
    <div class="container my-4">
        <button class="btn btn-primary" type="submit">Save</button>
        <button class="btn btn-secondary" type="button" onclick={oncancel}>Cancel</button>    
        {#if !newRecipe}
        <button class="btn btn-error" type="button" onclick={() => delete_warning.showModal()}>Delete</button>
        <dialog bind:this={delete_warning} class="modal modal-bottom sm:modal-middle">
            <div class="p-4 modal-box">
                <h2>Are you sure you want to delete this recipe?</h2>
                <button class="btn btn-error" formaction="?/delete">Delete</button>
                <button class="btn btn-secondary" onclick={(e) => {e.preventDefault(); delete_warning.close();}}>No</button>
            </div>
        </dialog>
        {/if}
    </div>
</form>
