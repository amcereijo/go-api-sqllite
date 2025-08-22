package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/angel/go-api-sqlite/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	tokenLength = 32 // Length of the random token in bytes
)

// generateToken creates a new random API token
func generateToken() (string, string, error) {
	// Generate random bytes for the token
	tokenBytes := make([]byte, tokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	// Convert to base64 for the actual token
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create hash of the token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	return token, tokenHash, nil
}

// CreateAPIToken handles the creation of new API tokens
func (h *Handler) CreateAPIToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Token name is required", http.StatusBadRequest)
		return
	}

	// Generate token and hash
	token, tokenHash, err := generateToken()
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userId").(string)
	if !ok {
		http.Error(w, "Unauthorized - User not authenticated", http.StatusUnauthorized)
		return
	}

	// Create API token record
	apiToken := models.APIToken{
		ID:           uuid.New().String(),
		Name:         req.Name,
		TokenHash:    tokenHash,
		CreatedAt:    time.Now(),
		CreatedByUID: userID,
	}

	// Insert into database
	_, err = h.db.Exec(`
		INSERT INTO api_tokens (id, name, token_hash, created_at, created_by_uid)
		VALUES (?, ?, ?, ?, ?)`,
		apiToken.ID, apiToken.Name, apiToken.TokenHash, apiToken.CreatedAt, apiToken.CreatedByUID)

	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	// Return the token only once
	response := struct {
		models.APIToken
		Token string `json:"token"`
	}{
		APIToken: apiToken,
		Token:    token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListAPITokens returns all API tokens for the authenticated user
func (h *Handler) ListAPITokens(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userId").(string)
	if !ok {
		http.Error(w, "Unauthorized - User not authenticated", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, last_used_at, created_at, created_by_uid
		FROM api_tokens
		WHERE created_by_uid = ?
		ORDER BY created_at DESC`,
		userID)

	if err != nil {
		http.Error(w, "Error retrieving tokens", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tokens []models.APIToken
	for rows.Next() {
		var token models.APIToken

		err := rows.Scan(&token.ID, &token.Name, &token.LastUsedAt, &token.CreatedAt, &token.CreatedByUID)
		if err != nil {
			http.Error(w, "Error scanning tokens", http.StatusInternalServerError)
			return
		}

		tokens = append(tokens, token)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// DeleteAPIToken deletes an API token
func (h *Handler) DeleteAPIToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userId").(string)
	if !ok {
		http.Error(w, "Unauthorized - User not authenticated", http.StatusUnauthorized)
		return
	}

	tokenID := mux.Vars(r)["id"]

	result, err := h.db.Exec(`
		DELETE FROM api_tokens
		WHERE id = ? AND created_by_uid = ?`,
		tokenID, userID)

	if err != nil {
		http.Error(w, "Error deleting token", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AuthenticateToken validates an API token and returns true if valid
func (h *Handler) AuthenticateToken(token string) bool {
	if token == "" {
		return false
	}

	// Calculate token hash
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Check if token exists and update last_used_at
	result, err := h.db.Exec(`
		UPDATE api_tokens
		SET last_used_at = CURRENT_TIMESTAMP
		WHERE token_hash = ?`,
		tokenHash)

	if err != nil {
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}
