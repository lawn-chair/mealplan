package models

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Household struct {
	ID      int      `db:"id" json:"id"`
	Name    string   `db:"name" json:"name"`
	Members []string `json:"members"`
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

func JoinHouseholdByCode(db *sqlx.DB, userID, code string) error {
	var householdID int
	err := db.Get(&householdID, `SELECT household_id FROM household_join_codes WHERE code=$1 AND expires_at > NOW()`, code)
	if err != nil {
		return errors.New("invalid or expired code")
	}
	_, err = db.Exec(`INSERT INTO household_members (household_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, householdID, userID)
	return err
}

func LeaveHousehold(db *sqlx.DB, userID string) error {
	_, err := db.Exec(`DELETE FROM household_members WHERE user_id=$1`, userID)
	return err
}

func RemoveHouseholdMember(db *sqlx.DB, householdID int, actorUserID, targetUserID string) error {
	if actorUserID == targetUserID {
		return errors.New("cannot remove yourself")
	}
	_, err := db.Exec(`DELETE FROM household_members WHERE household_id=$1 AND user_id=$2`, householdID, targetUserID)
	return err
}

func GetHousehold(db *sqlx.DB, user *clerk.User) (*Household, error) {
	var household Household
	row := db.QueryRow(`SELECT h.id, h.name, ARRAY_AGG(m.user_id) as members
		FROM households h
		LEFT JOIN household_members m ON h.id = m.household_id
		WHERE m.user_id = $1
		GROUP BY h.id`, user.ID)

	err := row.Scan(&household.ID, &household.Name, pq.Array(&household.Members))
	if err != nil {
		return nil, err
	}

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
