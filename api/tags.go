package api

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
)

func ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	tags, err := models.GetAllTags(db)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tags)
}
