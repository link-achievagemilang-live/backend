// Package domain defines the core domain models and entities.
package domain

import "time"

// URL represents a shortened URL entity
type URL struct {
	ID           int64      `json:"id"`
	ShortCode    string     `json:"short_code"`
	OriginalURL  string     `json:"original_url"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	UserID       *string    `json:"user_id,omitempty"`
	ClickCount   int64      `json:"click_count"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
}

// Analytics represents analytics data for a short URL
type Analytics struct {
	ShortCode    string     `json:"short_code"`
	ClickCount   int64      `json:"click_count"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
}

// CreateURLRequest represents the request to create a short URL
type CreateURLRequest struct {
	LongURL     string  `json:"long_url"`
	CustomAlias *string `json:"custom_alias,omitempty"`
	TTLDays     *int    `json:"ttl_days,omitempty"`
}

// CreateURLResponse represents the response after creating a short URL
type CreateURLResponse struct {
	ShortURL  string     `json:"short_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
