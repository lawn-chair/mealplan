package models

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestGetMeal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Test data
	mealID := 1
	name := "Test Meal"
	description := "A delicious test meal"
	slug := "test-meal"
	image := sql.NullString{String: "test-meal.jpg", Valid: true}

	// Mock the meal query
	mealRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		mealID, name, description, slug, image,
	)

	mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
		WithArgs(mealID).
		WillReturnRows(mealRows)

	// Mock for ingredients
	ingredientRows := sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}).
		AddRow(1, mealID, "Ingredient1", "1 cup").
		AddRow(2, mealID, "Ingredient2", "2 tbsp")

	mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnRows(ingredientRows)

	// Mock for steps
	stepsRows := sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}).
		AddRow(1, mealID, 1, "Step 1").
		AddRow(2, mealID, 2, "Step 2")

	mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1 ORDER BY").
		WithArgs(mealID).
		WillReturnRows(stepsRows)

	// Mock for recipes
	recipesRows := sqlmock.NewRows([]string{"meal_id", "recipe_id"}).
		AddRow(mealID, 101).
		AddRow(mealID, 102)

	mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnRows(recipesRows)

	// Call the function
	meal, err := GetMeal(sqlxDB, mealID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, meal)
	assert.Equal(t, mealID, meal.ID)
	assert.Equal(t, name, meal.Name)
	assert.Equal(t, description, meal.Description)
	assert.Equal(t, slug, meal.Slug)
	assert.Equal(t, image, meal.Image)
	assert.Equal(t, 2, len(meal.Ingredients))
	assert.Equal(t, 2, len(meal.Steps))
	assert.Equal(t, 2, len(meal.MealRecipes))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetMeals(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Setup mock query responses
	mealRows := sqlmock.NewRows([]string{"id", "name", "description", "slug", "image"}).
		AddRow(1, "Meal 1", "Description 1", "meal-1", sql.NullString{String: "image1.jpg", Valid: true}).
		AddRow(2, "Meal 2", "Description 2", "meal-2", sql.NullString{String: "image2.jpg", Valid: true})

	mock.ExpectQuery("SELECT \\* FROM meals").
		WillReturnRows(mealRows)

	// Call the function
	meals, err := GetMeals(sqlxDB)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, meals)
	assert.Equal(t, 2, len(*meals))
	assert.Equal(t, "Meal 1", (*meals)[0].Name)
	assert.Equal(t, "Meal 2", (*meals)[1].Name)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestCreateMeal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Test data
	mealName := "New Test Meal"
	description := "A delicious new test meal"
	image := sql.NullString{String: "new-test-meal.jpg", Valid: true}
	slug := "new-test-meal"

	newMeal := &Meal{
		Name:        mealName,
		Description: description,
		Image:       image,
		Slug:        slug,
		Ingredients: []MealIngredient{
			{Name: "Ing1", Amount: "1 cup"},
			{Name: "Ing2", Amount: "2 tsp"},
		},
		Steps: []MealStep{
			{Order: 1, Text: "Step 1"},
			{Order: 2, Text: "Step 2"},
		},
		MealRecipes: []MealRecipes{
			{RecipeID: 101},
		},
	}

	// First check if slug exists - should return error to continue with create
	mock.ExpectQuery("SELECT id FROM meals WHERE slug=\\$1").
		WithArgs(slug).
		WillReturnError(sql.ErrNoRows)

	// For MustBegin() in CreateMeal
	mock.ExpectBegin()

	// For the INSERT in CreateMeal
	mock.ExpectExec("INSERT INTO meals").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// For GetMealIdFromSlug
	mock.ExpectQuery("SELECT id FROM meals WHERE slug=\\$1").
		WithArgs(slug).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// For UpdateMeal
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE meals SET").
		WithArgs(mealName, description, image, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete old ingredients and insert new
	mock.ExpectExec("DELETE FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	for _, ing := range newMeal.Ingredients {
		mock.ExpectExec("INSERT INTO meal_ingredients").
			WithArgs(1, ing.Name, ing.Amount).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// Delete old steps and insert new
	mock.ExpectExec("DELETE FROM meal_steps WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	for _, step := range newMeal.Steps {
		mock.ExpectExec("INSERT INTO meal_steps").
			WithArgs(1, step.Text, step.Order).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// Delete old recipes and insert new
	mock.ExpectExec("DELETE FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO meal_recipes").
		WithArgs(1, 101).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// For the GetMeal call at the end
	mealRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		1, mealName, description, slug, image,
	)

	mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(mealRows)

	// Mock for ingredients in final GetMeal
	ingredientRows := sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}).
		AddRow(1, 1, "Ing1", "1 cup").
		AddRow(2, 1, "Ing2", "2 tsp")

	mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(ingredientRows)

	// Mock for steps in final GetMeal
	stepsRows := sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}).
		AddRow(1, 1, 1, "Step 1").
		AddRow(2, 1, 2, "Step 2")

	mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1 ORDER BY").
		WithArgs(1).
		WillReturnRows(stepsRows)

	// Mock for recipes in final GetMeal
	recipesRows := sqlmock.NewRows([]string{"meal_id", "recipe_id"}).
		AddRow(1, 101)

	mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(recipesRows)

	// Call the function
	createdMeal, err := CreateMeal(sqlxDB, newMeal)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, createdMeal)
	assert.Equal(t, 1, createdMeal.ID)
	assert.Equal(t, mealName, createdMeal.Name)
	assert.Equal(t, slug, createdMeal.Slug)
	assert.Equal(t, 2, len(createdMeal.Ingredients))
	assert.Equal(t, 2, len(createdMeal.Steps))
	assert.Equal(t, 1, len(createdMeal.MealRecipes))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUpdateMeal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Test data
	mealID := 1
	updatedName := "Updated Test Meal"
	description := "An updated test meal"
	slug := "updated-test-meal"
	image := sql.NullString{String: "updated-test-meal.jpg", Valid: true}

	updateMeal := &Meal{
		ID:          mealID,
		Name:        updatedName,
		Description: description,
		Slug:        slug,
		Image:       image,
		Ingredients: []MealIngredient{
			{Name: "UpdatedIng1", Amount: "2 cups"},
			{Name: "UpdatedIng2", Amount: "3 tbsp"},
		},
		Steps: []MealStep{
			{Order: 1, Text: "Updated Step 1"},
			{Order: 2, Text: "Updated Step 2"},
		},
		MealRecipes: []MealRecipes{
			{RecipeID: 201},
			{RecipeID: 202},
		},
	}

	// Mock the transaction for UpdateMeal
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE meals SET").
		WithArgs(updatedName, description, image, mealID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete old ingredients and insert new
	mock.ExpectExec("DELETE FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	for _, ing := range updateMeal.Ingredients {
		mock.ExpectExec("INSERT INTO meal_ingredients").
			WithArgs(mealID, ing.Name, ing.Amount).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// Delete old steps and insert new
	mock.ExpectExec("DELETE FROM meal_steps WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	for _, step := range updateMeal.Steps {
		mock.ExpectExec("INSERT INTO meal_steps").
			WithArgs(mealID, step.Text, step.Order).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// Delete old recipes and insert new
	mock.ExpectExec("DELETE FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	for _, recipe := range updateMeal.MealRecipes {
		mock.ExpectExec("INSERT INTO meal_recipes").
			WithArgs(mealID, recipe.RecipeID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	// For the final GetMeal call
	mealRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		mealID, updatedName, description, slug, image,
	)

	mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
		WithArgs(mealID).
		WillReturnRows(mealRows)

	// Mock for ingredients in final GetMeal
	ingredientRows := sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}).
		AddRow(3, mealID, "UpdatedIng1", "2 cups").
		AddRow(4, mealID, "UpdatedIng2", "3 tbsp")

	mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnRows(ingredientRows)

	// Mock for steps in final GetMeal
	stepsRows := sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}).
		AddRow(3, mealID, 1, "Updated Step 1").
		AddRow(4, mealID, 2, "Updated Step 2")

	mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1 ORDER BY").
		WithArgs(mealID).
		WillReturnRows(stepsRows)

	// Mock for recipes in final GetMeal
	recipesRows := sqlmock.NewRows([]string{"meal_id", "recipe_id"}).
		AddRow(mealID, 201).
		AddRow(mealID, 202)

	mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnRows(recipesRows)

	// Call the function
	updatedMeal, err := UpdateMeal(sqlxDB, mealID, updateMeal)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedMeal)
	assert.Equal(t, mealID, updatedMeal.ID)
	assert.Equal(t, updatedName, updatedMeal.Name)
	assert.Equal(t, 2, len(updatedMeal.Ingredients))
	assert.Equal(t, 2, len(updatedMeal.Steps))
	assert.Equal(t, 2, len(updatedMeal.MealRecipes))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestDeleteMeal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mealID := 1

	// Mock the transaction for DeleteMeal
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("DELETE FROM meal_steps WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM plan_meals WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM meals WHERE id=\\$1").
		WithArgs(mealID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Call the function
	err = DeleteMeal(sqlxDB, mealID)

	// Assertions
	assert.NoError(t, err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetMealBySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	slug := "test-meal"
	mealID := 1
	name := "Test Meal"
	description := "A delicious test meal"
	image := sql.NullString{String: "test-meal.jpg", Valid: true}

	// First, mock the query for GetMealIdFromSlug
	mock.ExpectQuery("SELECT id FROM meals WHERE slug=\\$1").
		WithArgs(slug).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(mealID))

	// Now mock the queries for GetMeal
	mealRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		mealID, name, description, slug, image,
	)

	mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
		WithArgs(mealID).
		WillReturnRows(mealRows)

	// Mock for ingredients
	ingredientRows := sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}).
		AddRow(1, mealID, "Ingredient1", "1 cup").
		AddRow(2, mealID, "Ingredient2", "2 tbsp")

	mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnRows(ingredientRows)

	// Mock for steps
	stepsRows := sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}).
		AddRow(1, mealID, 1, "Step 1").
		AddRow(2, mealID, 2, "Step 2")

	mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1 ORDER BY").
		WithArgs(mealID).
		WillReturnRows(stepsRows)

	// Mock for recipes
	recipesRows := sqlmock.NewRows([]string{"meal_id", "recipe_id"}).
		AddRow(mealID, 101).
		AddRow(mealID, 102)

	mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(mealID).
		WillReturnRows(recipesRows)

	// Call the function
	id, err := GetMealIdFromSlug(sqlxDB, slug)
	assert.NoError(t, err)
	assert.Equal(t, mealID, id)

	meal, err := GetMeal(sqlxDB, id)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, meal)
	assert.Equal(t, mealID, meal.ID)
	assert.Equal(t, name, meal.Name)
	assert.Equal(t, slug, meal.Slug)
	assert.Equal(t, 2, len(meal.Ingredients))
	assert.Equal(t, 2, len(meal.Steps))
	assert.Equal(t, 2, len(meal.MealRecipes))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
