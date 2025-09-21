package sqlite

import (
	"context"
	"database/sql"

	"github.com/angel/go-api-sqlite/internal/domain/models"
)

// TokenRepository implements domain.TokenRepository interface
type TokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates a new SQLite token repository
func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// Create inserts a new token into the database
func (r *TokenRepository) Create(ctx context.Context, token *models.APIToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO api_tokens (id, name, token_hash, created_at, created_by_uid)
		 VALUES (?, ?, ?, ?, ?)`,
		token.ID, token.Name, token.TokenHash, token.CreatedAt, token.CreatedByUID)
	return err
}

// GetAll retrieves all tokens
func (r *TokenRepository) GetAll(ctx context.Context) ([]*models.APIToken, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, token_hash, last_used_at, created_at, created_by_uid FROM api_tokens`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]*models.APIToken, 0)
	for rows.Next() {
		token := &models.APIToken{}
		err := rows.Scan(
			&token.ID,
			&token.Name,
			&token.TokenHash,
			&token.LastUsedAt,
			&token.CreatedAt,
			&token.CreatedByUID,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

// Delete removes a token from the database
func (r *TokenRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM api_tokens WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return models.ErrTokenNotFound
	}

	return nil
}
