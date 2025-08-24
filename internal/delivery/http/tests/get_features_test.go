package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	delivery "github.com/angel/go-api-sqlite/internal/delivery/http"
	"github.com/angel/go-api-sqlite/internal/domain/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetFeatures(t *testing.T) {
	// Setup
	mockUseCase := new(MockFeatureUseCase)
	handler := delivery.NewFeatureHandler(mockUseCase)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test features
	testTime := time.Now()
	features := []*models.Feature{
		{
			ID:         "1",
			Name:       "Feature 1",
			Value:      json.RawMessage(`"value-1"`),
			ResourceID: "resource-1",
			Active:     true,
			CreatedAt:  testTime,
		},
		{
			ID:         "2",
			Name:       "Feature 2",
			Value:      json.RawMessage(`42`),
			ResourceID: "resource-1",
			Active:     false,
			CreatedAt:  testTime,
		},
		{
			ID:         "3",
			Name:       "Feature 3",
			Value:      json.RawMessage(`{"key":"value"}`),
			ResourceID: "resource-2",
			Active:     true,
			CreatedAt:  testTime,
		},
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
			// Setup mock expectations
			mockUseCase.On("GetAllFeatures", mock.Anything).Return(features, nil)

			// Create request
			url := "/features"
			if tt.resourceID != "" {
				url = fmt.Sprintf("%s?resource_id=%s", url, tt.resourceID)
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Call handler through router
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			// Parse response
			var resp []*models.Feature
			err := json.NewDecoder(w.Body).Decode(&resp)
			assert.NoError(t, err)

			// Check response
			filteredFeatures := make([]*models.Feature, 0)
			if tt.resourceID == "" {
				filteredFeatures = features
			} else {
				for _, f := range features {
					if f.ResourceID == tt.resourceID {
						filteredFeatures = append(filteredFeatures, f)
					}
				}
			}
			assert.Equal(t, len(filteredFeatures), len(resp))

			for _, f := range resp {
				if tt.wantResourceID != "" {
					assert.Equal(t, tt.wantResourceID, f.ResourceID)
				}
			}

			// Verify mock expectations
			mockUseCase.AssertExpectations(t)
		})
	}
}
