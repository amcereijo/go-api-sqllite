package http

import (
	"encoding/json"
	"net/http"

	"github.com/angel/go-api-sqlite/internal/domain/models"
	"github.com/angel/go-api-sqlite/internal/usecases/interfaces"
	"github.com/gorilla/mux"
)

// TokenHandler handles HTTP requests for API tokens
type TokenHandler struct {
	useCase interfaces.TokenUseCase
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(useCase interfaces.TokenUseCase) *TokenHandler {
	return &TokenHandler{
		useCase: useCase,
	}
}

// RegisterRoutes registers the token routes
func (h *TokenHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/tokens", h.CreateAPIToken).Methods("POST")
	router.HandleFunc("/tokens", h.ListAPITokens).Methods("GET")
	router.HandleFunc("/tokens/{id}", h.DeleteAPIToken).Methods("DELETE")
}

// CreateAPIToken handles API token creation
func (h *TokenHandler) CreateAPIToken(w http.ResponseWriter, r *http.Request) {
	var token models.APIToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userId").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	token.CreatedByUID = userID

	tokenValue, err := h.useCase.CreateAPIToken(r.Context(), &token)
	if err != nil {
		if err == models.ErrEmptyTokenName || err == models.ErrEmptyCreatedByUID {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		models.APIToken
		Token string `json:"token"`
	}{
		APIToken: token,
		Token:    tokenValue,
	}
	json.NewEncoder(w).Encode(response)
}

// ListAPITokens handles retrieving all API tokens
func (h *TokenHandler) ListAPITokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.useCase.ListAPITokens(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// DeleteAPIToken handles API token deletion
func (h *TokenHandler) DeleteAPIToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := h.useCase.DeleteAPIToken(r.Context(), vars["id"]); err != nil {
		if err == models.ErrTokenNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
