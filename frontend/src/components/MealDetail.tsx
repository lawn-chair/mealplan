import React, { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { getMealBySlug, deleteMeal, Meal } from '../api';

const MealDetail: React.FC = () => {
  const { slug: mealSlug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const [meal, setMeal] = useState<Meal | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMeal = async () => {
      if (!mealSlug) {
        setError('No meal slug provided.');
        setLoading(false);
        return;
      }
      try {
        setLoading(true);
        const response = await getMealBySlug(mealSlug);
        setMeal(response.data);
        setError(null);
      } catch (err) {
        setError('Failed to fetch meal.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchMeal();
  }, [mealSlug]);

  const handleDelete = async () => {
    if (meal && meal.id) {
      if (window.confirm(`Are you sure you want to delete meal \\"${meal.name}\\"?`)) {
        try {
          setLoading(true);
          await deleteMeal(meal.id);
          navigate('/meals');
        } catch (err) {
          setError('Failed to delete meal.');
          console.error(err);
        } finally {
          setLoading(false);
        }
      }
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading meal details...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div role="alert" className="alert alert-error shadow-lg max-w-md mx-auto mt-10">
        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Error: {error}</span>
        <button className="btn btn-sm btn-ghost" onClick={() => navigate('/meals')}>Back to Meals</button>
      </div>
    );
  }

  if (!meal) {
    return (
      <div className="text-center py-10">
        <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-base-content/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
          <path strokeLinecap="round" strokeLinejoin="round" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p className="mt-4 text-xl text-base-content/70">Meal not found.</p>
        <Link to="/meals" className="btn btn-sm btn-outline btn-primary mt-4">Back to Meals</Link>
      </div>
    );
  }

  const imageUrl = typeof meal.image === 'string'
    ? meal.image
    : (meal.image?.Valid ? meal.image.String : '/meal-blank.jpg');

  return (
    <div className="container mx-auto px-4 py-8">
      <article className="bg-base-100 shadow-xl rounded-lg p-6 md:p-10">
        <div className="flex flex-col md:flex-row md:items-start gap-8 mb-8">
          {imageUrl && (
            <figure className="md:w-1/3 lg:w-1/4 flex-shrink-0">
              <img
                src={imageUrl}
                alt={meal.name || 'Meal image'}
                className="rounded-lg shadow-md w-full h-auto object-cover aspect-square"
                onError={(e) => { (e.target as HTMLImageElement).src = '/meal-blank.jpg'; }}
              />
            </figure>
          )}
          <div className="flex-grow">
            <h1 className="text-3xl lg:text-4xl font-bold mb-2 !mt-0">{meal.name}</h1>
            {meal.description && <p className="text-base-content/80 mb-6 prose">{meal.description}</p>}
            <div className="flex flex-wrap gap-2 mt-auto">
              <Link to={`/meals/${meal.slug}/edit`} className="btn btn-primary btn-outline">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                Edit Meal
              </Link>
              <button onClick={handleDelete} className="btn btn-error btn-outline" disabled={loading}>
                {loading ? <span className="loading loading-spinner loading-xs"></span> :
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>}
                Delete Meal
              </button>
            </div>
          </div>
        </div>

        <div className="divider my-8"></div>

        {/* Ingredients and Steps Section - Side by side on larger screens */}
        <div className="grid md:grid-cols-3 gap-8">
          {/* Ingredients Section */}
          {meal.ingredients && meal.ingredients.length > 0 && (
            <section className="md:col-span-1 mb-8 md:mb-0">
              <h2 className="text-2xl font-semibold mb-4">Ingredients</h2>
              <ul className="list-none p-0 m-0 grid grid-cols-1 gap-3">
                {meal.ingredients.map((ingredient, index) => (
                  <li key={ingredient.id || index} className="p-3 bg-base-200 rounded-md shadow-sm">
                    <span className="font-medium">{ingredient.name}:</span> {ingredient.amount}
                  </li>
                ))}
              </ul>
            </section>
          )}

          {/* Steps Section */}
          {meal.steps && meal.steps.length > 0 && (
            <section className="md:col-span-2 mb-8 md:mb-0">
              <h2 className="text-2xl font-semibold mb-4">Steps</h2>
              <ol className="list-decimal list-inside space-y-3 prose max-w-none">
                {meal.steps.sort((a,b) => a.order - b.order).map((step) => (
                  <li key={step.id} className="p-3 bg-base-200 rounded-md shadow-sm text-base-content leading-relaxed not-prose">
                    {step.text}
                  </li>
                ))}
              </ol>
            </section>
          )}
        </div>

        {/* Fallback for when one of them is missing but not the other, to prevent layout shift */}
        {((meal.ingredients && meal.ingredients.length > 0 && (!meal.steps || meal.steps.length === 0)) ||
         (meal.steps && meal.steps.length > 0 && (!meal.ingredients || meal.ingredients.length === 0))) && (
          <div className="md:col-span-3"></div> // Empty div to maintain grid structure if one section is missing
        )}

        {/* Associated Recipes Section */}
        {meal.recipes && meal.recipes.length > 0 && (
          <section className="mb-8">
            <h2 className="text-2xl font-semibold mb-4">Included Recipes</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {meal.recipes.map((recipeItem) => (
                <div key={recipeItem.recipe_id} className="card card-compact bg-base-200 shadow">
                  <div className="card-body">
                    {/* Ideally, you'd fetch recipe details (name, image, slug) to make this a proper DisplayCard or similar */}
                    <h3 className="card-title text-sm">Recipe ID: {recipeItem.recipe_id}</h3>
                    {/* Example: Link to a recipe if you had its slug */}
                    {/* <Link to={`/recipes/${recipeItem.recipe_slug}`} className="btn btn-xs btn-outline btn-primary mt-2">View Recipe</Link> */}
                    <p className="text-xs text-base-content/60">Full recipe details would appear here.</p>
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}
        
        <div className="mt-8 pt-6 border-t border-base-300">
            <Link to="/meals" className="btn btn-ghost">&larr; Back to Meals List</Link>
        </div>
      </article>
    </div>
  );
};

export default MealDetail;
