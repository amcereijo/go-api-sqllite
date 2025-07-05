package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/angel/go-api-sqlite/internal/handlers"
	"github.com/angel/go-api-sqlite/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDeleteItem(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(db)

	// Insert a test item
	testItem := models.Item{
		ID:        uuid.New().String(),
		Name:      "Test Item",
		Value:     29.99,
		CreatedAt: time.Now(),
	}

	_, err := db.Exec(
		"INSERT INTO items (id, name, value, created_at) VALUES (?, ?, ?, ?)",
		testItem.ID, testItem.Name, testItem.Value, testItem.CreatedAt,
	)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		itemID     string
		wantStatus int
	}{
		{
			name:       "Existing item",
			itemID:     testItem.ID,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Non-existent item",
			itemID:     uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/items/%s", tt.itemID), nil)

			// Add URL parameters to request
			vars := map[string]string{
				"id": tt.itemID,
			}
			req = mux.SetURLVars(req, vars)

			w := httptest.NewRecorder()

			// Call handler
			h.DeleteItem(w, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusNoContent {
				// Verify item was deleted
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM items WHERE id = ?", tt.itemID).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 0, count)
			}
		})
	}
}
