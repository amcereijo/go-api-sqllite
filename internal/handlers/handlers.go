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

// CreateItem handles POST requests to create a new item
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling CreateItem request from %s", r.RemoteAddr)
	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if item.Name == "" {
		log.Printf("Invalid request: name is required")
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Generate UUID for new item
	item.ID = uuid.New().String()
	item.CreatedAt = time.Now()

	// Insert into database
	_, err := h.db.Exec("INSERT INTO items (id, name, value, created_at) VALUES (?, ?, ?, ?)",
		item.ID, item.Name, item.Value, item.CreatedAt)
	if err != nil {
		log.Printf("Error inserting item into database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully created item with ID: %s", item.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// GetItems handles GET requests to retrieve all items
func (h *Handler) GetItems(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GetItems request from %s", r.RemoteAddr)
	rows, err := h.db.Query("SELECT id, name, value, created_at FROM items")
	if err != nil {
		log.Printf("Error querying items: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Value, &item.CreatedAt); err != nil {
			log.Printf("Error scanning item row: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}
	log.Printf("Successfully retrieved %d items", len(items))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GetItem handles GET requests to retrieve a specific item
func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling GetItem request for ID: %s from %s", id, r.RemoteAddr)

	var item models.Item
	err := h.db.QueryRow("SELECT id, name, value, created_at FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Value, &item.CreatedAt)

	if err == sql.ErrNoRows {
		log.Printf("Item not found with ID: %s", id)
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error retrieving item with ID %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully retrieved item with ID: %s", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// UpdateItem handles PUT requests to update an existing item
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling UpdateItem request for ID: %s from %s", id, r.RemoteAddr)

	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("UPDATE items SET name = ?, value = ? WHERE id = ?",
		item.Name, item.Value, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("No item found to update with ID: %s", id)
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	log.Printf("Successfully updated item with ID: %s", id)
	item.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DeleteItem handles DELETE requests to remove an item
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling DeleteItem request for ID: %s from %s", id, r.RemoteAddr)

	result, err := h.db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("No item found to delete with ID: %s", id)
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	log.Printf("Successfully deleted item with ID: %s", id)
	w.WriteHeader(http.StatusNoContent)
}
