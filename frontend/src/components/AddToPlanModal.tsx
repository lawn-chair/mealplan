import React, { useEffect, useState } from 'react';
import { getUpcomingPlans, updatePlan, Plan } from '../api';

interface AddToPlanModalProps {
  mealId: number;
  open: boolean;
  onClose: () => void;
}

const AddToPlanModal: React.FC<AddToPlanModalProps> = ({ mealId, open, onClose }) => {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [selectedPlanId, setSelectedPlanId] = useState<number | null>(null);
  const [addToPlanError, setAddToPlanError] = useState<string | null>(null);
  const [addToPlanSuccess, setAddToPlanSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (open) {
      getUpcomingPlans().then(res => setPlans(res.data)).catch(() => setPlans([]));
      setSelectedPlanId(null);
      setAddToPlanError(null);
      setAddToPlanSuccess(null);
    }
  }, [open]);

  const handleAddToPlan = async () => {
    if (!selectedPlanId) return;
    setAddToPlanError(null);
    setAddToPlanSuccess(null);
    setLoading(true);
    try {
      const plan = plans.find(p => p.id === selectedPlanId);
      if (!plan) throw new Error('Plan not found');
      if (plan.meals.includes(mealId)) {
        setAddToPlanError('Meal already in plan.');
        setLoading(false);
        return;
      }
      const updatedMeals = [...plan.meals, mealId];
      await updatePlan(selectedPlanId, { meals: updatedMeals });
      setAddToPlanSuccess('Meal added to plan!');
      setTimeout(() => {
        setLoading(false);
        onClose();
      }, 1000);
    } catch (err: any) {
      setAddToPlanError(err.message || 'Failed to add meal to plan.');
      setLoading(false);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
      <div className="bg-base-100 p-6 rounded-lg shadow-lg w-full max-w-xs">
        <h3 className="text-lg font-bold mb-2">Add Meal to Plan</h3>
        <select
          className="select select-bordered w-full mb-4"
          value={selectedPlanId ?? ''}
          onChange={e => setSelectedPlanId(Number(e.target.value))}
        >
          <option value="" disabled>Select a plan</option>
          {plans.map(plan => (
            <option key={plan.id} value={plan.id}>
              {new Date(plan.start_date).toLocaleDateString()} - {new Date(plan.end_date).toLocaleDateString()}
            </option>
          ))}
        </select>
        {addToPlanError && <div className="alert alert-error py-1 mb-2">{addToPlanError}</div>}
        {addToPlanSuccess && <div className="alert alert-success py-1 mb-2">{addToPlanSuccess}</div>}
        <div className="flex gap-2 justify-end">
          <button className="btn btn-sm btn-ghost" onClick={onClose} disabled={loading}>Cancel</button>
          <button
            className="btn btn-sm btn-primary"
            disabled={!selectedPlanId || loading}
            onClick={handleAddToPlan}
          >
            {loading ? <span className="loading loading-spinner loading-xs"></span> : 'Add'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default AddToPlanModal;
