package interfaces

import (
	"context"

	"github.com/angel/go-api-sqlite/internal/domain/models"
)

// FeatureUseCase defines the interface for feature business operations
type FeatureUseCase interface {
	CreateFeature(ctx context.Context, feature *models.Feature) error
	GetFeatureByID(ctx context.Context, id string) (*models.Feature, error)
	GetAllFeatures(ctx context.Context) ([]*models.Feature, error)
	UpdateFeature(ctx context.Context, feature *models.Feature) error
	DeleteFeature(ctx context.Context, id string) error
	ToggleFeature(ctx context.Context, id string, active bool) error
}
