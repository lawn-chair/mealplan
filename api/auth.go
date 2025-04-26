package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

func RequiresAuthentication(r *http.Request) (*clerk.User, error) {
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
