package tests

import (
	"bytes"
	"encoding/json"
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

func TestCreateFeature(t *testing.T) {
	// Setup
	testTime := time.Now()
	mockUseCase := new(MockFeatureUseCase)
	handler := delivery.NewFeatureHandler(mockUseCase)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

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
			// Setup expectations
			if tt.wantErr {
				if tt.input.Name == "" {
					mockUseCase.On("CreateFeature", mock.Anything, &tt.input).Return(models.ErrEmptyFeatureName)
				} else if tt.input.ResourceID == "" {
					mockUseCase.On("CreateFeature", mock.Anything, &tt.input).Return(models.ErrEmptyResourceID)
				}
			} else {
				mockUseCase.On("CreateFeature", mock.Anything, &tt.input).Run(func(args mock.Arguments) {
					feature := args.Get(1).(*models.Feature)
					feature.ID = "test-id"
					feature.CreatedAt = testTime
				}).Return(nil)
			}

			// Create request body
			jsonData, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Create request and execute
			req := httptest.NewRequest("POST", "/features", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler through router
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.wantStatus, w.Code)

			// Verify mock expectations
			mockUseCase.AssertExpectations(t)

			if tt.wantErr {
				var errorResponse struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(w.Body).Decode(&errorResponse)
				assert.NoError(t, err)

				if tt.input.Name == "" {
					assert.Contains(t, errorResponse.Error, "feature name cannot be empty")
				} else if tt.input.ResourceID == "" {
					assert.Contains(t, errorResponse.Error, "resource ID cannot be empty")
				}
			}

			// Check response body
			if !tt.wantErr {
				var responseFeature models.Feature
				err := json.NewDecoder(w.Body).Decode(&responseFeature)
				assert.NoError(t, err)
				assert.Equal(t, "test-id", responseFeature.ID)
				assert.Equal(t, tt.input.Name, responseFeature.Name)
				assert.Equal(t, tt.input.ResourceID, responseFeature.ResourceID)
				assert.Equal(t, tt.input.Active, responseFeature.Active)
				assert.True(t, testTime.Equal(responseFeature.CreatedAt))

				var expectedValue, actualValue interface{}
				err = json.Unmarshal(tt.input.Value, &expectedValue)
				assert.NoError(t, err)
				err = json.Unmarshal(responseFeature.Value, &actualValue)
				assert.NoError(t, err)
				assert.Equal(t, expectedValue, actualValue)

			}
		})
	}
}
