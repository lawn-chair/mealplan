package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestListTagsHandler(t *testing.T) {
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

	req := httptest.NewRequest("GET", "/api/tags", nil)
	w := httptest.NewRecorder()

	// Inject db into context
	req = req.WithContext(context.WithValue(req.Context(), "db", sqlxDB))

	ListTagsHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.ElementsMatch(t, tags, resp)
	assert.NoError(t, mock.ExpectationsWereMet())
}
