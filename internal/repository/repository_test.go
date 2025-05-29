package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dBiTech/go-apiTemplate/internal/models"
	"github.com/dBiTech/go-apiTemplate/internal/repository"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

func TestMemoryRepository(t *testing.T) {
	log := logger.Default()
	repo := repository.NewMemoryRepository(log)

	ctx := context.Background()

	// Test CreateExample
	t.Run("CreateExample", func(t *testing.T) {
		id := uuid.New().String()
		example := &models.Example{
			BaseModel: models.BaseModel{
				ID:        id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Test Example",
			Description: "Test description",
			Status:      "active",
		}

		err := repo.CreateExample(ctx, example)
		require.NoError(t, err)

		// Test duplicate entry
		err = repo.CreateExample(ctx, example)
		assert.Equal(t, repository.ErrAlreadyExists, err)
	})

	// Test GetExample
	t.Run("GetExample", func(t *testing.T) {
		// Create example first
		id := uuid.New().String()
		example := &models.Example{
			BaseModel: models.BaseModel{
				ID:        id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Get Example",
			Description: "Test description",
			Status:      "active",
		}

		err := repo.CreateExample(ctx, example)
		require.NoError(t, err)

		// Get the created example
		retrieved, err := repo.GetExample(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, example.ID, retrieved.ID)
		assert.Equal(t, example.Name, retrieved.Name)
		assert.Equal(t, example.Description, retrieved.Description)

		// Test getting non-existent example
		_, err = repo.GetExample(ctx, "non-existent-id")
		assert.Equal(t, repository.ErrNotFound, err)
	})

	// Test ListExamples
	t.Run("ListExamples", func(t *testing.T) {
		// Clear repo and add some examples
		repo = repository.NewMemoryRepository(log)

		for i := 0; i < 5; i++ {
			id := uuid.New().String()
			example := &models.Example{
				BaseModel: models.BaseModel{
					ID:        id,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:        "List Example",
				Description: "Test description",
				Status:      "active",
			}

			err := repo.CreateExample(ctx, example)
			require.NoError(t, err)
		}

		// List examples
		examples, err := repo.ListExamples(ctx, 3, 0)
		require.NoError(t, err)
		assert.Len(t, examples, 3)

		// List with offset
		examples, err = repo.ListExamples(ctx, 3, 3)
		require.NoError(t, err)
		assert.Len(t, examples, 2)

		// List with no limit
		examples, err = repo.ListExamples(ctx, 0, 0)
		require.NoError(t, err)
		assert.Len(t, examples, 5)
	})

	// Test UpdateExample
	t.Run("UpdateExample", func(t *testing.T) {
		// Create example first
		id := uuid.New().String()
		example := &models.Example{
			BaseModel: models.BaseModel{
				ID:        id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Update Example",
			Description: "Original description",
			Status:      "active",
		}

		err := repo.CreateExample(ctx, example)
		require.NoError(t, err)

		// Update it
		example.Name = "Updated Example"
		example.Description = "Updated description"

		err = repo.UpdateExample(ctx, example)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetExample(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "Updated Example", retrieved.Name)
		assert.Equal(t, "Updated description", retrieved.Description)

		// Test updating non-existent example
		nonExistentExample := &models.Example{
			BaseModel: models.BaseModel{
				ID:        "non-existent-id",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		err = repo.UpdateExample(ctx, nonExistentExample)
		assert.Equal(t, repository.ErrNotFound, err)
	})

	// Test DeleteExample
	t.Run("DeleteExample", func(t *testing.T) {
		// Create example first
		id := uuid.New().String()
		example := &models.Example{
			BaseModel: models.BaseModel{
				ID:        id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Delete Example",
			Description: "Test description",
			Status:      "active",
		}

		err := repo.CreateExample(ctx, example)
		require.NoError(t, err)

		// Delete it
		err = repo.DeleteExample(ctx, id)
		require.NoError(t, err)

		// Verify it's gone
		_, err = repo.GetExample(ctx, id)
		assert.Equal(t, repository.ErrNotFound, err)

		// Test deleting non-existent example
		err = repo.DeleteExample(ctx, "non-existent-id")
		assert.Equal(t, repository.ErrNotFound, err)
	})

	// Test Ping
	t.Run("Ping", func(t *testing.T) {
		err := repo.Ping(ctx)
		require.NoError(t, err)
	})
}
