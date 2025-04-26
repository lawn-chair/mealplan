
export interface MealData {
    id?: number;
    slug: string;
    name: string;
    description: string;
    image?: {Valid: boolean, String: string};
    ingredients: {amount: string, name: string, calories?: number}[]; // Replace `any` with a more specific type if possible
    steps: {id?: number, text: string, order: number}[]; // Replace `any` with a more specific type if possible
    recipes: {recipe_id: number}[];
}

export interface RecipeData {
    id?: number;
    slug: string;
    name: string;
    description: string;
    image?: {Valid: boolean, String: string};
    ingredients: {amount: string, name: string, calories?: number}[]; 
    steps: {text: string, order: number}[]; 
}

export interface PlanData {
    id?: number;
    start_date: string;
    end_date: string;
    user_id: string;
    meals?: number[];
}

export interface Pantry {
    id?: number;
    user_id: string;
    items: string[]; 
}