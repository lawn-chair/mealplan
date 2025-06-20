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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDbCtx is a helper to create a context with a mock DB
func mockDbCtx(next http.Handler, db *sqlx.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TestGetShoppingList(t *testing.T) {
	const householdID = 42
	const userID = "test-user-id"
	const planID = 1
	startDate := time.Now().AddDate(0, 0, +1)
	endDate := time.Now().AddDate(0, 0, +8)

	// Common plan object
	expectedPlan := models.Plan{
		ID:          planID,
		HouseholdID: householdID,
		StartDate:   models.Date{Time: startDate},
		EndDate:     models.Date{Time: endDate},
	}

	t.Run("success - no existing shopping status", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "sqlmock")

		// Mock GetNextPlan
		planRows := sqlmock.NewRows([]string{"id", "household_id", "start_date", "end_date"}).
			AddRow(planID, householdID, startDate, endDate)
		mock.ExpectQuery(`SELECT \* FROM plans WHERE household_id`).
			WithArgs(householdID).
			WillReturnRows(planRows)

		// Mock GetPantry
		pantryRows := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, householdID)
		mock.ExpectQuery(`SELECT \* FROM pantry WHERE household_id = \$1`).WithArgs(householdID).WillReturnRows(pantryRows)
		pantryItemsRows := sqlmock.NewRows([]string{"item_name"}).AddRow("sugar") // Pantry contains "sugar"
		mock.ExpectQuery(`SELECT item_name FROM pantry_items WHERE pantry_id = \$1`).WithArgs(1).WillReturnRows(pantryItemsRows)

		// Mocks for models.GetShoppingList
		// 1. Internal plan fetch
		internalPlanRows := sqlmock.NewRows([]string{"id", "household_id", "start_date", "end_date"}).
			AddRow(planID, householdID, startDate, endDate)
		mock.ExpectQuery(`SELECT \* FROM plans WHERE id = \$1`).
			WithArgs(planID).
			WillReturnRows(internalPlanRows)

		// 2. Shopping status fetch (not found)
		mock.ExpectQuery(`SELECT \* FROM shopping_status WHERE plan_id = \$1`).
			WithArgs(planID).
			WillReturnError(sql.ErrNoRows)

		// 3. Insertion of new empty shopping status
		emptyStatus := models.Status{Items: []models.StatusItem{}}
		emptyStatusJSON, _ := json.Marshal(emptyStatus)
		mock.ExpectExec(`INSERT INTO shopping_status \(plan_id, status\) VALUES`).
			WithArgs(planID, emptyStatusJSON).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 4. GetPlanIngredients
		ingredientRows := sqlmock.NewRows([]string{"name", "amount"}).
			AddRow("Flour", "2 cups").
			AddRow("Sugar", "1 cup"). // Will be filtered by pantry
			AddRow("Eggs", "2")
		mock.ExpectQuery(`SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=\$1 AND pm.household_id=\$2`).
			WithArgs(planID, householdID).
			WillReturnRows(ingredientRows)

		// Setup router and request
		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return mockDbCtx(next, sqlxDB)
		})

		r.Get("/shopping-list", GetShoppingList)

		req := httptest.NewRequest("GET", "/shopping-list", nil)
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", &clerk.User{ID: userID})
		ctx = context.WithValue(ctx, "household", householdID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		fmt.Println(rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)

		var respBody models.ShoppingList
		err = json.Unmarshal(rr.Body.Bytes(), &respBody)
		require.NoError(t, err)

		assert.Equal(t, expectedPlan.ID, respBody.Plan.ID)
		assert.Equal(t, expectedPlan.HouseholdID, respBody.Plan.HouseholdID)

		require.Len(t, respBody.Ingredients, 2) // Flour, Eggs (Sugar filtered)
		assert.Equal(t, "Flour", respBody.Ingredients[0].Name)
		assert.False(t, respBody.Ingredients[0].Checked)
		assert.Equal(t, "Eggs", respBody.Ingredients[1].Name)
		assert.False(t, respBody.Ingredients[1].Checked)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - with existing shopping status", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "sqlmock")

		// Mock GetNextPlan
		planRows := sqlmock.NewRows([]string{"id", "household_id", "start_date", "end_date"}).
			AddRow(planID, householdID, startDate, endDate)
		mock.ExpectQuery(`SELECT \* FROM plans WHERE household_id`).
			WithArgs(householdID).
			WillReturnRows(planRows)

		// Mock GetPantry (empty for this test to simplify)
		pantryRows := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, householdID)
		mock.ExpectQuery(`SELECT \* FROM pantry WHERE household_id = \$1`).WithArgs(householdID).WillReturnRows(pantryRows)
		mock.ExpectQuery(`SELECT item_name FROM pantry_items WHERE pantry_id = \$1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"item_name"}))

		// Mocks for models.GetShoppingList
		// 1. Internal plan fetch
		internalPlanRows := sqlmock.NewRows([]string{"id", "household_id", "start_date", "end_date"}).
			AddRow(planID, householdID, startDate, endDate)
		mock.ExpectQuery(`SELECT \* FROM plans WHERE id = \$1`).
			WithArgs(planID).
			WillReturnRows(internalPlanRows)

		// 2. Shopping status fetch (found)
		existingStatus := models.Status{
			Items: []models.StatusItem{
				{Name: "Flour", Amount: "2 cups"},
			},
		}
		existingStatusJSON, _ := json.Marshal(existingStatus)
		statusRows := sqlmock.NewRows([]string{"plan_id", "status"}).AddRow(planID, existingStatusJSON)
		mock.ExpectQuery(`SELECT \* FROM shopping_status WHERE plan_id = \$1`).
			WithArgs(planID).
			WillReturnRows(statusRows)

		// 3. GetPlanIngredients
		ingredientRows := sqlmock.NewRows([]string{"name", "amount"}).
			AddRow("Flour", "2 cups").
			AddRow("Eggs", "2")
		mock.ExpectQuery(`SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=\$1 AND pm.household_id=\$2`).
			WithArgs(planID, householdID).
			WillReturnRows(ingredientRows)

		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return mockDbCtx(next, sqlxDB)
		})
		r.Get("/shopping-list", GetShoppingList)

		req := httptest.NewRequest("GET", "/shopping-list", nil)
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", &clerk.User{ID: userID})
		ctx = context.WithValue(ctx, "household", householdID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody models.ShoppingList
		err = json.Unmarshal(rr.Body.Bytes(), &respBody)
		require.NoError(t, err)

		assert.Equal(t, expectedPlan.ID, respBody.Plan.ID)
		require.Len(t, respBody.Ingredients, 2)
		assert.Equal(t, "Flour", respBody.Ingredients[0].Name)
		assert.True(t, respBody.Ingredients[0].Checked) // Checked from existing status
		assert.Equal(t, "Eggs", respBody.Ingredients[1].Name)
		assert.False(t, respBody.Ingredients[1].Checked)

		require.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestUpdateShoppingList(t *testing.T) {
	const householdID = 42
	const planID = 1
	const userID = "test-user-id"
	startDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 6, 23, 0, 0, 0, 0, time.UTC)

	payload := models.ShoppingList{
		Plan: models.Plan{ID: planID, HouseholdID: householdID, StartDate: models.Date{Time: startDate}, EndDate: models.Date{Time: endDate}},
		Ingredients: []models.ShoppingListItem{
			{Name: "Flour", Amount: "2 cups", Checked: true},
			{Name: "Eggs", Amount: "2", Checked: false},
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "sqlmock")

		// Mock for plan validation in models.UpdateShoppingList
		mock.ExpectQuery(`SELECT id FROM plans WHERE id = \$1 AND household_id = \$2`).
			WithArgs(planID, householdID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))

		// Mock for DB update in models.UpdateShoppingList
		expectedStatus := models.Status{
			Items: []models.StatusItem{{Name: "Flour", Amount: "2 cups"}},
		}

		mock.ExpectExec(`UPDATE shopping_status SET status = \$1 WHERE plan_id = \$2`).
			WithArgs(expectedStatus, planID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return mockDbCtx(next, sqlxDB)
		})

		r.Put("/shopping-list", UpdateShoppingList)

		req := httptest.NewRequest("PUT", "/shopping-list", bytes.NewBuffer(payloadBytes))
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", &clerk.User{ID: userID})
		ctx = context.WithValue(ctx, "household", householdID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var respBody models.ShoppingList
		err = json.Unmarshal(rr.Body.Bytes(), &respBody)
		require.NoError(t, err)
		assert.Equal(t, planID, respBody.Plan.ID)
		require.Len(t, respBody.Ingredients, 2)
		assert.True(t, respBody.Ingredients[0].Checked)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error on update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "sqlmock")

		// Add plan validation query mock
		mock.ExpectQuery(`SELECT id FROM plans WHERE id = \$1 AND household_id = \$2`).
			WithArgs(planID, householdID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))

		mock.ExpectExec(`UPDATE shopping_status SET status = \$1 WHERE plan_id = \$2`).
			WillReturnError(sql.ErrConnDone)

		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return mockDbCtx(next, sqlxDB)
		})
		r.Put("/shopping-list", UpdateShoppingList)

		req := httptest.NewRequest("PUT", "/shopping-list", bytes.NewBuffer(payloadBytes))
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", &clerk.User{ID: userID})
		ctx = context.WithValue(ctx, "household", householdID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		// Further check error response if needed
		require.NoError(t, mock.ExpectationsWereMet())
	})

}

// Note: The original RequiresAuthentication function is not directly tested here,
// but its behavior is mocked. If it were part of the api package and exported,
// it could be unit tested separately.
// Similarly, the DbCtx and IdCtx from mealplan.go are used implicitly by the router.
// For a production system, these might also have their own focused tests.
