package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/angel/go-api-sqlite/internal/handlers"
	"github.com/angel/go-api-sqlite/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetItems(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(db)

	// Insert test items
	testItems := []models.Item{
		{
			ID:        uuid.New().String(),
			Name:      "Test Item 1",
			Value:     29.99,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Name:      "Test Item 2",
			Value:     39.99,
			CreatedAt: time.Now(),
		},
	}

	for _, item := range testItems {
		_, err := db.Exec(
			"INSERT INTO items (id, name, value, created_at) VALUES (?, ?, ?, ?)",
			item.ID, item.Name, item.Value, item.CreatedAt,
		)
		assert.NoError(t, err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/items", nil)
	w := httptest.NewRecorder()

	// Call handler
	h.GetItems(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Item
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, len(testItems))

	// Test empty response
	// Clear the table
	_, err = db.Exec("DELETE FROM items")
	assert.NoError(t, err)

	// Make another request
	w = httptest.NewRecorder()
	h.GetItems(w, req)

	// Assert empty array response
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 0)
}
