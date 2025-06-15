package api

import (
	"encoding/json"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
)

func GetPantryHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)
	householdID, err := GetHouseholdIDForUser(db, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pantry, err := models.GetPantry(db, householdID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(pantry)
}

func UpdatePantryHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)
	householdID, err := GetHouseholdIDForUser(db, user.ID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := new(models.Pantry)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pantry, err := models.UpdatePantry(db, householdID, data)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(pantry)
}

func DeletePantryHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)
	householdID, err := GetHouseholdIDForUser(db, user.ID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = models.DeletePantry(db, householdID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CreatePantryHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)
	householdID, err := GetHouseholdIDForUser(db, user.ID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := new(models.Pantry)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pantry, err := models.CreatePantry(db, householdID, data)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(pantry)
}
