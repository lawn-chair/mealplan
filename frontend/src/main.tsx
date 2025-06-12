import React from 'react';
import ReactDOM from 'react-dom/client';
import App from '@/App'; // Using path alias
import '@/index.css'; // Using path alias
import { ClerkProvider } from '@clerk/clerk-react'; // Uncommented

// Import your publishable key from .env.local
const PUBLISHABLE_KEY = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY; // Uncommented

if (!PUBLISHABLE_KEY) { // Uncommented
  throw new Error("Missing Publishable Key. Check .env.local and VITE_CLERK_PUBLISHABLE_KEY"); // Added more detail to error
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ClerkProvider publishableKey={PUBLISHABLE_KEY}> {/* Using the variable */} 
      <App />
    </ClerkProvider>
  </React.StrictMode>,
);
