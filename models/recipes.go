package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var ErrValidation = errors.New("name and description are required")

type RecipeIngredient struct {
	ID       int    `db:"id" json:"id"`
	RecipeID int    `db:"recipe_id" json:"recipe_id"`
	Name     string `db:"name" json:"name"`
	Amount   string `db:"amount" json:"amount"`
	Calories *int   `db:"calories" json:"calories"`
}

type RecipeStep struct {
	ID       int    `db:"id" json:"id"`
	RecipeID int    `db:"recipe_id" json:"recipe_id"`
	Order    int    `db:"order" json:"order"`
	Text     string `db:"text" json:"text"`
}

type Recipe struct {
	ID          int                `db:"id" json:"id"`
	Name        string             `db:"name" json:"name"`
	Description string             `db:"description" json:"description"`
	Slug        string             `db:"slug" json:"slug"`
	Image       sql.NullString     `db:"image" json:"image"`
	Ingredients []RecipeIngredient `json:"ingredients"`
	Steps       []RecipeStep       `json:"steps"`
	Tags        []string           `json:"tags"`
}

func GetRecipes(db *sqlx.DB) (*[]Recipe, error) {
	recipes := []Recipe{}
	err := db.Select(&recipes, "SELECT * FROM recipes")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Load tags for each recipe
	for i := range recipes {
		tags := []string{}
		err := db.Select(&tags, `SELECT t.name FROM tags t INNER JOIN recipe_tags rt ON t.id = rt.tag_id WHERE rt.recipe_id = $1`, recipes[i].ID)
		if err == nil {
			recipes[i].Tags = tags
		}
	}

	return &recipes, nil
}

func GetRecipeIdFromSlug(db *sqlx.DB, slug string) (int, error) {
	var id int
	err := db.Get(&id, "SELECT id FROM recipes WHERE slug=$1", slug)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return id, nil
}

func GetRecipe(db *sqlx.DB, i int) (*Recipe, error) {
	recipe := Recipe{}
	err := db.Get(&recipe, "SELECT * FROM recipes WHERE id=$1", i)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	ingredients := []RecipeIngredient{}
	err = db.Select(&ingredients, "SELECT * FROM recipe_ingredients WHERE recipe_id=$1", recipe.ID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	steps := []RecipeStep{}
	err = db.Select(&steps, "SELECT * FROM recipe_steps WHERE recipe_id=$1 ORDER BY \"order\" ASC", recipe.ID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	tags := []string{}
	err = db.Select(&tags, `SELECT t.name FROM tags t INNER JOIN recipe_tags rt ON t.id = rt.tag_id WHERE rt.recipe_id = $1`, recipe.ID)
	if err == nil {
		recipe.Tags = tags
	}

	recipe.Ingredients = ingredients
	recipe.Steps = steps

	return &recipe, nil
}

func NullStringWrapper(s string) sql.NullString {
	if s == "" {
		return sql.NullString{String: s, Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func UpdateRecipe(db *sqlx.DB, i int, r *Recipe) (*Recipe, error) {
	if r.Name == "" || r.Description == "" {
		return nil, ErrValidation
	}

	tx, err := db.Beginx()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = tx.Exec("UPDATE recipes SET name=$1, description=$2, image=$3 WHERE id=$4", r.Name, r.Description, r.Image, i)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	_, err = tx.Exec("DELETE FROM recipe_ingredients WHERE recipe_id=$1", i)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	for _, ingredient := range r.Ingredients {
		_, err = tx.Exec("INSERT INTO recipe_ingredients (recipe_id, name, amount, calories) VALUES ($1, $2, $3, $4)", i, ingredient.Name, ingredient.Amount, ingredient.Calories)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return nil, err
		}
	}

	_, err = tx.Exec("DELETE FROM recipe_steps WHERE recipe_id=$1", i)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	for _, step := range r.Steps {
		_, err = tx.Exec("INSERT INTO recipe_steps (recipe_id, \"order\", text) VALUES ($1, $2, $3)", i, step.Order, step.Text)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return nil, err
		}
	}

	// Handle tags if present
	if r.Tags != nil {
		// Remove all existing tags for this recipe
		_, err = tx.Exec("DELETE FROM recipe_tags WHERE recipe_id=$1", i)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return nil, err
		}
		for _, tag := range r.Tags {
			tag = strings.ToLower(tag)
			var tagID int
			err = tx.Get(&tagID, "SELECT id FROM tags WHERE name=$1", tag)
			if err != nil {
				// Tag does not exist, insert it
				err = tx.Get(&tagID, "INSERT INTO tags (name) VALUES ($1) RETURNING id", tag)
				if err != nil {
					tx.Rollback()
					fmt.Println(err)
					return nil, err
				}
			}
			_, err = tx.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES ($1, $2)", i, tagID)
			if err != nil {
				tx.Rollback()
				fmt.Println(err)
				return nil, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return GetRecipe(db, i)
}

func CreateRecipe(db *sqlx.DB, r *Recipe) (*Recipe, error) {
	if r.Name == "" || r.Description == "" {
		return nil, ErrValidation
	}

	r.Slug = slug.Make(r.Name)
	var id int
	err := db.Get(&id, "SELECT id FROM recipes WHERE slug=$1", r.Slug)
	if err == nil {
		i := 1
		for err == nil {
			fmt.Printf("slug %s already exists\n", r.Slug)
			r.Slug = slug.Make(r.Name + "-" + fmt.Sprint(i))
			fmt.Printf("trying %s\n", r.Slug)
			err = db.Get(&id, "SELECT id FROM recipes WHERE slug=$1", r.Slug)
			i++
		}
	}

	tx, err := db.Beginx()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = tx.Exec("INSERT INTO recipes (name, description, slug) VALUES ($1, $2, $3)", r.Name, r.Description, r.Slug)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	err = tx.Get(&id, "SELECT id FROM recipes WHERE slug=$1", r.Slug)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Now update tags, ingredients, steps, etc.
	return UpdateRecipe(db, id, r)
}

func DeleteRecipe(db *sqlx.DB, i int) error {
	tx, err := db.Beginx()
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec("DELETE FROM recipe_ingredients WHERE recipe_id=$1", i)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec("DELETE FROM recipe_steps WHERE recipe_id=$1", i)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec("DELETE FROM meal_recipes WHERE recipe_id=$1", i)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec("DELETE FROM recipes WHERE id=$1", i)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
