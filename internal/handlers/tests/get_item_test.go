package tests

import (
	"encoding/json"
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

func TestGetItem(t *testing.T) {
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
		wantErr    bool
	}{
		{
			name:       "Existing item",
			itemID:     testItem.ID,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Non-existent item",
			itemID:     uuid.New().String(),
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/items/%s", tt.itemID), nil)

			// Add URL parameters to request
			vars := map[string]string{
				"id": tt.itemID,
			}
			req = mux.SetURLVars(req, vars)

			w := httptest.NewRecorder()

			// Call handler
			h.GetItem(w, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantErr {
				var response models.Item
				err = json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, testItem.ID, response.ID)
				assert.Equal(t, testItem.Name, response.Name)
				assert.Equal(t, testItem.Value, response.Value)
			}
		})
	}
}
