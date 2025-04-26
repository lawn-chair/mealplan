/** @type {import('./$types').PageLoad} */
import { API } from '$lib/api.js';

export async function load({fetch}) {
	const response = await fetch(API + '/recipes');
    return {recipeData: await response.json()};
}