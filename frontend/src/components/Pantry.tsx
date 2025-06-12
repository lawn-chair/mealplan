import { useState, useEffect, FormEvent } from 'react';
import { getPantry, updatePantry, clearPantry } from '@/api';

function Pantry() {
  const [pantryItems, setPantryItems] = useState<string[]>([]);
  const [newItem, setNewItem] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPantry();
  }, []);

  const fetchPantry = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await getPantry();
      setPantryItems(response.data && response.data.items ? response.data.items : []);
    } catch (err: any) {
      setError(`Failed to fetch pantry items: ${err.response?.data?.error || err.message}`);
      console.error(err);
    }
    setIsLoading(false);
  };

  const handleAddItem = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!newItem.trim()) return;
    setIsLoading(true);
    try {
      const updatedPantryData: { items: string[] } = { items: [...pantryItems, newItem.trim()] };
      await updatePantry(updatedPantryData);
      setPantryItems(updatedPantryData.items);
      setNewItem('');
    } catch (err: any) {
      setError(`Failed to add item: ${err.response?.data?.error || err.message}`);
      console.error(err);
    }
    setIsLoading(false);
  };

  const handleRemoveItem = async (itemToRemove: string) => {
    setIsLoading(true);
    try {
      const updatedPantryItems = pantryItems.filter(item => item !== itemToRemove);
      await updatePantry({ items: updatedPantryItems });
      setPantryItems(updatedPantryItems);
    } catch (err: any) {
      setError(`Failed to remove item: ${err.response?.data?.error || err.message}`);
      console.error(err);
    }
    setIsLoading(false);
  };

  const handleClearPantry = async () => {
    if (!window.confirm("Are you sure you want to clear your entire pantry?")) return;
    setIsLoading(true);
    try {
      await clearPantry();
      setPantryItems([]);
    } catch (err: any) {
      setError(`Failed to clear pantry: ${err.response?.data?.error || err.message}`);
      console.error(err);
    }
    setIsLoading(false);
  };

  return (
    <div className="container mx-auto px-4 py-8 max-w-2xl">
      <h2 className="text-3xl font-bold mb-6 text-center">My Pantry</h2>

      {error && (
        <div role="alert" className="alert alert-error shadow-lg mb-4">
          <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleAddItem} className="mb-6 p-4 bg-base-200 rounded-lg shadow">
        <div className="form-control">
          <label htmlFor="newItem" className="label">
            <span className="label-text">Add New Pantry Item</span>
          </label>
          <div className="join w-full">
            <input
              type="text"
              id="newItem"
              value={newItem}
              onChange={(e) => setNewItem(e.target.value)}
              placeholder="E.g., Flour, Sugar, Olive Oil"
              className="input input-bordered join-item flex-grow"
              disabled={isLoading}
            />
            <button type="submit" className="btn btn-primary join-item" disabled={isLoading || !newItem.trim()}>
              {isLoading && pantryItems.length > 0 ? <span className="loading loading-spinner loading-xs"></span> : 'Add Item'}
            </button>
          </div>
        </div>
      </form>

      {isLoading && pantryItems.length === 0 && (
        <div className="flex flex-col items-center justify-center py-10">
          <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
          <p className="text-lg">Loading pantry items...</p>
        </div>
      )}

      {!isLoading && pantryItems.length === 0 && (
        <div className="text-center py-10 bg-base-200 rounded-lg shadow">
          <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <p className="mt-4 text-xl text-base-content/70">Your pantry is empty.</p>
          <p className="text-sm text-base-content/50">Add items using the form above.</p>
        </div>
      )}

      {pantryItems.length > 0 && (
        <div className="bg-base-100 p-4 sm:p-6 rounded-lg shadow-xl">
          <ul className="space-y-2">
            {pantryItems.map((item, index) => (
              <li key={index} className="flex justify-between items-center p-3 bg-base-200 rounded-md shadow-sm hover:bg-base-300 transition-colors">
                <span className="text-base-content">{item}</span>
                <button 
                  onClick={() => handleRemoveItem(item)} 
                  className="btn btn-error btn-sm btn-outline"
                  disabled={isLoading}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  Remove
                </button>
              </li>
            ))}
          </ul>
          {pantryItems.length > 0 && (
            <div className="mt-6 pt-4 border-t border-base-300 flex justify-end">
              <button 
                onClick={handleClearPantry} 
                className="btn btn-error btn-outline" 
                disabled={isLoading}
              >
                {isLoading ? <span className="loading loading-spinner loading-xs"></span> : 'Clear Entire Pantry'}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Pantry;
