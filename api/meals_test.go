package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestGetMealsHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	t.Run("Get all meals", func(t *testing.T) {
		// Set up mock expectations
		rows := sqlmock.NewRows([]string{"id", "name", "description", "slug", "image"}).
			AddRow(1, "Test Meal 1", "Description 1", "test-meal-1", nil).
			AddRow(2, "Test Meal 2", "Description 2", "test-meal-2", nil)

		mock.ExpectQuery("SELECT \\* FROM meals").
			WillReturnRows(rows)

		// Create request
		req := httptest.NewRequest("GET", "/api/meals", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		req = req.WithContext(ctx)

		// Call the handler
		GetMealsHandler(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response body
		var response []models.Meal
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response data
		assert.Len(t, response, 2)
		assert.Equal(t, "Test Meal 1", response[0].Name)
		assert.Equal(t, "Test Meal 2", response[1].Name)
	})

	t.Run("Get meal by slug", func(t *testing.T) {
		// Mock for GetMealIdFromSlug
		mock.ExpectQuery("SELECT id FROM meals WHERE slug=\\$1").
			WithArgs("test-meal-slug").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Mock for GetMeal
		mealRow := sqlmock.NewRows([]string{
			"id", "name", "description", "slug", "image",
		}).AddRow(
			1, "Test Meal", "Description", "test-meal-slug", nil,
		)

		mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
			WithArgs(1).
			WillReturnRows(mealRow)

		// Mock for ingredients, steps, and recipes would go here in a complete test
		mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}))

		mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}))

		mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"meal_id", "recipe_id"}))

		// Create request with slug query parameter
		req := httptest.NewRequest("GET", "/api/meals?slug=test-meal-slug", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		req = req.WithContext(ctx)

		// Call the handler
		GetMealsHandler(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response body
		var meal models.Meal
		err := json.Unmarshal(rec.Body.Bytes(), &meal)
		require.NoError(t, err)

		// Verify response data
		assert.Equal(t, 1, meal.ID)
		assert.Equal(t, "Test Meal", meal.Name)
		assert.Equal(t, "test-meal-slug", meal.Slug)
	})

	t.Run("Error when slug not found", func(t *testing.T) {
		// Mock for GetMealIdFromSlug - returns error
		mock.ExpectQuery("SELECT id FROM meals WHERE slug=\\$1").
			WithArgs("nonexistent-slug").
			WillReturnError(fmt.Errorf("not found"))

		// Create request with non-existent slug
		req := httptest.NewRequest("GET", "/api/meals?slug=nonexistent-slug", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		req = req.WithContext(ctx)

		// Call the handler
		GetMealsHandler(rec, req)

		// Verify response status code
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestGetMealHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Mock for GetMeal
	mealRow := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		1, "Test Meal", "Description", "test-meal-slug", nil,
	)

	mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(mealRow)

	// Mock for ingredients, steps, and recipes
	mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}))

	mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}))

	mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"meal_id", "recipe_id"}))

	// Create request
	req := httptest.NewRequest("GET", "/api/meals/1", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB and ID
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	req = req.WithContext(ctx)

	// Call the handler
	GetMealHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var meal models.Meal
	err := json.Unmarshal(rec.Body.Bytes(), &meal)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, meal.ID)
	assert.Equal(t, "Test Meal", meal.Name)
	assert.Equal(t, "test-meal-slug", meal.Slug)
}

func TestCreateMealHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: "test-user-id"}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Create a test meal
	newMeal := models.Meal{
		Name:        "New Test Meal",
		Description: "New Description",
		Slug:        "new-test-meal",
		Ingredients: []models.MealIngredient{
			{Name: "Ingredient 1", Amount: "1 cup"},
		},
		Steps: []models.MealStep{
			{Order: 1, Text: "Step 1"},
		},
		MealRecipes: []models.MealRecipes{
			{RecipeID: 101},
		},
	}

	// Mock for slug check and insert
	mock.ExpectQuery("SELECT id FROM meals WHERE slug=\\$1").
		WithArgs("new-test-meal").
		WillReturnError(fmt.Errorf("not found"))

	// Mock transaction for CreateMeal
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO meals").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Mock GetMealIdFromSlug
	mock.ExpectQuery("SELECT id FROM meals WHERE slug=\\$1").
		WithArgs("new-test-meal").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Mock for UpdateMeal
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE meals SET").
		WithArgs(newMeal.Name, newMeal.Description, newMeal.Image, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for ingredients
	mock.ExpectExec("DELETE FROM meal_ingredients").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO meal_ingredients").
		WithArgs(1, "Ingredient 1", "1 cup").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for steps
	mock.ExpectExec("DELETE FROM meal_steps").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO meal_steps").
		WithArgs(1, "Step 1", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for recipes
	mock.ExpectExec("DELETE FROM meal_recipes").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO meal_recipes").
		WithArgs(1, 101).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// For the final GetMeal call
	mealRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		1, newMeal.Name, newMeal.Description, newMeal.Slug, newMeal.Image,
	)

	mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(mealRows)

	// Mock for ingredients in final GetMeal
	ingredientRows := sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}).
		AddRow(1, 1, "Ingredient 1", "1 cup")

	mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(ingredientRows)

	// Mock for steps in final GetMeal
	stepsRows := sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}).
		AddRow(1, 1, 1, "Step 1")

	mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1 ORDER BY").
		WithArgs(1).
		WillReturnRows(stepsRows)

	// Mock for recipes in final GetMeal
	recipesRows := sqlmock.NewRows([]string{"meal_id", "recipe_id"}).
		AddRow(1, 101)

	mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(recipesRows)

	// Create request with meal data
	mealJSON, err := json.Marshal(newMeal)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/meals", bytes.NewBuffer(mealJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	req = req.WithContext(ctx)

	// Call the handler
	CreateMealHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var createdMeal models.Meal
	err = json.Unmarshal(rec.Body.Bytes(), &createdMeal)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, createdMeal.ID)
	assert.Equal(t, newMeal.Name, createdMeal.Name)
	assert.Equal(t, newMeal.Slug, createdMeal.Slug)
	assert.Equal(t, 1, len(createdMeal.Ingredients))
	assert.Equal(t, 1, len(createdMeal.Steps))
	assert.Equal(t, 1, len(createdMeal.MealRecipes))
}

func TestUpdateMealHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: "test-user-id"}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Create a test meal for update
	updateMeal := models.Meal{
		ID:          1,
		Name:        "Updated Test Meal",
		Description: "Updated Description",
		Slug:        "updated-test-meal",
		Ingredients: []models.MealIngredient{
			{Name: "Updated Ingredient", Amount: "2 cups"},
		},
		Steps: []models.MealStep{
			{Order: 1, Text: "Updated Step"},
		},
		MealRecipes: []models.MealRecipes{
			{RecipeID: 202},
		},
	}

	// Mock for UpdateMeal
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE meals SET").
		WithArgs(updateMeal.Name, updateMeal.Description, updateMeal.Image, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for ingredients
	mock.ExpectExec("DELETE FROM meal_ingredients").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO meal_ingredients").
		WithArgs(1, "Updated Ingredient", "2 cups").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for steps
	mock.ExpectExec("DELETE FROM meal_steps").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO meal_steps").
		WithArgs(1, "Updated Step", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for recipes
	mock.ExpectExec("DELETE FROM meal_recipes").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO meal_recipes").
		WithArgs(1, 202).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// For the final GetMeal call
	mealRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		1, updateMeal.Name, updateMeal.Description, updateMeal.Slug, updateMeal.Image,
	)

	mock.ExpectQuery("SELECT \\* FROM meals WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(mealRows)

	// Mock for ingredients in final GetMeal
	ingredientRows := sqlmock.NewRows([]string{"id", "meal_id", "name", "amount"}).
		AddRow(1, 1, "Updated Ingredient", "2 cups")

	mock.ExpectQuery("SELECT \\* FROM meal_ingredients WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(ingredientRows)

	// Mock for steps in final GetMeal
	stepsRows := sqlmock.NewRows([]string{"id", "meal_id", "order", "text"}).
		AddRow(1, 1, 1, "Updated Step")

	mock.ExpectQuery("SELECT \\* FROM meal_steps WHERE meal_id=\\$1 ORDER BY").
		WithArgs(1).
		WillReturnRows(stepsRows)

	// Mock for recipes in final GetMeal
	recipesRows := sqlmock.NewRows([]string{"meal_id", "recipe_id"}).
		AddRow(1, 202)

	mock.ExpectQuery("SELECT \\* FROM meal_recipes WHERE meal_id=\\$1").
		WithArgs(1).
		WillReturnRows(recipesRows)

	// Create request with meal data
	mealJSON, err := json.Marshal(updateMeal)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/meals/1", bytes.NewBuffer(mealJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB and ID
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	req = req.WithContext(ctx)

	// Call the handler
	UpdateMealHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var updatedMeal models.Meal
	err = json.Unmarshal(rec.Body.Bytes(), &updatedMeal)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, updatedMeal.ID)
	assert.Equal(t, updateMeal.Name, updatedMeal.Name)
	assert.Equal(t, updateMeal.Slug, updatedMeal.Slug)
	assert.Equal(t, 1, len(updatedMeal.Ingredients))
	assert.Equal(t, "Updated Ingredient", updatedMeal.Ingredients[0].Name)
}

func TestDeleteMealHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: "test-user-id"}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Mock for DeleteMeal
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM meal_ingredients").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM meal_steps").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM meal_recipes").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM plan_meals").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM meals").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Create request
	req := httptest.NewRequest("DELETE", "/api/meals/1", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB and ID
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	req = req.WithContext(ctx)

	// Call the handler
	DeleteMealHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
