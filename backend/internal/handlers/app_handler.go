// internal/handlers/app_handler.go
package handlers

import (
	"net/http"
)

// AppHandler handles non-domain routes: home page, health check, docs, etc.
type AppHandler struct{}

// NewAppHandler constructs your AppHandler, wiring in any static FS you need.
func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

// Home serves GET /
func (h *AppHandler) Home(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, "./static/index.html") // or use embed if you like
	enableCORS(w)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to Kanban-Sim!"))
}

// Ping serves GET /ping
func (h *AppHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// OpenAPI serves GET /openapi.yaml
func (h *AppHandler) OpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./docs/openapi.yaml")
}

// DocsRedirect handles GET /docs → /docs/
func (h *AppHandler) DocsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
}

// Docs serves everything under GET /docs/*
func (h *AppHandler) Docs(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))).ServeHTTP(w, r)
}
