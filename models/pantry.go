package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Pantry struct {
	ID     uint     `db:"id" json:"id"`
	UserID string   `db:"user_id" json:"user_id"`
	Items  []string `json:"items"`
}

func GetPantry(db *sqlx.DB, userID string) (*Pantry, error) {
	pantry := Pantry{}

	err := db.Get(&pantry, "SELECT * FROM pantry WHERE user_id = $1", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pantry, nil
		}
		fmt.Println("1", err)
		return nil, err
	}

	err = db.Select(&pantry.Items, "SELECT item_name FROM pantry_items WHERE pantry_id = $1", pantry.ID)
	if err != nil {
		fmt.Println("2", err)
		return nil, err
	}

	return &pantry, nil
}

func CreatePantry(db *sqlx.DB, userId string, pantry *Pantry) (*Pantry, error) {
	pantry.UserID = userId

	_, err := db.Exec(`INSERT INTO pantry (user_id) VALUES ($1)`, pantry.UserID)
	if err != nil {
		fmt.Println(err)
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Handle unique constraint violation
			fmt.Println("pgError", pgErr)
			return UpdatePantry(db, userId, pantry)
		}
		return nil, err
	}

	return UpdatePantry(db, userId, pantry)
}

func UpdatePantry(db *sqlx.DB, userId string, pantry *Pantry) (*Pantry, error) {

	user_pantry, err := GetPantry(db, userId)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("DELETE FROM pantry_items WHERE pantry_id=$1", user_pantry.ID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, item := range pantry.Items {
		_, err = db.Exec("INSERT INTO pantry_items (pantry_id, item_name) VALUES ($1, $2)", user_pantry.ID, strings.ToLower(item))
		if err != nil {
			var pgErr *pq.Error
			// Handle unique constraint violation
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				fmt.Println("pgError", pgErr)
			} else {
				fmt.Println(err)
				return nil, err
			}
		}
	}

	return GetPantry(db, userId)
}

func DeletePantry(db *sqlx.DB, userId string) error {
	pantry, err := GetPantry(db, userId)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM pantry_items WHERE pantry_id=$1", pantry.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
