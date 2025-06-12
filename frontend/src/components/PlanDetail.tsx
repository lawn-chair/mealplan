import React, { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { getPlanById, deletePlan, getMealById, Plan, Meal } from '../api'; // Assuming Meal type is available
import { formatDate, formatDateLong } from '@/utils'; // Assuming you have a utility function for date formatting

const PlanDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [plan, setPlan] = useState<Plan | null>(null);
  const [mealsDetails, setMealsDetails] = useState<Meal[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) {
      setError('No plan ID provided.');
      setLoading(false);
      return;
    }

    const fetchPlanAndMeals = async () => {
      try {
        setLoading(true);
        const planResponse = await getPlanById(Number(id));
        setPlan(planResponse.data);

        if (planResponse.data && planResponse.data.meals) {
          // Fetch details for each meal in the plan
          const mealPromises = planResponse.data.meals.map(mealId => getMealById(mealId));
          const mealsResponses = await Promise.all(mealPromises);
          setMealsDetails(mealsResponses.map(res => res.data));
        }
        setError(null);
      } catch (err: any) {
        setError(`Failed to fetch plan details: ${err.response?.data?.error || err.message}`);
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchPlanAndMeals();
  }, [id]);

  const handleDelete = async () => {
    if (plan && plan.id) {
      if (window.confirm(`Are you sure you want to delete this meal plan?`)) {
        try {
          setLoading(true);
          await deletePlan(plan.id);
          navigate('/plans');
        } catch (err: any) {
          setError(`Failed to delete plan: ${err.response?.data?.error || err.message}`);
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
        <p className="text-lg">Loading plan details...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div role="alert" className="alert alert-error shadow-lg max-w-md mx-auto mt-10">
        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Error: {error}</span>
        <button className="btn btn-sm btn-ghost" onClick={() => navigate('/plans')}>Back to Plans</button>
      </div>
    );
  }

  if (!plan) {
    return (
      <div className="text-center py-10">
        <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-base-content/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
          <path strokeLinecap="round" strokeLinejoin="round" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p className="mt-4 text-xl text-base-content/70">Meal plan not found.</p>
        <Link to="/plans" className="btn btn-sm btn-outline btn-primary mt-4">Back to Plans</Link>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <article className="bg-base-100 shadow-xl rounded-lg p-6 md:p-10">
        <div className="mb-8">
          <h1 className="text-3xl lg:text-4xl font-bold mb-2 !mt-0">
            Meal Plan: {formatDate(plan.start_date)} - {formatDate(plan.end_date)}
          </h1>
          <p className="text-base-content/80 mb-6 prose">
            This plan covers the week from {formatDateLong(plan.start_date)} to {formatDateLong(plan.end_date)}.
          </p>
          <div className="flex flex-wrap gap-2 mt-auto">
            <Link to={`/plans/${plan.id}/edit`} className="btn btn-primary btn-outline">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
              Edit Plan
            </Link>
            <button onClick={handleDelete} className="btn btn-error btn-outline" disabled={loading}>
              {loading ? <span className="loading loading-spinner loading-xs"></span> :
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>}
              Delete Plan
            </button>
          </div>
        </div>

        <div className="divider my-8"></div>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Scheduled Meals</h2>
          {mealsDetails.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {mealsDetails.map(meal => (
                <div key={meal.id} className="card bg-base-200 shadow-md">
                  <div className="card-body">
                    <h3 className="card-title">{meal.name}</h3>
                    <p className="text-sm opacity-70 truncate">{meal.description}</p>
                    <div className="card-actions justify-end mt-2">
                      <Link to={`/meals/${meal.slug}`} className="btn btn-sm btn-outline btn-secondary">
                        View Meal
                      </Link>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-base-content/70 italic">No meals have been added to this plan yet.</p>
          )}
        </section>
        
        <div className="mt-8 pt-6 border-t border-base-300">
            <Link to="/plans" className="btn btn-ghost">&larr; Back to Plans List</Link>
        </div>
      </article>
    </div>
  );
};

export default PlanDetail;
