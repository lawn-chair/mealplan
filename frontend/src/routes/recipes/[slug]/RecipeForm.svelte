<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { onMount } from "svelte";
    import Sortable from 'sortablejs';
    import { API } from '$lib/api';
	import { invalidateAll, goto } from '$app/navigation';
	import { applyAction, deserialize } from '$app/forms';

    import type { RecipeData } from '$lib/types.js';
	import Ingredient from '$lib/Ingredient.svelte';
	import { parseFormValues } from "$lib/utils";

    const dispatch = createEventDispatcher();

    function cancelForm() {
        dispatch('cancel');
    }

    export let newRecipe = false;
    export let data: RecipeData = {
        slug: '',
        id: 0,
        name: '',
        description: '',
        ingredients: [{amount: '', name: ''}],
        steps: [{text: '', order: 0}],
        image: {String: '', Valid: false}
    };

    let steps : HTMLElement;
    let sortable : Sortable;
	onMount(async function() {
		sortable = Sortable.create(steps, {
            handle: '.handle',
            dragClass: 'drag',
            ghostClass: 'ghost',
			animation: 200,
		});
    });

    function updateSteps() {
        const order = sortable.toArray();  
        console.log("before: ", data.steps);                  
        data.steps = order.map((step, i) => {
            return {
                text: data.steps[parseInt(step)].text, 
                order: i
            };
        });
        console.log("after: ", data.steps);
    }

	/** @param {SubmitEvent & { currentTarget: EventTarget & HTMLFormElement}} event */
    async function handleSubmit(event) {
		event.preventDefault();
        updateSteps();
		const data = new FormData(event.currentTarget);
        const parsedData = parseFormValues(data);

		const response = await fetch(API + '/recipes', {
			method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
			body: JSON.stringify(parsedData),
		});

		/** @type {import('@sveltejs/kit').ActionResult} */
		const result = deserialize(await response.text());

		if (result.type === 'success') {
			// rerun all `load` functions, following the successful update
			await invalidateAll();
		}

		applyAction(result);
	}

    let delete_warning : HTMLDialogElement;

</script>
<form method="POST" enctype="multipart/form-data" action={newRecipe ? "/recipes" : "?/update"} on:submit={updateSteps}>
    <fieldset>
    <input type="hidden" id="slug" name="slug" value={data.slug} />
    <input type="hidden" id="id" name="id" value={data.id} />

    <label class="form-control" for="name">
        <div class="label"><span class="label-text">Name:</span></div>
        <input class="input input-bordered" type="text" id="name" name="name" value={data.name}>
    </label>

    <label class="form-control" for="description">
       <span class="label label-text">Description:</span>
        <textarea class="textarea textarea-bordered" id="description" name="description">{data.description}</textarea>
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
              <input type="file" id="image_file" name="image_file" accept="image/*">
            </div>
          </div>
        
    </label>

    </fieldset>
    <label class="form-control" for="ingredients">
        <span class="label label-text">Ingredients:</span>

        {#each data.ingredients as ingredient, i}
            <Ingredient {ingredient} {i} 
                on:remove={function () {data.ingredients.splice(i, 1); data.ingredients = data.ingredients}}/> 
        {/each}
    </label>
    <div class="my-5">
    <button class="btn" type="button" on:click={function () {data.ingredients = [...data.ingredients, {amount: '', name: '', calories: 0}]}}>Add Ingredient</button>
    </div>

    <label class="form-control" for="steps">
        <span class="label label-text">Steps:</span>
        <section bind:this={steps}>
        {#each data.steps as step, i}
            <div class="flex" data-id={i}>

                <input type="text" class="input input-bordered w-full" name="step.{step.order}.text" value={step.text}>
                <input class="ghost drag" type="hidden" name="step.{step.order}.order" value={step.order}>
                <button class="btn btn-ghost" type="button" on:click={function () {data.steps.splice(i, 1); data.steps = data.steps;}}>
                    <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 36 36"><path fill="currentColor" d="M27.14 34H8.86A2.93 2.93 0 0 1 6 31V11.23h2V31a.93.93 0 0 0 .86 1h18.28a.93.93 0 0 0 .86-1V11.23h2V31a2.93 2.93 0 0 1-2.86 3" class="clr-i-outline clr-i-outline-path-1"/><path fill="currentColor" d="M30.78 9H5a1 1 0 0 1 0-2h25.78a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-2"/><path fill="currentColor" d="M21 13h2v15h-2z" class="clr-i-outline clr-i-outline-path-3"/><path fill="currentColor" d="M13 13h2v15h-2z" class="clr-i-outline clr-i-outline-path-4"/><path fill="currentColor" d="M23 5.86h-1.9V4h-6.2v1.86H13V4a2 2 0 0 1 1.9-2h6.2A2 2 0 0 1 23 4Z" class="clr-i-outline clr-i-outline-path-5"/><path fill="none" d="M0 0h36v36H0z"/></svg>
                </button>
                <span class="handle"><svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 36 36"><path fill="currentColor" d="M32 29H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-1"/><path fill="currentColor" d="M32 19H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-2"/><path fill="currentColor" d="M32 9H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-3"/><path fill="none" d="M0 0h36v36H0z"/></svg></span>
            </div>
        {/each}
        </section>
    </label>
    <div class="my-5">
        <button class="btn" type="button" on:click={function () {data.steps = [...data.steps, {text: '', order: data.steps.length}]}}>Add Step</button>
    </div>
    <div class="container my-4">
        <button class="btn btn-primary" type="submit">Save</button>
        <button class="btn btn-secondary" type="button" on:click={cancelForm}>Cancel</button>    
        {#if !newRecipe}
        <button class="btn btn-error" type="button" on:click={() => delete_warning.showModal()}>Delete</button>
        <dialog bind:this={delete_warning} class="modal modal-bottom sm:modal-middle">
            <div class="p-4 modal-box">
                <h2>Are you sure you want to delete this recipe?</h2>
                <button class="btn btn-error" formaction="?/delete">Delete</button>
                <button class="btn btn-secondary" on:click|preventDefault={() => delete_warning.close()}>No</button>
            </div>
        </dialog>
        {/if}
    </div>
</form>
