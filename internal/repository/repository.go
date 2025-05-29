package repository

import (
	"context"
	"time"

	"github.com/dBiTech/go-apiTemplate/internal/models"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

// Repository defines the interface for data access
type Repository interface {
	// Examples
	GetExample(ctx context.Context, id string) (*models.Example, error)
	ListExamples(ctx context.Context, limit, offset int) ([]*models.Example, error)
	CreateExample(ctx context.Context, example *models.Example) error
	UpdateExample(ctx context.Context, example *models.Example) error
	DeleteExample(ctx context.Context, id string) error

	// Health check
	Ping(ctx context.Context) error
}

// MemoryRepository implements the Repository interface with in-memory storage
// This is just for the template, in a real app you would implement a database repository
type MemoryRepository struct {
	examples map[string]*models.Example
	log      logger.Logger
}

// NewMemoryRepository creates a new memory repository
func NewMemoryRepository(log logger.Logger) *MemoryRepository {
	return &MemoryRepository{
		examples: make(map[string]*models.Example),
		log:      log,
	}
}

// GetExample gets an example by ID
func (r *MemoryRepository) GetExample(ctx context.Context, id string) (*models.Example, error) {
	r.log.Debug("getting example", logger.String("id", id))

	if example, ok := r.examples[id]; ok {
		return example, nil
	}

	return nil, ErrNotFound
}

// ListExamples lists examples
func (r *MemoryRepository) ListExamples(ctx context.Context, limit, offset int) ([]*models.Example, error) {
	r.log.Debug("listing examples", logger.Int("limit", limit), logger.Int("offset", offset))

	examples := make([]*models.Example, 0, len(r.examples))

	i := 0
	for _, example := range r.examples {
		if i >= offset && (limit <= 0 || len(examples) < limit) {
			examples = append(examples, example)
		}
		i++
	}

	return examples, nil
}

// CreateExample creates a new example
func (r *MemoryRepository) CreateExample(ctx context.Context, example *models.Example) error {
	r.log.Debug("creating example", logger.String("id", example.ID))

	if _, ok := r.examples[example.ID]; ok {
		return ErrAlreadyExists
	}

	r.examples[example.ID] = example

	return nil
}

// UpdateExample updates an example
func (r *MemoryRepository) UpdateExample(ctx context.Context, example *models.Example) error {
	r.log.Debug("updating example", logger.String("id", example.ID))

	if _, ok := r.examples[example.ID]; !ok {
		return ErrNotFound
	}

	example.UpdatedAt = time.Now()
	r.examples[example.ID] = example

	return nil
}

// DeleteExample deletes an example
func (r *MemoryRepository) DeleteExample(ctx context.Context, id string) error {
	r.log.Debug("deleting example", logger.String("id", id))

	if _, ok := r.examples[id]; !ok {
		return ErrNotFound
	}

	delete(r.examples, id)

	return nil
}

// Ping checks database connectivity
func (r *MemoryRepository) Ping(ctx context.Context) error {
	// For memory repository, this always succeeds
	return nil
}
