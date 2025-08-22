package models

import "time"

// APIToken represents an API authentication token
type APIToken struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	TokenHash    string     `json:"-"` // Stored hashed token, not exposed in JSON
	LastUsedAt   *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	CreatedByUID string     `json:"createdByUID"`
}
