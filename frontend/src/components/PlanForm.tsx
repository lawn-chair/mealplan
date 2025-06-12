import React, { useState, useEffect, FormEvent } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { createPlan, updatePlan, getPlanById, getMeals, Plan, Meal } from '../api';
import Select from 'react-select'; // Using react-select for multi-select

interface PlanFormProps {
  isEditMode?: boolean;
}

interface MealOption {
  value: number;
  label: string;
}

const PlanForm: React.FC<PlanFormProps> = ({ isEditMode }) => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [plan, setPlan] = useState<Omit<Plan, 'id' | 'user_id'>>({
    start_date: '',
    end_date: '',
    meals: [],
  });
  const [allMeals, setAllMeals] = useState<Meal[]>([]);
  const [selectedMeals, setSelectedMeals] = useState<MealOption[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);


  useEffect(() => {
      const fetchMeals = async () => {
          try {
              const response = await getMeals();
              setAllMeals(response.data || []);
          } catch (err: any) {
              setError(`Failed to fetch meals for selection: ${err.response?.data?.error || err.message}`);
              console.error(err);
          }
      };

      fetchMeals();
  }, []);

  useEffect(() => {

    if (isEditMode && id) {
      const fetchPlan = async () => {
        try {
          setLoading(true);
          const response = await getPlanById(Number(id));
          const fetchedPlan = response.data;
          setPlan({
            start_date: fetchedPlan.start_date.split('T')[0], // Format for date input
            end_date: fetchedPlan.end_date.split('T')[0],   // Format for date input
            meals: fetchedPlan.meals || [],
          });
          // Pre-populate selectedMeals for react-select
          if (fetchedPlan.meals && allMeals.length > 0) {
            const preselected = allMeals
              .filter(meal => fetchedPlan.meals.includes(meal.id!))
              .map(meal => ({ value: meal.id!, label: meal.name }));
            setSelectedMeals(preselected);
          }
          setError(null);
        } catch (err: any) {
          setError(`Failed to fetch plan details: ${err.response?.data?.error || err.message}`);
          console.error(err);
        } finally {
          setLoading(false);
        }
      };
      fetchPlan();
    }
  }, [id, isEditMode, allMeals]); // Added allMeals to dependency array for pre-selection logic

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setPlan((prevPlan) => ({
      ...prevPlan,
      [name]: value,
    }));
  };

  const handleMealSelectionChange = (selectedOptions: readonly MealOption[] | null) => {
    setSelectedMeals(selectedOptions ? [...selectedOptions] : []);
    setPlan(prevPlan => ({
      ...prevPlan,
      meals: selectedOptions ? selectedOptions.map(option => option.value) : [],
    }));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    // Ensure dates are in YYYY-MM-DD format and then append time for ISO 8601
    const planData = {
      ...plan,
      start_date: `${plan.start_date}`,
      end_date: `${plan.end_date}`,
    };

    try {
      if (isEditMode && id) {
        await updatePlan(Number(id), planData);
        navigate(`/plans/${id}`);
      } else {
        const response = await createPlan(planData);
        navigate(`/plans/${response.data.id}`);
      }
    } catch (err: any) {
      setError((isEditMode ? 'Failed to update plan: ' : 'Failed to create plan: ') + (err.response?.data?.error || err.message));
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const mealOptions: MealOption[] = allMeals.map(meal => ({ value: meal.id!, label: meal.name }));

  if (loading && isEditMode && !plan.start_date) { // Show loading only if plan data hasn't been populated
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading plan details...</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-2xl">
      <form onSubmit={handleSubmit} className="space-y-6 bg-base-100 p-6 md:p-8 rounded-lg shadow-xl">
        <h2 className="text-2xl font-bold mb-6 text-center">{isEditMode ? 'Edit Meal Plan' : 'Create New Meal Plan'}</h2>
        
        {error && (
          <div role="alert" className="alert alert-error">
            <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <span>{error}</span>
          </div>
        )}

        <div className="form-control">
          <label htmlFor="start_date" className="label">
            <span className="label-text">Start Date</span>
          </label>
          <input
            type="date"
            id="start_date"
            name="start_date"
            value={plan.start_date}
            onChange={handleChange}
            required
            className="input input-bordered w-full"
          />
        </div>

        <div className="form-control">
          <label htmlFor="end_date" className="label">
            <span className="label-text">End Date</span>
          </label>
          <input
            type="date"
            id="end_date"
            name="end_date"
            value={plan.end_date}
            onChange={handleChange}
            required
            className="input input-bordered w-full"
          />
        </div>

        <div className="form-control">
          <label htmlFor="meals" className="label">
            <span className="label-text">Select Meals for this Plan</span>
          </label>
          <Select
            id="meals"
            isMulti
            name="meals"
            options={mealOptions}
            value={selectedMeals}
            onChange={handleMealSelectionChange}
            className="basic-multi-select"
            classNamePrefix="select"
            isLoading={allMeals.length === 0 && !error} // Show loading indicator in select if meals are still fetching
            placeholder="Select meals..."
            noOptionsMessage={() => allMeals.length === 0 && !error ? 'Loading meals...' : 'No meals available'}
            styles={{
              control: (base) => ({ ...base, backgroundColor: 'var(--fallback-b1,oklch(var(--b1)/1))' , borderColor: 'var(--fallback-bc,oklch(var(--bc)/0.2))' }),
              menu: (base) => ({ ...base, backgroundColor: 'var(--fallback-b1,oklch(var(--b1)/1))' }),
              option: (base, { isFocused, isSelected }) => ({
                ...base,
                backgroundColor: isSelected ? 'var(--fallback-p,oklch(var(--p)/1))' : isFocused ? 'var(--fallback-b2,oklch(var(--b2)/1))' : 'var(--fallback-b1,oklch(var(--b1)/1))',
                color: isSelected ? 'var(--fallback-pc,oklch(var(--pc)/1))' : 'var(--fallback-bc,oklch(var(--bc)/1))' 
              }),
              multiValue: (base) => ({ ...base, backgroundColor: 'var(--fallback-s,oklch(var(--s)/1))' }),
              multiValueLabel: (base) => ({ ...base, color: 'var(--fallback-sc,oklch(var(--sc)/1))' }),
            }}
          />
        </div>

        <div className="flex justify-end gap-4 pt-4">
          <button type="button" onClick={() => navigate(isEditMode && id ? `/plans/${id}` : '/plans')} className="btn btn-ghost">
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={loading || !plan.start_date || !plan.end_date}>
            {loading ? <span className="loading loading-spinner loading-xs"></span> : (isEditMode ? 'Update Plan' : 'Create Plan')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default PlanForm;
