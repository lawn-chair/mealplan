
import { API } from '$lib/api.js';

export async function load({ fetch }) {
    const response = await fetch(API + '/plans?future=true');
    return {planData: await response.json()}
}