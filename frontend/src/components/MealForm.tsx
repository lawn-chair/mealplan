import React, { useState, useEffect, ChangeEvent } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { 
  DndContext, 
  closestCenter, 
  KeyboardSensor, 
  PointerSensor, 
  useSensor, 
  useSensors,
  DragEndEvent 
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { getMealBySlug, createMeal, updateMeal, Meal, MealIngredient, MealStep, uploadImage, Recipe, getRecipes, getRecipeById } from '../api';
import { SortableStepItem } from './SortableStepItem';
import TagInput from './TagInput';

interface MealFormProps {
  isEditMode?: boolean;
}

// Helper type to ensure step ID is always defined when interacting with SortableStepItem
type MealStepWithDefinedId = Omit<MealStep, 'id'> & { id: string | number };

const MealForm: React.FC<MealFormProps> = ({ isEditMode }) => {
  const { slug: mealSlug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const [meal, setMeal] = useState<Meal>({
    id: undefined,
    slug: undefined,
    name: '',
    description: '',
    ingredients: [],
    steps: [], 
    recipes: [],
    image: { Valid: false, String: '' },
  });
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null);
  const [currentImageUrl, setCurrentImageUrl] = useState<string | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [tags, setTags] = useState<string[]>([]);
  const [allRecipes, setAllRecipes] = useState<Recipe[]>([]);
  const [recipeSearch, setRecipeSearch] = useState('');
  const [showRecipePicker, setShowRecipePicker] = useState(false);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  useEffect(() => {
    if (isEditMode && mealSlug) {
      const fetchMeal = async () => {
        try {
          setLoading(true);
          const response = await getMealBySlug(mealSlug);
          const fetchedMeal = response.data;
          setMeal({
            ...fetchedMeal,
            ingredients: fetchedMeal.ingredients || [],
            steps: (fetchedMeal.steps || []).map((step, index) => ({
              ...step,
              id: step.id || `loaded-step-${index}-${Date.now()}`,
            })),
            recipes: fetchedMeal.recipes || [],
            image: fetchedMeal.image || { Valid: false, String: '' },
          });
          setTags(fetchedMeal.tags || []);
          if (fetchedMeal.image && fetchedMeal.image.Valid && fetchedMeal.image.String) {
            setCurrentImageUrl(fetchedMeal.image.String);
            setImagePreview(fetchedMeal.image.String);
          }
          setError(null);
        } catch (err) {
          setError('Failed to fetch meal details.');
          console.error(err);
        } finally {
          setLoading(false);
        }
      };
      fetchMeal();
    } else {
      setMeal({
        id: undefined,
        slug: undefined,
        name: '',
        description: '',
        ingredients: [],
        steps: [{ id: `new-step-${Date.now()}`, order: 1, text: '' }],
        recipes: [],
        image: { Valid: false, String: '' },
      });
      setSelectedImageFile(null);
      setCurrentImageUrl(null);
      setImagePreview(null);
      setTags([]);
    }

    // Fetch all recipes for picker
    getRecipes().then(res => setAllRecipes(res.data || []));
  }, [mealSlug, isEditMode]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setMeal((prevMeal) => ({
      ...prevMeal,
      [name]: value,
    }));
  };

  const handleIngredientChange = (index: number, field: keyof MealIngredient, value: string) => {
    const newIngredients = [...meal.ingredients];
    newIngredients[index] = { ...newIngredients[index], [field]: value };
    setMeal((prevMeal) => ({ ...prevMeal, ingredients: newIngredients }));
  };

  const addIngredient = () => {
    setMeal((prevMeal) => ({
      ...prevMeal,
      ingredients: [...prevMeal.ingredients, { name: '', amount: '' }],
    }));
  };

  const removeIngredient = (index: number) => {
    const newIngredients = meal.ingredients.filter((_, i) => i !== index);
    setMeal((prevMeal) => ({ ...prevMeal, ingredients: newIngredients }));
  };

  const handleStepTextChange = (itemId: string | number, newText: string) => {
    setMeal((prevMeal) => ({
      ...prevMeal,
      steps: prevMeal.steps.map(step => 
        step.id === itemId ? { ...step, text: newText, id: step.id || `ensure-${itemId}` } : step
      ) as MealStepWithDefinedId[],
    }));
  };

  const addStep = () => {
    setMeal((prevMeal) => ({
      ...prevMeal,
      steps: [...prevMeal.steps, { id: `new-step-${Date.now()}`, order: prevMeal.steps.length + 1, text: '' }] as MealStepWithDefinedId[],
    }));
  };

  const removeStepById = (idToRemove: number | string) => {
    setMeal((prevMeal) => {
      const newSteps = prevMeal.steps.filter((step) => step.id !== idToRemove);
      const reorderedSteps = newSteps.map((step, index) => ({ ...step, order: index + 1, id: step.id || `ensure-removed-${index}` }));
      return { ...prevMeal, steps: reorderedSteps as MealStepWithDefinedId[] };
    });
  };


  const removeRecipe = (recipeIdToRemove: number) => {
    setMeal(prevMeal => ({
        ...prevMeal,
        recipes: prevMeal.recipes.filter(recipe => recipe.recipe_id !== recipeIdToRemove)
    }));
  };

  const handleImageChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      const file = event.target.files[0];
      setSelectedImageFile(file);
      setImagePreview(URL.createObjectURL(file));
    } else {
      setSelectedImageFile(null);
      setImagePreview(currentImageUrl);
    }
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const {active, over} = event;
    if (over && active.id !== over.id) {
      setMeal((prevMeal) => {
        const stepsWithGuaranteedIds = prevMeal.steps.map((s, i) => ({ ...s, id: s.id || `temp-dnd-${i}-${Date.now()}`})) as MealStepWithDefinedId[];
        const oldIndex = stepsWithGuaranteedIds.findIndex((step) => step.id === active.id);
        const newIndex = stepsWithGuaranteedIds.findIndex((step) => step.id === over.id);
        if (oldIndex === -1 || newIndex === -1) return prevMeal;

        const reorderedSteps = arrayMove(stepsWithGuaranteedIds, oldIndex, newIndex);
        const finalSteps = reorderedSteps.map((step, index) => ({ ...step, order: index + 1 }));
        return { ...prevMeal, steps: finalSteps };
      });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    let finalImage = meal.image;

    if (selectedImageFile) {
      try {
        const formData = new FormData();
        formData.append('file', selectedImageFile);
        const uploadResponse = await uploadImage(formData);
        if (uploadResponse.data.url) {
          finalImage = { Valid: true, String: uploadResponse.data.url };
        } else {
          throw new Error('Image URL not found in upload response.');
        }
      } catch (err) {
        console.error("Failed to upload image:", err);
        setError('Failed to upload image. Please try again.');
        setLoading(false);
        return;
      }
    } else if (!imagePreview && meal.image && meal.image.Valid) {
        finalImage = { Valid: false, String: '' };
    }

    const processedSteps = meal.steps.map((step, index) => ({
      id: typeof step.id === 'string' && (step.id.startsWith('new-step-') || step.id.startsWith('loaded-step-') || step.id.startsWith('temp-dnd-') || step.id.startsWith('ensure-')) ? undefined : step.id,
      order: index + 1,
      text: step.text,
    }));

    const mealDataPayload = {
      name: meal.name,
      description: meal.description,
      ingredients: meal.ingredients.map(ing => ({ name: ing.name, amount: ing.amount })),
      steps: processedSteps,
      recipes: meal.recipes.map(r => ({ recipe_id: r.recipe_id })),
      image: finalImage,
      tags: tags.length > 0 ? tags : undefined,
    };

    try {
      if (isEditMode && meal.id) {
        await updateMeal(meal.id, mealDataPayload);
        navigate(`/meals/${meal.slug || meal.id}`);
      } else {
        const response = await createMeal(mealDataPayload);
        navigate(`/meals/${response.data.slug}`);
      }
    } catch (err: any) {
      setError((isEditMode ? 'Failed to update meal: ' : 'Failed to create meal: ') + (err.response?.data?.error || err.message));
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  if (loading && isEditMode && !meal.name) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)]">
        <span className="loading loading-lg loading-spinner text-primary mb-4"></span>
        <p className="text-lg">Loading meal details...</p>
      </div>
    );
  }

  // Ensure meal.steps always contains items with defined IDs for SortableContext and SortableStepItem
  const stepsForRender = meal.steps.map((step, index) => ({
    ...step,
    id: step.id || `render-step-${index}-${Date.now()}`,
  })) as MealStepWithDefinedId[];

  return (
    <div className="container mx-auto px-4 py-8 max-w-3xl">
      <form onSubmit={handleSubmit} className="space-y-6 bg-base-100 p-6 md:p-8 rounded-lg shadow-xl">
        <h2 className="text-2xl font-bold mb-6 text-center">{isEditMode ? 'Edit Meal' : 'Create New Meal'}</h2>
        
        {error && (
          <div role="alert" className="alert alert-error">
            <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <span>{error}</span>
          </div>
        )}

        {/* TagInput for tags */}
        <TagInput tags={tags} setTags={setTags} label="Tags" placeholder="Add a tag and press Enter" />

        {/* Name Field */}
        <div className="form-control">
          <label htmlFor="name" className="label">
            <span className="label-text">Name</span>
          </label>
          <input
            type="text"
            id="name"
            name="name"
            value={meal.name}
            onChange={handleChange}
            required
            className="input input-bordered input-primary w-full"
          />
        </div>

        {/* Description Field */}
        <div className="form-control">
          <label htmlFor="description" className="label">
            <span className="label-text">Description</span>
          </label>
          <textarea
            id="description"
            name="description"
            value={meal.description}
            onChange={handleChange}
            required
            className="textarea textarea-bordered textarea-primary w-full h-32"
          />
        </div>

        {/* Image Upload Section */}
        <div className="form-control">
          <label htmlFor="mealImage" className="label">
            <span className="label-text">Meal Image</span>
          </label>
          <input
            type="file"
            id="mealImage"
            accept="image/*"
            onChange={handleImageChange}
            className="file-input file-input-bordered file-input-primary w-full"
          />
          {imagePreview && (
            <div className="mt-4">
              <p className="label-text mb-2">Image Preview:</p>
              <img src={imagePreview} alt="Preview" className="max-w-xs rounded-lg shadow-md" />
            </div>
          )}
        </div>

        {/* Ingredients Section */}
        <div className="space-y-4 p-4 border border-base-300 rounded-lg">
          <h3 className="text-lg font-semibold">Ingredients</h3>
          {meal.ingredients.map((ingredient, index) => (
            <div key={index} className="flex items-center gap-2 p-2 bg-base-200 rounded">
              <input
                type="text"
                placeholder="Ingredient Name"
                value={ingredient.name}
                onChange={(e) => handleIngredientChange(index, 'name', e.target.value)}
                required
                className="input input-bordered input-primary input-sm flex-grow"
              />
              <input
                type="text"
                placeholder="Amount"
                value={ingredient.amount}
                onChange={(e) => handleIngredientChange(index, 'amount', e.target.value)}
                required
                className="input input-bordered input-primary input-sm w-1/3"
              />
              <button type="button" onClick={() => removeIngredient(index)} className="btn btn-error btn-sm btn-outline">Remove</button>
            </div>
          ))}
          <button type="button" onClick={addIngredient} className="btn btn-secondary btn-sm">+ Add Ingredient</button>
        </div>


        {/* Steps Section */}
        <div className="space-y-4 p-4 border border-base-300 rounded-lg">
          <h3 className="text-lg font-semibold">Steps</h3>
          <DndContext 
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <SortableContext 
              items={stepsForRender.map(step => step.id)} 
              strategy={verticalListSortingStrategy}
            >
              {stepsForRender.map((stepWithId, index) => (
                <SortableStepItem<MealStepWithDefinedId>
                  key={stepWithId.id}
                  item={stepWithId}
                  index={index} 
                  onTextChange={handleStepTextChange} 
                  onRemove={removeStepById} 
                  canRemove={stepsForRender.length > 1}
                />
              ))}
            </SortableContext>
          </DndContext>
          <button type="button" onClick={addStep} className="btn btn-secondary btn-sm mt-2">+ Add Step</button>
        </div>

        {/* Recipes Section */}
        <div className="space-y-4 p-4 border border-base-300 rounded-lg">
          <h3 className="text-lg font-semibold">Associated Recipes</h3>
          <button type="button" className="btn btn-secondary mb-2" onClick={() => setShowRecipePicker(true)}>
            + Add Recipe
          </button>
          {showRecipePicker && (
            <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
              <div className="bg-base-100 p-6 rounded-lg shadow-lg w-full max-w-2xl relative">
                <button className="btn btn-sm btn-ghost absolute top-2 right-2" onClick={() => setShowRecipePicker(false)}>✕</button>
                <h4 className="text-lg font-bold mb-2">Select a Recipe</h4>
                <input
                  type="text"
                  className="input input-bordered w-full mb-4"
                  placeholder="Search recipes..."
                  value={recipeSearch}
                  onChange={e => setRecipeSearch(e.target.value)}
                />
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 max-h-96 overflow-y-auto">
                  {allRecipes.filter(recipe => {
                    const q = recipeSearch.trim().toLowerCase();
                    if (!q) return true;
                    return (
                      recipe.name?.toLowerCase().includes(q) ||
                      recipe.description?.toLowerCase().includes(q) ||
                      recipe.tags?.some(tag => tag.toLowerCase().includes(q))
                    );
                  }).map(recipe => {
                    const imageUrl = typeof recipe.image === 'string'
                      ? recipe.image
                      : (recipe.image?.Valid ? recipe.image.String : '/recipe-blank.jpg');
                    return (
                      <div key={recipe.id} className="card bg-base-200 shadow-sm flex flex-col">
                        <figure className="w-full h-32 bg-base-300 flex items-center justify-center overflow-hidden rounded-t-lg">
                          <img src={imageUrl} alt={recipe.name} className="object-cover w-full h-full" />
                        </figure>
                        <div className="card-body p-4 flex-1 flex flex-col">
                          <h5 className="font-bold text-base mb-1">{recipe.name}</h5>
                          <p className="text-xs text-gray-500 mb-2 line-clamp-2">{recipe.description}</p>
                          <button
                            type="button"
                            className="btn btn-primary btn-sm mt-auto"
                            onClick={async () => {
                              setShowRecipePicker(false);
                              setLoading(true);
                              try {
                                // Add recipe reference and steps/ingredients as before
                                const foundRecipe = (await getRecipeById(recipe.id!)).data;
                                console.log('Adding recipe:', foundRecipe);
                                if (!meal.recipes.find(r => r.recipe_id === foundRecipe.id)) {
                                  setMeal(prevMeal => {
                                    return {
                                      ...prevMeal,
                                      recipes: [...prevMeal.recipes, { recipe_id: foundRecipe.id!, recipe_slug: foundRecipe.slug, recipe_name: foundRecipe.name }],
                                      ingredients: [...prevMeal.ingredients, ...foundRecipe.ingredients],
                                      steps: [...prevMeal.steps, ...foundRecipe.steps]
                                    };
                                  });
                                }
                              } finally {
                                setLoading(false);
                              }
                            }}
                          >
                            Add This Recipe
                          </button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}
          {meal.recipes.length > 0 && (
            <ul className="list-disc list-inside pl-2 space-y-1 mt-2">
              {meal.recipes.map((recipe) => (
                <li key={recipe.recipe_id} className="text-sm flex justify-between items-center bg-base-200 p-2 rounded">
                  <span>{recipe.recipe_name ? `${recipe.recipe_name} (Slug: ${recipe.recipe_slug || 'N/A'})` : `Recipe ID: ${recipe.recipe_id}`}</span>
                  <button type="button" onClick={() => removeRecipe(recipe.recipe_id)} className="btn btn-xs btn-error btn-outline">Remove</button>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="flex justify-end gap-4 pt-4">
          <button type="button" onClick={() => navigate(isEditMode && mealSlug ? `/meals/${mealSlug}` : '/meals')} className="btn btn-ghost">
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? <span className="loading loading-spinner loading-xs"></span> : (isEditMode ? 'Update Meal' : 'Create Meal')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default MealForm;
