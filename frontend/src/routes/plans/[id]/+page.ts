
import { API } from '$lib/api.js';

export async function load({ params, fetch }) {
    const response = await fetch(API + '/plans/' + params.id);
    const meals = await fetch(API + '/meals');
    return {planData: await response.json(), meals: await meals.json()}
}