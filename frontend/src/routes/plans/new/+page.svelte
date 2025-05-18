<script lang="ts">
    import { API } from '$lib/api.js';
    import { useClerkContext } from 'svelte-clerk';
    import { addDays, format, parse, nextDay, isPast } from "date-fns";
    import { goto } from '$app/navigation';
    import type { PlanData } from '$lib/types.js';
	import { onMount } from 'svelte';
    import { toaster } from '$lib/toaster-svelte';

    const ctx = useClerkContext();
    const userId = ctx.auth.userId || '';

    let data: PlanData = $state({
        id: 0,
        start_date: "",
        end_date: "",
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
            
            toaster.create({
                title: 'Success',
                description: 'Plan created successfully',
                type: 'success'
            });
            
            goto(`/plans/${result.id}`);
		} else {
            toaster.create({
                title: 'Error',
                description: 'Failed to create plan',
                type: 'error'
            });
        }
	}

    function setDefaultDate() {
        const nextMonday = nextDay(new Date, 1)
        data.start_date = format(nextMonday, "yyyy-MM-dd");
        data.end_date = format(addDays(nextMonday, 6), "yyyy-MM-dd");
    }

    async function getLastPlan() {
        if (!ctx.session) {
            console.error('Session is not available');
            return;
        }
        const token = await ctx.session.getToken();
        const response = await fetch(API + `/plans?last=true`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        if (response.ok) {
            const result = await JSON.parse(await response.text());
            const endDate = parse(result.end_date, "yyyy-MM-dd", new Date);
            if (!isPast(endDate)) {
                data.start_date = format(addDays(endDate, 1), "yyyy-MM-dd");
                data.end_date = format(addDays(endDate, 7), "yyyy-MM-dd");
            } else {
                setDefaultDate();
            }
        } else {
            setDefaultDate();
        }
    }

    onMount(() => {
        getLastPlan();
    });

    $inspect(data);
</script>

<svelte:head>
    <title>Yum! - Plans - New Plan</title>
</svelte:head>

<div class="container mx-auto">
    <h1 class="h3 p-4">Add New Meal Plan</h1>
    <form class="p-4" onsubmit={handleSubmit}>
        <input type="hidden" id="id" bind:value={data.id}>
        <label for="start_date">Start Date:</label>
        <input type="date" class="input" id="start_date" bind:value={data.start_date} required>
        <label for="end_date">End Date:</label>
        <input type="date" class="input" id="end_date" bind:value={data.end_date} required>
        <div class="py-4"></div>
        <button class="btn preset-filled-primary-500" type="submit">Create Plan</button>
    </form>
</div>