/** @type {import('./$types').Actions} */
import { API } from '$lib/api.js';
import { fail } from '@sveltejs/kit';
import { parseMealFormValues } from '$lib/utils.js';

export const actions = {
	default: async (event) => {
		const submit = await event.request.formData();
        const data = await parseMealFormValues(submit);
        console.log(data);
        
        const response = await fetch(API + '/meals', {
            method: 'POST',
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
	}
};