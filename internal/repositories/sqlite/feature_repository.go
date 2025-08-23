package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/angel/go-api-sqlite/internal/domain/models"
	"github.com/google/uuid"
)

// FeatureRepository implements the domain.FeatureRepository interface
type FeatureRepository struct {
	db *sql.DB
}

// NewFeatureRepository creates a new SQLite feature repository
func NewFeatureRepository(db *sql.DB) *FeatureRepository {
	return &FeatureRepository{
		db: db,
	}
}

// Create inserts a new feature into the database
func (r *FeatureRepository) Create(ctx context.Context, feature *models.Feature) error {
	feature.ID = uuid.New().String()
	feature.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO features (id, name, value, resource_id, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		feature.ID, feature.Name, feature.Value, feature.ResourceID, feature.Active, feature.CreatedAt)
	return err
}

// GetByID retrieves a feature by its ID
func (r *FeatureRepository) GetByID(ctx context.Context, id string) (*models.Feature, error) {
	feature := &models.Feature{}
	var valueBytes []byte

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, value, resource_id, active, created_at
		 FROM features WHERE id = ?`, id).
		Scan(&feature.ID, &feature.Name, &valueBytes, &feature.ResourceID, &feature.Active, &feature.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, models.ErrFeatureNotFound
	}
	if err != nil {
		return nil, err
	}

	feature.Value = json.RawMessage(valueBytes)
	return feature, nil
}

// GetAll retrieves all features
func (r *FeatureRepository) GetAll(ctx context.Context) ([]*models.Feature, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, value, resource_id, active, created_at FROM features`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)
	for rows.Next() {
		feature := &models.Feature{}
		var valueBytes []byte

		err := rows.Scan(&feature.ID, &feature.Name, &valueBytes, &feature.ResourceID, &feature.Active, &feature.CreatedAt)
		if err != nil {
			return nil, err
		}

		feature.Value = json.RawMessage(valueBytes)
		features = append(features, feature)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return features, nil
}

// Update updates an existing feature
func (r *FeatureRepository) Update(ctx context.Context, feature *models.Feature) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE features
		 SET name = ?, value = ?, resource_id = ?, active = ?
		 WHERE id = ?`,
		feature.Name, feature.Value, feature.ResourceID, feature.Active, feature.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return models.ErrFeatureNotFound
	}

	return nil
}

// Delete removes a feature from the database
func (r *FeatureRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM features WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return models.ErrFeatureNotFound
	}

	return nil
}
