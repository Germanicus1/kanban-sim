package middleware_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/middleware"
	"github.com/joho/godotenv"
)

func TestAuth(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Failed to load .env: %v", err)
	}

	token := os.Getenv("API_KEY")

	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		headerValue    string
		wantStatus     int
		expectNextCall bool
	}{
		{
			name:           "malformatted header (no space)",
			headerValue:    "Bearertoken123",
			wantStatus:     http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:           "no authorization header",
			headerValue:    "",
			wantStatus:     http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:           "wrong scheme prefix",
			headerValue:    "Token sometoken123",
			wantStatus:     http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:           "empty beearer token",
			headerValue:    "Bearer ",
			wantStatus:     http.StatusForbidden,
			expectNextCall: false,
		},
		{
			name:           "wrong token",
			headerValue:    "Bearer not-correct",
			wantStatus:     http.StatusForbidden,
			expectNextCall: false,
		},
		{
			name:           "correct token",
			headerValue:    "Bearer " + token,
			wantStatus:     http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "correct token, but with wrong prefix case",
			headerValue:    "bearer " + token,
			wantStatus:     http.StatusOK,
			expectNextCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called = false
			req := httptest.NewRequest("GET", "/any", nil)
			if tt.headerValue != "" {
				req.Header.Set("Authorization", tt.headerValue)
			}
			rec := httptest.NewRecorder()
			// Create a new instance of the middleware with the next handler
			handler := middleware.APIKeyAuth(next)
			// Serve the HTTP request using the middleware
			handler.ServeHTTP(rec, req)
			// Check the response status code and body
			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if tt.expectNextCall && !called {
				t.Error("expected next handler to be called, but it was not")
			}
			if !tt.expectNextCall && called {
				t.Error("expected next handler not to be called, but it was")
			}
			if rec.Body.String() != "OK" && tt.expectNextCall {
				t.Errorf("expected response body 'OK', got '%s'", rec.Body.String())
			}
			if rec.Body.String() == "OK" && !tt.expectNextCall {
				t.Errorf("expected no response body, got '%s'", rec.Body.String())
			}
		})
	}

}
