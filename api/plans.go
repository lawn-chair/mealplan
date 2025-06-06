package api

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
)

func GetPlans(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)

	if r.URL.Query().Get("last") != "" {
		user, err := RequiresAuthentication(r)
		if err != nil {
			ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
			return
		}

		plan, err := models.GetLastPlan(db, user.ID)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	} else if r.URL.Query().Get("next") != "" {
		user, err := RequiresAuthentication(r)
		if err != nil {
			ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
			return
		}

		plan, err := models.GetNextPlan(db, user.ID)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	} else if r.URL.Query().Get("future") != "" {
		plans, err := models.GetFuturePlans(db)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plans)
		return
	}

	plans, err := models.GetPlans(db)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plans)
}

func GetPlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	plan, err := models.GetPlan(db, id)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plan)
}

func UpdatePlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	user, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Plan)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, err := models.GetPlan(db, id)
	if err != nil || (plan.UserID != user.ID) || (data.UserID != "" && data.UserID != user.ID) {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
	}
	data.UserID = user.ID

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
	user, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Plan)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	data.UserID = user.ID

	plan, err := models.CreatePlan(db, data)
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

	user, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	plan, err := models.GetPlan(db, id)
	if err != nil || (plan.UserID != user.ID) {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
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

	ingredients, err := models.GetPlanIngredients(db, id)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(*ingredients)
}
