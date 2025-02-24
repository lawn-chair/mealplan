
import { API } from '$lib/api.js';

export async function load({ params, fetch }) {
    const response = await fetch(API + '/plans/' + params.id);
    return {planData: await response.json()}
}