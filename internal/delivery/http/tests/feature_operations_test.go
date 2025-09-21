package tests

import (
	"bytes"
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

func TestFeatureOperations(t *testing.T) {
	// Setup
	mockUseCase := new(MockFeatureUseCase)
	handler := delivery.NewFeatureHandler(mockUseCase)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create a test feature
	testTime := time.Now()
	feature := &models.Feature{
		ID:         "test-id",
		Name:       "Test Feature",
		Value:      json.RawMessage(`"test-value"`),
		ResourceID: "resource-1",
		Active:     true,
		CreatedAt:  testTime,
	}

	t.Run("Get Feature", func(t *testing.T) {
		// Setup mock expectations
		mockUseCase.On("GetFeatureByID", mock.Anything, feature.ID).Return(feature, nil)

		// Create request
		req := httptest.NewRequest("GET", fmt.Sprintf("/features/%s", feature.ID), nil)
		w := httptest.NewRecorder()

		// Call handler through router
		router.ServeHTTP(w, req)

		// Check status code
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var resp models.Feature
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)

		// Check fields
		assert.Equal(t, feature.ID, resp.ID)
		assert.Equal(t, feature.Name, resp.Name)
		assert.Equal(t, feature.ResourceID, resp.ResourceID)
		assert.Equal(t, feature.Active, resp.Active)

		// Compare JSON values
		var expectedValue, actualValue interface{}
		err = json.Unmarshal(feature.Value, &expectedValue)
		assert.NoError(t, err)
		err = json.Unmarshal(resp.Value, &actualValue)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, actualValue)

		// Verify mock expectations
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Update Feature", func(t *testing.T) {
		// Create updated feature
		updatedFeature := &models.Feature{
			ID:         feature.ID,
			Name:       "Updated Feature",
			Value:      json.RawMessage(`{"key":"updated-value"}`),
			ResourceID: "resource-2",
			Active:     false,
		}

		// Setup mock expectations
		mockUseCase.On("UpdateFeature", mock.Anything, updatedFeature).Return(nil)

		// Create request
		jsonData, err := json.Marshal(updatedFeature)
		assert.NoError(t, err)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/features/%s", feature.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Call handler through router
		router.ServeHTTP(w, req)

		// Check status code
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify mock expectations
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Delete Feature", func(t *testing.T) {
		// Setup mock expectations
		mockUseCase.On("DeleteFeature", mock.Anything, feature.ID).Return(nil)

		// Create request
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/features/%s", feature.ID), nil)
		w := httptest.NewRecorder()

		// Call handler through router
		router.ServeHTTP(w, req)

		// Check status code
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify mock expectations
		mockUseCase.AssertExpectations(t)
	})
}
