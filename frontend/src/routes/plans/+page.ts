
import { API } from '$lib/api.js';

export async function load({ fetch }) {
    const response = await fetch(API + '/plans');
    return {planData: await response.json()}
}