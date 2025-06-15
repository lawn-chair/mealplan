package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// var ErrValidation = errors.New("name, description, and slug are required")
type Date struct {
	time.Time
}

func (t *Date) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02"`, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t Date) MarshalJSON() ([]byte, error) {
	return []byte(t.Time.Format(`"2006-01-02"`)), nil
}

func (t *Date) Scan(value interface{}) error {
	t.Time = value.(time.Time)
	return nil
}

func (t Date) Value() (driver.Value, error) {
	return t.Time, nil
}

type Plan struct {
	ID          int   `db:"id" json:"id"`
	StartDate   Date  `db:"start_date" json:"start_date"`
	EndDate     Date  `db:"end_date" json:"end_date"`
	HouseholdID int   `db:"household_id" json:"household_id"`
	Meals       []int `json:"meals,omitempty"`
}

type PlanMeals struct {
	ID     int `db:"id" json:"id"`
	PlanID int `db:"plan_id" json:"plan_id"`
	MealID int `db:"meal_id" json:"meal_id"`
}

type Ingredient struct {
	Name   string `db:"name" json:"name"`
	Amount string `db:"amount" json:"amount"`
}

func GetPlans(db *sqlx.DB, householdID int) (*[]Plan, error) {
	plans := []Plan{}
	err := db.Select(&plans, "SELECT * FROM plans WHERE household_id = $1 ORDER BY start_date ASC", householdID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &plans, nil
}

func GetLastPlan(db *sqlx.DB, householdID int) (*Plan, error) {
	plan := Plan{}
	err := db.Get(&plan, "SELECT * FROM plans WHERE household_id=$1 ORDER BY start_date DESC LIMIT 1", householdID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &plan, nil
}

func GetNextPlan(db *sqlx.DB, householdID int) (*Plan, error) {
	plan := Plan{}
	err := db.Get(&plan, "SELECT * FROM plans WHERE household_id=$1 AND start_date > NOW() ORDER BY start_date ASC LIMIT 1", householdID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &plan, nil
}

func GetPlan(db *sqlx.DB, id int) (*Plan, error) {
	plan := Plan{}
	err := db.Get(&plan, "SELECT * FROM plans WHERE id=$1", id)
	if err != nil {
		fmt.Println("GetPlan error:", err)
		return nil, err
	}

	planMeals := []PlanMeals{}
	err = db.Select(&planMeals, "SELECT * FROM plan_meals WHERE plan_id=$1", plan.ID)
	if err != nil {
		fmt.Println("GetPlan error:", err)
		return nil, err
	}

	plan.Meals = make([]int, len(planMeals))
	for i, pm := range planMeals {
		plan.Meals[i] = pm.MealID
	}

	return &plan, nil
}

func ValidatePlan(p *Plan) error {
	now := time.Now()
	// Get the beginning of today (00:00:00)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if p.StartDate.Before(todayStart) || p.EndDate.Before(todayStart) {
		return fmt.Errorf("start date and end date must be in the future")
	} else if p.StartDate.After(p.EndDate.Time) {
		return fmt.Errorf("start date must be before end date")
	}

	return nil
}

func CreatePlan(db *sqlx.DB, p *Plan, householdID int) (*Plan, error) {
	if err := ValidatePlan(p); err != nil {
		return nil, err
	}

	p.HouseholdID = householdID

	tx, err := db.Beginx()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = tx.Exec("INSERT INTO plans (start_date, end_date, household_id) VALUES ($1, $2, $3)", p.StartDate, p.EndDate, p.HouseholdID)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	err = tx.Get(&p.ID, "SELECT id FROM plans WHERE start_date=$1 AND household_id=$2", p.StartDate, p.HouseholdID)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}
	tx.Commit()

	return UpdatePlan(db, p.ID, p)
}

func UpdatePlan(db *sqlx.DB, id int, p *Plan) (*Plan, error) {

	tx, err := db.Beginx()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = tx.Exec("DELETE FROM plan_meals WHERE plan_id=$1", id)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	for _, meal := range p.Meals {
		_, err = tx.Exec("INSERT INTO plan_meals (plan_id, meal_id) VALUES ($1, $2)", id, meal)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return GetPlan(db, id)
}

func DeletePlan(db *sqlx.DB, id int) error {
	tx, err := db.Beginx()
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec("DELETE FROM plan_meals WHERE plan_id=$1", id)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec("DELETE FROM plans WHERE id=$1", id)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetPlanIngredients(db *sqlx.DB, id int, householdID int) (*[]Ingredient, error) {
	ingredients := []Ingredient{}
	err := db.Select(&ingredients, "SELECT i.name, i.amount FROM meal_ingredients i JOIN plan_meals pm ON pm.meal_id = i.meal_id WHERE pm.plan_id=$1 AND pm.household_id=$2", id, householdID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &ingredients, nil
}

func GetFuturePlans(db *sqlx.DB, householdID int) (*[]Plan, error) {
	plans := []Plan{}
	plan_ids := []int{}
	err := db.Select(&plan_ids, "SELECT id FROM plans WHERE end_date > NOW() AND household_id=$1 ORDER BY start_date ASC", householdID)
	if err != nil {
		fmt.Println("Error fetching plan IDs:", err)
		return nil, err
	}

	for _, id := range plan_ids {
		plan, err := GetPlan(db, id)
		if err != nil {
			fmt.Printf("Error fetching plan with ID %d: %v\n", id, err)
			continue // Skip this plan if there's an error
		}
		plans = append(plans, *plan)
	}

	return &plans, nil
}
