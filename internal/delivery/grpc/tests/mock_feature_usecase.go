package tests

import (
	"context"

	"github.com/angel/go-api-sqlite/internal/domain/models"
	"github.com/stretchr/testify/mock"
)

type MockFeatureUseCase struct {
	mock.Mock
}

func (m *MockFeatureUseCase) CreateFeature(ctx context.Context, feature *models.Feature) error {
	args := m.Called(ctx, feature)
	return args.Error(0)
}

func (m *MockFeatureUseCase) GetFeatureByID(ctx context.Context, id string) (*models.Feature, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Feature), args.Error(1)
}

func (m *MockFeatureUseCase) GetAllFeatures(ctx context.Context) ([]*models.Feature, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Feature), args.Error(1)
}

func (m *MockFeatureUseCase) UpdateFeature(ctx context.Context, feature *models.Feature) error {
	args := m.Called(ctx, feature)
	return args.Error(0)
}

func (m *MockFeatureUseCase) DeleteFeature(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFeatureUseCase) ToggleFeature(ctx context.Context, id string, active bool) error {
	args := m.Called(ctx, id, active)
	return args.Error(0)
}
