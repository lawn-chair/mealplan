<script lang="ts">
    import { Pencil } from 'lucide-svelte'

    interface Props {
        image?: any;
        fallback?: string;
        name?: string;
        description?: string;
        editing?: boolean;
        userId?: string;
        onclick?: () => void;
    }

    let {
        image = {String: '', Valid: false},
        fallback = '/recipe-blank.jpg',
        name = '',
        description = '',
        editing = false,
        userId = '',
        onclick = () => {}
    }: Props = $props();

    console.log(userId);
</script>

<div class="w-full items-center  bg-base-200">
    <div class="hero-content justify-center flex flex-col lg:flex-row">
        <img src={image && image.Valid ? image.String : fallback} alt={name} class="max-w-sm rounded-lg shadow-2xl" />
        <div>
            <h1 class="text-xl font-bold lg:text-4xl">{name}
            {#if !editing && userId !== ''} 
                <button class="btn btn-ghost" aria-label="edit" onclick={onclick}>
                    <Pencil />
                </button>
            {/if}
            </h1>
            {#if !editing}
                <p class="py-6">{description}</p>
            {/if}
        </div>
    </div>
</div>
<style>

    /* borrowed from Daisy UI */
    .hero-content {
        justify-content: left;
        align-items: center;
        gap: 1rem;
        max-width: 80rem;
        padding-top: 1rem;
        padding-bottom: 1rem;
        display: flex;
    }
</style>