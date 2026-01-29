package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHealthCheck(t *testing.T) {
	handler := &URLHandler{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}

func TestCreateShortURL_InvalidJSON(t *testing.T) {
	handler := &URLHandler{}

	invalidJSON := []byte(`{"long_url": invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(invalidJSON))
	w := httptest.NewRecorder()

	handler.CreateShortURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateShortURL_MissingLongURL(t *testing.T) {
	handler := &URLHandler{}

	requestBody := map[string]interface{}{
		"custom_alias": "test",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateShortURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetAnalytics_MissingShortCode(t *testing.T) {
	handler := &URLHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/", nil)
	w := httptest.NewRecorder()

	// Create a chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", "")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	handler.GetAnalytics(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
