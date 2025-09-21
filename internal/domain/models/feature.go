package models

import (
	"encoding/json"
	"errors"
	"time"
)

// Feature represents a feature flag in the database
type Feature struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Value      json.RawMessage `json:"value"`
	ResourceID string          `json:"resourceId"`
	Active     bool            `json:"active"`
	CreatedAt  time.Time       `json:"createdAt"`
}

// Validate validates the Feature entity
func (f *Feature) Validate() error {
	if f.Name == "" {
		return ErrEmptyFeatureName
	}
	if f.ResourceID == "" {
		return ErrEmptyResourceID
	}
	return nil
}

// ErrEmptyFeatureName is returned when feature name is empty
var ErrEmptyFeatureName = errors.New("feature name cannot be empty")

// ErrEmptyResourceID is returned when resource ID is empty
var ErrEmptyResourceID = errors.New("resource ID cannot be empty")

// ErrFeatureNotFound is returned when a feature is not found
var ErrFeatureNotFound = errors.New("feature not found")
