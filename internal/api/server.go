package api

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/stefanbeyeler/loxone2velux/internal/config"
	"github.com/stefanbeyeler/loxone2velux/internal/gateway"
)

// Server represents the HTTP API server
type Server struct {
	cfg     *config.ServerConfig
	gateway *gateway.Service
	logger  zerolog.Logger
	server  *http.Server
}

// NewServer creates a new API server
func NewServer(cfg *config.ServerConfig, gw *gateway.Service, logger zerolog.Logger) *Server {
	return &Server{
		cfg:     cfg,
		gateway: gw,
		logger:  logger.With().Str("component", "api").Logger(),
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	router := s.setupRoutes()

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}

	s.logger.Info().Str("addr", addr).Msg("Starting HTTP server")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(NewLoggingMiddleware(s.logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// Handlers
	h := NewHandlers(s.gateway, s.logger)

	// Public routes (no auth required)
	r.Get("/health", h.Health)

	// Auth status endpoint - tells frontend if auth is required
	r.Get("/api/auth/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if s.cfg.APIToken == "" {
			w.Write([]byte(`{"required":false}`))
		} else {
			w.Write([]byte(`{"required":true}`))
		}
	})

	// API routes - protected only if token is configured
	r.Route("/api", func(r chi.Router) {
		if s.cfg.APIToken != "" {
			r.Use(NewTokenAuthMiddleware(s.cfg.APIToken, s.logger))
		}
		r.Route("/nodes", func(r chi.Router) {
			r.Get("/", h.ListNodes)
			r.Get("/{nodeID}", h.GetNode)
			r.Post("/{nodeID}/position", h.SetPosition)
			r.Post("/{nodeID}/open", h.OpenNode)
			r.Post("/{nodeID}/close", h.CloseNode)
			r.Post("/{nodeID}/stop", h.StopNode)
		})
	})

	// Loxone-friendly endpoints - protected only if token is configured
	// Token via query param: /loxone/node/1/open?token=YOUR_TOKEN
	r.Route("/loxone", func(r chi.Router) {
		if s.cfg.APIToken != "" {
			r.Use(NewTokenAuthMiddleware(s.cfg.APIToken, s.logger))
		}
		r.Get("/node/{nodeID}/set/{position}", h.LoxoneSetPosition)
		r.Get("/node/{nodeID}/open", h.LoxoneOpen)
		r.Get("/node/{nodeID}/close", h.LoxoneClose)
		r.Get("/node/{nodeID}/stop", h.LoxoneStop)
	})

	// Static files for web frontend
	staticDir := "./web/dist"
	if _, err := os.Stat(staticDir); err == nil {
		s.logger.Info().Str("dir", staticDir).Msg("Serving static files")
		fileServer(r, "/", http.Dir(staticDir))
	} else {
		s.logger.Debug().Msg("No static files directory found, skipping frontend")
	}

	return r
}

// fileServer sets up a http.FileServer handler to serve static files from a directory
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fsPath := strings.TrimPrefix(r.URL.Path, pathPrefix)

		// Try to open the file
		f, err := root.Open(fsPath)
		if err != nil {
			// If file not found, serve index.html for SPA routing
			if os.IsNotExist(err) {
				indexPath := filepath.Join(".", "index.html")
				if f, err = root.Open(indexPath); err != nil {
					http.NotFound(w, r)
					return
				}
			} else {
				http.NotFound(w, r)
				return
			}
		}
		defer f.Close()

		// Check if it's a directory
		stat, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if stat.IsDir() {
			// Try to serve index.html from directory
			indexPath := filepath.Join(fsPath, "index.html")
			if f, err = root.Open(indexPath); err != nil {
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			stat, _ = f.Stat()
		}

		// Serve the file
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), f.(fs.File).(interface {
			Seek(offset int64, whence int) (int64, error)
			Read(p []byte) (n int, err error)
		}).(http.File))
	})
}
