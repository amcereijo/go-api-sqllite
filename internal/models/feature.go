package models

import (
	"encoding/json"
	"time"
)

// Feature represents a feature flag in the database
type Feature struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Value      json.RawMessage `json:"value"`
	ResourceID string         `json:"resourceId"`
	Active     bool           `json:"active"`
	CreatedAt  time.Time      `json:"created_at"`
}
