package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/dBiTech/go-apiTemplate/internal/models"
	"github.com/dBiTech/go-apiTemplate/internal/repository"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"
	"github.com/dBiTech/go-apiTemplate/pkg/telemetry"
)

// Service provides business logic operations
type Service struct {
	repo repository.Repository
	log  logger.Logger
	tel  *telemetry.Telemetry
}

// New creates a new service instance
func New(repo repository.Repository, log logger.Logger, tel *telemetry.Telemetry) *Service {
	return &Service{
		repo: repo,
		log:  log,
		tel:  tel,
	}
}

// GetExample gets an example by ID
func (s *Service) GetExample(ctx context.Context, id string) (*models.Example, error) {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.GetExample")
	defer span.End()
	span.SetAttributes(attribute.String("example.id", id))

	s.log.Debug("getting example", logger.String("id", id))

	example, err := s.repo.GetExample(ctx, id)
	if err != nil {
		s.log.Error("failed to get example", logger.String("id", id), logger.Error(err))
		span.RecordError(err)
		return nil, err
	}

	return example, nil
}

// ListExamples lists examples
func (s *Service) ListExamples(ctx context.Context, limit, offset int) ([]*models.Example, error) {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.ListExamples")
	defer span.End()
	span.SetAttributes(attribute.Int("limit", limit), attribute.Int("offset", offset))

	s.log.Debug("listing examples", logger.Int("limit", limit), logger.Int("offset", offset))

	examples, err := s.repo.ListExamples(ctx, limit, offset)
	if err != nil {
		s.log.Error("failed to list examples", logger.Error(err))
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("count", len(examples)))
	return examples, nil
}

// CreateExample creates a new example
func (s *Service) CreateExample(ctx context.Context, req *models.ExampleRequest) (*models.Example, error) {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.CreateExample")
	defer span.End()
	span.SetAttributes(attribute.String("example.name", req.Name))

	s.log.Debug("creating example", logger.String("name", req.Name))

	// Generate a new UUID
	id := uuid.New().String()

	example := models.NewExample(id, req.Name, req.Description)

	if err := s.repo.CreateExample(ctx, example); err != nil {
		s.log.Error("failed to create example", logger.String("name", req.Name), logger.Error(err))
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.String("example.id", example.ID))
	return example, nil
}

// UpdateExample updates an existing example
func (s *Service) UpdateExample(ctx context.Context, id string, req *models.ExampleRequest) (*models.Example, error) {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.UpdateExample")
	defer span.End()
	span.SetAttributes(
		attribute.String("example.id", id),
		attribute.String("example.name", req.Name),
	)

	s.log.Debug("updating example",
		logger.String("id", id),
		logger.String("name", req.Name),
	)

	// Get existing example
	example, err := s.repo.GetExample(ctx, id)
	if err != nil {
		s.log.Error("failed to get example for update", logger.String("id", id), logger.Error(err))
		span.RecordError(err)
		return nil, err
	}

	// Update fields
	example.Name = req.Name
	example.Description = req.Description
	example.UpdatedAt = time.Now()

	if err := s.repo.UpdateExample(ctx, example); err != nil {
		s.log.Error("failed to update example", logger.String("id", id), logger.Error(err))
		span.RecordError(err)
		return nil, err
	}

	return example, nil
}

// DeleteExample deletes an example
func (s *Service) DeleteExample(ctx context.Context, id string) error {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.DeleteExample")
	defer span.End()
	span.SetAttributes(attribute.String("example.id", id))

	s.log.Debug("deleting example", logger.String("id", id))

	if err := s.repo.DeleteExample(ctx, id); err != nil {
		s.log.Error("failed to delete example", logger.String("id", id), logger.Error(err))
		span.RecordError(err)
		return err
	}

	return nil
}

// GetUserProfile gets a user profile by ID
func (s *Service) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.GetUserProfile")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID))

	s.log.Debug("getting user profile", logger.String("userID", userID))

	// This is a mock implementation. In a real app, you would fetch from a database
	profile := &models.UserProfile{
		ID:       userID,
		Username: "user" + userID,
		Email:    "user" + userID + "@example.com",
		Roles:    []string{"user"},
		Scopes:   []string{"read", "write"},
	}

	return profile, nil
}

// GetProtectedResource gets a protected resource by ID
func (s *Service) GetProtectedResource(ctx context.Context, id string) (*models.ProtectedResource, error) {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.GetProtectedResource")
	defer span.End()
	span.SetAttributes(attribute.String("resource.id", id))

	s.log.Debug("getting protected resource", logger.String("id", id))

	// This is a mock implementation. In a real app, you would fetch from a database
	resource := &models.ProtectedResource{
		ID:        id,
		Name:      "Protected Resource " + id,
		Content:   "This is a protected resource that requires authentication.",
		CreatedAt: time.Now(),
		OwnerID:   "user123",
	}

	return resource, nil
}

// ListProtectedResources lists protected resources
func (s *Service) ListProtectedResources(ctx context.Context) ([]*models.ProtectedResource, error) {
	ctx, span := s.tel.Tracer("service").Start(ctx, "Service.ListProtectedResources")
	defer span.End()

	s.log.Debug("listing protected resources")

	// This is a mock implementation. In a real app, you would fetch from a database
	resources := []*models.ProtectedResource{
		{
			ID:        uuid.New().String(),
			Name:      "Protected Resource 1",
			Content:   "This is protected resource 1.",
			CreatedAt: time.Now(),
			OwnerID:   "user123",
		},
		{
			ID:        uuid.New().String(),
			Name:      "Protected Resource 2",
			Content:   "This is protected resource 2.",
			CreatedAt: time.Now(),
			OwnerID:   "user456",
		},
	}

	return resources, nil
}
