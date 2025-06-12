import { useState, ChangeEvent, FormEvent, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { createRecipe, getRecipeBySlug, updateRecipe, Recipe, RecipeIngredient, RecipeStep, uploadImage } from '@/api';
import { AxiosResponse } from 'axios'; 
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { SortableStepItem, CommonStep } from './SortableStepItem'; // Import generic component

// Define a type for the ingredient part of the form
interface RecipeIngredientForm {
  name: string;
  amount: string;
  calories?: string; // Keep as string for form input, convert to number on submit
}

// Helper type for steps in the form, ensuring ID is always defined for SortableStepItem
// and includes 'order' for internal management, though SortableStepItem doesn't need it directly.
type RecipeStepWithDefinedId = Omit<RecipeStep, 'id'> & { id: string | number } & CommonStep;


interface RecipeFormProps {
  isEditMode?: boolean; // Changed from mode to isEditMode
}

function RecipeForm({ isEditMode = false }: RecipeFormProps) { // Default to false
  const navigate = useNavigate();
  const { slug: recipeSlug } = useParams<{ slug: string }>(); 
  const [recipeName, setRecipeName] = useState<string>('');
  const [recipeDescription, setRecipeDescription] = useState<string>('');
  const [ingredients, setIngredients] = useState<RecipeIngredientForm[]>([{ name: '', amount: '', calories: '' }]);
  // Use RecipeStepWithDefinedId for the steps state
  const [steps, setSteps] = useState<RecipeStepWithDefinedId[]>([{ id: `new-step-${Date.now()}`, text: '', order: 1 }]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [currentRecipeId, setCurrentRecipeId] = useState<number | null>(null); 
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null); 
  const [currentImageUrl, setCurrentImageUrl] = useState<string | null>(null); 

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  useEffect(() => {
    if (isEditMode && recipeSlug) { 
      setIsLoading(true);
      getRecipeBySlug(recipeSlug)
        .then((response: AxiosResponse<Recipe>) => {
          const recipe = response.data;
          setRecipeName(recipe.name);
          setRecipeDescription(recipe.description);
          setIngredients(recipe.ingredients.map((ing: RecipeIngredient) => ({
            name: ing.name,
            amount: ing.amount,
            calories: ing.calories?.toString() || ''
          })));
          
          const fetchedSteps = (recipe.steps || []).map((step, index) => ({
            ...step,
            id: step.id || `loaded-step-${index}-${Date.now()}`, // Ensure ID is string or number
            text: step.text,
            order: step.order,
          })).sort((a,b) => a.order - b.order); // Sort by order
          setSteps(fetchedSteps as RecipeStepWithDefinedId[]);

          if (recipe.id) {
            setCurrentRecipeId(recipe.id);
          }
          if (recipe.image && recipe.image.Valid && recipe.image.String) {
            setCurrentImageUrl(recipe.image.String); 
          }
        })
        .catch((err: any) => {
          console.error("Failed to fetch recipe for editing:", err);
          setError("Failed to load recipe data. Please try again.");
        })
        .finally(() => setIsLoading(false));
    } else {
      // Reset form for create mode
      setRecipeName('');
      setRecipeDescription('');
      setIngredients([{ name: '', amount: '', calories: '' }]);
      setSteps([{ id: `new-step-${Date.now()}`, text: '', order: 1 }]);
      setSelectedImageFile(null);
      setCurrentImageUrl(null);
      setCurrentRecipeId(null);
      setError(null);
    }
  }, [isEditMode, recipeSlug]); // Depend on isEditMode and recipeSlug

  const handleIngredientChange = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    const newIngredients = ingredients.map((ingredient, i) => {
      if (index === i) {
        return { ...ingredient, [event.target.name]: event.target.value };
      }
      return ingredient;
    });
    setIngredients(newIngredients);
  };

  const handleAddIngredient = () => {
    setIngredients([...ingredients, { name: '', amount: '', calories: '' }]);
  };

  const handleRemoveIngredient = (index: number) => {
    const newIngredients = ingredients.filter((_, i) => i !== index);
    setIngredients(newIngredients);
  };

  // Updated to match SortableStepItem's onTextChange prop
  const handleStepTextChange = (itemId: string | number, newText: string) => {
    setSteps(prevSteps => 
      prevSteps.map(step => 
        step.id === itemId ? { ...step, text: newText } : step
      )
    );
  };

  const handleAddStep = () => {
    const newStepId = `new-step-${Date.now()}`;
    setSteps(prevSteps => [
      ...prevSteps, 
      { id: newStepId, text: '', order: prevSteps.length + 1 }
    ]);
  };

  // Updated to match SortableStepItem's onRemove prop
  const handleRemoveStepById = (idToRemove: string | number) => {
    setSteps(prevSteps => {
      const newSteps = prevSteps.filter(step => step.id !== idToRemove);
      // Re-order remaining steps
      return newSteps.map((step, index) => ({ ...step, order: index + 1 }));
    });
  };

  const handleImageChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setSelectedImageFile(event.target.files[0]);
      setCurrentImageUrl(null); 
      // Preview for new image
      const reader = new FileReader();
      reader.onloadend = () => {
        // setPreviewImage(reader.result as string); // If you add a preview state
      };
      reader.readAsDataURL(event.target.files[0]);
    } else {
      setSelectedImageFile(null);
      // setPreviewImage(null); // Clear preview if file is deselected
    }
  };
  
  const [imagePreview, setImagePreview] = useState<string | null>(null);

  useEffect(() => {
    if (selectedImageFile) {
      const objectUrl = URL.createObjectURL(selectedImageFile);
      setImagePreview(objectUrl);
      return () => URL.revokeObjectURL(objectUrl); // Clean up
    } else if (currentImageUrl) {
      setImagePreview(currentImageUrl);
    } else {
      setImagePreview(null);
    }
  }, [selectedImageFile, currentImageUrl]);


  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (over && active.id !== over.id) {
      setSteps((items) => {
        const oldIndex = items.findIndex(item => item.id === active.id);
        const newIndex = items.findIndex(item => item.id === over.id);
        const reorderedItems = arrayMove(items, oldIndex, newIndex);
        // Update order property after reordering
        return reorderedItems.map((item, index) => ({ ...item, order: index + 1 }));
      });
    }
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsLoading(true);
    setError(null);

    let finalImageUrl: string | undefined = currentImageUrl || undefined;

    if (selectedImageFile) {
      try {
        const formData = new FormData();
        formData.append('file', selectedImageFile);
        const uploadResponse = await uploadImage(formData);
        if (uploadResponse.data.url) {
          finalImageUrl = uploadResponse.data.url;
        } else {
          throw new Error('Image URL not found in upload response.');
        }
      } catch (err) {
        console.error("Failed to upload image:", err);
        setError('Failed to upload image. Please try again.');
        setIsLoading(false);
        return;
      } 
    } else if (!imagePreview && currentRecipeId && recipeSlug) { // If preview is null (image removed) and it's edit mode
      finalImageUrl = undefined; // Explicitly set to undefined to remove image
    }


    const finalIngredients = ingredients.map(ing => ({
      name: ing.name,
      amount: ing.amount,
      calories: ing.calories && ing.calories.trim() !== '' ? parseInt(ing.calories, 10) : null,
    }));

    const finalStepsForApi = steps.map((step, index) => ({
      // For new steps, ID should be undefined. For existing, it should be the number ID.
      id: typeof step.id === 'string' && (step.id.startsWith('new-step-') || step.id.startsWith('loaded-step-')) ? undefined : Number(step.id),
      text: step.text,
      order: index + 1, 
    }));
    
    // Filter out steps with potentially problematic IDs before sending to API
    const cleanSteps = finalStepsForApi.map(s => ({
        text: s.text,
        order: s.order,
        // Ensure id is number | undefined, matching RecipeStep in api.ts
        id: (typeof s.id === 'number' && !isNaN(s.id)) ? s.id : undefined 
    }));


    const recipePayload = {
      name: recipeName,
      description: recipeDescription,
      ingredients: finalIngredients,
      steps: cleanSteps,
      image: finalImageUrl ? { Valid: true, String: finalImageUrl } : { Valid: false, String: '' },
    };

    try {
      let response: AxiosResponse<Recipe>;
      if (isEditMode && currentRecipeId) {
        response = await updateRecipe(currentRecipeId, recipePayload);
      } else {
        response = await createRecipe(recipePayload);
      }
      navigate(`/recipes/${response.data.slug || response.data.id}`);
    } catch (err: any) {
      setError((isEditMode ? 'Failed to update recipe: ' : 'Failed to create recipe: ') + (err.response?.data?.error || err.message));
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading && isEditMode) { 
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading recipe for editing...</p>
      </div>
    );
  }
  
  // Ensure steps always have a defined ID for SortableContext items prop
  const stepsForRender = steps.map((step, index) => ({
    ...step,
    id: step.id || `render-step-${index}-${Date.now()}`, // Fallback, though state should manage this
  })) as RecipeStepWithDefinedId[];


  return (
    <div className="container mx-auto px-4 py-8 max-w-3xl">
      <h2 className="text-3xl font-bold mb-8 text-center">{isEditMode ? 'Edit Recipe' : 'Create New Recipe'}</h2>
      {error && (
        <div role="alert" className="alert alert-error shadow-lg mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
          <span>Error: {error}</span>
        </div>
      )}
      <form onSubmit={handleSubmit} className="space-y-6 bg-base-100 p-6 sm:p-8 rounded-lg shadow-xl">
        
        <div className="form-control">
          <label htmlFor="recipeName" className="label">
            <span className="label-text text-lg">Recipe Name</span>
          </label>
          <input
            type="text"
            id="recipeName"
            value={recipeName}
            onChange={(e) => setRecipeName(e.target.value)}
            required
            className="input input-bordered input-primary w-full"
            placeholder="e.g., Chocolate Chip Cookies"
          />
        </div>

        <div className="form-control">
          <label htmlFor="recipeDescription" className="label">
            <span className="label-text text-lg">Description</span>
          </label>
          <textarea
            id="recipeDescription"
            value={recipeDescription}
            onChange={(e) => setRecipeDescription(e.target.value)}
            required
            className="textarea textarea-bordered textarea-primary w-full h-32"
            placeholder="A short summary of your recipe..."
          />
        </div>

        {/* Image Upload Section */}
        <div className="form-control">
          <label htmlFor="recipeImage" className="label">
            <span className="label-text text-lg">Recipe Image</span>
          </label>
          <input
            type="file"
            id="recipeImage"
            accept="image/*"
            onChange={handleImageChange}
            className="file-input file-input-bordered file-input-primary w-full"
          />
          {imagePreview && (
            <div className="mt-4 p-4 border border-base-300 rounded-lg bg-base-200">
              <p className="font-semibold mb-2">Image Preview:</p>
              <img src={imagePreview} alt={recipeName || 'Recipe image preview'} className="max-w-xs rounded-md shadow" />
            </div>
          )}
           {!imagePreview && currentRecipeId && recipeSlug && ( // Show button to remove image only in edit mode if there was an image
            <button 
              type="button" 
              onClick={() => {
                setCurrentImageUrl(null); 
                setSelectedImageFile(null);
                //setImagePreview(null); // This will be handled by the useEffect for imagePreview
              }} 
              className="btn btn-outline btn-error btn-sm mt-2"
            >
              Remove Image
            </button>
          )}
        </div>

        {/* Ingredients Section */}
        <div className="space-y-4">
          <h3 className="text-xl font-semibold border-b border-base-300 pb-2 mb-4">Ingredients</h3>
          {ingredients.map((ingredient, index) => (
            <div key={index} className="p-4 border border-base-300 rounded-lg space-y-3 bg-base-200 shadow-sm">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                <input
                  type="text"
                  name="name"
                  placeholder="Ingredient Name (e.g., Flour)"
                  value={ingredient.name}
                  onChange={(e) => handleIngredientChange(index, e)}
                  required
                  className="input input-bordered input-primary w-full col-span-1 md:col-span-2"
                />
                <input
                  type="text"
                  name="amount"
                  placeholder="Amount (e.g., 2 cups)"
                  value={ingredient.amount}
                  onChange={(e) => handleIngredientChange(index, e)}
                  required
                  className="input input-bordered input-primary w-full"
                />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3 items-center">
                 <input
                    type="text"
                    name="calories"
                    placeholder="Calories (optional)"
                    value={ingredient.calories}
                    onChange={(e) => handleIngredientChange(index, e)}
                    className="input input-bordered input-sm w-full" // Changed to input-sm
                  />
                <div className="md:col-start-3 flex justify-end">
                  {ingredients.length > 0 && ( // Show remove only if there's at least one, though typically you'd want > 1
                    <button 
                      type="button" 
                      onClick={() => handleRemoveIngredient(index)} 
                      className="btn btn-sm btn-outline btn-error"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                      Remove
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
          <button 
            type="button" 
            onClick={handleAddIngredient} 
            className="btn btn-outline btn-accent w-full md:w-auto"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
            Add Ingredient
          </button>
        </div>

        {/* Steps Section - Now using generic SortableStepItem */}
        <div className="space-y-4">
          <h3 className="text-xl font-semibold border-b border-base-300 pb-2 mb-4">Steps</h3>
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <SortableContext
              items={stepsForRender.map(step => step.id)} // Use IDs from stepsForRender
              strategy={verticalListSortingStrategy}
            >
              {stepsForRender.map((step, index) => (
                <SortableStepItem<RecipeStepWithDefinedId>
                  key={step.id}
                  item={step}
                  index={index} // Pass the current visual index
                  onTextChange={handleStepTextChange}
                  onRemove={handleRemoveStepById}
                  canRemove={stepsForRender.length > 1} // Can remove if more than one step
                />
              ))}
            </SortableContext>
          </DndContext>
          <button 
            type="button" 
            onClick={handleAddStep} 
            className="btn btn-outline btn-accent w-full md:w-auto mt-2" // Added mt-2
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
            Add Step
          </button>
        </div>

        <div className="form-control mt-8 pt-6 border-t border-base-300">
          <button 
            type="submit" 
            disabled={isLoading} 
            className="btn btn-primary btn-lg w-full md:w-auto md:self-end"
          >
            {isLoading ? (
              <><span className="loading loading-spinner"></span> {isEditMode ? 'Updating Recipe...' : 'Creating Recipe...'}</>
            ) : (
              isEditMode ? 'Update Recipe' : 'Create Recipe'
            )}
          </button>
        </div>
      </form>
    </div>
  );
}

export default RecipeForm;
