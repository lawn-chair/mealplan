package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lawn-chair/mealplan/models"
)

// POST /api/household/join-code
func GenerateJoinCodeHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	householdID := r.Context().Value("household").(int)

	code, err := models.GenerateJoinCode(db, householdID, 60*time.Minute)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"code": code})
}

// POST /api/household/join
func JoinHouseholdHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := models.JoinHouseholdByCode(db, user, req.Code); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/household/leave
func LeaveHouseholdHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)
	if err := models.LeaveHousehold(db, user); err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/household/remove-member
func RemoveHouseholdMemberHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)
	householdID := r.Context().Value("household").(int)
	var req struct {
		TargetUserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := models.RemoveHouseholdMember(db, householdID, user.ID, req.TargetUserID); err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetUserHouseholdHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user := r.Context().Value("user").(*clerk.User)

	household, err := models.GetHousehold(db, user)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(household)
}

// GET /api/household/members
func ListHouseholdMembersHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	householdID := r.Context().Value("household").(int)

	members, err := models.ListHouseholdMembers(db, householdID)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(members)
}
