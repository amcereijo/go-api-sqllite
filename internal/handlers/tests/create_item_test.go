package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angel/go-api-sqlite/internal/handlers"
	"github.com/angel/go-api-sqlite/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateItem(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(db)

	tests := []struct {
		name       string
		input      models.Item
		wantStatus int
		wantErr    bool
	}{
		{
			name: "Valid item",
			input: models.Item{
				Name:  "Test Item",
				Value: 29.99,
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "Missing name",
			input: models.Item{
				Value: 29.99,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/api/items", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler
			h.CreateItem(w, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantErr {
				var response models.Item
				err = json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.input.Name, response.Name)
				assert.Equal(t, tt.input.Value, response.Value)
				assert.NotZero(t, response.CreatedAt)
			}
		})
	}
}
