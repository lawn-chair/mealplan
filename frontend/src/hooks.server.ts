// hooks.server.ts
import { withClerkHandler } from 'svelte-clerk/server';

export const handle = withClerkHandler({debug: true, publishableKey: process.env.PUBLIC_CLERK_PUBLISHABLE_KEY, secretKey: process.env.CLERK_SECRET_KEY});
