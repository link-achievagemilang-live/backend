package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"url-shortener/internal/domain"
	"url-shortener/internal/repository"
)

// URLService handles business logic for URL operations
type URLService struct {
	pgRepo    *repository.PostgresRepository
	redisRepo *repository.RedisRepository
	baseURL   string
}

// NewURLService creates a new URL service
func NewURLService(pgRepo *repository.PostgresRepository, redisRepo *repository.RedisRepository, baseURL string) *URLService {
	return &URLService{
		pgRepo:    pgRepo,
		redisRepo: redisRepo,
		baseURL:   baseURL,
	}
}

// ShortenURL creates a shortened URL
func (s *URLService) ShortenURL(ctx context.Context, req *domain.CreateURLRequest) (*domain.CreateURLResponse, error) {
	// Validate URL format
	if !isValidURL(req.LongURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	var shortCode string
	var expiresAt *time.Time

	// Calculate expiration time if TTL is provided
	if req.TTLDays != nil && *req.TTLDays > 0 {
		expiry := time.Now().Add(time.Duration(*req.TTLDays) * 24 * time.Hour)
		expiresAt = &expiry
	}

	// Handle custom alias
	if req.CustomAlias != nil && *req.CustomAlias != "" {
		// Validate custom alias (alphanumeric only, 3-20 chars)
		if !isValidCustomAlias(*req.CustomAlias) {
			return nil, fmt.Errorf("invalid custom alias: must be 3-20 alphanumeric characters")
		}

		// Check if custom alias already exists
		exists, err := s.pgRepo.CheckShortCodeExists(ctx, *req.CustomAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom alias: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("custom alias already exists")
		}

		shortCode = *req.CustomAlias
	} else {
		// Generate short code using Base62 encoding
		id, err := s.pgRepo.GetNextID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}
		shortCode = Encode(id)
	}

	// Create URL entity
	urlEntity := &domain.URL{
		ShortCode:   shortCode,
		OriginalURL: req.LongURL,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
	}

	// Save to database
	err := s.pgRepo.CreateURL(ctx, urlEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Cache in Redis
	cacheTTL := 24 * time.Hour // Default cache TTL
	if expiresAt != nil {
		cacheTTL = time.Until(*expiresAt)
	}

	err = s.redisRepo.Set(ctx, shortCode, req.LongURL, cacheTTL)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to cache URL in Redis: %v", err)
	}

	// Build response
	shortURL := fmt.Sprintf("%s/%s", s.baseURL, shortCode)
	return &domain.CreateURLResponse{
		ShortURL:  shortURL,
		ExpiresAt: expiresAt,
	}, nil
}

// GetOriginalURL retrieves the original URL for a short code (cache-first)
func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// Try cache first
	originalURL, err := s.redisRepo.Get(ctx, shortCode)
	if err == nil {
		log.Printf("Cache hit for short code: %s", shortCode)
		// Asynchronously increment click count
		go s.incrementClickCountAsync(shortCode)
		return originalURL, nil
	}

	log.Printf("Cache miss for short code: %s", shortCode)

	// Fallback to database
	urlEntity, err := s.pgRepo.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("URL not found")
	}

	// Populate cache for future requests
	cacheTTL := 24 * time.Hour
	if urlEntity.ExpiresAt != nil {
		cacheTTL = time.Until(*urlEntity.ExpiresAt)
	}

	err = s.redisRepo.Set(ctx, shortCode, urlEntity.OriginalURL, cacheTTL)
	if err != nil {
		log.Printf("Failed to populate cache: %v", err)
	}

	// Asynchronously increment click count
	go s.incrementClickCountAsync(shortCode)

	return urlEntity.OriginalURL, nil
}

// GetAnalytics retrieves analytics for a short code
func (s *URLService) GetAnalytics(ctx context.Context, shortCode string) (*domain.Analytics, error) {
	analytics, err := s.pgRepo.GetAnalytics(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("analytics not found")
	}
	return analytics, nil
}

// incrementClickCountAsync increments click count asynchronously
func (s *URLService) incrementClickCountAsync(shortCode string) {
	ctx := context.Background()
	err := s.pgRepo.IncrementClickCount(ctx, shortCode)
	if err != nil {
		log.Printf("Failed to increment click count for %s: %v", shortCode, err)
	}
}

// isValidURL checks if a string is a valid URL
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// isValidCustomAlias checks if a custom alias is valid
func isValidCustomAlias(alias string) bool {
	if len(alias) < 3 || len(alias) > 20 {
		return false
	}
	for _, c := range alias {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
