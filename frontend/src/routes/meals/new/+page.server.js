/** @type {import('./$types').Actions} */
import { API } from '$lib/api.js';
import { fail } from '@sveltejs/kit';

export const actions = {
	default: async (event) => {
		const submit = await event.request.formData();
        const name = submit.get('name');
        const description = submit.get('description');
        
        const response = await fetch(API + '/meals', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: name,
                description: description,
            }), 
        });

        const res = await response;
        console.log(res);

        if(!res.ok) {
            const message = await res.text();
            console.log(message);
            return fail(res.status, { message: message});
        }
        return {success: true};
	}
};