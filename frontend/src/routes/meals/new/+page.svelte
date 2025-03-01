<script lang="ts">
    import MealForm from '../MealForm.svelte';
    import {goto} from '$app/navigation';
</script>

<main class="py-6 px-4 sm:p-6 md:py-10 md:px-8">
    <div class="container">

        <MealForm newMeal={true} oncancel={() => { goto('/meals') }} onsave={(result: any) => goto('/meals/' + result.slug)}/>
    
    </div>
</main>


<style>

</style>