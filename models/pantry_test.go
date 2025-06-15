package models

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestGetPantry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	testHouseholdID := 42
	rows1 := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, testHouseholdID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE household_id").
		WithArgs(testHouseholdID).
		WillReturnRows(rows1)

	rows2 := sqlmock.NewRows([]string{"item_name"}).AddRow("salt").AddRow("pepper").AddRow("sugar")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(rows2)

	pantry, err := GetPantry(sqlxDB, testHouseholdID)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), pantry.ID)
	assert.Equal(t, testHouseholdID, pantry.HouseholdID)
	assert.Equal(t, 3, len(pantry.Items))
	assert.Equal(t, "salt", pantry.Items[0])
	assert.Equal(t, "pepper", pantry.Items[1])
	assert.Equal(t, "sugar", pantry.Items[2])

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreatePantry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	testHouseholdID := 42
	testPantry := &Pantry{
		Items: []string{"flour", "sugar", "salt"},
	}

	// First, the INSERT into pantry table
	mock.ExpectExec("INSERT INTO pantry").
		WithArgs(testHouseholdID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// First GetPantry call within UpdatePantry after INSERT
	rows1 := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, testHouseholdID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE household_id").
		WithArgs(testHouseholdID).
		WillReturnRows(rows1)

	// Get existing items (empty at first)
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"item_name"}))

	// Then, DELETE FROM pantry_items
	mock.ExpectExec("DELETE FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Then INSERT for each item
	for _, item := range testPantry.Items {
		mock.ExpectExec("INSERT INTO pantry_items").
			WithArgs(1, item).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Final GetPantry call - after UpdatePantry
	rows2 := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, testHouseholdID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE household_id").
		WithArgs(testHouseholdID).
		WillReturnRows(rows2)

	rows3 := sqlmock.NewRows([]string{"item_name"}).AddRow("flour").AddRow("sugar").AddRow("salt")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(rows3)

	result, err := CreatePantry(sqlxDB, testHouseholdID, testPantry)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, testHouseholdID, result.HouseholdID)
	assert.Equal(t, 3, len(result.Items))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePantry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	testHouseholdID := 42
	testPantry := &Pantry{
		Items: []string{"updated-item-1", "updated-item-2"},
	}

	// First GetPantry call
	rows1 := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, testHouseholdID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE household_id").
		WithArgs(testHouseholdID).
		WillReturnRows(rows1)

	// Get existing items
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"item_name"}).
			AddRow("old-item-1").
			AddRow("old-item-2"))

	// Then DELETE FROM pantry_items
	mock.ExpectExec("DELETE FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Then INSERT for each item
	for _, item := range testPantry.Items {
		mock.ExpectExec("INSERT INTO pantry_items").
			WithArgs(1, item).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Final GetPantry call
	rows2 := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, testHouseholdID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE household_id").
		WithArgs(testHouseholdID).
		WillReturnRows(rows2)

	rows3 := sqlmock.NewRows([]string{"item_name"}).AddRow("updated-item-1").AddRow("updated-item-2")
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(rows3)

	result, err := UpdatePantry(sqlxDB, testHouseholdID, testPantry)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, testHouseholdID, result.HouseholdID)
	assert.Equal(t, 2, len(result.Items))
	assert.Equal(t, "updated-item-1", result.Items[0])
	assert.Equal(t, "updated-item-2", result.Items[1])

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeletePantry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	testHouseholdID := 42

	// First GetPantry call
	rows := sqlmock.NewRows([]string{"id", "household_id"}).AddRow(1, testHouseholdID)
	mock.ExpectQuery("SELECT \\* FROM pantry WHERE household_id").
		WithArgs(testHouseholdID).
		WillReturnRows(rows)

	// GetPantryItems call from inside GetPantry
	mock.ExpectQuery("SELECT item_name FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"item_name"}).
			AddRow("item-1").
			AddRow("item-2"))

	// Then DELETE FROM pantry_items
	mock.ExpectExec("DELETE FROM pantry_items WHERE pantry_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Note: The actual DeletePantry function doesn't delete the pantry record
	// itself, it only deletes the pantry items

	err = DeletePantry(sqlxDB, testHouseholdID)
	assert.NoError(t, err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
