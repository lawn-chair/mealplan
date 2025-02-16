/** @type {import('./$types').PageLoad} */
import { API } from '$lib/api.js';

export async function load({ params, fetch }) {
	const response = await fetch(API + '/meals?slug=' + params.slug);
    return response.json();
}