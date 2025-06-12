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

	user, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	plan, err := models.GetNextPlan(db, user.ID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pantry, err := models.GetPantry(db, user.ID)
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

	user, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	var list models.ShoppingList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := models.UpdateShoppingList(db, user.ID, &list); err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}
