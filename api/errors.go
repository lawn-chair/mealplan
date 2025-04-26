package api

import (
	"encoding/json"
	"net/http"
)

func ErrorResponse(w http.ResponseWriter, err string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": err})
}
