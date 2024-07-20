/** @type {import('./$types').Actions} */
import { API } from '$lib/api.js';
import { parseMealFormValues } from '$lib/utils.js';
import { redirect, fail } from '@sveltejs/kit';

export const actions = {
	update: async (event) => {
        const submit = await event.request.formData();
        const data = await parseMealFormValues(submit);
        console.log(data);
        const response = await fetch(API + '/meals/' + submit.get('id'), {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data), 
        });

        const res = await response;
        console.log(res);

        if(!res.ok) {
            const message = await res.text();
            console.log(message);
            return fail(res.status, { message: message});
        }
        return {success: true};
	},
    delete: async (event) => {
        const submit = await event.request.formData();
        const response = await fetch(API + '/meals/' + submit.get('id'), {
            method: 'DELETE',
        });

        const res = await response;
        console.log(res);

        if(!res.ok) {
            const message = await res.text();
            console.log(message);
            return fail(res.status, { message: message});
        }
        
        redirect(303, '/meals');

        return {success: true};
    }
};