package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"url-shortener/internal/domain"

	_ "github.com/lib/pq"
)

// PostgresRepository handles database operations for URLs
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateURL inserts a new URL into the database
func (r *PostgresRepository) CreateURL(ctx context.Context, url *domain.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, created_at, expires_at, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		url.ShortCode,
		url.OriginalURL,
		url.CreatedAt,
		url.ExpiresAt,
		url.UserID,
	).Scan(&url.ID, &url.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

// GetURLByShortCode retrieves a URL by its short code
func (r *PostgresRepository) GetURLByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, user_id, click_count, last_accessed
		FROM urls
		WHERE short_code = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`

	url := &domain.URL{}
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.UserID,
		&url.ClickCount,
		&url.LastAccessed,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("URL not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return url, nil
}

// CheckShortCodeExists checks if a short code already exists
func (r *PostgresRepository) CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check short code existence: %w", err)
	}

	return exists, nil
}

// IncrementClickCount increments the click count for a URL
func (r *PostgresRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	query := `
		UPDATE urls
		SET click_count = click_count + 1, last_accessed = $1
		WHERE short_code = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	return nil
}

// GetAnalytics retrieves analytics data for a short code
func (r *PostgresRepository) GetAnalytics(ctx context.Context, shortCode string) (*domain.Analytics, error) {
	query := `
		SELECT short_code, click_count, last_accessed
		FROM urls
		WHERE short_code = $1
	`

	analytics := &domain.Analytics{}
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&analytics.ShortCode,
		&analytics.ClickCount,
		&analytics.LastAccessed,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("URL not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

// DeleteExpiredURLs removes expired URLs from the database
func (r *PostgresRepository) DeleteExpiredURLs(ctx context.Context) error {
	query := `DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired URLs: %w", err)
	}

	return nil
}

// GetNextID gets the next available ID (for Base62 encoding)
func (r *PostgresRepository) GetNextID(ctx context.Context) (int64, error) {
	query := `SELECT nextval('urls_id_seq')`

	var id int64
	err := r.db.QueryRowContext(ctx, query).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get next ID: %w", err)
	}

	return id, nil
}
