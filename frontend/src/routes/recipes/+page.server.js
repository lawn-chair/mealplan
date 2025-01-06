/** @type {import('./$types').Actions} */
import { API } from '$lib/api.js';
import { fail } from '@sveltejs/kit';
import { parseFormValues } from '$lib/utils';

export const actions = {
	default: async (event) => {
		const submit = await event.request.formData();
        const data = await parseFormValues(submit);     
        //console.log(data);   
        console.log(event.locals.session);
        console.log(event);
        const response = await fetch(API + '/recipes', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + event.locals.session.getToken(),
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