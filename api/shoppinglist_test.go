package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDBForShoppingList(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestGetShoppingList(t *testing.T) {
	sqlxDB, mock := setupMockDBForShoppingList(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: "test-user-id"}, nil
	}
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Mock for GetNextPlan - use actual time.Time objects
	startDate := time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 6, 13, 0, 0, 0, 0, time.UTC)
	planRows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "user_id"}).AddRow(1, startDate, endDate, "test-user-id")
	mock.ExpectQuery("SELECT \\* FROM plans WHERE user_id").WithArgs("test-user-id").WillReturnRows(planRows)

	// Mock for GetPlanIngredients
	ingredientRows := sqlmock.NewRows([]string{"name", "amount"}).
		AddRow("Flour", "2 cups").
		AddRow("Sugar", "1 cup").
		AddRow("Eggs", "2")
	mock.ExpectQuery(`SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=\$1`).WithArgs(1).WillReturnRows(ingredientRows)

	// Mock for GetPantry
	pantryRows := sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, "test-user-id")
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE user_id").WithArgs("test-user-id").WillReturnRows(pantryRows)
	pantryItemsRows := sqlmock.NewRows([]string{"item_name"}).AddRow("sugar")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").WithArgs(1).WillReturnRows(pantryItemsRows)

	// Create request
	req := httptest.NewRequest("GET", "/api/shopping-list", nil)
	rec := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	req = req.WithContext(ctx)

	// Call the handler
	GetShoppingList(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var list models.ShoppingList
	err := json.Unmarshal(rec.Body.Bytes(), &list)
	require.NoError(t, err)

	assert.Equal(t, 1, list.Plan.ID)
	assert.Equal(t, 2, len(list.Ingredients)) // "Sugar" should be filtered out by pantry
	assert.Equal(t, "Flour", list.Ingredients[0].Name)
	assert.Equal(t, "Eggs", list.Ingredients[1].Name)
}
