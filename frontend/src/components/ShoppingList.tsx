import React, { useEffect, useState } from 'react';
import { getShoppingList, updateShoppingList, ShoppingList as IShoppingList, ShoppingListUpdatePayload } from '../api';
import { formatDate } from '../utils';

const ShoppingList: React.FC = () => {
  const [shoppingList, setShoppingList] = useState<IShoppingList | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [updatingItemId, setUpdatingItemId] = useState<number | null>(null);

  useEffect(() => {
    const fetchShoppingList = async () => {
      try {
        setIsLoading(true);
        const response = await getShoppingList();
        setShoppingList(response.data);
        setError(null);
      } catch (err) {
        console.error("Error fetching shopping list:", err);
        setError("Failed to load shopping list. Please try again later.");
        setShoppingList(null);
      } finally {
        setIsLoading(false);
      }
    };

    fetchShoppingList();
  }, []);

  const handleToggleChecked = async (index: number) => {
    if (!shoppingList || !shoppingList.plan || typeof shoppingList.plan.id === 'undefined') {
      setError("Cannot update item: Plan information is missing.");
      console.log(shoppingList);
      return;
    }

    setUpdatingItemId(index); 

    const updatedIngredients = shoppingList.ingredients.map((item, i) => 
      i === index ? { ...item, checked: !item.checked } : item
    );

    const updatedShoppingListState: IShoppingList = {
      ...shoppingList,
      ingredients: updatedIngredients,
    };

    setShoppingList(updatedShoppingListState);

    try {
      const payload: ShoppingListUpdatePayload = {
        plan: {id: shoppingList.plan.id}, // shoppingList.plan.id is now guaranteed to be a number
        ingredients: updatedIngredients,
      };
      await updateShoppingList(payload);
      setError(null); 
    } catch (err) {
      console.error("Error updating shopping list item:", err);
      setError("Failed to update item. Please try again.");
      // Revert to the previous state if update fails by refetching or using the original list
      // For simplicity, let's revert to the state before this specific toggle attempt.
      const revertedIngredients = shoppingList.ingredients.map((item, i) => 
        i === index ? { ...item, checked: !updatedIngredients[i].checked } : item
      );
      setShoppingList({
        ...shoppingList,
        ingredients: revertedIngredients,
      }); 
    } finally {
      setUpdatingItemId(null); 
    }
  };

  if (isLoading) {
    return (
      <div className="text-center py-10">
        <span className="loading loading-spinner loading-lg"></span>
        <p className="mt-2">Loading your shopping list...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-10 bg-error/10 text-error-content p-4 rounded-lg shadow mt-6">
        <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <p className="mt-4 text-xl">Oops! Something went wrong.</p>
        <p className="text-sm">{error}</p>
      </div>
    );
  }

  return (
    <div className="bg-base-100 p-4 sm:p-6 rounded-lg shadow-xl mt-6">
      <h2 className="text-2xl font-bold mb-2 text-center">Shopping List</h2>
      
      {shoppingList && shoppingList.plan && (
        <div className="text-center mb-4">
          <p className="text-sm text-base-content/70">
            For plan period: {formatDate(shoppingList.plan.start_date)} - {formatDate(shoppingList.plan.end_date)}
          </p>
        </div>
      )}

      {(!shoppingList || shoppingList.ingredients.length === 0) ? (
        <div className="text-center py-6">
          <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <p className="mt-4 text-xl text-base-content/70">Your shopping list is empty!</p>
          <p className="text-sm text-base-content/50">
            {shoppingList && shoppingList.plan 
              ? "It looks like you have everything you need for this plan period."
              : "It looks like you have everything you need, or no plan is active for shopping."}
          </p>
        </div>
      ) : (
        <>
          <ul className="space-y-2 mt-4">
            {shoppingList.ingredients.map((ingredient, index) => (
              <li key={index} className={`flex items-center justify-between p-3 rounded-lg shadow ${ingredient.checked ? 'bg-success/10 line-through text-base-content/60' : 'bg-base-200'}`}> 
                <div className="flex items-center">
                  <input 
                    type="checkbox" 
                    checked={ingredient.checked} 
                    onChange={() => handleToggleChecked(index)} 
                    disabled={updatingItemId === index} 
                    className={`checkbox checkbox-primary mr-3 ${updatingItemId === index ? 'opacity-50 cursor-not-allowed' : ''}`}
                  />
                  <span className={updatingItemId === index ? 'opacity-50' : ''}>
                    {ingredient.name} - {ingredient.amount}
                  </span>
                </div>
                {updatingItemId === index && <span className="loading loading-spinner loading-xs ml-2"></span>}
              </li>
            ))}
          </ul>
          <div className="mt-6 pt-4 border-t border-base-300 text-center">
            <p className="text-sm text-base-content/60">This shopping list is automatically generated based on your current meal plans.</p>
          </div>
        </>
      )}
    </div>
  );
};

export default ShoppingList;
