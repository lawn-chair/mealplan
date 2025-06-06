package models

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestGetRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Test data
	recipeID := 1
	name := "Test Recipe"
	slug := "test-recipe"
	description := "A delicious test recipe"
	image := sql.NullString{String: "test-recipe.jpg", Valid: true}

	// Setup mock query responses
	recipeRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		recipeID, name, description, slug, image,
	)

	mock.ExpectQuery("SELECT \\* FROM recipes WHERE id=\\$1").
		WithArgs(recipeID).
		WillReturnRows(recipeRows)

	// Mock for ingredients
	ingredientRows := sqlmock.NewRows([]string{"id", "recipe_id", "name", "amount", "calories"}).
		AddRow(1, recipeID, "Ingredient1", "1 cup", nil).
		AddRow(2, recipeID, "Ingredient2", "2 tbsp", nil)

	mock.ExpectQuery("SELECT \\* FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(recipeID).
		WillReturnRows(ingredientRows)

	// Mock for steps
	stepsRows := sqlmock.NewRows([]string{"id", "recipe_id", "order", "text"}).
		AddRow(1, recipeID, 1, "Step 1").
		AddRow(2, recipeID, 2, "Step 2")

	mock.ExpectQuery("SELECT \\* FROM recipe_steps WHERE recipe_id=\\$1 ORDER BY").
		WithArgs(recipeID).
		WillReturnRows(stepsRows)

	// Call the function
	recipe, err := GetRecipe(sqlxDB, recipeID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, recipe)
	assert.Equal(t, recipeID, recipe.ID)
	assert.Equal(t, name, recipe.Name)
	assert.Equal(t, slug, recipe.Slug)
	assert.Equal(t, description, recipe.Description)
	assert.Equal(t, image, recipe.Image)
	assert.Equal(t, 2, len(recipe.Ingredients))
	assert.Equal(t, "Ingredient1", recipe.Ingredients[0].Name)
	assert.Equal(t, "1 cup", recipe.Ingredients[0].Amount)
	assert.Equal(t, 2, len(recipe.Steps))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetRecipes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Setup mock query responses
	recipeRows := sqlmock.NewRows([]string{"id", "name", "description", "slug", "image"}).
		AddRow(1, "Recipe 1", "Description 1", "recipe-1", sql.NullString{String: "image1.jpg", Valid: true}).
		AddRow(2, "Recipe 2", "Description 2", "recipe-2", sql.NullString{String: "image2.jpg", Valid: true})

	mock.ExpectQuery("SELECT \\* FROM recipes").
		WillReturnRows(recipeRows)

	// Call the function
	recipes, err := GetRecipes(sqlxDB)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, recipes)
	assert.Equal(t, 2, len(*recipes))
	assert.Equal(t, "Recipe 1", (*recipes)[0].Name)
	assert.Equal(t, "Recipe 2", (*recipes)[1].Name)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestCreateRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Test data
	name := "New Test Recipe"
	slug := "new-test-recipe"
	description := "A delicious new test recipe"
	image := sql.NullString{String: "new-test-recipe.jpg", Valid: true}

	newRecipe := &Recipe{
		Name:        name,
		Slug:        slug,
		Description: description,
		Image:       image,
		Ingredients: []RecipeIngredient{
			{Name: "Ing1", Amount: "1 cup"},
			{Name: "Ing2", Amount: "2 tsp"},
		},
		Steps: []RecipeStep{
			{Order: 1, Text: "Step 1"},
			{Order: 2, Text: "Step 2"},
		},
	}

	// First check if slug exists
	mock.ExpectQuery("SELECT id FROM recipes WHERE slug=\\$1").
		WithArgs(slug).
		WillReturnError(sql.ErrNoRows)

	// Mock the transaction for insert
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO recipes").
		WithArgs(name, description, slug).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT id FROM recipes WHERE slug=\\$1").
		WithArgs(slug).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	// For UpdateRecipe call within CreateRecipe
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes SET").
		WithArgs(name, description, image, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete old ingredients and insert new
	mock.ExpectExec("DELETE FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	for _, ing := range newRecipe.Ingredients {
		mock.ExpectExec("INSERT INTO recipe_ingredients").
			WithArgs(1, ing.Name, ing.Amount, ing.Calories).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// Delete old steps and insert new
	mock.ExpectExec("DELETE FROM recipe_steps WHERE recipe_id=\\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	for _, step := range newRecipe.Steps {
		mock.ExpectExec("INSERT INTO recipe_steps").
			WithArgs(1, step.Order, step.Text).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	// For the GetRecipe call at the end
	recipeRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(1, name, description, slug, image)

	mock.ExpectQuery("SELECT \\* FROM recipes WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(recipeRows)

	ingredientRows := sqlmock.NewRows([]string{"id", "recipe_id", "name", "amount", "calories"}).
		AddRow(1, 1, "Ing1", "1 cup", nil).
		AddRow(2, 1, "Ing2", "2 tsp", nil)

	mock.ExpectQuery("SELECT \\* FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(1).
		WillReturnRows(ingredientRows)

	stepsRows := sqlmock.NewRows([]string{"id", "recipe_id", "order", "text"}).
		AddRow(1, 1, 1, "Step 1").
		AddRow(2, 1, 2, "Step 2")

	mock.ExpectQuery("SELECT \\* FROM recipe_steps WHERE recipe_id=\\$1 ORDER BY").
		WithArgs(1).
		WillReturnRows(stepsRows)

	// Call the function
	createdRecipe, err := CreateRecipe(sqlxDB, newRecipe)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, createdRecipe)
	assert.Equal(t, 1, createdRecipe.ID)
	assert.Equal(t, name, createdRecipe.Name)
	assert.Equal(t, slug, createdRecipe.Slug)
	assert.Equal(t, 2, len(createdRecipe.Ingredients))
	assert.Equal(t, 2, len(createdRecipe.Steps))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUpdateRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Test data
	recipeID := 1
	name := "Updated Test Recipe"
	slug := "updated-test-recipe"
	description := "An updated test recipe"
	image := sql.NullString{String: "updated-test-recipe.jpg", Valid: true}

	updateRecipe := &Recipe{
		ID:          recipeID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Image:       image,
		Ingredients: []RecipeIngredient{
			{Name: "UpdatedIng1", Amount: "2 cups"},
			{Name: "UpdatedIng2", Amount: "3 tbsp"},
			{Name: "UpdatedIng3", Amount: "1/2 tsp"},
		},
		Steps: []RecipeStep{
			{Order: 1, Text: "Updated Step 1"},
			{Order: 2, Text: "Updated Step 2"},
		},
	}

	// Mock the transaction for UpdateRecipe
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes SET").
		WithArgs(name, description, image, recipeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM recipe_ingredients WHERE recipe_id").
		WithArgs(recipeID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	for _, ing := range updateRecipe.Ingredients {
		mock.ExpectExec("INSERT INTO recipe_ingredients").
			WithArgs(recipeID, ing.Name, ing.Amount, ing.Calories).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectExec("DELETE FROM recipe_steps WHERE recipe_id").
		WithArgs(recipeID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	for _, step := range updateRecipe.Steps {
		mock.ExpectExec("INSERT INTO recipe_steps").
			WithArgs(recipeID, step.Order, step.Text).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	// For the GetRecipe call at the end
	recipeRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(recipeID, name, description, slug, image)

	mock.ExpectQuery("SELECT \\* FROM recipes WHERE id=\\$1").
		WithArgs(recipeID).
		WillReturnRows(recipeRows)

	ingredientRows := sqlmock.NewRows([]string{"id", "recipe_id", "name", "amount", "calories"}).
		AddRow(3, recipeID, "UpdatedIng1", "2 cups", nil).
		AddRow(4, recipeID, "UpdatedIng2", "3 tbsp", nil).
		AddRow(5, recipeID, "UpdatedIng3", "1/2 tsp", nil)

	mock.ExpectQuery("SELECT \\* FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(recipeID).
		WillReturnRows(ingredientRows)

	stepsRows := sqlmock.NewRows([]string{"id", "recipe_id", "order", "text"}).
		AddRow(3, recipeID, 1, "Updated Step 1").
		AddRow(4, recipeID, 2, "Updated Step 2")

	mock.ExpectQuery("SELECT \\* FROM recipe_steps WHERE recipe_id=\\$1 ORDER BY").
		WithArgs(recipeID).
		WillReturnRows(stepsRows)

	// Call the function
	updatedRecipe, err := UpdateRecipe(sqlxDB, recipeID, updateRecipe)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedRecipe)
	assert.Equal(t, recipeID, updatedRecipe.ID)
	assert.Equal(t, name, updatedRecipe.Name)
	assert.Equal(t, 3, len(updatedRecipe.Ingredients))
	assert.Equal(t, 2, len(updatedRecipe.Steps))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestDeleteRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	recipeID := 1

	// Mock the transaction for DeleteRecipe
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(recipeID).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("DELETE FROM recipe_steps WHERE recipe_id=\\$1").
		WithArgs(recipeID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM meal_recipes WHERE recipe_id=\\$1").
		WithArgs(recipeID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM recipes WHERE id=\\$1").
		WithArgs(recipeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Call the function
	err = DeleteRecipe(sqlxDB, recipeID)

	// Assertions
	assert.NoError(t, err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetRecipeIdFromSlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	slug := "test-recipe"
	expectedID := 1

	mock.ExpectQuery("SELECT id FROM recipes WHERE slug=\\$1").
		WithArgs(slug).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	// Call the function
	id, err := GetRecipeIdFromSlug(sqlxDB, slug)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
