package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/handlers"
)

func TestNewRouter_Patterns(t *testing.T) {
	ah := handlers.NewAppHandler()
	// GameHandler can be constructed with a nil service since we only inspect routing, not handler execution
	gh := handlers.NewGameHandler(nil)
	router := NewRouter(ah, gh)

	mux, ok := router.(*http.ServeMux)
	if !ok {
		t.Fatalf("expected *http.ServeMux, got %T", router)
	}

	tests := []struct {
		name        string
		method      string
		target      string
		wantPattern string
	}{
		{"Home", "GET", "/", "GET /"},
		{"Ping", "GET", "/ping", "GET /ping"},
		{"OpenAPI", "GET", "/openapi.yaml", "GET /openapi.yaml"},
		{"DocsRedirect", "GET", "/docs", "GET /docs"},
		{"DocsPrefix", "GET", "/docs/foo.txt", "GET /docs/"},

		{"CreateGame", "POST", "/games", "POST /games"},
		{"GetGame", "GET", "/games/123", "GET /games/{id}"},
		{"GetBoard", "GET", "/games/123/board", "GET /games/{id}/board"},
		{"UpdateGame", "PATCH", "/games/123", "PATCH /games/{id}"},
		{"DeleteGame", "DELETE", "/games/123", "DELETE /games/{id}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			h, pattern := mux.Handler(req)
			if pattern != tt.wantPattern {
				t.Errorf("for %s, pattern = %q; want %q", tt.name, pattern, tt.wantPattern)
			}
			if h == nil {
				t.Errorf("handler for %s is nil", tt.name)
			}
		})
	}
}
