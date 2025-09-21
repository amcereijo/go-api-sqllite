package interfaces

import (
	"context"

	"github.com/angel/go-api-sqlite/internal/domain/models"
)

// TokenRepository defines the interface for token data operations
type TokenRepository interface {
	Create(ctx context.Context, token *models.APIToken) error
	GetAll(ctx context.Context) ([]*models.APIToken, error)
	Delete(ctx context.Context, id string) error
}

// TokenUseCase defines the interface for token business operations
type TokenUseCase interface {
	CreateAPIToken(ctx context.Context, token *models.APIToken) (string, error)
	ListAPITokens(ctx context.Context) ([]*models.APIToken, error)
	DeleteAPIToken(ctx context.Context, id string) error
}
