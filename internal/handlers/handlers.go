package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/angel/go-api-sqlite/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler holds the database connection
type Handler struct {
	db *sql.DB
}

// NewHandler creates a new handler with the database connection
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// HealthCheck handles the health check endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// CreateFeature handles POST requests to create a new feature flag
func (h *Handler) CreateFeature(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling CreateFeature request from %s", r.RemoteAddr)
	var feature models.Feature
	if err := json.NewDecoder(r.Body).Decode(&feature); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if feature.Name == "" || feature.ResourceID == "" {
		log.Printf("Invalid request: name and resourceId are required")
		http.Error(w, "name and resourceId are required", http.StatusBadRequest)
		return
	}

	// Generate UUID for new feature
	feature.ID = uuid.New().String()
	feature.CreatedAt = time.Now()
	if !feature.Active {
		feature.Active = true // Set default to true
	}

	valueStr, err := feature.Value.MarshalJSON()
	if err != nil {
		log.Printf("Error marshaling value: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert into database
	_, err = h.db.Exec("INSERT INTO features (id, name, value, resource_id, active, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		feature.ID, feature.Name, string(valueStr), feature.ResourceID, feature.Active, feature.CreatedAt)
	if err != nil {
		log.Printf("Error inserting feature into database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully created feature with ID: %s", feature.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(feature)
}

// GetFeatures handles GET requests to retrieve all features
func (h *Handler) GetFeatures(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GetFeatures request from %s", r.RemoteAddr)

	resourceID := r.URL.Query().Get("resourceId")
	var rows *sql.Rows
	var err error

	if resourceID != "" {
		rows, err = h.db.Query("SELECT id, name, value, resource_id, active, created_at FROM features WHERE resource_id = ?", resourceID)
	} else {
		rows, err = h.db.Query("SELECT id, name, value, resource_id, active, created_at FROM features")
	}

	if err != nil {
		log.Printf("Error querying features: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	features := make([]models.Feature, 0)
	for rows.Next() {
		var feature models.Feature
		var valueStr string
		if err := rows.Scan(&feature.ID, &feature.Name, &valueStr, &feature.ResourceID, &feature.Active, &feature.CreatedAt); err != nil {
			log.Printf("Error scanning feature row: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		feature.Value = json.RawMessage(valueStr)
		features = append(features, feature)
	}
	log.Printf("Successfully retrieved %d features", len(features))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(features)
}

// GetFeature handles GET requests to retrieve a specific feature
func (h *Handler) GetFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling GetFeature request for ID: %s from %s", id, r.RemoteAddr)

	var feature models.Feature
	var valueStr string
	err := h.db.QueryRow("SELECT id, name, value, resource_id, active, created_at FROM features WHERE id = ?", id).
		Scan(&feature.ID, &feature.Name, &valueStr, &feature.ResourceID, &feature.Active, &feature.CreatedAt)

	if err == sql.ErrNoRows {
		log.Printf("Feature not found with ID: %s", id)
		http.Error(w, "Feature not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error retrieving feature with ID %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feature.Value = json.RawMessage(valueStr)
	log.Printf("Successfully retrieved feature with ID: %s", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feature)
}

// UpdateFeature handles PUT requests to update an existing feature
func (h *Handler) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling UpdateFeature request for ID: %s from %s", id, r.RemoteAddr)

	var feature models.Feature
	if err := json.NewDecoder(r.Body).Decode(&feature); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	feature.ID = id // Set ID from URL parameter

	// Validate required fields
	if feature.Name == "" || feature.ResourceID == "" {
		log.Printf("Invalid request: name and resourceId are required")
		http.Error(w, "name and resourceId are required", http.StatusBadRequest)
		return
	}

	valueStr, err := feature.Value.MarshalJSON()
	if err != nil {
		log.Printf("Error marshaling value: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("UPDATE features SET name = ?, value = ?, resource_id = ?, active = ? WHERE id = ?",
		feature.Name, string(valueStr), feature.ResourceID, feature.Active, id)
	if err != nil {
		log.Printf("Error updating feature: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("No feature found to update with ID: %s", id)
		http.Error(w, "Feature not found", http.StatusNotFound)
		return
	}

	log.Printf("Successfully updated feature with ID: %s", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feature)
}

// DeleteFeature handles DELETE requests to remove a feature
func (h *Handler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling DeleteFeature request for ID: %s from %s", id, r.RemoteAddr)

	result, err := h.db.Exec("DELETE FROM features WHERE id = ?", id)
	if err != nil {
		log.Printf("Error deleting feature: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("No feature found to delete with ID: %s", id)
		http.Error(w, "Feature not found", http.StatusNotFound)
		return
	}

	log.Printf("Successfully deleted feature with ID: %s", id)
	w.WriteHeader(http.StatusNoContent)
}
