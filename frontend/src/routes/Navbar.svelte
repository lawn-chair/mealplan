<script lang="ts">
    import { SignedIn, SignedOut, UserButton, SignInButton } from 'svelte-clerk';
</script>
<header class="preset-filled-tonal-50">
<nav class="container flex flex-row items-center">
    <div>
        <img width="225px" src="/yum-scheduler-website-header.svg" alt="logo" />
    </div>
    <div>
        <a class="btn btn-ghost text-xl" href="/">Home</a>
    </div>
    <a class="btn btn-ghost text-xl float-left" href="/plans/">Plans</a>
    <a class="btn btn-ghost text-xl float-left" href="/meals/">Meals</a>
    <a class="btn btn-ghost text-xl float-left" href="/recipes/">Recipes</a>
    <SignedIn>
        <a class="btn btn-ghost text-xl float-left" href="/pantry/">Pantry</a>
    </SignedIn>

    <div class="absolute right-0 pr-2">
        <SignedIn>
            <UserButton afterSignOutUrl="/" />
        </SignedIn>
        <SignedOut>
            <SignInButton withSignUp mode="modal"/>
            <!-- You could also use <SignInButton mode="modal" /> and <SignUpButton mode="modal" /> here -->
        </SignedOut>
    </div>
</nav>
</header>