package models

import (
	"github.com/jmoiron/sqlx"
)

// GetAllTags returns all unique tags in the tags table, sorted alphabetically
func GetAllTags(db *sqlx.DB) ([]string, error) {
	tags := []string{}
	err := db.Select(&tags, "SELECT name FROM tags ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	return tags, nil
}
