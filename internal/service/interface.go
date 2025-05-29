// Package service provides business logic and service layer implementations.
// It defines service interfaces and handles the core application logic between handlers and repositories.
package service

import (
	"context"

	"github.com/dBiTech/go-apiTemplate/internal/models"
)

// Interface defines methods for the service layer
type Interface interface {
	// Examples
	GetExample(ctx context.Context, id string) (*models.Example, error)
	ListExamples(ctx context.Context, limit, offset int) ([]*models.Example, error)
	CreateExample(ctx context.Context, req *models.ExampleRequest) (*models.Example, error)
	UpdateExample(ctx context.Context, id string, req *models.ExampleRequest) (*models.Example, error)
	DeleteExample(ctx context.Context, id string) error

	// Protected Resources
	ListProtectedResources(ctx context.Context) ([]*models.ProtectedResource, error)

	// User Profile
	GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error)
}

// Ensure Service implements Interface
var _ Interface = (*Service)(nil)
