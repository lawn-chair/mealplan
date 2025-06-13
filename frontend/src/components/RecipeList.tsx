import { useState, useEffect } from 'react';
import { getRecipes, Recipe } from '@/api';
import { Link } from 'react-router-dom';
import DisplayCard from './DisplayCard'; // Import the new DisplayCard component

function RecipeList() {
  const [recipes, setRecipes] = useState<Recipe[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>("");

  useEffect(() => {
    fetchRecipes();
  }, []);

  const fetchRecipes = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await getRecipes();
      setRecipes(Array.isArray(response.data) ? response.data : []);
    } catch (err) {
      setError('Failed to fetch recipes. Please try again.');
      console.error(err);
    }
    setIsLoading(false);
  };

  // Filter recipes by search (name, description, or tags)
  const filteredRecipes = recipes.filter((recipe) => {
    const q = search.trim().toLowerCase();
    if (!q) return true;
    const nameMatch = recipe.name?.toLowerCase().includes(q);
    const descMatch = recipe.description?.toLowerCase().includes(q);
    const tagMatch = recipe.tags?.some(tag => tag.toLowerCase().includes(q));
    return nameMatch || descMatch || tagMatch;
  });

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading recipes...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div role="alert" className="alert alert-error shadow-lg max-w-md mx-auto mt-10">
        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Error: {error}</span>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-8">
        <h2 className="text-4xl font-bold tracking-tight">Recipes</h2>
        <div className="flex-1 flex justify-end">
          <input
            type="text"
            className="input input-bordered w-full max-w-xs"
            placeholder="Search recipes..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            aria-label="Search recipes"
          />
        </div>
        <Link to="/recipes/new" className="btn btn-primary btn-md shadow-md ml-0 sm:ml-4">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
          </svg>
          Add New Recipe
        </Link>
      </div>

      {filteredRecipes.length === 0 && (
        <div className="text-center py-10">
          <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 12c0 6-4.03 10-9 10s-9-4-9-10c0-6 4.03-10 9-10s9 4 9 10zM14 14c0 1.104-.896 2-2 2s-2-.896-2-2 0-4 2-4 2 .896 2 2z" />
          </svg>
          <p className="mt-4 text-xl text-gray-500">No recipes found.</p>
          <p className="mt-2 text-sm text-gray-400">Why not add the first one?</p>
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {filteredRecipes.map((recipe) => {
          const imageUrl = (recipe.image?.Valid ? recipe.image.String : undefined);
          return (
            <DisplayCard
              key={recipe.id}
              id={recipe.id}
              imageUrl={imageUrl}
              title={recipe.name || 'Untitled Recipe'}
              description={recipe.description}
              viewLink={`/recipes/${recipe.slug}`}
              editLink={`/recipes/${recipe.slug}/edit`}
              imageAltText={recipe.name || 'Recipe image'}
              type="Recipe"
              tags={recipe.tags}
              onTagClick={(tag: string) => {
                // If tag is not already in search, add it (append with space if search is not empty)
                const tagLower = tag.toLowerCase();
                const searchLower = search.toLowerCase();
                if (!searchLower.split(/\s+/).includes(tagLower)) {
                  setSearch(search ? search + ' ' + tag : tag);
                }
              }}
            />
          );
        })}
      </div>
    </div>
  );
}

export default RecipeList;
