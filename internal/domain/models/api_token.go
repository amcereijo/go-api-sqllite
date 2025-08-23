package models

import (
	"errors"
	"time"
)

// APIToken represents an API authentication token
type APIToken struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	TokenHash    string     `json:"-"` // Stored hashed token, not exposed in JSON
	LastUsedAt   *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	CreatedByUID string     `json:"createdByUID"`
}

// Validate validates the API token
func (t *APIToken) Validate() error {
	if t.Name == "" {
		return ErrEmptyTokenName
	}
	if t.CreatedByUID == "" {
		return ErrEmptyCreatedByUID
	}
	return nil
}

// ErrEmptyTokenName is returned when token name is empty
var ErrEmptyTokenName = errors.New("token name cannot be empty")

// ErrEmptyCreatedByUID is returned when created by UID is empty
var ErrEmptyCreatedByUID = errors.New("created by UID cannot be empty")

// ErrTokenNotFound is returned when a token is not found
var ErrTokenNotFound = errors.New("token not found")
