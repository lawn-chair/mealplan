package models

import (
	"database/sql" // Added for sql.ErrNoRows
	"database/sql/driver"
	"encoding/json"
	"errors" // Already present, used for errors.Is
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Status struct {
	Items []StatusItem `json:"items"`
}

type StatusItem struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
}

func (s *Status) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

func (s Status) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type ShoppingStatus struct {
	PlanID int    `db:"plan_id"`
	Status Status `db:"status"`
}

type ShoppingListItem struct {
	Name    string `json:"name"`
	Amount  string `json:"amount"`
	Checked bool   `json:"checked"`
}

type ShoppingList struct {
	Plan        Plan               `json:"plan"`
	Ingredients []ShoppingListItem `json:"ingredients"`
}

func GetShoppingList(db *sqlx.DB, planID int) (*ShoppingList, error) {
	// Fetch full plan details
	var plan Plan
	err := db.Get(&plan, "SELECT * FROM plans WHERE id = $1", planID)
	if err != nil {
		fmt.Printf("Error fetching plan details for shopping list (planID: %d): %v\\n", planID, err)
		return nil, fmt.Errorf("failed to fetch plan %d: %w", planID, err)
	}

	var shoppingStatus ShoppingStatus

	err = db.Get(&shoppingStatus, "SELECT * FROM shopping_status WHERE plan_id = $1", planID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			shoppingStatus = ShoppingStatus{
				PlanID: planID,
				Status: Status{Items: []StatusItem{}},
			}
			_, err = db.Exec("INSERT INTO shopping_status (plan_id, status) VALUES ($1, $2)", planID, Status{Items: []StatusItem{}})
			if err != nil {
				fmt.Printf("Error inserting empty shopping status for planID %d: %v\\n", planID, err)
				return nil, fmt.Errorf("failed to create shopping list for plan %d: %w", planID, err)
			}
			// No error to return in this specific case, proceed with empty status.
		} else {
			fmt.Printf("Error fetching shopping status from DB (planID: %d): %v\\n", planID, err)
			return nil, fmt.Errorf("failed to fetch shopping status for plan %d: %w", planID, err)
		}
	}

	ingredients, err := GetPlanIngredients(db, planID, plan.HouseholdID)
	if err != nil {
		fmt.Println("Error fetching plan ingredients:", err) // Keep original logging style
		return nil, err
	}

	shoppingList := &ShoppingList{
		Plan:        plan,
		Ingredients: make([]ShoppingListItem, len(*ingredients)),
	}

	for i, item := range *ingredients {
		checked := false
		for _, status := range shoppingStatus.Status.Items {
			if status.Name == item.Name && status.Amount == item.Amount {
				checked = true
				break
			}
		}

		shoppingList.Ingredients[i] = ShoppingListItem{
			Name:    item.Name,
			Amount:  item.Amount,
			Checked: checked,
		}
	}

	return shoppingList, nil
}

func UpdateShoppingList(db *sqlx.DB, householdID int, list *ShoppingList) error {

	if list.Plan.ID <= 0 {
		return fmt.Errorf("invalid plan ID: %d", list.Plan.ID)
	}

	var planID int
	err := db.Get(&planID, "SELECT id FROM plans WHERE id = $1 AND household_id = $2", list.Plan.ID, householdID)
	if err != nil {
		return fmt.Errorf("plan not found or unauthorized: %w", err)
	}

	status := Status{Items: []StatusItem{}}
	for _, item := range list.Ingredients {
		if item.Checked {
			status.Items = append(status.Items, StatusItem{Name: item.Name, Amount: item.Amount})
		}
	}
	_, err = db.Exec("UPDATE shopping_status SET status = $1 WHERE plan_id = $2", status, planID)
	return err
}
