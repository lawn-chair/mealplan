/** @type {import('./$types').Actions} */
import { API } from '$lib/api';
import { fail, redirect } from '@sveltejs/kit';

import { parseFormValues } from '$lib/utils'

export const actions = {
	update: async (event) => {
		const submit = await event.request.formData();
        const data = await parseFormValues(submit);
        console.log(data);
        const response = await fetch(API + '/recipes/' + submit.get('id'), {
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
        const response = await fetch(API + '/recipes/' + submit.get('id'), {
            method: 'DELETE',
        });

        const res = await response;
        console.log(res);

        if(!res.ok) {
            const message = await res.text();
            console.log(message);
            return fail(res.status, { message: message});
        }
        redirect(303, '/recipes');
        return {success: true};
    }
};