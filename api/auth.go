package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/jmoiron/sqlx"
)

var RequiresAuthentication = func(r *http.Request) (*clerk.User, error) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		return nil, errors.New("unauthorized")
	}

	usr, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		return nil, errors.New("user not found")
	}
	fmt.Printf(`{"user_id": "%s", "email": "%s", "user_banned": "%t"}`, usr.ID, usr.EmailAddresses[0].EmailAddress, usr.Banned)
	fmt.Println()
	return usr, nil
}

// GetHouseholdIDForUser returns the household_id for a given user_id
func GetHouseholdIDForUser(db *sqlx.DB, userID string) (int, error) {
	var householdID int
	err := db.Get(&householdID, "SELECT household_id FROM household_members WHERE user_id = $1", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("user is not a member of any household")
		}
		return 0, err
	}
	return householdID, nil
}
