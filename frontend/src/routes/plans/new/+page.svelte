<script lang="ts">
    import { API } from '$lib/api.js';
    import { useClerkContext } from 'svelte-clerk';
    import { goto, invalidateAll } from '$app/navigation';
	import { applyAction, deserialize } from '$app/forms';
    import type { PlanData } from '$lib/types.js';

    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';

    let data: PlanData = $state({
        id: 0,
        start_date: '',
        end_date: '',
        meals: [],
        user_id: userId,
    });

    async function handleSubmit(event : SubmitEvent) {
        event.preventDefault();

        if (!ctx.session) {
            console.error('Session is not available');
            return;
        }
        const token = await ctx.session.getToken();
        
        const response = await fetch(API + '/plans', {
			method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
			body: JSON.stringify(data),
		});

		if (response.ok) {
		    const result : PlanData = await JSON.parse(await response.text());
            console.log(result);
            goto(`/plans/${result.id}`);
		}
	}
</script>

<svelte:head>
    <title>Yum! - Plans - New Plan</title>
</svelte:head>

<div class="container mx-auto">
    <h1>New Plan</h1>
    <form onsubmit={handleSubmit}>
        <input type="hidden" id="id" bind:value={data.id}>
        <label for="start_date">Start Date:</label>
        <input type="date" class="input" id="start_date" bind:value={data.start_date} required>
        <label for="end_date">End Date:</label>
        <input type="date" class="input" id="end_date" bind:value={data.end_date} required>
        <button class="btn preset-filled-primary-500" type="submit">Create Plan</button>
    </form>
</div>