import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getUpcomingPlans, Plan } from '../api';
import DisplayCard from './DisplayCard'; // Assuming DisplayCard can be adapted or a similar card is used

const PlanList: React.FC = () => {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPlans = async () => {
      try {
        setLoading(true);
        const response = await getUpcomingPlans();
        setPlans(response.data);
        setError(null);
      } catch (err: any) {
        setError(`Failed to fetch plans: ${err.response?.data?.error || err.message}`);
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchPlans();
  }, []);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading meal plans...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div role="alert" className="alert alert-error shadow-lg max-w-md mx-auto mt-10">
        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Error: {error}</span>
        <Link to="/plans" className="btn btn-sm btn-ghost" onClick={() => { /* Allow re-fetch or similar */ }}>Try Again</Link>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Meal Plans</h1>
        <Link to="/plans/new" className="btn btn-primary">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" />
          </svg>
          Create New Plan
        </Link>
      </div>

      {plans.length === 0 ? (
        <div className="text-center py-10 bg-base-200 rounded-lg shadow">
           <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <p className="mt-4 text-xl text-base-content/70">No meal plans found.</p>
          <p className="text-sm text-base-content/50">Get started by creating a new meal plan.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {plans.map((plan) => (
            <DisplayCard
              key={plan.id}
              id={plan.id}
              title={`Plan: ${new Date(plan.start_date).toLocaleDateString()} - ${new Date(plan.end_date).toLocaleDateString()}`}
              description={`Contains ${plan.meals?.length || 0} meal(s).`} // Adjust if plan.meals is not directly available or needs fetching
              viewLink={`/plans/${plan.id}`}
              // editLink={`/plans/${plan.id}/edit`} // Add if edit functionality is separate from view
              type="Item" // Using 'Item' as a generic type for now
              // imageUrl={undefined} // Plans might not have images, or you can assign a generic one
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default PlanList;
