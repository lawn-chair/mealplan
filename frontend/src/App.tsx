import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom';
import { SignedIn, SignedOut, UserButton, useAuth, SignIn, SignUp } from "@clerk/clerk-react";
import RecipeList from './components/RecipeList';
import RecipeDetail from './components/RecipeDetail';
import RecipeForm from './components/RecipeForm';
import MealList from './components/MealList';
import MealDetail from './components/MealDetail';
import MealForm from './components/MealForm';
import Pantry from './components/Pantry';
import PlanList from './components/PlanList';
import PlanDetail from './components/PlanDetail';
import PlanForm from './components/PlanForm';
import Home from './components/Home';
import NotFound from './components/NotFound';
import PageTitleUpdater from './components/PageTitleUpdater'; // Import the new component
import { setupAuthTokenInterceptor } from './api';
import HouseholdManager from './components/HouseholdManager';
import './App.css';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { userId } = useAuth();

  if (!userId) {
    return <Navigate to="/sign-in" replace />;
  }

  return <>{children}</>;
};

function App() {
  const { getToken } = useAuth();

  useEffect(() => {
    if (getToken) {
      setupAuthTokenInterceptor(getToken);
    }
  }, [getToken]);

  return (
    <BrowserRouter>
      <header className="navbar bg-base-100 shadow-md">
        <div className="navbar-start">
          <div className="flex-1">
            <Link to="/" className="btn btn-ghost normal-case text-xl hover:bg-transparent focus:bg-transparent active:bg-transparent hover:text-inherit focus:text-inherit active:text-inherit">
              <img src="/yum-scheduler-favicon.svg" alt="Yum!" className="h-12 w-auto" />
              <span className="ml-2 text-primary">Yum!</span>
            </Link>
          </div>
          <div className="flex-none">
            <ul className="menu menu-horizontal px-1">
              <li><Link to="/" className="link-primary">Home</Link></li>
              <li><Link to="/recipes" className="link-primary">Recipes</Link></li>
              <li><Link to="/meals" className="link-primary">Meals</Link></li>
              <SignedIn>
                <li><Link to="/plans" className="link-primary">Plans</Link></li>
                <li><Link to="/pantry" className="link-primary">Pantry</Link></li>
                <li><Link to="/household" className="link-primary">Household</Link></li>
              </SignedIn>
            </ul>
          </div>
          
        </div>
        <div className="navbar-end">
          <SignedIn>
            <UserButton afterSignOutUrl="/sign-in" />
          </SignedIn>
          <SignedOut>
            <Link to="/sign-in" className="btn btn-ghost">Sign In</Link>
            <Link to="/sign-up" className="btn btn-primary">Sign Up</Link>
          </SignedOut>
        </div>
      </header>
      <main className="container mx-auto px-4 py-8">
        <Routes>
          <Route path="/sign-in/*" element={<><PageTitleUpdater section="Sign In" /><SignIn routing="path" path="/sign-in" /></>} />
          <Route path="/sign-up/*" element={<><PageTitleUpdater section="Sign Up" /><SignUp routing="path" path="/sign-up" /></>} />

          <Route
            path="/"
            element={<><PageTitleUpdater section="Home" /><Home /></>}
          />
          <Route
            path="/recipes"
            element={<><PageTitleUpdater section="Recipes" /><RecipeList /></>}
          />
          <Route
            path="/recipes/new"
            element={<ProtectedRoute><PageTitleUpdater section="New Recipe" /><RecipeForm /></ProtectedRoute>}
          />
          <Route
            path="/recipes/:slug/edit"
            element={<ProtectedRoute><PageTitleUpdater section="Edit Recipe" /><RecipeForm isEditMode={true} /></ProtectedRoute>}
          />
          <Route
            path="/recipes/:slug"
            element={<><PageTitleUpdater section="Recipe Details" /><RecipeDetail /></>}
          />
          <Route
            path="/meals"
            element={<><PageTitleUpdater section="Meals" /><MealList /></>}
          />
          <Route
            path="/meals/new"
            element={<ProtectedRoute><PageTitleUpdater section="New Meal" /><MealForm /></ProtectedRoute>}
          />
          <Route
            path="/meals/:slug"
            element={<><PageTitleUpdater section="Meal Details" /><MealDetail /></>}
          />
          <Route
            path="/meals/:slug/edit"
            element={<ProtectedRoute><PageTitleUpdater section="Edit Meal" /><MealForm isEditMode={true} /></ProtectedRoute>}
          />
          <Route
            path="/pantry"
            element={<ProtectedRoute><PageTitleUpdater section="Pantry" /><Pantry /></ProtectedRoute>}
          />
          <Route
            path="/plans"
            element={<ProtectedRoute><PageTitleUpdater section="Plans" /><PlanList /></ProtectedRoute>}
          />
          <Route
            path="/plans/new"
            element={<ProtectedRoute><PageTitleUpdater section="New Plan" /><PlanForm /></ProtectedRoute>}
          />
          <Route
            path="/plans/:id"
            element={<ProtectedRoute><PageTitleUpdater section="Plan Details" /><PlanDetail /></ProtectedRoute>}
          />
          <Route
            path="/plans/:id/edit"
            element={<ProtectedRoute><PageTitleUpdater section="Edit Plan" /><PlanForm isEditMode={true} /></ProtectedRoute>}
          />
          <Route
            path="/household"
            element={<ProtectedRoute><PageTitleUpdater section="Household" /><HouseholdManager /></ProtectedRoute>}
          />
          <Route path="*" element={<><PageTitleUpdater section="Not Found" /><NotFound /></>} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}

export default App;
