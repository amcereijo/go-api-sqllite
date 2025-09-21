package feature

import (
	"context"

	"github.com/angel/go-api-sqlite/internal/domain/interfaces"
	"github.com/angel/go-api-sqlite/internal/domain/models"
)

// UseCase implements the feature business logic
type UseCase struct {
	repo interfaces.FeatureRepository
}

// NewUseCase creates a new feature use case
func NewUseCase(repo interfaces.FeatureRepository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// CreateFeature validates and creates a new feature
func (uc *UseCase) CreateFeature(ctx context.Context, feature *models.Feature) error {
	if err := feature.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, feature)
}

// GetFeatureByID retrieves a feature by its ID
func (uc *UseCase) GetFeatureByID(ctx context.Context, id string) (*models.Feature, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetAllFeatures retrieves all features
func (uc *UseCase) GetAllFeatures(ctx context.Context) ([]*models.Feature, error) {
	return uc.repo.GetAll(ctx)
}

// UpdateFeature validates and updates a feature
func (uc *UseCase) UpdateFeature(ctx context.Context, feature *models.Feature) error {
	if err := feature.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, feature)
}

// DeleteFeature removes a feature
func (uc *UseCase) DeleteFeature(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// ToggleFeature changes the active status of a feature
func (uc *UseCase) ToggleFeature(ctx context.Context, id string, active bool) error {
	feature, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	feature.Active = active
	return uc.repo.Update(ctx, feature)
}
