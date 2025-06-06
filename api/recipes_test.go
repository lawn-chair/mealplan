package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/lawn-chair/mealplan/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecipes(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	t.Run("Get all recipes", func(t *testing.T) {
		// Set up mock expectations
		rows := sqlmock.NewRows([]string{"id", "name", "description", "slug", "image"}).
			AddRow(1, "Test Recipe 1", "Description 1", "test-recipe-1", nil).
			AddRow(2, "Test Recipe 2", "Description 2", "test-recipe-2", nil)

		mock.ExpectQuery("SELECT \\* FROM recipes").
			WillReturnRows(rows)

		// Create request
		req := httptest.NewRequest("GET", "/api/recipes", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		req = req.WithContext(ctx)

		// Call the handler
		GetRecipes(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response body
		var response []models.Recipe
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response data
		assert.Len(t, response, 2)
		assert.Equal(t, "Test Recipe 1", response[0].Name)
		assert.Equal(t, "Test Recipe 2", response[1].Name)
	})

	t.Run("Get recipe by slug", func(t *testing.T) {
		// Mock for GetRecipeIdFromSlug
		mock.ExpectQuery("SELECT id FROM recipes WHERE slug=\\$1").
			WithArgs("test-recipe-slug").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Mock for GetRecipe
		recipeRow := sqlmock.NewRows([]string{
			"id", "name", "description", "slug", "image",
		}).AddRow(
			1, "Test Recipe", "Description", "test-recipe-slug", nil,
		)

		mock.ExpectQuery("SELECT \\* FROM recipes WHERE id=\\$1").
			WithArgs(1).
			WillReturnRows(recipeRow)

		// Mock for ingredients and steps
		mock.ExpectQuery("SELECT \\* FROM recipe_ingredients WHERE recipe_id=\\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "name", "amount"}))

		mock.ExpectQuery("SELECT \\* FROM recipe_steps WHERE recipe_id=\\$1 ORDER BY").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "order", "text"}))

		// Create request with slug query parameter
		req := httptest.NewRequest("GET", "/api/recipes?slug=test-recipe-slug", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		req = req.WithContext(ctx)

		// Call the handler
		GetRecipes(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response body
		var recipe models.Recipe
		err := json.Unmarshal(rec.Body.Bytes(), &recipe)
		require.NoError(t, err)

		// Verify response data
		assert.Equal(t, 1, recipe.ID)
		assert.Equal(t, "Test Recipe", recipe.Name)
		assert.Equal(t, "test-recipe-slug", recipe.Slug)
	})

	t.Run("Error when slug not found", func(t *testing.T) {
		// Mock for GetRecipeIdFromSlug - returns error
		mock.ExpectQuery("SELECT id FROM recipes WHERE slug=\\$1").
			WithArgs("nonexistent-slug").
			WillReturnError(fmt.Errorf("not found"))

		// Create request with non-existent slug
		req := httptest.NewRequest("GET", "/api/recipes?slug=nonexistent-slug", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		req = req.WithContext(ctx)

		// Call the handler
		GetRecipes(rec, req)

		// Verify response status code
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestGetRecipe(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Mock for GetRecipe
	recipeRow := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		1, "Test Recipe", "Description", "test-recipe-slug", nil,
	)

	mock.ExpectQuery("SELECT \\* FROM recipes WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(recipeRow)

	// Mock for ingredients and steps
	mock.ExpectQuery("SELECT \\* FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "name", "amount"}))

	mock.ExpectQuery("SELECT \\* FROM recipe_steps WHERE recipe_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "order", "text"}))

	// Create request
	req := httptest.NewRequest("GET", "/api/recipes/1", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB and ID
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	req = req.WithContext(ctx)

	// Call the handler
	GetRecipe(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var recipe models.Recipe
	err := json.Unmarshal(rec.Body.Bytes(), &recipe)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, recipe.ID)
	assert.Equal(t, "Test Recipe", recipe.Name)
	assert.Equal(t, "test-recipe-slug", recipe.Slug)
}

func TestCreateRecipe(t *testing.T) {
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

	// Create a test recipe
	newRecipe := models.Recipe{
		Name:        "New Test Recipe",
		Description: "New Description",
		Ingredients: []models.RecipeIngredient{
			{Name: "Ingredient 1", Amount: "1 cup"},
		},
		Steps: []models.RecipeStep{
			{Order: 1, Text: "Step 1"},
		},
	}

	// Mock for slug check - recipe doesn't exist yet
	mock.ExpectQuery("SELECT id FROM recipes WHERE slug=\\$1").
		WithArgs("new-test-recipe").
		WillReturnError(fmt.Errorf("no rows in result set"))

	// Mock transaction for CreateRecipe
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO recipes").
		WithArgs("New Test Recipe", "New Description", "new-test-recipe").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock getting the ID after insert
	mock.ExpectQuery("SELECT id FROM recipes WHERE slug=\\$1").
		WithArgs("new-test-recipe").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	// Mock for UpdateRecipe (called by CreateRecipe)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes SET").
		WithArgs(newRecipe.Name, newRecipe.Description, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for ingredients
	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO recipe_ingredients").
		WithArgs(1, "Ingredient 1", "1 cup", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for steps
	mock.ExpectExec("DELETE FROM recipe_steps").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO recipe_steps").
		WithArgs(1, 1, "Step 1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// For the final GetRecipe call
	recipeRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		1, newRecipe.Name, newRecipe.Description, "new-test-recipe", sql.NullString{},
	)

	mock.ExpectQuery("SELECT \\* FROM recipes WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(recipeRows)

	// Mock for ingredients in final GetRecipe
	ingredientRows := sqlmock.NewRows([]string{"id", "recipe_id", "name", "amount", "calories"}).
		AddRow(1, 1, "Ingredient 1", "1 cup", nil)

	mock.ExpectQuery("SELECT \\* FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(1).
		WillReturnRows(ingredientRows)

	// Mock for steps in final GetRecipe
	stepsRows := sqlmock.NewRows([]string{"id", "recipe_id", "order", "text"}).
		AddRow(1, 1, 1, "Step 1")

	mock.ExpectQuery("SELECT \\* FROM recipe_steps WHERE recipe_id=\\$1 ORDER BY").
		WithArgs(1).
		WillReturnRows(stepsRows)

	// Create request with recipe data
	recipeJSON, err := json.Marshal(newRecipe)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/recipes", bytes.NewBuffer(recipeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	req = req.WithContext(ctx)

	// Call the handler
	CreateRecipe(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var createdRecipe models.Recipe
	err = json.Unmarshal(rec.Body.Bytes(), &createdRecipe)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, createdRecipe.ID)
	assert.Equal(t, newRecipe.Name, createdRecipe.Name)
	assert.Equal(t, "new-test-recipe", createdRecipe.Slug)
	assert.Equal(t, 1, len(createdRecipe.Ingredients))
	assert.Equal(t, 1, len(createdRecipe.Steps))
}

func TestUpdateRecipe(t *testing.T) {
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

	// Create a test recipe for update
	updateRecipe := models.Recipe{
		ID:          1,
		Name:        "Updated Test Recipe",
		Description: "Updated Description",
		Slug:        "updated-test-recipe",
		Ingredients: []models.RecipeIngredient{
			{Name: "Updated Ingredient", Amount: "2 cups"},
		},
		Steps: []models.RecipeStep{
			{Order: 1, Text: "Updated Step"},
		},
	}

	// Mock for UpdateRecipe
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes SET").
		WithArgs(updateRecipe.Name, updateRecipe.Description, updateRecipe.Image, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for ingredients
	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO recipe_ingredients").
		WithArgs(1, "Updated Ingredient", "2 cups", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock DELETE and INSERT for steps
	mock.ExpectExec("DELETE FROM recipe_steps").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO recipe_steps").
		WithArgs(1, 1, "Updated Step").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// For the final GetRecipe call
	recipeRows := sqlmock.NewRows([]string{
		"id", "name", "description", "slug", "image",
	}).AddRow(
		1, updateRecipe.Name, updateRecipe.Description, updateRecipe.Slug, sql.NullString{},
	)

	mock.ExpectQuery("SELECT \\* FROM recipes WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(recipeRows)

	// Mock for ingredients in final GetRecipe
	ingredientRows := sqlmock.NewRows([]string{"id", "recipe_id", "name", "amount", "calories"}).
		AddRow(1, 1, "Updated Ingredient", "2 cups", nil)

	mock.ExpectQuery("SELECT \\* FROM recipe_ingredients WHERE recipe_id=\\$1").
		WithArgs(1).
		WillReturnRows(ingredientRows)

	// Mock for steps in final GetRecipe
	stepsRows := sqlmock.NewRows([]string{"id", "recipe_id", "order", "text"}).
		AddRow(1, 1, 1, "Updated Step")

	mock.ExpectQuery("SELECT \\* FROM recipe_steps WHERE recipe_id=\\$1 ORDER BY").
		WithArgs(1).
		WillReturnRows(stepsRows)

	// Create request with recipe data
	recipeJSON, err := json.Marshal(updateRecipe)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/recipes/1", bytes.NewBuffer(recipeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB and ID
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	req = req.WithContext(ctx)

	// Call the handler
	UpdateRecipe(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var updatedRecipe models.Recipe
	err = json.Unmarshal(rec.Body.Bytes(), &updatedRecipe)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, updatedRecipe.ID)
	assert.Equal(t, updateRecipe.Name, updatedRecipe.Name)
	assert.Equal(t, updateRecipe.Slug, updatedRecipe.Slug)
	assert.Equal(t, 1, len(updatedRecipe.Ingredients))
	assert.Equal(t, "Updated Ingredient", updatedRecipe.Ingredients[0].Name)
}

func TestDeleteRecipe(t *testing.T) {
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

	// Mock for DeleteRecipe
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM recipe_steps").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("DELETE FROM meal_recipes").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM recipes").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Create request
	req := httptest.NewRequest("DELETE", "/api/recipes/1", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB and ID
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	req = req.WithContext(ctx)

	// Call the handler
	DeleteRecipe(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
