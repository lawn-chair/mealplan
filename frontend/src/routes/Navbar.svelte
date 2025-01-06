<script lang="ts">
    import UserButton from 'clerk-sveltekit/client/UserButton.svelte'
    import SignedIn from 'clerk-sveltekit/client/SignedIn.svelte'
    import SignedOut from 'clerk-sveltekit/client/SignedOut.svelte'
	import SignInButton from 'clerk-sveltekit/client/SignInButton.svelte';
	import SignUpButton from 'clerk-sveltekit/client/SignUpButton.svelte';
</script>

<div class="navbar bg-base-100">
    <div class="left-0">
        <div>
            <img width="225px" src="/yum-scheduler-website-header.svg" alt="logo" />
        </div>
    </div>
    <a class="btn btn-ghost text-xl" href="/meals/">Meals</a>
    <a class="btn btn-ghost text-xl" href="/recipes/">Recipes</a>

    <div class="absolute right-0 pr-2">
        <SignedIn>
            <UserButton afterSignOutUrl="/" />
        </SignedIn>
        <SignedOut>
            <SignInButton mode="modal" /> <span>|</span> <SignUpButton mode="modal" />
            <!-- You could also use <SignInButton mode="modal" /> and <SignUpButton mode="modal" /> here -->
        </SignedOut>
    </div>
</div>