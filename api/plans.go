package api

import (
	"encoding/json"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
)

func GetPlans(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user, ok := r.Context().Value("user").(*clerk.User)
	householdID := r.Context().Value("household").(int)

	if !ok || user == nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	if r.URL.Query().Get("last") != "" {
		plan, err := models.GetLastPlan(db, householdID)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	} else if r.URL.Query().Get("next") != "" {
		plan, err := models.GetNextPlan(db, householdID)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	} else if r.URL.Query().Get("future") != "" {
		plans, err := models.GetFuturePlans(db, householdID)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plans)
		return
	}

	plans, err := models.GetPlans(db, householdID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plans)
}

func GetPlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)
	household := r.Context().Value("household").(int)

	plan, err := models.GetPlan(db, id)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan.HouseholdID != household {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(plan)
}

func UpdatePlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)
	user, ok := r.Context().Value("user").(*clerk.User)
	householdID := r.Context().Value("household").(int)

	if !ok || user == nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Plan)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, err := models.GetPlan(db, id)
	if err != nil || (plan.HouseholdID != householdID) {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}
	data.HouseholdID = householdID

	plan, err = models.UpdatePlan(db, id, data)
	if err != nil {
		if err == models.ErrValidation {
			ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(plan)
}

func CreatePlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user, ok := r.Context().Value("user").(*clerk.User)
	householdID := r.Context().Value("household").(int)
	if !ok || user == nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Plan)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, err := models.CreatePlan(db, data, householdID)
	if err != nil {
		if err == models.ErrValidation {
			ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(plan)
}

func DeletePlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)
	householdID := r.Context().Value("household").(int)
	user, ok := r.Context().Value("user").(*clerk.User)
	if !ok || user == nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	plan, err := models.GetPlan(db, id)
	if err != nil || (plan.HouseholdID != householdID) {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	err = models.DeletePlan(db, id)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetPlanIngredients(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)
	householdID := r.Context().Value("household").(int)

	ingredients, err := models.GetPlanIngredients(db, id, householdID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(*ingredients)
}
