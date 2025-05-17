package api

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
)

func GetRecipes(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)

	if slug := r.URL.Query().Get("slug"); slug != "" {
		id, err := models.GetRecipeIdFromSlug(db, slug)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusNotFound)
			return
		}
		recipe, err := models.GetRecipe(db, id)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recipe)
	} else {
		recipes, err := models.GetRecipes(db)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recipes)
	}
}

func GetRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	recipe, err := models.GetRecipe(db, id)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	_, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Recipe)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipe, err := models.UpdateRecipe(db, id, data)
	if err != nil {
		if err == models.ErrValidation {
			ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func CreateRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	_, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Recipe)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipe, err := models.CreateRecipe(db, data)
	if err != nil {
		if err == models.ErrValidation {
			ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	_, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	err = models.DeleteRecipe(db, id)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
