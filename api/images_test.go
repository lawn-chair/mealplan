package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/stretchr/testify/assert"
)

// Mock RequiresAuthentication for testing
func mockAuth(r *http.Request) (*clerk.User, error) {
	return &clerk.User{ID: "test-user"}, nil
}

func TestPostImageHandler_Unauthorized(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/images", nil)
	orig := RequiresAuthentication
	RequiresAuthentication = func(r *http.Request) (*clerk.User, error) { return nil, assert.AnError }
	defer func() { RequiresAuthentication = orig }()

	PostImageHandler(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPostImageHandler_BadRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/images", nil)
	orig := RequiresAuthentication
	RequiresAuthentication = mockAuth
	defer func() { RequiresAuthentication = orig }()

	PostImageHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostImageHandler_MissingEnvOrMinio(t *testing.T) {
	// This test checks that if env or minio fails, we get a 500 error
	// We can't patch minio.New, so we just check that the handler doesn't panic and returns 500 for a valid form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", "test.jpg")
	assert.NoError(t, err)
	_, err = fw.Write([]byte("fake image data"))
	assert.NoError(t, err)
	w.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/images", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	orig := RequiresAuthentication
	RequiresAuthentication = mockAuth
	defer func() { RequiresAuthentication = orig }()

	PostImageHandler(rec, req)
	// Should be 500 because minio.New will fail with default env
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
