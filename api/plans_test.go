package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/lawn-chair/mealplan/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPlans(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	t.Run("Get all plans", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, 1)
		endDate := time.Now().AddDate(0, 0, 7)
		testUserID := "user1"
		householdID := 42
		mockUser := &clerk.User{ID: testUserID}

		// Household lookup expectation
		rows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
			AddRow(1, startDate, endDate, 42).
			AddRow(2, startDate.AddDate(0, 0, 7), endDate.AddDate(0, 0, 7), 43)

		mock.ExpectQuery("SELECT \\* FROM plans WHERE household_id").
			WithArgs(42).
			WillReturnRows(rows)

		req := httptest.NewRequest("GET", "/api/plans", nil)
		rec := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", mockUser)
		ctx = context.WithValue(ctx, "household", householdID)

		req = req.WithContext(ctx)

		GetPlans(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []models.Plan
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response, 2)
		assert.Equal(t, 42, response[0].HouseholdID)
		assert.Equal(t, 43, response[1].HouseholdID)
	})

	t.Run("Get last plan", func(t *testing.T) {
		// Setup mock authentication
		testUserID := "user1"
		householdID := 42
		mockUser := &clerk.User{ID: testUserID}

		// Set up mock expectations for GetLastPlan
		startDate := time.Now().AddDate(0, 0, -7)
		endDate := time.Now().AddDate(0, 0, -1)
		rows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
			AddRow(1, startDate, endDate, householdID)

		mock.ExpectQuery("SELECT \\* FROM plans WHERE household_id").
			WithArgs(householdID).
			WillReturnRows(rows)

		// Create request with last query parameter
		req := httptest.NewRequest("GET", "/api/plans?last=true", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", mockUser)
		ctx = context.WithValue(ctx, "household", householdID)
		req = req.WithContext(ctx)

		// Call the handler
		GetPlans(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response body
		var plan models.Plan
		err := json.Unmarshal(rec.Body.Bytes(), &plan)
		require.NoError(t, err)

		// Verify response data
		assert.Equal(t, 1, plan.ID)
		assert.Equal(t, 42, plan.HouseholdID)
	})

	t.Run("Get next plan", func(t *testing.T) {
		// Setup mock authentication
		testUserID := "user1"
		householdID := 42
		mockUser := &clerk.User{ID: testUserID}

		// Set up mock expectations for GetNextPlan
		startDate := time.Now().AddDate(0, 0, 1)
		endDate := time.Now().AddDate(0, 0, 7)
		rows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
			AddRow(1, startDate, endDate, householdID)

		mock.ExpectQuery("SELECT \\* FROM plans WHERE household_id").
			WithArgs(householdID).
			WillReturnRows(rows)

		// Create request with next query parameter
		req := httptest.NewRequest("GET", "/api/plans?next=true", nil)
		rec := httptest.NewRecorder()

		// Set up context with mocked DB
		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", mockUser)
		ctx = context.WithValue(ctx, "household", householdID)
		req = req.WithContext(ctx)

		// Call the handler
		GetPlans(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response body
		var plan models.Plan
		err := json.Unmarshal(rec.Body.Bytes(), &plan)
		require.NoError(t, err)

		// Verify response data
		assert.Equal(t, 1, plan.ID)
		assert.Equal(t, 42, plan.HouseholdID)
	})

	t.Run("Get future plans", func(t *testing.T) {
		// Setup mock authentication
		testUserID := "user1"
		householdID := 42
		mockUser := &clerk.User{ID: testUserID}

		// Set up mock expectations for GetFuturePlans
		// First, mock the plan IDs query (should only return 'id' column)
		planIDRows := sqlmock.NewRows([]string{"id"}).
			AddRow(1).
			AddRow(2)
		mock.ExpectQuery("SELECT id FROM plans WHERE end_date >").
			WithArgs(householdID).
			WillReturnRows(planIDRows)

		// Then, mock the full plan fetch for each ID
		rows1 := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
			AddRow(1, time.Now().AddDate(0, 0, 1), time.Now().AddDate(0, 0, 7), householdID)
		mock.ExpectQuery("SELECT \\* FROM plans WHERE id").
			WithArgs(1).
			WillReturnRows(rows1)
		rows2 := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
			AddRow(1, 1, 101).
			AddRow(2, 1, 102)
		mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id").
			WithArgs(1).
			WillReturnRows(rows2)

		// Create request with future query parameter
		req := httptest.NewRequest("GET", "/api/plans?future=true", nil)
		rec := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), "db", sqlxDB)
		ctx = context.WithValue(ctx, "user", mockUser)
		ctx = context.WithValue(ctx, "household", householdID)
		req = req.WithContext(ctx)

		GetPlans(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var plans []models.Plan
		err := json.Unmarshal(rec.Body.Bytes(), &plans)
		require.NoError(t, err)

		assert.Len(t, plans, 1)
		assert.Equal(t, householdID, plans[0].HouseholdID)
	})
}

func TestGetPlan(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Set up mock expectations for GetPlan
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 0, 7)
	rows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(1, startDate, endDate, 42)

	mock.ExpectQuery("SELECT \\* FROM plans WHERE id").
		WithArgs(1).
		WillReturnRows(rows)

	// Mock for plan_meals query
	mealRows := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, 1, 101).
		AddRow(2, 1, 102)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id").
		WithArgs(1).
		WillReturnRows(mealRows)

	// Create request
	req := httptest.NewRequest("GET", "/api/plans/1", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB, ID, and household
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	ctx = context.WithValue(ctx, "household", 42)
	req = req.WithContext(ctx)

	// Call the handler
	GetPlan(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var plan models.Plan
	err := json.Unmarshal(rec.Body.Bytes(), &plan)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, plan.ID)
	assert.Equal(t, 42, plan.HouseholdID)
	assert.Equal(t, 2, len(plan.Meals))
}

func TestCreatePlan(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: "user1"}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Create a test plan
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 0, 7)
	newPlan := models.Plan{
		StartDate:   models.Date{Time: startDate},
		EndDate:     models.Date{Time: endDate},
		HouseholdID: 42,
		Meals:       []int{101, 102},
	}

	// Mock for CreatePlan
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO plans").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 42).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT id FROM plans WHERE start_date").
		WithArgs(sqlmock.AnyArg(), 42).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Mock for UpdatePlan (called within CreatePlan)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM plan_meals WHERE plan_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	for _, mealID := range newPlan.Meals {
		mock.ExpectExec("INSERT INTO plan_meals").
			WithArgs(1, mealID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	// For the returned plan after creation (GetPlan call)
	rows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(1, startDate, endDate, 42)
	mock.ExpectQuery("SELECT \\* FROM plans WHERE id").
		WithArgs(1).
		WillReturnRows(rows)

	mealRows := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, 1, 101).
		AddRow(2, 1, 102)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id").
		WithArgs(1).
		WillReturnRows(mealRows)

	// Create request with plan data
	planJSON, err := json.Marshal(newPlan)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/plans", bytes.NewBuffer(planJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "household", 42)
	ctx = context.WithValue(ctx, "user", &clerk.User{ID: "user1"})
	req = req.WithContext(ctx)

	// Call the handler
	CreatePlan(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var createdPlan models.Plan
	err = json.Unmarshal(rec.Body.Bytes(), &createdPlan)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, 1, createdPlan.ID)
	assert.Equal(t, 42, createdPlan.HouseholdID)
	assert.Equal(t, 2, len(createdPlan.Meals))
}

func TestUpdatePlan(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: "user1"}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// First mock for getting the existing plan for authorization check
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 0, 7)
	existingPlanRows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(1, startDate, endDate, 42)

	mock.ExpectQuery("SELECT \\* FROM plans WHERE id").
		WithArgs(1).
		WillReturnRows(existingPlanRows)

	// Mock for plan_meals query for existing plan
	mealRows := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, 1, 101)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id").
		WithArgs(1).
		WillReturnRows(mealRows)

	// Create an updated plan
	updatePlan := models.Plan{
		StartDate:   models.Date{Time: time.Now().AddDate(0, 0, 2)},
		EndDate:     models.Date{Time: time.Now().AddDate(0, 0, 9)},
		HouseholdID: 42,
		Meals:       []int{201, 202},
	}

	// Mock for UpdatePlan - it only updates meals, not the plan dates
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM plan_meals WHERE plan_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	for _, mealID := range updatePlan.Meals {
		mock.ExpectExec("INSERT INTO plan_meals").
			WithArgs(1, mealID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	// For the returned plan after update (GetPlan call) - returns original dates
	updatedPlanRows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(1, startDate, endDate, 42)

	mock.ExpectQuery("SELECT \\* FROM plans WHERE id").
		WithArgs(1).
		WillReturnRows(updatedPlanRows)

	updatedMealRows := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, 1, 201).
		AddRow(2, 1, 202)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id").
		WithArgs(1).
		WillReturnRows(updatedMealRows)

	// Create request with plan data
	planJSON, err := json.Marshal(updatePlan)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/plans/1", bytes.NewBuffer(planJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB, ID, and household
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	ctx = context.WithValue(ctx, "household", 42)
	ctx = context.WithValue(ctx, "user", &clerk.User{ID: "user1"})
	req = req.WithContext(ctx)

	// Call the handler
	UpdatePlan(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var updatedPlan models.Plan
	err = json.Unmarshal(rec.Body.Bytes(), &updatedPlan)
	require.NoError(t, err)

	// Verify response data - UpdatePlan only updates meals, not dates
	assert.Equal(t, 1, updatedPlan.ID)
	assert.Equal(t, 42, updatedPlan.HouseholdID)
	assert.Equal(t, 2, len(updatedPlan.Meals))
}

func TestDeletePlan(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: "user1"}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// First mock for getting the existing plan for authorization check
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 0, 7)
	existingPlanRows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(1, startDate, endDate, 42)

	mock.ExpectQuery("SELECT \\* FROM plans WHERE id").
		WithArgs(1).
		WillReturnRows(existingPlanRows)

	// Mock for plan_meals query for existing plan
	mealRows := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, 1, 101)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id").
		WithArgs(1).
		WillReturnRows(mealRows)

	// Mock for DeletePlan
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM plan_meals").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM plans").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Create request
	req := httptest.NewRequest("DELETE", "/api/plans/1", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB, ID, and household
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	ctx = context.WithValue(ctx, "household", 42)
	ctx = context.WithValue(ctx, "user", &clerk.User{ID: "user1"})
	req = req.WithContext(ctx)

	// Call the handler
	DeletePlan(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestGetPlanIngredients(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Set up mock expectations for GetPlanIngredients
	rows := sqlmock.NewRows([]string{"name", "amount"}).
		AddRow("Ingredient 1", "1 cup").
		AddRow("Ingredient 2", "2 tbsp")

	mock.ExpectQuery("SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE").
		WithArgs(1, 42).
		WillReturnRows(rows)

	// Create request
	req := httptest.NewRequest("GET", "/api/plans/1/ingredients", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB, ID, and household
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	ctx = context.WithValue(ctx, "id", 1)
	ctx = context.WithValue(ctx, "household", 42)
	req = req.WithContext(ctx)

	// Call the handler
	GetPlanIngredients(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var ingredients []models.Ingredient
	err := json.Unmarshal(rec.Body.Bytes(), &ingredients)
	require.NoError(t, err)

	// Verify response data
	assert.Len(t, ingredients, 2)
	assert.Equal(t, "Ingredient 1", ingredients[0].Name)
	assert.Equal(t, "1 cup", ingredients[0].Amount)
	assert.Equal(t, "Ingredient 2", ingredients[1].Name)
	assert.Equal(t, "2 tbsp", ingredients[1].Amount)
}
