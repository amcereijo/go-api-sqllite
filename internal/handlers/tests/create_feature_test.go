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

func TestCreateFeature(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(db)

	tests := []struct {
		name       string
		input      models.Feature
		wantStatus int
		wantErr    bool
	}{
		{
			name: "Valid feature with string value",
			input: models.Feature{
				Name:       "Test Feature",
				Value:      json.RawMessage(`"test-value"`),
				ResourceID: "resource-1",
				Active:     true,
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "Valid feature with number value",
			input: models.Feature{
				Name:       "Number Feature",
				Value:      json.RawMessage(`42`),
				ResourceID: "resource-1",
				Active:     true,
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "Valid feature with object value",
			input: models.Feature{
				Name:       "Object Feature",
				Value:      json.RawMessage(`{"key":"value","enabled":true}`),
				ResourceID: "resource-1",
				Active:     true,
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "Missing name",
			input: models.Feature{
				Value:      json.RawMessage(`"test-value"`),
				ResourceID: "resource-1",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "Missing resourceId",
			input: models.Feature{
				Name:  "Test Feature",
				Value: json.RawMessage(`"test-value"`),
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
			req := httptest.NewRequest("POST", "/api/features", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler
			h.CreateFeature(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantErr {
				// Parse response
				var resp models.Feature
				err = json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)

				// Check fields
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, tt.input.Name, resp.Name)
				assert.Equal(t, tt.input.ResourceID, resp.ResourceID)
				assert.Equal(t, tt.input.Active, resp.Active)
				assert.NotZero(t, resp.CreatedAt)

				// Compare JSON values
				var expectedValue, actualValue interface{}
				err = json.Unmarshal(tt.input.Value, &expectedValue)
				assert.NoError(t, err)
				err = json.Unmarshal(resp.Value, &actualValue)
				assert.NoError(t, err)
				assert.Equal(t, expectedValue, actualValue)
			}
		})
	}
}
