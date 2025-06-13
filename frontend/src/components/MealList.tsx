import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getMeals, Meal } from '../api';
import DisplayCard from './DisplayCard';
import AddToPlanModal from './AddToPlanModal';

const MealList: React.FC = () => {
  const [meals, setMeals] = useState<Meal[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>("");
  const [showAddToPlan, setShowAddToPlan] = useState<{ mealId: number | null, open: boolean }>({ mealId: null, open: false });

  useEffect(() => {
    const fetchMeals = async () => {
      try {
        setLoading(true);
        const response = await getMeals();
        setMeals(Array.isArray(response.data) ? response.data : []);
        setError(null);
      } catch (err) {
        setError('Failed to fetch meals.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchMeals();
  }, []);

  // Filter meals by search (name, description, or tags)
  const filteredMeals = meals.filter((meal) => {
    const q = search.trim().toLowerCase();
    if (!q) return true;
    const nameMatch = meal.name?.toLowerCase().includes(q);
    const descMatch = meal.description?.toLowerCase().includes(q);
    const tagMatch = meal.tags?.some(tag => tag.toLowerCase().includes(q));
    return nameMatch || descMatch || tagMatch;
  });

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading meals...</p>
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
        <h2 className="text-4xl font-bold tracking-tight">Meals</h2>
        <div className="flex-1 flex justify-end">
          <input
            type="text"
            className="input input-bordered w-full max-w-xs"
            placeholder="Search meals..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            aria-label="Search meals"
          />
        </div>
        <Link to="/meals/new" className="btn btn-primary btn-md shadow-md ml-0 sm:ml-4">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
          </svg>
          Create New Meal
        </Link>
      </div>

      {filteredMeals.length === 0 ? (
        <div className="text-center py-10">
          <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p className="mt-4 text-xl text-gray-500">No meals found.</p>
          <p className="mt-2 text-sm text-gray-400">Ready to plan some meals?</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {filteredMeals.map((meal) => {
            const imageUrl = typeof meal.image === 'string' 
              ? meal.image 
              : (meal.image?.Valid ? meal.image.String : undefined);
            return (
              <DisplayCard
                key={meal.id}
                id={meal.id}
                imageUrl={imageUrl}
                title={meal.name || 'Untitled Meal'}
                description={meal.description}
                viewLink={`/meals/${meal.slug}`}
                imageAltText={meal.name || 'Meal image'}
                type="Meal"
                tags={meal.tags}
                onTagClick={(tag: string) => {
                  const tagLower = tag.toLowerCase();
                  const searchLower = search.toLowerCase();
                  if (!searchLower.split(/\s+/).includes(tagLower)) {
                    setSearch(search ? search + ' ' + tag : tag);
                  }
                }}
                onAddToPlan={() => setShowAddToPlan({ mealId: meal.id!, open: true })}
              />
            );
          })}
        </div>
      )}
      {/* AddToPlanModal for the selected meal */}
      {showAddToPlan.open && showAddToPlan.mealId !== null && (
        <AddToPlanModal
          mealId={showAddToPlan.mealId}
          open={showAddToPlan.open}
          onClose={() => setShowAddToPlan({ mealId: null, open: false })}
        />
      )}
    </div>
  );
};

export default MealList;
