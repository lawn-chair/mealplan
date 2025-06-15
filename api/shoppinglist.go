package api

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
)

func filter[T any](ss *[]T, test func(T) bool) *[]T {
	rval := make([]T, 0)
	for _, s := range *ss {
		if test(s) {
			rval = append(rval, s)
		}
	}
	return &rval
}

func GetShoppingList(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	householdID := r.Context().Value("household").(int)

	plan, err := models.GetNextPlan(db, householdID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pantry, err := models.GetPantry(db, householdID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := models.GetShoppingList(db, plan.ID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list.Ingredients = *filter(&list.Ingredients, func(i models.ShoppingListItem) bool {
		return slices.ContainsFunc(pantry.Items, func(a string) bool {
			return strings.Contains(strings.ToLower(i.Name), a)
		}) == false
	})

	json.NewEncoder(w).Encode(list)
}

func UpdateShoppingList(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	householdID := r.Context().Value("household").(int)

	var list models.ShoppingList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := models.UpdateShoppingList(db, householdID, &list); err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}
