package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("SERVER_HOST", "testhost")
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("DB_HOST", "testdb")
	os.Setenv("REDIS_HOST", "testredis")
	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("REDIS_HOST")
	}()

	cfg := Load()

	if cfg.Server.Host != "testhost" {
		t.Errorf("Expected server host 'testhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != "9999" {
		t.Errorf("Expected server port '9999', got '%s'", cfg.Server.Port)
	}

	if cfg.Database.Host != "testdb" {
		t.Errorf("Expected database host 'testdb', got '%s'", cfg.Database.Host)
	}

	if cfg.Redis.Host != "testredis" {
		t.Errorf("Expected redis host 'testredis', got '%s'", cfg.Redis.Host)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Clear all environment variables
	os.Clearenv()

	cfg := Load()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default server host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default server port '8080', got '%s'", cfg.Server.Port)
	}

	if cfg.Database.Port != "5432" {
		t.Errorf("Expected default database port '5432', got '%s'", cfg.Database.Port)
	}

	if cfg.Redis.Port != "6379" {
		t.Errorf("Expected default redis port '6379', got '%s'", cfg.Redis.Port)
	}

	if cfg.RateLimit.RequestsPerMinute != 10 {
		t.Errorf("Expected default rate limit 10, got %d", cfg.RateLimit.RequestsPerMinute)
	}
}
