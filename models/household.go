package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/jmoiron/sqlx"
)

type Household struct {
	ID      int               `db:"id" json:"id"`
	Name    string            `db:"name" json:"name"`
	Members []HouseholdMember `json:"members"`
}

type HouseholdMember struct {
	HouseholdID int    `db:"household_id" json:"household_id"`
	UserID      string `db:"user_id" json:"user_id"`
	Email       string `db:"email" json:"email"`
}

type HouseholdJoinCode struct {
	Code        string    `db:"code" json:"code"`
	HouseholdID int       `db:"household_id" json:"household_id"`
	ExpiresAt   time.Time `db:"expires_at" json:"expires_at"`
}

func GenerateJoinCode(db *sqlx.DB, householdID int, duration time.Duration) (string, error) {
	code := strings.ToUpper(randomString(8))
	expires := time.Now().Add(duration)
	_, err := db.Exec(`INSERT INTO household_join_codes (code, household_id, expires_at) VALUES ($1, $2, $3)`, code, householdID, expires)
	if err != nil {
		return "", err
	}
	return code, nil
}

func getPrimaryEmailAddress(u *clerk.User) string {
	if len(u.EmailAddresses) == 0 {
		return ""
	}
	for _, email := range u.EmailAddresses {
		if email.ID == *u.PrimaryEmailAddressID {
			return email.EmailAddress
		}
	}
	return u.EmailAddresses[0].EmailAddress
}

func createHousehold(db *sqlx.DB, u *clerk.User) (int, error) {
	fmt.Println("Creating household for user:", u.ID, u.LastName)
	var householdID int
	err := db.QueryRow(`INSERT INTO households (name) VALUES ($1) RETURNING id`, *u.LastName+" Household").Scan(&householdID)
	if err != nil {
		fmt.Println("Error creating household:", err)
		return 0, err
	}
	fmt.Println("Created household with ID:", householdID)
	email := getPrimaryEmailAddress(u)
	_, err = db.Exec(`INSERT INTO household_members (household_id, user_id, email) VALUES ($1, $2, $3)`, householdID, u.ID, email)
	if err != nil {
		fmt.Println("Error adding user to household:", err)
		return 0, err
	}

	return householdID, nil
}

func JoinHouseholdByCode(db *sqlx.DB, user *clerk.User, code string) error {
	var householdID int
	err := db.Get(&householdID, `SELECT household_id FROM household_join_codes WHERE code=$1 AND expires_at > NOW()`, code)
	if err != nil {
		return errors.New("invalid or expired code")
	}
	_, err = db.Exec(`DELETE FROM household_members WHERE user_id=$1`, user.ID)
	if err != nil {
		fmt.Println("Error removing user from previous household:", err)
	}
	_, err = db.Exec(`INSERT INTO household_members (household_id, user_id, email) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, householdID, user.ID, user.PrimaryEmailAddressID)
	return err
}

func LeaveHousehold(db *sqlx.DB, user *clerk.User) error {
	_, err := db.Exec(`DELETE FROM household_members WHERE user_id=$1`, user.ID)
	createHousehold(db, user)
	return err
}

func RemoveHouseholdMember(db *sqlx.DB, householdID int, actorUserID, targetUserID string) error {
	if actorUserID == targetUserID {
		return errors.New("cannot remove yourself")
	}
	ctx := context.Background()

	user, err := user.Get(ctx, targetUserID)
	if err != nil {
		fmt.Println("Error fetching user:", err)
		return err
	}
	createHousehold(db, user)

	_, err = db.Exec(`DELETE FROM household_members WHERE household_id=$1 AND user_id=$2`, householdID, targetUserID)
	return err
}

func GetHousehold(db *sqlx.DB, user *clerk.User) (*Household, error) {
	var household Household
	// Query the household
	row := db.QueryRow(`SELECT h.id, h.name
		FROM households h
		JOIN household_members m ON h.id = m.household_id
		WHERE m.user_id = $1
		LIMIT 1`, user.ID)
	err := row.Scan(&household.ID, &household.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no household found, create a new one
			householdID, err := createHousehold(db, user)
			if err != nil {
				fmt.Println("Error creating household:", err)
				return nil, err
			}
			household.ID = householdID
			household.Name = *user.LastName + " Household"
		}
		return nil, err
	}

	// Query the members
	var members []HouseholdMember
	err = db.Select(&members, `SELECT household_id, user_id, email FROM household_members WHERE household_id = $1`, household.ID)
	if err != nil {
		return nil, err
	}
	household.Members = members

	fmt.Println("Household found:", household.ID, household.Name, household.Members)
	return &household, nil
}

func ListHouseholdMembers(db *sqlx.DB, householdID int) ([]string, error) {
	var members []string
	err := db.Select(&members, `SELECT user_id FROM household_members WHERE household_id=$1`, householdID)
	return members, err
}

func randomString(n int) string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)

	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
