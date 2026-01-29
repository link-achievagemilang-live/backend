package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize PostgreSQL
	db, err := initPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	redisClient := initRedis(cfg.Redis)
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repositories
	pgRepo := repository.NewPostgresRepository(db)
	redisRepo := repository.NewRedisRepository(redisClient)

	// Initialize services
	urlService := service.NewURLService(pgRepo, redisRepo, cfg.Server.BaseURL)

	// Initialize handlers
	urlHandler := handler.NewURLHandler(urlService)

	// Initialize rate limiter
	rateLimiter := handler.NewRateLimiter(cfg.RateLimit.RequestsPerMinute)

	// Setup router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", urlHandler.HealthCheck)

	// API endpoints
	mux.HandleFunc("/api/v1/urls", urlHandler.CreateShortURL)
	mux.HandleFunc("/api/v1/analytics/", urlHandler.GetAnalytics)

	// Redirect endpoint (catch-all for short codes)
	mux.HandleFunc("/", urlHandler.RedirectToOriginal)

	// Apply rate limiting only to the create URL endpoint
	rateLimitedMux := http.NewServeMux()
	rateLimitedMux.HandleFunc("/health", urlHandler.HealthCheck)
	rateLimitedMux.Handle("/api/v1/urls", handler.RateLimitMiddleware(rateLimiter)(http.HandlerFunc(urlHandler.CreateShortURL)))
	rateLimitedMux.HandleFunc("/api/v1/analytics/", urlHandler.GetAnalytics)
	rateLimitedMux.HandleFunc("/", urlHandler.RedirectToOriginal)

	// Apply global middleware
	finalHandler := handler.CORSMiddleware(
		handler.LoggingMiddleware(
			handler.RecoveryMiddleware(rateLimitedMux),
		),
	)

	// Create server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// initPostgres initializes PostgreSQL connection
func initPostgres(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Connected to PostgreSQL")
	return db, nil
}

// initRedis initializes Redis connection
func initRedis(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	log.Println("Connected to Redis")
	return client
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB) error {
	migration := `
		CREATE TABLE IF NOT EXISTS urls (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(8) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP,
			user_id UUID,
			click_count BIGINT NOT NULL DEFAULT 0,
			last_accessed TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
		CREATE INDEX IF NOT EXISTS idx_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
	`

	// Split by semicolon and execute each statement
	statements := strings.Split(migration, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}
