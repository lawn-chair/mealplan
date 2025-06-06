package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	ErrorResponse(rec, "something went wrong", http.StatusBadRequest)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "something went wrong", resp["error"])
}
