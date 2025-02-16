/** @type {import('./$types').PageLoad} */
import { API } from '$lib/api.js';

export async function load( {fetch}) {
	const response = await fetch(API + '/meals');
    return {mealData: await response.json()};
}