import ShoppingList from './ShoppingList'; // Import the new component

function Home() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4">Welcome to YumScheduler!</h1>
        <p className="text-lg text-base-content/80">Your friendly meal planning assistant.</p>
      </div>
      
      {/* Current Meal Plan Summary - Placeholder */}
      {/* You could fetch and display the current or next upcoming meal plan here */}
      {/* For example:
      <div className="mb-12 p-6 bg-base-200 rounded-lg shadow">
        <h2 className="text-2xl font-semibold mb-4">This Week's Plan</h2>
        <p className="text-base-content/70">Details about the current meal plan...</p>
        <Link to="/plans" className="btn btn-sm btn-outline btn-primary mt-4">View Plans</Link>
      </div>
      */}

      {/* Display the Shopping List */}
      <ShoppingList />

      {/* Quick Links - Placeholder */}
      {/* 
      <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-6">
        <Link to="/recipes" className="btn btn-lg btn-outline btn-secondary">Browse Recipes</Link>
        <Link to="/meals" className="btn btn-lg btn-outline btn-accent">Manage Meals</Link>
        <Link to="/pantry" className="btn btn-lg btn-outline">Check Pantry</Link>
      </div>
      */}
    </div>
  );
}
export default Home;
