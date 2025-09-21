package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/angel/go-api-sqlite/internal/domain/models"
	"github.com/angel/go-api-sqlite/internal/usecases/interfaces"
	"github.com/google/uuid"
)

// UseCase implements the token business logic
type UseCase struct {
	repo interfaces.TokenRepository
}

// NewUseCase creates a new token use case
func NewUseCase(repo interfaces.TokenRepository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// CreateAPIToken validates and creates a new API token
func (uc *UseCase) CreateAPIToken(ctx context.Context, token *models.APIToken) (string, error) {
	if err := token.Validate(); err != nil {
		return "", err
	}

	rawToken, tokenHash, err := generateToken()
	if err != nil {
		return "", err
	}

	token.ID = uuid.New().String()
	token.TokenHash = tokenHash
	token.CreatedAt = time.Now()

	uc.repo.Create(ctx, token)

	return rawToken, nil
}

// ListAPITokens retrieves all API tokens
func (uc *UseCase) ListAPITokens(ctx context.Context) ([]*models.APIToken, error) {
	return uc.repo.GetAll(ctx)
}

// DeleteAPIToken removes an API token
func (uc *UseCase) DeleteAPIToken(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

const (
	tokenLength = 32 // Length of the random token in bytes
)

func generateToken() (string, string, error) {
	// Generate random bytes for the token
	tokenBytes := make([]byte, tokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	// Convert to base64 for the actual token
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create hash of the token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	return token, tokenHash, nil
}
