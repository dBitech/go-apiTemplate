package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dBiTech/go-apiTemplate/internal/handlers"
	"github.com/dBiTech/go-apiTemplate/internal/models"
	"github.com/dBiTech/go-apiTemplate/internal/repository"
	"github.com/dBiTech/go-apiTemplate/internal/service"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

// MockService mocks the service layer for testing handlers
type MockService struct {
	mock.Mock
}

// Ensure MockService implements service.Interface
var _ service.Interface = (*MockService)(nil)

func (m *MockService) GetExample(ctx context.Context, id string) (*models.Example, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Example), args.Error(1)
}

func (m *MockService) ListExamples(ctx context.Context, limit, offset int) ([]*models.Example, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Example), args.Error(1)
}

func (m *MockService) CreateExample(ctx context.Context, req *models.ExampleRequest) (*models.Example, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Example), args.Error(1)
}

func (m *MockService) UpdateExample(ctx context.Context, id string, req *models.ExampleRequest) (*models.Example, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Example), args.Error(1)
}

func (m *MockService) DeleteExample(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) ListProtectedResources(ctx context.Context) ([]*models.ProtectedResource, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProtectedResource), args.Error(1)
}

func (m *MockService) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserProfile), args.Error(1)
}

func TestHandlers(t *testing.T) {
	log := logger.Default()
	mockService := new(MockService)
	handler := handlers.NewHandler(log, mockService)

	// Test HelloHandler
	t.Run("HelloHandler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
		w := httptest.NewRecorder()

		handler.HelloHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", resp["message"])
	})

	// Test GetExampleHandler
	t.Run("GetExampleHandler", func(t *testing.T) {
		id := uuid.New().String()
		example := &models.Example{
			BaseModel: models.BaseModel{ID: id},
			Name:      "Test Example",
		}

		// Create a new request with the ID as a URL parameter
		req := httptest.NewRequest(http.MethodGet, "/api/v1/examples/"+id, nil)

		// Create a new Chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)

		// Set the route context
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Set up mock expectations
		mockService.On("GetExample", mock.Anything, id).Return(example, nil)

		handler.GetExampleHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.Example
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, id, resp.ID)
		assert.Equal(t, example.Name, resp.Name)
	})

	// Test ListExamplesHandler
	t.Run("ListExamplesHandler", func(t *testing.T) {
		examples := []*models.Example{
			{BaseModel: models.BaseModel{ID: uuid.New().String()}, Name: "Example 1"},
			{BaseModel: models.BaseModel{ID: uuid.New().String()}, Name: "Example 2"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/examples?limit=10&offset=0", nil)
		w := httptest.NewRecorder()

		// Set up mock expectations
		mockService.On("ListExamples", mock.Anything, 10, 0).Return(examples, nil)

		handler.ListExamplesHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []*models.Example
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, examples[0].ID, resp[0].ID)
		assert.Equal(t, examples[1].Name, resp[1].Name)
	})

	// Test CreateExampleHandler
	t.Run("CreateExampleHandler", func(t *testing.T) {
		id := uuid.New().String()
		reqBody := models.ExampleRequest{
			Name:        "New Example",
			Description: "Test Description",
		}

		example := &models.Example{
			BaseModel:   models.BaseModel{ID: id},
			Name:        reqBody.Name,
			Description: reqBody.Description,
		}

		reqBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/examples", bytes.NewBuffer(reqBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Set up mock expectations
		mockService.On("CreateExample", mock.Anything, mock.MatchedBy(func(r *models.ExampleRequest) bool {
			return r.Name == reqBody.Name && r.Description == reqBody.Description
		})).Return(example, nil)

		handler.CreateExampleHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.Example
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, id, resp.ID)
		assert.Equal(t, reqBody.Name, resp.Name)
		assert.Equal(t, reqBody.Description, resp.Description)
	})

	// Test UpdateExampleHandler
	t.Run("UpdateExampleHandler", func(t *testing.T) {
		id := uuid.New().String()
		reqBody := models.ExampleRequest{
			Name:        "Updated Example",
			Description: "Updated Description",
		}

		example := &models.Example{
			BaseModel:   models.BaseModel{ID: id},
			Name:        reqBody.Name,
			Description: reqBody.Description,
		}

		reqBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/examples/"+id, bytes.NewBuffer(reqBytes))
		req.Header.Set("Content-Type", "application/json")

		// Create a new Chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)

		// Set the route context
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Set up mock expectations
		mockService.On("UpdateExample", mock.Anything, id, mock.MatchedBy(func(r *models.ExampleRequest) bool {
			return r.Name == reqBody.Name && r.Description == reqBody.Description
		})).Return(example, nil)

		handler.UpdateExampleHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.Example
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, id, resp.ID)
		assert.Equal(t, reqBody.Name, resp.Name)
		assert.Equal(t, reqBody.Description, resp.Description)
	})

	// Test DeleteExampleHandler
	t.Run("DeleteExampleHandler", func(t *testing.T) {
		id := uuid.New().String()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/examples/"+id, nil)

		// Create a new Chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)

		// Set the route context
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Set up mock expectations
		mockService.On("DeleteExample", mock.Anything, id).Return(nil)

		handler.DeleteExampleHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	// Test error handling in GetExampleHandler
	t.Run("GetExampleHandler_NotFound", func(t *testing.T) {
		id := uuid.New().String()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/examples/"+id, nil)

		// Create a new Chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)

		// Set the route context
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Set up mock expectations
		mockService.On("GetExample", mock.Anything, id).Return(nil, repository.ErrNotFound)

		handler.GetExampleHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
