package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/lawn-chair/mealplan/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPantryHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	testUserID := "test-user-id"
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: testUserID}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Mock for GetPantry
	rows := sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, testUserID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE user_id").
		WithArgs(testUserID).
		WillReturnRows(rows)

	// Mock for pantry items
	itemRows := sqlmock.NewRows([]string{"item_name"}).
		AddRow("salt").
		AddRow("pepper").
		AddRow("sugar")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(itemRows)

	// Create request
	req := httptest.NewRequest("GET", "/api/pantry", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	req = req.WithContext(ctx)

	// Call the handler
	GetPantryHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var pantry models.Pantry
	err := json.Unmarshal(rec.Body.Bytes(), &pantry)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, uint(1), pantry.ID)
	assert.Equal(t, testUserID, pantry.UserID)
	assert.Len(t, pantry.Items, 3)
	assert.Equal(t, "salt", pantry.Items[0])
	assert.Equal(t, "pepper", pantry.Items[1])
	assert.Equal(t, "sugar", pantry.Items[2])
}

func TestUpdatePantryHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	testUserID := "test-user-id"
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: testUserID}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Create test pantry for update
	updatePantry := models.Pantry{
		Items: []string{"flour", "sugar", "salt"},
	}

	// Mock for GetPantry inside UpdatePantry
	rows := sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, testUserID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE user_id").
		WithArgs(testUserID).
		WillReturnRows(rows)

	// Mock for existing pantry items
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"item_name"}).
			AddRow("old-item-1").
			AddRow("old-item-2"))

	// Mock DELETE and INSERT for pantry items
	mock.ExpectExec("DELETE FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Mock INSERTs for new items
	for _, item := range updatePantry.Items {
		mock.ExpectExec("INSERT INTO pantry_items").
			WithArgs(1, item).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Mock final GetPantry call
	finalRows := sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, testUserID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE user_id").
		WithArgs(testUserID).
		WillReturnRows(finalRows)

	// Mock for updated pantry items
	finalItemRows := sqlmock.NewRows([]string{"item_name"}).
		AddRow("flour").
		AddRow("sugar").
		AddRow("salt")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(finalItemRows)

	// Create request with pantry data
	pantryJSON, err := json.Marshal(updatePantry)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/pantry", bytes.NewBuffer(pantryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	req = req.WithContext(ctx)

	// Call the handler
	UpdatePantryHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var updatedPantry models.Pantry
	err = json.Unmarshal(rec.Body.Bytes(), &updatedPantry)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, uint(1), updatedPantry.ID)
	assert.Equal(t, testUserID, updatedPantry.UserID)
	assert.Len(t, updatedPantry.Items, 3)
	assert.Contains(t, updatedPantry.Items, "flour")
	assert.Contains(t, updatedPantry.Items, "sugar")
	assert.Contains(t, updatedPantry.Items, "salt")
}

func TestCreatePantryHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	testUserID := "test-user-id"
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: testUserID}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Create test pantry
	newPantry := models.Pantry{
		Items: []string{"apple", "banana", "orange"},
	}

	// Mock for INSERT into pantry
	mock.ExpectExec("INSERT INTO pantry").
		WithArgs(testUserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock for GetPantry inside UpdatePantry (called by CreatePantry)
	rows := sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, testUserID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE user_id").
		WithArgs(testUserID).
		WillReturnRows(rows)

	// Mock for existing pantry items (empty at first)
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"item_name"}))

	// Mock DELETE and INSERT for pantry items
	mock.ExpectExec("DELETE FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock INSERTs for new items
	for _, item := range newPantry.Items {
		mock.ExpectExec("INSERT INTO pantry_items").
			WithArgs(1, item).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Mock final GetPantry call
	finalRows := sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, testUserID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE user_id").
		WithArgs(testUserID).
		WillReturnRows(finalRows)

	// Mock for updated pantry items
	finalItemRows := sqlmock.NewRows([]string{"item_name"}).
		AddRow("apple").
		AddRow("banana").
		AddRow("orange")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(finalItemRows)

	// Create request with pantry data
	pantryJSON, err := json.Marshal(newPantry)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/pantry", bytes.NewBuffer(pantryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Set up context with mocked DB
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	req = req.WithContext(ctx)

	// Call the handler
	CreatePantryHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var createdPantry models.Pantry
	err = json.Unmarshal(rec.Body.Bytes(), &createdPantry)
	require.NoError(t, err)

	// Verify response data
	assert.Equal(t, uint(1), createdPantry.ID)
	assert.Equal(t, testUserID, createdPantry.UserID)
	assert.Len(t, createdPantry.Items, 3)
	assert.Contains(t, createdPantry.Items, "apple")
	assert.Contains(t, createdPantry.Items, "banana")
	assert.Contains(t, createdPantry.Items, "orange")
}

func TestDeletePantryHandler(t *testing.T) {
	sqlxDB, mock := setupMockDB(t)
	defer sqlxDB.Close()

	// Setup mock authentication
	testUserID := "test-user-id"
	mockAuthFunc := func(r *http.Request) (*clerk.User, error) {
		return &clerk.User{ID: testUserID}, nil
	}

	// Save original RequiresAuthentication function and restore it after the test
	originalFunc := RequiresAuthentication
	RequiresAuthentication = mockAuthFunc
	defer func() { RequiresAuthentication = originalFunc }()

	// Mock for GetPantry inside DeletePantry
	rows := sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, testUserID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE user_id").
		WithArgs(testUserID).
		WillReturnRows(rows)

	// Mock for pantry items
	itemRows := sqlmock.NewRows([]string{"item_name"}).
		AddRow("item-1").
		AddRow("item-2")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(itemRows)

	// Mock DELETE for pantry items
	mock.ExpectExec("DELETE FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/pantry", nil)
	rec := httptest.NewRecorder()

	// Set up context with mocked DB
	ctx := context.WithValue(req.Context(), "db", sqlxDB)
	req = req.WithContext(ctx)

	// Call the handler
	DeletePantryHandler(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
