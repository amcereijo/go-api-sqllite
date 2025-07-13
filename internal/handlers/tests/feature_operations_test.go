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

func TestFeatureOperations(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(db)

	// Create a test feature
	feature := models.Feature{
		ID:         uuid.New().String(),
		Name:       "Test Feature",
		Value:      json.RawMessage(`"test-value"`),
		ResourceID: "resource-1",
		Active:     true,
		CreatedAt:  time.Now(),
	}

	valueStr, err := feature.Value.MarshalJSON()
	assert.NoError(t, err)
	_, err = db.Exec(
		"INSERT INTO features (id, name, value, resource_id, active, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		feature.ID, feature.Name, string(valueStr), feature.ResourceID, feature.Active, feature.CreatedAt,
	)
	assert.NoError(t, err)

	t.Run("Get Feature", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/features/%s", feature.ID), nil)
		req = mux.SetURLVars(req, map[string]string{"id": feature.ID})
		w := httptest.NewRecorder()

		// Call handler
		h.GetFeature(w, req)

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
	})

	t.Run("Update Feature", func(t *testing.T) {
		// Create updated feature
		updatedFeature := models.Feature{
			ID:         feature.ID,
			Name:       "Updated Feature",
			Value:      json.RawMessage(`{"key":"updated-value"}`),
			ResourceID: "resource-2",
			Active:     false,
		}

		// Create request body
		body, err := json.Marshal(updatedFeature)
		assert.NoError(t, err)

		// Create request
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/features/%s", feature.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"id": feature.ID})
		w := httptest.NewRecorder()

		// Call handler
		h.UpdateFeature(w, req)

		// Check status code
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var resp models.Feature
		err = json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)

		// Check fields
		assert.Equal(t, updatedFeature.ID, resp.ID)
		assert.Equal(t, updatedFeature.Name, resp.Name)
		assert.Equal(t, updatedFeature.ResourceID, resp.ResourceID)
		assert.Equal(t, updatedFeature.Active, resp.Active)

		// Compare JSON values
		var expectedValue, actualValue interface{}
		err = json.Unmarshal(updatedFeature.Value, &expectedValue)
		assert.NoError(t, err)
		err = json.Unmarshal(resp.Value, &actualValue)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, actualValue)
	})

	t.Run("Delete Feature", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/features/%s", feature.ID), nil)
		req = mux.SetURLVars(req, map[string]string{"id": feature.ID})
		w := httptest.NewRecorder()

		// Call handler
		h.DeleteFeature(w, req)

		// Check status code
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify feature is deleted
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM features WHERE id = ?", feature.ID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
