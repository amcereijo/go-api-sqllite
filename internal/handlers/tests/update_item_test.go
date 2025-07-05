package tests

import (
	"bytes"
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

func TestUpdateItem(t *testing.T) {
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
		updates    models.Item
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "Valid update",
			itemID: testItem.ID,
			updates: models.Item{
				Name:  "Updated Item",
				Value: 39.99,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "Non-existent item",
			itemID: uuid.New().String(),
			updates: models.Item{
				Name:  "Updated Item",
				Value: 39.99,
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.updates)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/items/%s", tt.itemID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Add URL parameters to request
			vars := map[string]string{
				"id": tt.itemID,
			}
			req = mux.SetURLVars(req, vars)

			w := httptest.NewRecorder()

			// Call handler
			h.UpdateItem(w, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantErr {
				var response models.Item
				err = json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.itemID, response.ID)
				assert.Equal(t, tt.updates.Name, response.Name)
				assert.Equal(t, tt.updates.Value, response.Value)
			}
		})
	}
}
