package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/dBiTech/go-apiTemplate/internal/auth"
	"github.com/dBiTech/go-apiTemplate/internal/config"
	"github.com/dBiTech/go-apiTemplate/internal/handlers"
	appmiddleware "github.com/dBiTech/go-apiTemplate/internal/middleware"
	"github.com/dBiTech/go-apiTemplate/internal/repository"
	"github.com/dBiTech/go-apiTemplate/internal/service"
	"github.com/dBiTech/go-apiTemplate/pkg/health"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"
	"github.com/dBiTech/go-apiTemplate/pkg/metrics"
	"github.com/dBiTech/go-apiTemplate/pkg/telemetry"
)

const (
	appName        = "api-template"
	appVersion     = "0.1.0"
	appDescription = "API Template Application"
)

// Server represents the API server
type Server struct {
	config     *config.Config
	router     *chi.Mux
	httpServer *http.Server
	log        logger.Logger
	metrics    *metrics.Metrics
	telemetry  *telemetry.Telemetry
	health     *health.HealthCheck
	auth       *auth.Authenticator
}

// NewServer creates a new API server
func NewServer(cfg *config.Config) (*Server, error) {
	// Initialize logger
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	log.Info("initializing api server",
		logger.String("version", appVersion),
		logger.String("config", cfg.String()),
	)

	// Initialize metrics
	m := metrics.NewMetrics(appName)

	// Initialize telemetry
	tel, err := telemetry.New(context.Background(), telemetry.Config{
		ServiceName:    appName,
		ServiceVersion: appVersion,
		Environment:    "development", // TODO: Make configurable
		Endpoint:       cfg.Tracing.Endpoint,
		Enabled:        cfg.Tracing.Enabled,
	}, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry: %w", err)
	}

	// Initialize health check
	healthCheck := health.NewHealthCheck(appName, appVersion, appDescription)

	// Initialize authenticator
	authenticator, err := auth.NewAuthenticator(auth.AuthConfig{
		JWTSecret:          cfg.Auth.JWTSecret,
		JWTSigningMethod:   cfg.Auth.JWTSigningMethod,
		JWTExpirationTime:  cfg.Auth.JWTExpirationTime,
		JWTIssuer:          cfg.Auth.JWTIssuer,
		OAuth2ClientID:     cfg.Auth.OAuth2ClientID,
		OAuth2ClientSecret: cfg.Auth.OAuth2ClientSecret,
		OAuth2RedirectURL:  cfg.Auth.OAuth2RedirectURL,
		OAuth2AuthURL:      cfg.Auth.OAuth2AuthURL,
		OAuth2TokenURL:     cfg.Auth.OAuth2TokenURL,
		OAuth2Scopes:       cfg.Auth.OAuth2Scopes,
	}, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Initialize router
	router := chi.NewRouter()

	// Initialize server
	server := &Server{
		config:    cfg,
		router:    router,
		log:       log,
		metrics:   m,
		telemetry: tel,
		health:    healthCheck,
		auth:      authenticator,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	// Create repository
	repo := repository.NewMemoryRepository(s.log)

	// Create service
	svc := service.New(repo, s.log, s.telemetry)

	// Create handler
	handler := handlers.NewHandler(s.log, svc)

	// Add health check for database
	s.health.AddCheck(health.DBCheck("database", repo.Ping))

	// Middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(appmiddleware.RequestLogger(s.log))
	s.router.Use(appmiddleware.Tracing(s.telemetry))
	s.router.Use(appmiddleware.Metrics(s.metrics))
	s.router.Use(appmiddleware.Recover(s.log))
	s.router.Use(appmiddleware.CORS([]string{"*"})) // TODO: Make configurable

	// Health routes
	s.router.Get("/health", s.health.HealthHandler())
	s.router.Get("/health/liveness", s.health.LivenessHandler())
	s.router.Get("/health/readiness", s.health.ReadinessHandler())

	// Swagger UI route
	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Metrics route
	if s.config.Metrics.Enabled {
		s.router.Get("/metrics", s.metrics.Handler().ServeHTTP)
	}

	// API routes
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/hello", handler.HelloHandler())

		r.Route("/examples", func(r chi.Router) {
			r.Get("/", handler.ListExamplesHandler())
			r.Post("/", handler.CreateExampleHandler())
			r.Get("/{id}", handler.GetExampleHandler())
			r.Put("/{id}", handler.UpdateExampleHandler())
			r.Delete("/{id}", handler.DeleteExampleHandler())
		})

		// JWT protected route
		r.Route("/protected/jwt", func(r chi.Router) {
			// Apply JWT authentication middleware with required 'read' scope
			r.Use(s.auth.JWTAuthMiddleware([]string{"read"}))
			r.Get("/", handler.JWTProtectedResourceHandler())
		})

		// OAuth2 protected route
		r.Route("/protected/oauth2", func(r chi.Router) {
			// Apply OAuth2 authentication middleware with required 'read' scope
			r.Use(s.auth.OAuth2AuthMiddleware([]string{"read"}))
			r.Get("/", handler.OAuth2ProtectedResourceHandler())
		})

		// User profile route (requires either JWT or OAuth2)
		r.Route("/me", func(r chi.Router) {
			// This demonstrates how to use different auth methods for the same endpoint
			r.With(s.auth.JWTAuthMiddleware(nil)).Get("/", handler.UserProfileHandler())
			r.With(s.auth.OAuth2AuthMiddleware(nil)).Get("/oauth2", handler.UserProfileHandler())
		})
	})
}

// Start starts the API server
func (s *Server) Start() error {
	// Start server in a goroutine
	go func() {
		s.log.Info("starting server", logger.String("address", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatal("server failed", logger.Error(err))
		}
	}()

	return nil
}

// Stop gracefully stops the API server
func (s *Server) Stop() {
	s.log.Info("stopping server")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("server shutdown failed", logger.Error(err))
	}

	// Shutdown telemetry
	if err := s.telemetry.Shutdown(ctx); err != nil {
		s.log.Error("telemetry shutdown failed", logger.Error(err))
	}

	s.log.Info("server stopped")
}

// GetRouter returns the router for testing
func (s *Server) GetRouter() *chi.Mux {
	return s.router
}

// GetAuthenticator returns the authenticator for testing
func (s *Server) GetAuthenticator() *auth.Authenticator {
	return s.auth
}

// Run runs the API server until it receives a signal to stop
func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal is received
	sig := <-quit
	s.log.Info("received signal, shutting down server", logger.String("signal", sig.String()))

	// Shutdown server
	s.Stop()

	return nil
}
