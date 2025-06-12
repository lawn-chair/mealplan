import axios, { InternalAxiosRequestConfig, AxiosResponse } from 'axios';

// Type for the getToken function from Clerk's useAuth
export type GetTokenFn = () => Promise<string | null>;

let getClerkToken: GetTokenFn | null = null;

export const setupAuthTokenInterceptor = (tokenFn: GetTokenFn) => {
  getClerkToken = tokenFn;
};

// Define interfaces for your API data structures (mirroring openapi.yaml)

export interface ApiError {
  error: string;
}

// Exporting RecipeIngredient and RecipeStep for use in components
export interface RecipeIngredient {
  id?: number;
  recipe_id?: number;
  name: string;
  amount: string;
  calories?: number | null;
}

export interface RecipeStep {
  id?: number | string;
  recipe_id?: number;
  order: number;
  text: string;
}

export interface Recipe {
  id?: number;
  name: string;
  description: string;
  slug?: string;
  image?: { Valid: boolean; String: string };
  ingredients: RecipeIngredient[]; // Use exported type
  steps: RecipeStep[]; // Use exported type
  tags?: string[]; // Optional tags field
}

// Interfaces for Meal components
export interface MealIngredient {
  id?: number;
  meal_id?: number;
  name: string;
  amount: string;
}

export interface MealStep {
  id?: number | string;
  meal_id?: number;
  order: number;
  text: string;
}

export interface MealRecipe {
  meal_id?: number;
  recipe_id: number;
  recipe_name?: string; // Added for UI purposes
  recipe_slug?: string; // Added for UI purposes
}

export interface Meal {
  id?: number;
  name: string;
  description: string;
  slug?: string;
  image?: { Valid: boolean; String: string };
  ingredients: MealIngredient[];
  steps: MealStep[];
  recipes: MealRecipe[];
  tags?: string[]; // Optional tags field
}

export interface Plan {
  id?: number;
  start_date: string; // Consider using Date type and formatting before sending
  end_date: string;   // Consider using Date type and formatting before sending
  user_id?: string;
  meals: number[];
}

export interface Pantry {
  id?: number;
  user_id?: string;
  items: string[];
}

export interface ShoppingListItem {
  name: string;
  amount: string;
  checked: boolean;
}

export interface ShoppingList {
  plan: Plan; 
  ingredients: ShoppingListItem[];
}

export interface ShoppingListUpdatePayload {
  plan: { id: number }; 
  ingredients: ShoppingListItem[];
}

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Updated request interceptor to use the provided getToken function
apiClient.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    if (getClerkToken) {
      const token = await getClerkToken();
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Define types for API responses
export const getRecipes = (): Promise<AxiosResponse<Recipe[]>> => apiClient.get('/recipes');
export const getRecipeById = (id: number): Promise<AxiosResponse<Recipe>> => apiClient.get(`/recipes/${id}`);
// Ensure getRecipeBySlug uses a query parameter as per backend requirement
// Adjusted return type to Recipe[] as per OpenAPI spec for /recipes?slug=...
export const getRecipeBySlug = (slug: string): Promise<AxiosResponse<Recipe>> => apiClient.get(`/recipes?slug=${slug}`);
export const createRecipe = (recipeData: Omit<Recipe, 'id' | 'slug'>): Promise<AxiosResponse<Recipe>> => apiClient.post('/recipes', recipeData);
export const updateRecipe = (id: number, recipeData: Partial<Omit<Recipe, 'id' | 'slug'>>): Promise<AxiosResponse<Recipe>> => apiClient.put(`/recipes/${id}`, recipeData);
export const deleteRecipe = (id: number): Promise<AxiosResponse<void>> => apiClient.delete(`/recipes/${id}`);

export const getMeals = (): Promise<AxiosResponse<Meal[]>> => apiClient.get('/meals');
export const getMealById = (id: number): Promise<AxiosResponse<Meal>> => apiClient.get(`/meals/${id}`);
export const createMeal = (mealData: Omit<Meal, 'id' | 'slug'>): Promise<AxiosResponse<Meal>> => apiClient.post('/meals', mealData);
export const updateMeal = (id: number, mealData: Partial<Omit<Meal, 'id' | 'slug'>>): Promise<AxiosResponse<Meal>> => apiClient.put(`/meals/${id}`, mealData);
export const deleteMeal = (id: number): Promise<AxiosResponse<void>> => apiClient.delete(`/meals/${id}`);
export const getMealBySlug = (slug: string): Promise<AxiosResponse<Meal>> => apiClient.get(`/meals?slug=${slug}`);

export const getPlans = (params?: { last?: boolean; next?: boolean; future?: boolean }): Promise<AxiosResponse<Plan[]>> => apiClient.get('/plans', { params });
export const getUpcomingPlans = (): Promise<AxiosResponse<Plan[]>> => apiClient.get('/plans?future=true');
export const getPlanById = (id: number): Promise<AxiosResponse<Plan>> => apiClient.get(`/plans/${id}`);
export const createPlan = (planData: Omit<Plan, 'id' | 'user_id'>): Promise<AxiosResponse<Plan>> => apiClient.post('/plans', planData);
export const updatePlan = (id: number, planData: Partial<Omit<Plan, 'id' | 'user_id'>>): Promise<AxiosResponse<Plan>> => apiClient.put(`/plans/${id}`, planData);
export const deletePlan = (id: number): Promise<AxiosResponse<void>> => apiClient.delete(`/plans/${id}`);
export const getPlanIngredients = (id: number): Promise<AxiosResponse<Array<{name: string, amount: string}>>> => apiClient.get(`/plans/${id}/ingredients`);

export const getPantry = (): Promise<AxiosResponse<Pantry>> => apiClient.get('/pantry');
export const createPantry = (pantryData: { items: string[] }): Promise<AxiosResponse<Pantry>> => apiClient.post('/pantry', pantryData);
export const updatePantry = (pantryData: { items: string[] }): Promise<AxiosResponse<Pantry>> => apiClient.put('/pantry', pantryData);
export const clearPantry = (): Promise<AxiosResponse<void>> => apiClient.delete('/pantry');

export const getShoppingList = (): Promise<AxiosResponse<ShoppingList>> => apiClient.get('/shopping-list');
export const updateShoppingList = (payload: ShoppingListUpdatePayload): Promise<AxiosResponse<ShoppingList>> => apiClient.put('/shopping-list', payload);

export const uploadImage = (formData: FormData): Promise<AxiosResponse<{ url: string }>> => apiClient.post('/images', formData, {
  headers: {
    'Content-Type': 'multipart/form-data',
  },
});

export const getTags = (): Promise<AxiosResponse<string[]>> => apiClient.get('/tags');

export default apiClient;
