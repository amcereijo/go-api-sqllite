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
	"github.com/stretchr/testify/assert"
)

func TestGetFeatures(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(db)

	// Insert test features
	features := []models.Feature{
		{
			ID:         uuid.New().String(),
			Name:       "Feature 1",
			Value:      json.RawMessage(`"value-1"`),
			ResourceID: "resource-1",
			Active:     true,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New().String(),
			Name:       "Feature 2",
			Value:      json.RawMessage(`42`),
			ResourceID: "resource-1",
			Active:     false,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New().String(),
			Name:       "Feature 3",
			Value:      json.RawMessage(`{"key":"value"}`),
			ResourceID: "resource-2",
			Active:     true,
			CreatedAt:  time.Now(),
		},
	}

	for _, f := range features {
		valueStr, err := f.Value.MarshalJSON()
		assert.NoError(t, err)
		_, err = db.Exec(
			"INSERT INTO features (id, name, value, resource_id, active, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			f.ID, f.Name, string(valueStr), f.ResourceID, f.Active, f.CreatedAt,
		)
		assert.NoError(t, err)
	}

	tests := []struct {
		name           string
		resourceID     string
		wantCount      int
		wantResourceID string
		wantStatus     int
	}{
		{
			name:       "Get all features",
			resourceID: "",
			wantCount:  3,
			wantStatus: http.StatusOK,
		},
		{
			name:           "Get features by resource ID",
			resourceID:     "resource-1",
			wantCount:      2,
			wantResourceID: "resource-1",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "Get features by non-existent resource ID",
			resourceID:     "non-existent",
			wantCount:      0,
			wantResourceID: "non-existent",
			wantStatus:     http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			url := "/api/features"
			if tt.resourceID != "" {
				url = fmt.Sprintf("%s?resourceId=%s", url, tt.resourceID)
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Call handler
			h.GetFeatures(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			// Parse response
			var resp []models.Feature
			err := json.NewDecoder(w.Body).Decode(&resp)
			assert.NoError(t, err)

			// Check response
			assert.Len(t, resp, tt.wantCount)
			if tt.wantResourceID != "" {
				for _, f := range resp {
					assert.Equal(t, tt.wantResourceID, f.ResourceID)
				}
			}
		})
	}
}
