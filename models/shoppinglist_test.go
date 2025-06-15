package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestGetShoppingList(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	planID := 1
	householdID := 42
	startDate := time.Now()
	endDate := time.Now().Add(7 * 24 * time.Hour)

	t.Run("success_existing_status", func(t *testing.T) {
		// Mock for fetching plan
		planRows := sqlmock.NewRows([]string{"id", "household_id", "start_date", "end_date"}).
			AddRow(planID, householdID, startDate, endDate)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plans WHERE id = $1")).
			WithArgs(planID).
			WillReturnRows(planRows)

		// Mock for fetching existing shopping_status
		statusItems := []StatusItem{{Name: "Flour", Amount: "1kg"}}
		statusJSON, _ := json.Marshal(Status{Items: statusItems})
		statusRows := sqlmock.NewRows([]string{"plan_id", "status"}).
			AddRow(planID, statusJSON)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM shopping_status WHERE plan_id = $1")).
			WithArgs(planID).
			WillReturnRows(statusRows)

		// Mock for GetPlanIngredients
		ingredientRows := sqlmock.NewRows([]string{"name", "amount"}).
			AddRow("Flour", "1kg").
			AddRow("Sugar", "500g")
		mock.ExpectQuery(regexp.QuoteMeta("SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=$1 AND pm.household_id=$2")).
			WithArgs(planID, householdID).
			WillReturnRows(ingredientRows)

		list, err := GetShoppingList(sqlxDB, planID)
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Equal(t, planID, list.Plan.ID)
		assert.Equal(t, householdID, list.Plan.HouseholdID)
		require.Len(t, list.Ingredients, 2)
		assert.Equal(t, "Flour", list.Ingredients[0].Name)
		assert.True(t, list.Ingredients[0].Checked) // Checked because it's in statusItems
		assert.Equal(t, "Sugar", list.Ingredients[1].Name)
		assert.False(t, list.Ingredients[1].Checked)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success_no_existing_status_creates_new", func(t *testing.T) {
		// Mock for fetching plan
		planRows := sqlmock.NewRows([]string{"id", "household_id", "start_date", "end_date"}).
			AddRow(planID, householdID, startDate, endDate)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plans WHERE id = $1")).
			WithArgs(planID).
			WillReturnRows(planRows)

		// Mock for fetching shopping_status (not found)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM shopping_status WHERE plan_id = $1")).
			WithArgs(planID).
			WillReturnError(sql.ErrNoRows)

		// Mock for inserting new shopping_status
		emptyStatusJSON, _ := json.Marshal(Status{Items: []StatusItem{}})
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO shopping_status (plan_id, status) VALUES ($1, $2)")).
			WithArgs(planID, emptyStatusJSON).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock for GetPlanIngredients
		ingredientRows := sqlmock.NewRows([]string{"name", "amount"}).
			AddRow("Milk", "1L")
		mock.ExpectQuery(regexp.QuoteMeta("SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=$1 AND pm.household_id=$2")).
			WithArgs(planID, householdID).
			WillReturnRows(ingredientRows)

		list, err := GetShoppingList(sqlxDB, planID)
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Equal(t, planID, list.Plan.ID)
		require.Len(t, list.Ingredients, 1)
		assert.Equal(t, "Milk", list.Ingredients[0].Name)
		assert.False(t, list.Ingredients[0].Checked) // Not checked as status was new and empty

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error_fetching_plan", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plans WHERE id = $1")).
			WithArgs(planID).
			WillReturnError(errors.New("db error fetching plan"))

		list, err := GetShoppingList(sqlxDB, planID)
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "failed to fetch plan")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error_fetching_ingredients", func(t *testing.T) {
		planRows := sqlmock.NewRows([]string{"id", "household_id", "start_date", "end_date"}).
			AddRow(planID, householdID, startDate, endDate)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plans WHERE id = $1")).
			WithArgs(planID).
			WillReturnRows(planRows)

		// Mock for fetching shopping_status (not found, will try to create)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM shopping_status WHERE plan_id = $1")).
			WithArgs(planID).
			WillReturnError(sql.ErrNoRows)
		emptyStatusJSON, _ := json.Marshal(Status{Items: []StatusItem{}})
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO shopping_status (plan_id, status) VALUES ($1, $2)")).
			WithArgs(planID, emptyStatusJSON).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=$1 AND pm.household_id=$2")).
			WithArgs(planID, householdID).
			WillReturnError(errors.New("db error fetching ingredients"))

		list, err := GetShoppingList(sqlxDB, planID)
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "db error fetching ingredients") // Error from GetPlanIngredients
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateShoppingList(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	householdID := 42
	planID := 2
	listToUpdate := &ShoppingList{
		Plan: Plan{ID: planID, HouseholdID: householdID},
		Ingredients: []ShoppingListItem{
			{Name: "Eggs", Amount: "12", Checked: true},
			{Name: "Bacon", Amount: "500g", Checked: false},
			{Name: "Bread", Amount: "1 loaf", Checked: true},
		},
	}

	t.Run("success", func(t *testing.T) {
		// Mock for plan validation
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM plans WHERE id = $1 AND household_id = $2")).
			WithArgs(planID, householdID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))

		// Mock for DB update
		expectedStatusItems := []StatusItem{
			{Name: "Eggs", Amount: "12"},
			{Name: "Bread", Amount: "1 loaf"},
		}
		expectedStatusJSON, _ := json.Marshal(Status{Items: expectedStatusItems})
		mock.ExpectExec(regexp.QuoteMeta("UPDATE shopping_status SET status = $1 WHERE plan_id = $2")).
			WithArgs(expectedStatusJSON, planID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := UpdateShoppingList(sqlxDB, householdID, listToUpdate)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error_invalid_plan_id", func(t *testing.T) {
		invalidList := &ShoppingList{Plan: Plan{ID: 0}} // Invalid Plan ID
		err := UpdateShoppingList(sqlxDB, householdID, invalidList)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid plan ID")
		// No DB calls expected
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error_plan_not_found_or_unauthorized", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM plans WHERE id = $1 AND household_id = $2")).
			WithArgs(planID, householdID).
			WillReturnError(sql.ErrNoRows) // Simulate plan not found or not matching user

		err := UpdateShoppingList(sqlxDB, householdID, listToUpdate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan not found or unauthorized")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error_db_update_fails", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM plans WHERE id = $1 AND household_id = $2")).
			WithArgs(planID, householdID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))

		expectedStatusItems := []StatusItem{
			{Name: "Eggs", Amount: "12"},
			{Name: "Bread", Amount: "1 loaf"},
		}
		expectedStatusJSON, _ := json.Marshal(Status{Items: expectedStatusItems})
		mock.ExpectExec(regexp.QuoteMeta("UPDATE shopping_status SET status = $1 WHERE plan_id = $2")).
			WithArgs(expectedStatusJSON, planID).
			WillReturnError(errors.New("db update failed"))

		err := UpdateShoppingList(sqlxDB, householdID, listToUpdate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db update failed")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestStatus_Scan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := &Status{}
		items := []StatusItem{{Name: "Test", Amount: "1"}}
		jsonBytes, _ := json.Marshal(Status{Items: items})
		err := s.Scan(jsonBytes)
		require.NoError(t, err)
		assert.Equal(t, items, s.Items)
	})

	t.Run("failure_type_assertion", func(t *testing.T) {
		s := &Status{}
		err := s.Scan("not a byte slice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type assertion to []byte failed")
	})

	t.Run("failure_json_unmarshal", func(t *testing.T) {
		s := &Status{}
		err := s.Scan([]byte("invalid json"))
		assert.Error(t, err)
	})
}

func TestStatus_Value(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		items := []StatusItem{{Name: "Test", Amount: "1"}}
		s := Status{Items: items}
		expectedJSON, _ := json.Marshal(s)

		val, err := s.Value()
		require.NoError(t, err)
		assert.Equal(t, expectedJSON, val) // Changed from string(expectedJSON) to expectedJSON
	})

	// Test for json.Marshal error is hard to simulate without complex types
	// or invalid struct tags that might not be caught by compiler.
	// For now, we assume standard library json.Marshal works as expected for this struct.
}
