package models

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a mock sqlx database
func newMockDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock, nil
}

func TestGetPlan(t *testing.T) {
	sqlxDB, mock, err := newMockDB()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer sqlxDB.Close()

	testID := 1
	startDate := time.Now().AddDate(0, 0, 1) // tomorrow
	endDate := time.Now().AddDate(0, 0, 7)   // a week from today
	householdID := 42

	planRows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(testID, startDate, endDate, householdID)
	mock.ExpectQuery("SELECT \\* FROM plans WHERE id=\\$1").
		WithArgs(testID).
		WillReturnRows(planRows)

	mealRows := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, testID, 101).
		AddRow(2, testID, 102)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id=\\$1").
		WithArgs(testID).
		WillReturnRows(mealRows)

	plan, err := GetPlan(sqlxDB, testID)
	assert.NoError(t, err)
	assert.Equal(t, testID, plan.ID)
	assert.Equal(t, householdID, plan.HouseholdID)
	assert.Equal(t, 2, len(plan.Meals))
	assert.Equal(t, 101, plan.Meals[0])
	assert.Equal(t, 102, plan.Meals[1])

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetFuturePlans(t *testing.T) {
	sqlxDB, mock, err := newMockDB()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer sqlxDB.Close()

	householdID := 42

	// Mock the plans query
	startDate1 := time.Now().AddDate(0, 0, 1) // tomorrow
	endDate1 := time.Now().AddDate(0, 0, 7)   // a week from today
	startDate2 := time.Now().AddDate(0, 0, 8) // 8 days from today
	endDate2 := time.Now().AddDate(0, 0, 14)  // two weeks from today

	planIDsRows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM plans WHERE end_date > NOW() AND household_id=$1 ORDER BY start_date ASC")).
		WithArgs(householdID).
		WillReturnRows(planIDsRows)

	// Mock GetPlan for each id
	planRows1 := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(1, startDate1, endDate1, householdID)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plans WHERE id=$1")).
		WithArgs(1).
		WillReturnRows(planRows1)
	mealRows1 := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, 1, 101)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plan_meals WHERE plan_id=$1")).
		WithArgs(1).
		WillReturnRows(mealRows1)

	planRows2 := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(2, startDate2, endDate2, householdID)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plans WHERE id=$1")).
		WithArgs(2).
		WillReturnRows(planRows2)
	mealRows2 := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(2, 2, 102)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM plan_meals WHERE plan_id=$1")).
		WithArgs(2).
		WillReturnRows(mealRows2)

	plans, err := GetFuturePlans(sqlxDB, householdID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(*plans))
	assert.Equal(t, 1, (*plans)[0].ID)
	assert.Equal(t, 2, (*plans)[1].ID)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetPlanIngredients(t *testing.T) {
	sqlxDB, mock, err := newMockDB()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer sqlxDB.Close()

	planID := 1
	householdID := 42

	// Mock the ingredients query
	ingredientsRows := sqlmock.NewRows([]string{"name", "amount"}).
		AddRow("Flour", "2 cups").
		AddRow("Sugar", "1 cup").
		AddRow("Eggs", "2")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=$1 AND pm.household_id=$2")).
		WithArgs(planID, householdID).
		WillReturnRows(ingredientsRows)

	ingredients, err := GetPlanIngredients(sqlxDB, planID, householdID)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(*ingredients))
	assert.Equal(t, "Flour", (*ingredients)[0].Name)
	assert.Equal(t, "2 cups", (*ingredients)[0].Amount)
	assert.Equal(t, "Sugar", (*ingredients)[1].Name)
	assert.Equal(t, "Eggs", (*ingredients)[2].Name)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestCreatePlan(t *testing.T) {
	sqlxDB, mock, err := newMockDB()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer sqlxDB.Close()

	// Set up test data
	future := time.Now().AddDate(0, 0, 5)     // 5 days in the future
	endFuture := time.Now().AddDate(0, 0, 12) // 12 days in the future

	testPlan := &Plan{
		StartDate:   Date{Time: future},
		EndDate:     Date{Time: endFuture},
		HouseholdID: 42,
		Meals:       []int{101, 102},
	}

	// Mock the database interactions for CreatePlan
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO plans").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 42).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT id FROM plans WHERE start_date=\\$1 AND household_id=\\$2").
		WithArgs(sqlmock.AnyArg(), 42).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Mocks for the UpdatePlan call inside CreatePlan
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM plan_meals WHERE plan_id=\\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	for _, mealID := range testPlan.Meals {
		mock.ExpectExec("INSERT INTO plan_meals").
			WithArgs(1, mealID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	// Final GetPlan call
	planRows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(1, future, endFuture, 42)
	mock.ExpectQuery("SELECT \\* FROM plans WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(planRows)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
			AddRow(1, 1, 101).
			AddRow(2, 1, 102))

	// Execute the function
	plan, err := CreatePlan(sqlxDB, testPlan, 42)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, 1, plan.ID)
	assert.Equal(t, 42, plan.HouseholdID)
	assert.Equal(t, 2, len(plan.Meals))

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestValidatePlan(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		plan    *Plan
		wantErr bool
	}{
		{
			name: "Valid Plan",
			plan: &Plan{
				StartDate:   Date{Time: now.AddDate(0, 0, 1)}, // tomorrow
				EndDate:     Date{Time: now.AddDate(0, 0, 7)}, // a week from today
				HouseholdID: 42,
			},
			wantErr: false,
		},
		{
			name: "Start Date in Past",
			plan: &Plan{
				StartDate:   Date{Time: now.AddDate(0, 0, -1)}, // yesterday
				EndDate:     Date{Time: now.AddDate(0, 0, 7)},  // a week from today
				HouseholdID: 42,
			},
			wantErr: true,
		},
		{
			name: "End Date in Past",
			plan: &Plan{
				StartDate:   Date{Time: now.AddDate(0, 0, 1)},  // tomorrow
				EndDate:     Date{Time: now.AddDate(0, 0, -3)}, // 3 days ago
				HouseholdID: 42,
			},
			wantErr: true,
		},
		{
			name: "Start Date After End Date",
			plan: &Plan{
				StartDate:   Date{Time: now.AddDate(0, 0, 7)}, // a week from today
				EndDate:     Date{Time: now.AddDate(0, 0, 1)}, // tomorrow
				HouseholdID: 42,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlan(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdatePlan(t *testing.T) {
	sqlxDB, mock, err := newMockDB()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer sqlxDB.Close()

	// Set up test data
	planID := 1
	future := time.Now().AddDate(0, 0, 5)     // 5 days in the future
	endFuture := time.Now().AddDate(0, 0, 12) // 12 days in the future

	testPlan := &Plan{
		ID:          planID,
		StartDate:   Date{Time: future},
		EndDate:     Date{Time: endFuture},
		HouseholdID: 42,
		Meals:       []int{201, 202, 203},
	}

	// Mock the transaction for UpdatePlan
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM plan_meals WHERE plan_id=\\$1").
		WithArgs(planID).
		WillReturnResult(sqlmock.NewResult(0, 3))

	for _, mealID := range testPlan.Meals {
		mock.ExpectExec("INSERT INTO plan_meals \\(plan_id, meal_id\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs(planID, mealID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	// Mock the final GetPlan call
	planRows := sqlmock.NewRows([]string{"id", "start_date", "end_date", "household_id"}).
		AddRow(planID, future, endFuture, 42)
	mock.ExpectQuery("SELECT \\* FROM plans WHERE id=\\$1").
		WithArgs(planID).
		WillReturnRows(planRows)

	mealRows := sqlmock.NewRows([]string{"id", "plan_id", "meal_id"}).
		AddRow(1, planID, 201).
		AddRow(2, planID, 202).
		AddRow(3, planID, 203)
	mock.ExpectQuery("SELECT \\* FROM plan_meals WHERE plan_id=\\$1").
		WithArgs(planID).
		WillReturnRows(mealRows)

	// Execute the function
	updatedPlan, err := UpdatePlan(sqlxDB, planID, testPlan)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedPlan)
	assert.Equal(t, planID, updatedPlan.ID)
	assert.Equal(t, 42, updatedPlan.HouseholdID)
	assert.Equal(t, 3, len(updatedPlan.Meals))
	assert.Equal(t, 201, updatedPlan.Meals[0])
	assert.Equal(t, 202, updatedPlan.Meals[1])
	assert.Equal(t, 203, updatedPlan.Meals[2])

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestDeletePlan(t *testing.T) {
	sqlxDB, mock, err := newMockDB()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer sqlxDB.Close()

	planID := 1

	// Mock the transaction for DeletePlan
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM plan_meals WHERE plan_id=\\$1").
		WithArgs(planID).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("DELETE FROM plans WHERE id=\\$1").
		WithArgs(planID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Execute the function
	err = DeletePlan(sqlxDB, planID)

	// Assertions
	assert.NoError(t, err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

// Test for Date type methods
func TestDateMethods(t *testing.T) {
	// Test UnmarshalJSON
	t.Run("UnmarshalJSON", func(t *testing.T) {
		var d Date
		err := d.UnmarshalJSON([]byte(`"2025-01-15"`))
		assert.NoError(t, err)
		assert.Equal(t, 2025, d.Year())
		assert.Equal(t, time.January, d.Month())
		assert.Equal(t, 15, d.Day())
	})

	// Test MarshalJSON
	t.Run("MarshalJSON", func(t *testing.T) {
		d := Date{Time: time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC)}
		bytes, err := d.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, `"2025-07-04"`, string(bytes))
	})

	// Test Scan
	t.Run("Scan", func(t *testing.T) {
		d := Date{}
		timeVal := time.Date(2025, 8, 12, 0, 0, 0, 0, time.UTC)
		err := d.Scan(timeVal)
		assert.NoError(t, err)
		assert.Equal(t, timeVal, d.Time)
	})

	// Test Value
	t.Run("Value", func(t *testing.T) {
		timeVal := time.Date(2025, 9, 23, 0, 0, 0, 0, time.UTC)
		d := Date{Time: timeVal}
		val, err := d.Value()
		assert.NoError(t, err)
		assert.Equal(t, driver.Value(timeVal), val)
	})
}
