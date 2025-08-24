package http

import (
	"encoding/json"
	"net/http"

	"github.com/angel/go-api-sqlite/internal/domain/models"
	"github.com/angel/go-api-sqlite/internal/usecases/interfaces"
	"github.com/gorilla/mux"
)

// FeatureHandler handles HTTP requests for features
type FeatureHandler struct {
	useCase interfaces.FeatureUseCase
}

// NewFeatureHandler creates a new feature handler
func NewFeatureHandler(useCase interfaces.FeatureUseCase) *FeatureHandler {
	return &FeatureHandler{
		useCase: useCase,
	}
}

// RegisterRoutes registers the feature routes
func (h *FeatureHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/features", h.CreateFeature).Methods("POST")
	router.HandleFunc("/features", h.GetFeatures).Methods("GET")
	router.HandleFunc("/features/{id}", h.GetFeature).Methods("GET")
	router.HandleFunc("/features/{id}", h.UpdateFeature).Methods("PUT")
	router.HandleFunc("/features/{id}", h.DeleteFeature).Methods("DELETE")
	router.HandleFunc("/features/{id}/toggle", h.ToggleFeature).Methods("POST")
}

// CreateFeature handles feature creation
func (h *FeatureHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
	var feature models.Feature
	if err := json.NewDecoder(r.Body).Decode(&feature); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.useCase.CreateFeature(r.Context(), &feature); err != nil {
		if err == models.ErrEmptyFeatureName || err == models.ErrEmptyResourceID {
			writeError(w, err, http.StatusBadRequest)
			return
		}
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(feature)
}

// GetFeatures handles retrieving all features
func (h *FeatureHandler) GetFeatures(w http.ResponseWriter, r *http.Request) {
	features, err := h.useCase.GetAllFeatures(r.Context())
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	resourceID := r.URL.Query().Get("resource_id")
	if resourceID != "" {
		filteredFeatures := make([]*models.Feature, 0)
		for _, f := range features {
			if f.ResourceID == resourceID {
				filteredFeatures = append(filteredFeatures, f)
			}
		}
		features = filteredFeatures
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(features)
}

// GetFeature handles retrieving a single feature
func (h *FeatureHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	feature, err := h.useCase.GetFeatureByID(r.Context(), vars["id"])
	if err != nil {
		if err == models.ErrFeatureNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feature)
}

// UpdateFeature handles feature updates
func (h *FeatureHandler) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var feature models.Feature
	if err := json.NewDecoder(r.Body).Decode(&feature); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	feature.ID = vars["id"]
	if err := h.useCase.UpdateFeature(r.Context(), &feature); err != nil {
		if err == models.ErrFeatureNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == models.ErrEmptyFeatureName || err == models.ErrEmptyResourceID {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feature)
}

// DeleteFeature handles feature deletion
func (h *FeatureHandler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := h.useCase.DeleteFeature(r.Context(), vars["id"]); err != nil {
		if err == models.ErrFeatureNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// ToggleFeature handles feature activation/deactivation
func (h *FeatureHandler) ToggleFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var body struct {
		Active bool `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.useCase.ToggleFeature(r.Context(), vars["id"], body.Active); err != nil {
		if err == models.ErrFeatureNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
