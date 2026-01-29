package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testIP = "192.168.1.1:1234"

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2) // 2 requests per minute

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	req1.RemoteAddr = testIP
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", w1.Code)
	}

	// Second request should succeed
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	req2.RemoteAddr = testIP
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second request: expected status 200, got %d", w2.Code)
	}

	// Third request should be rate limited
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	req3.RemoteAddr = testIP
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Third request: expected status 429, got %d", w3.Code)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(1)

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request from IP 1
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	req1.RemoteAddr = testIP
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("IP1 request: expected status 200, got %d", w1.Code)
	}

	// Request from IP 2 should succeed (different IP)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("IP2 request: expected status 200, got %d", w2.Code)
	}
}

func TestCORS(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/urls", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header not set correctly")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
