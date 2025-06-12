import { useParams, Link } from 'react-router-dom';
import { getRecipeBySlug, Recipe } from '@/api';
import { useState, useEffect } from 'react';

function RecipeDetail() {
  const { slug: recipeSlug } = useParams<{ slug: string }>(); // Changed from recipeId to recipeSlug
  const [recipe, setRecipe] = useState<Recipe | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    if (recipeSlug || '') { // Changed from recipeId to recipeSlug
      fetchRecipeDetails((recipeSlug || '')); // Changed from recipeId to recipeSlug
    }
  }, [(recipeSlug || '')]); // Changed from recipeId to recipeSlug

  const fetchRecipeDetails = async (slug: string) => { // Changed parameter name to slug
    setIsLoading(true);
    setError(null);
    try {
      const response = await getRecipeBySlug(slug); // Use slug
      setRecipe(response.data);
    } catch (err) {
      setError('Failed to fetch recipe details. Please try again.');
      console.error(err);
    }
    setIsLoading(false);
  };

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading recipe details...</p>
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

  if (!recipe) {
    return (
      <div className="text-center py-10">
        <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
          <path strokeLinecap="round" strokeLinejoin="round" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p className="mt-4 text-xl text-gray-500">Recipe not found.</p>
        <Link to="/recipes" className="btn btn-sm btn-outline btn-primary mt-4">Back to Recipes</Link>
      </div>
    );
  }

  const imageUrl = typeof recipe.image === 'string' 
    ? recipe.image 
    : (recipe.image?.Valid ? recipe.image.String : '/recipe-blank.jpg');

  return (
    <div className="container mx-auto px-4 py-8">
      <article className="prose lg:prose-xl max-w-none bg-base-100 shadow-xl rounded-lg p-6 md:p-10">
        
        <div className="flex flex-col md:flex-row md:items-start gap-8 mb-8">
          {imageUrl && (
            <figure className="md:w-1/3 lg:w-1/4 flex-shrink-0">
              <img 
                src={imageUrl} 
                alt={recipe.name || 'Recipe image'} 
                className="rounded-lg shadow-md w-full h-auto object-cover aspect-square" 
                onError={(e) => { (e.target as HTMLImageElement).src = '/recipe-blank.jpg'; }}
              />
            </figure>
          )}
          <div className="flex-grow">
            <h1 className="text-4xl font-bold mb-2 !mt-0">{recipe.name}</h1>
            {recipe.description && <p className="text-lg text-base-content opacity-80 mb-6">{recipe.description}</p>}
            {recipe && recipe.id && ( 
              <div className="mt-auto"> {/* Changed from marginTop: '20px' */}
                <Link to={`/recipes/${recipeSlug || ''}/edit`} className="btn btn-primary btn-outline">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                  Edit Recipe
                </Link>
                {/* Delete button can be added here later */}
              </div>
            )}
          </div>
        </div>

        <div className="divider"></div>

        <div className="grid md:grid-cols-3 gap-8">
          <div className="md:col-span-1">
            <h2 className="text-2xl font-semibold mb-4">Ingredients</h2>
            {recipe.ingredients && recipe.ingredients.length > 0 ? (
              <ul className="list-none p-0 m-0">
                {recipe.ingredients.map((ingredient, index) => (
                  <li key={index} className="mb-2 p-3 bg-base-200 rounded-md shadow-sm">
                    <span className="font-medium">{ingredient.name}:</span> {ingredient.amount}
                    {ingredient.calories && <span className="text-xs text-base-content opacity-70"> ({ingredient.calories} kcal)</span>}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-base-content opacity-70">No ingredients listed.</p>
            )}
          </div>

          <div className="md:col-span-2">
            <h2 className="text-2xl font-semibold mb-4">Steps</h2>
            {recipe.steps && recipe.steps.length > 0 ? (
              <ol className="list-decimal list-inside space-y-3">
                {recipe.steps.sort((a, b) => a.order - b.order).map((step) => (
                  <li key={step.id} className="p-3 bg-base-200 rounded-md shadow-sm text-base-content leading-relaxed">
                    {step.text}
                  </li>
                ))}
              </ol>
            ) : (
              <p className="text-base-content opacity-70">No steps provided.</p>
            )}
          </div>
        </div>
      </article>
    </div>
  );
}

export default RecipeDetail;
