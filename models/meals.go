package models

import (
	"database/sql"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//var ErrValidation = errors.New("name, description, and slug are required")

type MealIngredient struct {
	ID     int    `db:"id" json:"id"`
	MealID int    `db:"meal_id" json:"meal_id"`
	Name   string `db:"name" json:"name"`
	Amount string `db:"amount" json:"amount"`
}

type MealStep struct {
	ID     int    `db:"id" json:"id"`
	MealID int    `db:"meal_id" json:"meal_id"`
	Order  int    `db:"order" json:"order"`
	Text   string `db:"text" json:"text"`
}

type MealRecipes struct {
	MealID   int `db:"meal_id" json:"meal_id"`
	RecipeID int `db:"recipe_id" json:"recipe_id"`
}

type Meal struct {
	ID          int              `db:"id" json:"id"`
	Name        string           `db:"name" json:"name"`
	Description string           `db:"description" json:"description"`
	Slug        string           `db:"slug" json:"slug"`
	Image       sql.NullString   `db:"image" json:"image"`
	Ingredients []MealIngredient `json:"ingredients"`
	Steps       []MealStep       `json:"steps"`
	MealRecipes []MealRecipes    `json:"recipes"`
}

func GetMeals(db *sqlx.DB) (*[]Meal, error) {

	meals := []Meal{}
	err := db.Select(&meals, "SELECT * FROM meals")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &meals, nil
}

func CreateMeal(db *sqlx.DB, meal *Meal) (*Meal, error) {
	meal.Slug = slug.Make(meal.Name)
	var id int
	err := db.Get(&id, "SELECT id FROM meals WHERE slug=$1", meal.Slug)
	if err == nil {
		i := 0
		for err == nil {
			i++
			fmt.Printf("slug %s already exists\n", meal.Slug)
			meal.Slug = slug.Make(meal.Name + "-" + fmt.Sprint(i))
			fmt.Printf("trying %s\n", meal.Slug)
			err = db.Get(&id, "SELECT id FROM meals WHERE slug=$1", meal.Slug)
		}
	}

	tx := db.MustBegin()
	tx.NamedExec("INSERT INTO meals (name, description, slug, image) VALUES (:name, :description, :slug, :image)", meal)
	tx.Commit()

	id, err = GetMealIdFromSlug(db, meal.Slug)
	if err != nil {
		return nil, err
	}

	return UpdateMeal(db, id, meal)
}

func GetMealIdFromSlug(db *sqlx.DB, slug string) (int, error) {
	var id int
	err := db.Get(&id, "SELECT id FROM meals WHERE slug=$1", slug)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return id, nil
}

func GetMeal(db *sqlx.DB, i int) (*Meal, error) {
	meal := Meal{}
	err := db.Get(&meal, "SELECT * FROM meals WHERE id=$1", i)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	ingredients := []MealIngredient{}
	err = db.Select(&ingredients, "SELECT * FROM meal_ingredients WHERE meal_id=$1", meal.ID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	steps := []MealStep{}
	err = db.Select(&steps, "SELECT * FROM meal_steps WHERE meal_id=$1 ORDER BY \"order\" ASC", meal.ID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	recipes := []MealRecipes{}
	err = db.Select(&recipes, "SELECT * FROM meal_recipes WHERE meal_id=$1", meal.ID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	meal.Ingredients = ingredients
	meal.Steps = steps
	meal.MealRecipes = recipes

	return &meal, nil
}

func UpdateMeal(db *sqlx.DB, i int, meal *Meal) (*Meal, error) {
	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	// Update the Meal table
	_, err = tx.Exec("UPDATE meals SET name=$1, description=$2, image=$3 WHERE id=$4", meal.Name, meal.Description, meal.Image, i)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		fmt.Println(err)
		return nil, err
	}
	// Update MealIngredients: Delete old and insert new
	_, err = tx.Exec("DELETE FROM meal_ingredients WHERE meal_id=$1", i)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, ingredient := range meal.Ingredients {
		_, err = tx.Exec("INSERT INTO meal_ingredients (meal_id, name, amount) VALUES ($1, $2, $3)", i, ingredient.Name, ingredient.Amount)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return nil, err
		}
	}
	// Update MealSteps: Delete old and insert new
	_, err = tx.Exec("DELETE FROM meal_steps WHERE meal_id=$1", i)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, step := range meal.Steps {
		_, err = tx.Exec("INSERT INTO meal_steps (meal_id, text, \"order\") VALUES ($1, $2, $3)", i, step.Text, step.Order)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// Update MealRecipes: Delete old and insert new
	_, err = tx.Exec("DELETE FROM meal_recipes WHERE meal_id=$1", i)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	fmt.Println(meal.MealRecipes)
	fmt.Println(i)
	for _, recipe := range meal.MealRecipes {
		_, err = tx.Exec("INSERT INTO meal_recipes (meal_id, recipe_id) VALUES ($1, $2)", i, recipe.RecipeID)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			fmt.Println(recipe)
			return nil, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return GetMeal(db, i)
}

func DeleteMeal(db *sqlx.DB, i int) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM meal_ingredients WHERE meal_id=$1", i)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM meal_steps WHERE meal_id=$1", i)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM meal_recipes WHERE meal_id=$1", i)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM plan_meals WHERE meal_id=$1", i)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM meals WHERE id=$1", i)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
