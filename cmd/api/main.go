// Package main provides the entry point for the Go API template application.
// This is the main command that starts the HTTP server with all configured middleware,
// handlers, and services for the API template.
package main

import (
	"os"

	"github.com/dBiTech/go-apiTemplate/internal/api"
	"github.com/dBiTech/go-apiTemplate/internal/config"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"

	// Swagger documentation
	_ "github.com/dBiTech/go-apiTemplate/docs"
)

// @title dBi Technologies API Template
// @version 1.0
// @description A production-ready Go API template following best practices and dBi Technologies API guidelines.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/dBiTech/go-apiTemplate
// @contact.email support@example.com

// @license.name MIT
// @license.url https://github.com/dBiTech/go-apiTemplate/blob/main/LICENSE

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer: ` prefix, e.g. "Bearer abcde12345".

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Default().Fatal("failed to load configuration", logger.Error(err))
	}

	// Create and start API server
	server, err := api.NewServer(cfg)
	if err != nil {
		logger.Default().Fatal("failed to create server", logger.Error(err))
	}

	// Run server
	if err := server.Run(); err != nil {
		logger.Default().Fatal("failed to run server", logger.Error(err))
		os.Exit(1)
	}
}
