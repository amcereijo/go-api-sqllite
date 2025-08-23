package interfaces

import (
	"context"

	"github.com/angel/go-api-sqlite/internal/domain/models"
)

// FeatureRepository defines the interface for feature data operations
type FeatureRepository interface {
	Create(ctx context.Context, feature *models.Feature) error
	GetByID(ctx context.Context, id string) (*models.Feature, error)
	GetAll(ctx context.Context) ([]*models.Feature, error)
	Update(ctx context.Context, feature *models.Feature) error
	Delete(ctx context.Context, id string) error
}
