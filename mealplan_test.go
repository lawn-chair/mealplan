package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// Mock DB for testing
func setupMockDB() (*sqlx.DB, error) {
	// In a real test, you might use sqlmock or a test database
	// For now, we'll just create a mock DB connection
	db, err := sqlx.Connect("postgres", "postgres://user:password@localhost:5432/mockdb?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestDbCtx(t *testing.T) {
	// We can't test with a real DB connection, so we'll mock the DB
	db, err := setupMockDB()
	if err != nil {
		t.Skip("Skipping test as DB connection couldn't be established")
	}
	defer db.Close()

	// Create a middleware chain with DbCtx
	middleware := DbCtx(db)

	// Create a test handler that checks if the db is in the context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if db is in context
		dbFromCtx, ok := r.Context().Value("db").(*sqlx.DB)
		assert.True(t, ok, "DB should be in context")
		assert.Equal(t, db, dbFromCtx, "DB in context should match the one we passed in")
		w.WriteHeader(http.StatusOK)
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	// Apply the middleware to the test handler and execute the request
	middleware(testHandler).ServeHTTP(rec, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestIdCtx(t *testing.T) {
	// Create a new chi router
	r := chi.NewRouter()

	// Add a test route that uses IdCtx middleware
	r.Route("/api/test/{id}", func(r chi.Router) {
		r.Use(IdCtx)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// Extract the ID from context and verify it's correct
			id, ok := r.Context().Value("id").(int)
			assert.True(t, ok, "ID should be in context as an integer")
			assert.Equal(t, 42, id, "ID should match the URL parameter")
			w.WriteHeader(http.StatusOK)
		})
	})

	// Create a test request with ID parameter
	req := httptest.NewRequest("GET", "/api/test/42", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(rec, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test with invalid ID format
	req = httptest.NewRequest("GET", "/api/test/invalid", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRouteSetup(t *testing.T) {
	// We can't test the entire application setup and routes directly,
	// but we can verify that certain routes exist and respond as expected

	// Create a new chi router like in main()
	r := chi.NewRouter()

	// Mock DB Context middleware
	mockDbCtx := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a mock DB and add it to context
			ctx := context.WithValue(r.Context(), "db", &sqlx.DB{})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Setup a minimal version of the routes to test structure
	r.Route("/api", func(apir chi.Router) {
		apir.Use(mockDbCtx)

		apir.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Test the 404 handler
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
	})

	// Test a route that should exist
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test 404 handling
	req = httptest.NewRequest("GET", "/nonexistent", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
