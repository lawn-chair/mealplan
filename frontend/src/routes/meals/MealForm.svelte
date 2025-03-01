<script lang="ts">
	import { onMount } from "svelte";
    
    import Card from "$lib/Card.svelte";
    import Ingredient from "$lib/Ingredient.svelte";
    import DeleteConfirm from "$lib/DeleteConfirm.svelte";

    import { API } from '$lib/api.js';
    import { invalidateAll, goto } from '$app/navigation';
	import { applyAction, deserialize } from '$app/forms';
    import { useClerkContext } from 'svelte-clerk';

    import type { RecipeData, MealData } from '$lib/types.js';
	import SortableRow from "$lib/SortableRow.svelte";
    import { updatePosition } from "$lib/utils.js";
	

    interface Props {
        newMeal: boolean;
        data: MealData;
        oncancel: () => void;
        onsave: (result: any) => void;
    };

    let { newMeal = false, data = $bindable({
        slug: '',
        id: 0,
        name: '',
        description: '',
        ingredients: [],
        steps: [],
        recipes: [],
        image: {String: '', Valid: false}
    }), oncancel, onsave }:Props = $props();

    // Do not destructure context or you'll lose reactivity!
    const ctx = useClerkContext();
    const userId = $derived(ctx.auth.userId);
        
    let recipes : RecipeData[] = $state([]);

    function getRecipes() {
        fetch(API + `/recipes`)
            .then(response => response.json())
            .then(data => {
                recipes = data;
            });
    }
    
    function handleAddRecipe(recipe : any) {
        const recipe_id = recipe.id;
        fetch(API + `/recipes/${recipe_id}`)
        .then(response => response.json())
        .then(apiData => {
            data.recipes = [...data.recipes, {recipe_id: recipe_id}];
            data.ingredients = [...data.ingredients, ...apiData.ingredients];
            data.steps = [...data.steps, ...apiData.steps];
        });
    }

	onMount(async function() {
        if(!ctx.auth.userId) {
            oncancel();
        }

        getRecipes();
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
    const response = await fetch(API + `/meals/${data.id}`, {
        method: 'DELETE',
        headers: {
            Authorization: `Bearer ${token}`
        }
    });
    goto('/meals');
}

async function handleSubmit(event : SubmitEvent) {
        event.preventDefault();
        //console.log(event);
        if (!ctx.session) {
            console.error('Session is not available');
            return;
        }
        const token = await ctx.session.getToken();
        console.log("Sending to: ", API + (newMeal ? '/meals' : `/meals/${data.id}`));
        
        const response = await fetch(API + (newMeal ? '/meals' : `/meals/${data.id}`), {
			method: newMeal ? 'POST' : 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
			body: JSON.stringify(data),
		});

		/** @type {import('@sveltejs/kit').ActionResult} */
		const result = deserialize(await response.text());

		if (result.type === 'success') {
			// rerun all `load` functions, following the successful update
			await invalidateAll();
		}
        console.log(result);
		applyAction(result);
        onsave(result);
	}

    let recipe_dialog : HTMLDialogElement;
    let delete_warning : HTMLDialogElement;
</script>

<form method="POST" enctype="multipart/form-data" action={newMeal ? "/meals/new":"?/update"} onsubmit={handleSubmit}>
    <fieldset>
    <input type="hidden" id="slug" name="slug" value={data.slug} />
    <input type="hidden" id="id" name="id" value={data.id} />

    <label class="form-control" for="name">
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
    </label>

    </fieldset>

    <fieldset>
        <label class="form-control" for="ingredients">
            <span class="label label-text">Ingredients:</span>

            {#each data.ingredients as ingredient, i}
                <Ingredient {ingredient} {i} 
                onremove={function () {data.ingredients.splice(i, 1); data.ingredients = data.ingredients}}/> 
            {/each}
        </label>      
        <button class="btn" type="button" onclick={function () {data.ingredients = [...data.ingredients, {amount: '', name: ''}]}}>Add Ingredient</button>
    </fieldset>
    <fieldset>
        <label class="form-control" for="steps">
            <span class="label label-text">Steps:</span>
            <section>
            {#each data.steps as step, i}
                <SortableRow class="flex" data-id={i} 
                    handleClass="handle" 
                    this="div"
                    drag={(from, to) => updatePosition(data.steps, from, to)}>
                    <input type="hidden" name="step.{step.order}.order" bind:value={step.order} />
                    <input type="text" class="input input-bordered w-full" name="step.{step.order}.text" bind:value={step.text}>
                    <button class="btn btn-ghost" type="button" aria-label='drag' onclick={function () {data.steps.splice(i, 1); data.steps = data.steps}}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 36 36"><path fill="currentColor" d="M27.14 34H8.86A2.93 2.93 0 0 1 6 31V11.23h2V31a.93.93 0 0 0 .86 1h18.28a.93.93 0 0 0 .86-1V11.23h2V31a2.93 2.93 0 0 1-2.86 3" class="clr-i-outline clr-i-outline-path-1"/><path fill="currentColor" d="M30.78 9H5a1 1 0 0 1 0-2h25.78a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-2"/><path fill="currentColor" d="M21 13h2v15h-2z" class="clr-i-outline clr-i-outline-path-3"/><path fill="currentColor" d="M13 13h2v15h-2z" class="clr-i-outline clr-i-outline-path-4"/><path fill="currentColor" d="M23 5.86h-1.9V4h-6.2v1.86H13V4a2 2 0 0 1 1.9-2h6.2A2 2 0 0 1 23 4Z" class="clr-i-outline clr-i-outline-path-5"/><path fill="none" d="M0 0h36v36H0z"/></svg>
                    </button>              
                    <span class="handle"><svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 36 36"><path fill="currentColor" d="M32 29H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-1"/><path fill="currentColor" d="M32 19H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-2"/><path fill="currentColor" d="M32 9H4a1 1 0 0 1 0-2h28a1 1 0 0 1 0 2" class="clr-i-outline clr-i-outline-path-3"/><path fill="none" d="M0 0h36v36H0z"/></svg></span>
                </SortableRow>
            {/each}
            </section>
        </label>
        <button class="btn" type="button" onclick={function () {data.steps = [...data.steps, {id: data.steps.length, text: '', order: data.steps.length}]}}>Add Step</button>
    </fieldset>
    <fieldset>
        <label class="form-control" for="recipes">
            <div class="label"><span class="label-text">Recipes:</span></div>
            {#await recipes}
                <p>Loading...</p>
            {:then recipes }
                <div class="flex flex-wrap gap-4">
                {#each data.recipes as recipe}
                    {#if recipes.find((x) => x.id === recipe.recipe_id)}
                        <Card 
                        obj={recipes.find((x) => x.id === recipe.recipe_id) || { id: 0, name: '', description: '', slug: '', ingredients: [], steps: [] }} 
                        compact={true}/>
                    {/if}
                    <input type="hidden" name="recipe.{recipe.recipe_id}" value={recipe.recipe_id} />
                {/each}
                </div>
            {/await}
        </label>

        <button class="btn my-4" type="button" onclick={function () {recipe_dialog.showModal()}}>Add Recipe</button>
    </fieldset>
    
    <div class="container my-4">
        <button class="btn btn-primary" type="submit">Save</button>
        <button class="btn btn-secondary" type="button" onclick={oncancel}>Cancel</button>
        {#if !newMeal}
        <button class="btn btn-error" type="button" onclick={() => delete_warning.showModal()}>Delete</button>
        {/if}
    </div>
</form>

<DeleteConfirm bind:dialog={delete_warning} onselect={handleDelete}>
    <h2>Are you sure you want to delete this meal?</h2>
</DeleteConfirm>

<dialog bind:this={recipe_dialog} class="modal modal-bottom sm:modal-middle">
    <div class="modal-box">
        <h2 class="modal-title">Add Recipe</h2>
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
    {#await recipes}
        <p>Loading...</p>
    {:then recipes }
        <div class="flex flex-wrap gap-4">
        {#each recipes as recipe}
            <Card obj={recipe} onselect={handleAddRecipe}/>
        {/each}
        </div>
        <button class="btn my-4" type="button" onclick={function (event) {event.preventDefault(); recipe_dialog.close()}}>Close</button>
    {/await}
    </div>
</dialog>

