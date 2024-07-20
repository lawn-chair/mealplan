<script>
    /** @type {import('./$types').PageData} */
    import MealForm from '../MealForm.svelte';
    import Navbar from '../../Navbar.svelte'
    import {goto} from '$app/navigation';

</script>
<Navbar />

<main class="py-6 px-4 sm:p-6 md:py-10 md:px-8">
    <div class="container">

        <MealForm newMeal={true} on:cancel={() => { goto('/meals') }}/>
    
    </div>
</main>


<style>

</style>