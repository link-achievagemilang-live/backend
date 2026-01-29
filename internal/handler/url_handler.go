package handler

import (
	"encoding/json"
	"net/http"

	"url-shortener/internal/domain"
	"url-shortener/internal/service"
)

// URLHandler handles HTTP requests for URL operations
type URLHandler struct {
	urlService *service.URLService
}

// NewURLHandler creates a new URL handler
func NewURLHandler(urlService *service.URLService) *URLHandler {
	return &URLHandler{urlService: urlService}
}

// CreateShortURL handles POST /api/v1/urls
func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if req.LongURL == "" {
		respondWithError(w, http.StatusBadRequest, "long_url is required")
		return
	}

	// Create short URL
	resp, err := h.urlService.ShortenURL(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

// RedirectToOriginal handles GET /{short_code}
func (h *URLHandler) RedirectToOriginal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract short code from path
	shortCode := r.URL.Path[1:] // Remove leading "/"
	if shortCode == "" || shortCode == "api" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Get original URL
	originalURL, err := h.urlService.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		// Redirect to frontend error page
		http.Redirect(w, r, "/not-found", http.StatusSeeOther)
		return
	}

	// Redirect with 302 (temporary) to track analytics
	http.Redirect(w, r, originalURL, http.StatusFound)
}

// GetAnalytics handles GET /api/v1/analytics/{short_code}
func (h *URLHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract short code from path
	// Path format: /api/v1/analytics/{short_code}
	shortCode := r.URL.Path[len("/api/v1/analytics/"):]
	if shortCode == "" {
		respondWithError(w, http.StatusBadRequest, "short_code is required")
		return
	}

	// Get analytics
	analytics, err := h.urlService.GetAnalytics(r.Context(), shortCode)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Analytics not found")
		return
	}

	respondWithJSON(w, http.StatusOK, analytics)
}

// HealthCheck handles GET /health
func (h *URLHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
