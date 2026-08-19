package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

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

// ListenAndServe starts the HTTP server on addr (e.g. ":8080") and automatically opens the browser.
func (s *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if os.Getenv("AUTO_OPEN") != "false" {
		url := getURL(addr)
		go func() {
			time.Sleep(100 * time.Millisecond)
			openBrowser(url)
		}()
	}

	return http.Serve(ln, s.handler)
}

func getURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://localhost" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		slog.Warn("unsupported platform for auto-opening browser", "os", runtime.GOOS)
		return
	}
	if err := cmd.Start(); err != nil {
		slog.Warn("failed to open browser", "err", err)
	}
}
