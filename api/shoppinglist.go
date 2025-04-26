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

	ingredients, err := models.GetPlanIngredients(db, plan.ID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pantry, err := models.GetPantry(db, user.ID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	ingredients = filter(ingredients, func(i models.Ingredient) bool {
		return slices.ContainsFunc(pantry.Items, func(a string) bool {
			return strings.Contains(strings.ToLower(i.Name), a)
		}) == false
	})

	list := models.ShoppingList{
		Plan:        *plan,
		Ingredients: *ingredients,
	}

	json.NewEncoder(w).Encode(list)
}
