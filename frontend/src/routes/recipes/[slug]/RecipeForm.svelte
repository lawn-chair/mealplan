<script lang="ts">
    import { onMount } from "svelte";
    import { API } from '$lib/api';
	import { invalidateAll, goto } from '$app/navigation';
	import { applyAction, deserialize } from '$app/forms';
    import { toaster } from '$lib/toaster-svelte';
    
    import { Trash2, Menu } from 'lucide-svelte';
    
    import type { RecipeData } from '$lib/types.js';
    import DeleteConfirm from "$lib/DeleteConfirm.svelte";
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
   
    async function handleDelete() {
        if (!ctx.session) {
            console.error('Session is not available');
            return;
        }
        const token = await ctx.session.getToken();
        const response = await fetch(API + `/recipes/${data.id}`, {
            method: 'DELETE',
            headers: {
                Authorization: `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            toaster.create({
                title: 'Success',
                description: 'Recipe deleted successfully',
                type: 'success'
            });
        } else {
            toaster.create({
                title: 'Error',
                description: 'Failed to delete recipe',
                type: 'error'
            });
        }
        
        goto('/recipes'); 
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
                
        const response = await fetch(API + (newRecipe ? '/recipes' : `/recipes/${data.id}`), {
			method: newRecipe ? 'POST' : 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
			body: JSON.stringify(data),
		});

		/** @type {import('@sveltejs/kit').ActionResult} */
        const result = deserialize(await response.text());

        if (response.ok) {
            if (result.type === 'success') {
                // rerun all `load` functions, following the successful update
                await invalidateAll();
                
                toaster.create({
                    title: 'Success',
                    description: newRecipe ? 'Recipe created successfully' : 'Recipe updated successfully',
                    type: 'success'
                });
            }
        } else {
            toaster.create({
                title: 'Error',
                description: `Failed to ${newRecipe ? 'create' : 'update'} recipe`,
                type: 'error'
            });
        }

        applyAction(result);
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
                    <Trash2 />
                </button>
                <span class="handle"><Menu /></span>
            </SortableRow>
        {/each}
    </label>
    <div class="my-5">
        <button class="btn" type="button" onclick={function () {data.steps = [...data.steps, {text: '', order: data.steps.length}]}}>Add Step</button>
    </div>
    <div class="container my-4">
        <button class="btn preset-filled-primary-500" type="submit">Save</button>
        <button class="btn preset-filled-secondary-500" type="button" onclick={oncancel}>Cancel</button>    
        {#if !newRecipe}
        <button class="btn preset-filled-error-500" type="button" onclick={() => delete_warning.showModal()}>Delete</button>
        {/if}
    </div>
</form>

<DeleteConfirm bind:dialog={delete_warning} onselect={handleDelete}>
    <h2>Are you sure you want to delete this recipe?</h2>
</DeleteConfirm>
