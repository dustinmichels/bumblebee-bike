package server

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server wraps the HTTP server and its router.
type Server struct {
	handler http.Handler
}

// New builds a Server with the embedded static FS for the frontend.
func New(staticFS fs.FS) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API — add sub-routers here as the backend grows.
	r.Mount("/api", apiRouter())

	// SPA catch-all: serves static assets, falls back to index.html.
	r.Handle("/*", spaHandler(staticFS))

	return &Server{handler: r}
}

// ListenAndServe starts the HTTP server on addr (e.g. ":8080").
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.handler)
}
