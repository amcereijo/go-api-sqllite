package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	delivery "github.com/angel/go-api-sqlite/internal/delivery/http"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	handler := delivery.NewHealthHandler()

	// Create a new HTTP request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler directly
	handler.HealthCheck(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}
