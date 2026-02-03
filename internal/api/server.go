package api

import (
	"context"
	"fmt"
	"net/http"
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

	// Routes
	r.Get("/health", h.Health)

	r.Route("/api", func(r chi.Router) {
		r.Route("/nodes", func(r chi.Router) {
			r.Get("/", h.ListNodes)
			r.Get("/{nodeID}", h.GetNode)
			r.Post("/{nodeID}/position", h.SetPosition)
			r.Post("/{nodeID}/open", h.OpenNode)
			r.Post("/{nodeID}/close", h.CloseNode)
			r.Post("/{nodeID}/stop", h.StopNode)
		})
	})

	// Loxone-friendly endpoints (simple query parameters)
	r.Route("/loxone", func(r chi.Router) {
		r.Get("/node/{nodeID}/set/{position}", h.LoxoneSetPosition)
		r.Get("/node/{nodeID}/open", h.LoxoneOpen)
		r.Get("/node/{nodeID}/close", h.LoxoneClose)
		r.Get("/node/{nodeID}/stop", h.LoxoneStop)
	})

	return r
}
