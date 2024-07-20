/** @type {import('./$types').PageLoad} */
import { API } from '$lib/api.js';

export async function load({ params }) {
	const response = await fetch(API + '/recipes?slug=' + params.slug);
    return response.json();
}