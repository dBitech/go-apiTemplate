package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dBiTech/go-apiTemplate/internal/models"
	"github.com/dBiTech/go-apiTemplate/internal/repository"
	"github.com/dBiTech/go-apiTemplate/internal/service"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"
	"github.com/dBiTech/go-apiTemplate/pkg/telemetry"
)

// MockRepository is a mock implementation of repository.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetExample(ctx context.Context, id string) (*models.Example, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Example), args.Error(1)
}

func (m *MockRepository) ListExamples(ctx context.Context, limit, offset int) ([]*models.Example, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Example), args.Error(1)
}

func (m *MockRepository) CreateExample(ctx context.Context, example *models.Example) error {
	args := m.Called(ctx, example)
	return args.Error(0)
}

func (m *MockRepository) UpdateExample(ctx context.Context, example *models.Example) error {
	args := m.Called(ctx, example)
	return args.Error(0)
}

func (m *MockRepository) DeleteExample(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestService(t *testing.T) {
	log := logger.Default()

	tel, err := telemetry.New(context.Background(), telemetry.Config{
		ServiceName:    "test-service",
		ServiceVersion: "test",
		Environment:    "test",
		Endpoint:       "",
		Enabled:        false,
	}, log)
	require.NoError(t, err)

	mockRepo := new(MockRepository)
	svc := service.New(mockRepo, log, tel)

	ctx := context.Background()

	// Test GetExample
	t.Run("GetExample", func(t *testing.T) {
		id := uuid.New().String()
		expected := &models.Example{
			BaseModel: models.BaseModel{ID: id},
			Name:      "Test Example",
		}

		// Setup expectations
		mockRepo.On("GetExample", ctx, id).Return(expected, nil)

		// Call service method
		result, err := svc.GetExample(ctx, id)

		// Assert expectations
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	// Test GetExample with error
	t.Run("GetExample_Error", func(t *testing.T) {
		id := uuid.New().String()

		// Setup expectations
		mockRepo.On("GetExample", ctx, id).Return(nil, repository.ErrNotFound)

		// Call service method
		result, err := svc.GetExample(ctx, id)

		// Assert expectations
		require.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	// Test ListExamples
	t.Run("ListExamples", func(t *testing.T) {
		limit, offset := 10, 0
		expected := []*models.Example{
			{BaseModel: models.BaseModel{ID: uuid.New().String()}, Name: "Example 1"},
			{BaseModel: models.BaseModel{ID: uuid.New().String()}, Name: "Example 2"},
		}

		// Setup expectations
		mockRepo.On("ListExamples", ctx, limit, offset).Return(expected, nil)

		// Call service method
		result, err := svc.ListExamples(ctx, limit, offset)

		// Assert expectations
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	// Test CreateExample
	t.Run("CreateExample", func(t *testing.T) {
		req := &models.ExampleRequest{
			Name:        "New Example",
			Description: "Description",
		}

		// Setup expectations - using mock.Anything for ID since it's generated
		mockRepo.On("CreateExample", ctx, mock.Anything).Return(nil)

		// Call service method
		result, err := svc.CreateExample(ctx, req)

		// Assert expectations
		require.NoError(t, err)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.Description, result.Description)
		mockRepo.AssertExpectations(t)
	})

	// Test UpdateExample
	t.Run("UpdateExample", func(t *testing.T) {
		id := uuid.New().String()
		req := &models.ExampleRequest{
			Name:        "Updated Example",
			Description: "Updated Description",
		}

		existingExample := &models.Example{
			BaseModel:   models.BaseModel{ID: id},
			Name:        "Original Example",
			Description: "Original Description",
		}

		// Setup expectations
		mockRepo.On("GetExample", ctx, id).Return(existingExample, nil)
		mockRepo.On("UpdateExample", ctx, mock.Anything).Return(nil)

		// Call service method
		result, err := svc.UpdateExample(ctx, id, req)

		// Assert expectations
		require.NoError(t, err)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.Description, result.Description)
		mockRepo.AssertExpectations(t)
	})

	// Test DeleteExample
	t.Run("DeleteExample", func(t *testing.T) {
		id := uuid.New().String()

		// Setup expectations
		mockRepo.On("DeleteExample", ctx, id).Return(nil)

		// Call service method
		err := svc.DeleteExample(ctx, id)

		// Assert expectations
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
