package models

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestGetAllTags(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tags := []string{"chicken", "quick", "dinner"}
	rows := sqlmock.NewRows([]string{"name"})
	for _, tag := range tags {
		rows.AddRow(tag)
	}
	mock.ExpectQuery("SELECT name FROM tags").WillReturnRows(rows)

	result, err := GetAllTags(sqlxDB)
	assert.NoError(t, err)
	assert.ElementsMatch(t, tags, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
