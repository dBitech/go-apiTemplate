package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dBiTech/go-apiTemplate/internal/api"
	"github.com/dBiTech/go-apiTemplate/internal/config"
	"github.com/dBiTech/go-apiTemplate/internal/models"
)

func TestAPIIntegration(t *testing.T) {
	// Skip in CI environment
	// You can implement a check here if needed

	// Create a test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Metrics: config.MetricsConfig{
			Enabled: false,
		},
		Tracing: config.TracingConfig{
			Enabled: false,
		},
		Auth: config.AuthConfig{
			Enabled:            true,
			JWTSecret:          "test-secret-key",
			JWTSigningMethod:   "HS256",
			JWTExpirationTime:  24 * 60 * 60 * 1000000000, // 24 hours in nanoseconds
			JWTIssuer:          "api-template-test",
			OAuth2ClientID:     "test-client-id",
			OAuth2ClientSecret: "test-client-secret",
			OAuth2RedirectURL:  "http://localhost:8080/auth/callback",
			OAuth2AuthURL:      "https://example.com/oauth/authorize",
			OAuth2TokenURL:     "https://example.com/oauth/token",
			OAuth2Scopes:       []string{"read", "write"},
		},
	}

	// Create a new server
	server, err := api.NewServer(cfg)
	require.NoError(t, err)

	// Create a router for testing
	router := server.GetRouter()

	// Test health endpoint
	t.Run("HealthEndpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "api-template", resp["name"])
		assert.Equal(t, "UP", resp["status"])
	})

	// Test hello endpoint
	t.Run("HelloEndpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "Hello, World!", resp["message"])
	})

	// Test create and get example
	t.Run("ExampleCRUD", func(t *testing.T) {
		// Create an example
		createReq := models.ExampleRequest{
			Name:        "Test Example",
			Description: "Test Description",
		}

		reqBytes, err := json.Marshal(createReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/examples", bytes.NewBuffer(reqBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createResp models.Example
		err = json.Unmarshal(w.Body.Bytes(), &createResp)
		require.NoError(t, err)

		// Get the created example
		req = httptest.NewRequest(http.MethodGet, "/api/v1/examples/"+createResp.ID, nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var getResp models.Example
		err = json.Unmarshal(w.Body.Bytes(), &getResp)
		require.NoError(t, err)

		assert.Equal(t, createResp.ID, getResp.ID)
		assert.Equal(t, createReq.Name, getResp.Name)
		assert.Equal(t, createReq.Description, getResp.Description)

		// Update the example
		updateReq := models.ExampleRequest{
			Name:        "Updated Example",
			Description: "Updated Description",
		}

		reqBytes, err = json.Marshal(updateReq)
		require.NoError(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/v1/examples/"+createResp.ID, bytes.NewBuffer(reqBytes))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updateResp models.Example
		err = json.Unmarshal(w.Body.Bytes(), &updateResp)
		require.NoError(t, err)

		assert.Equal(t, createResp.ID, updateResp.ID)
		assert.Equal(t, updateReq.Name, updateResp.Name)
		assert.Equal(t, updateReq.Description, updateResp.Description)

		// Delete the example
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/examples/"+createResp.ID, nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify it's deleted
		req = httptest.NewRequest(http.MethodGet, "/api/v1/examples/"+createResp.ID, nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test JWT protected endpoint (unauthorized)
	t.Run("JWTProtectedEndpoint_Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/jwt", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test OAuth2 protected endpoint (unauthorized)
	t.Run("OAuth2ProtectedEndpoint_Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/oauth2", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test JWT protected endpoint (authorized)
	t.Run("JWTProtectedEndpoint_Authorized", func(t *testing.T) {
		// Get a JWT token from the server's auth instance
		authInstance := server.GetAuthenticator() // Add this method to Server
		token, err := authInstance.GenerateJWTToken("test-user", []string{"user"}, []string{"read", "write"})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/jwt", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resources []*models.ProtectedResource
		err = json.Unmarshal(w.Body.Bytes(), &resources)
		require.NoError(t, err)

		assert.Len(t, resources, 2)
		assert.NotEmpty(t, resources[0].ID)
		assert.NotEmpty(t, resources[0].Name)
		assert.NotEmpty(t, resources[0].Content)
	})

	// Test user profile endpoint (authorized with JWT)
	t.Run("UserProfileEndpoint_JWTAuthorized", func(t *testing.T) {
		// Get a JWT token from the server's auth instance
		authInstance := server.GetAuthenticator() // Add this method to Server
		userID := "test-user-123"
		token, err := authInstance.GenerateJWTToken(userID, []string{"user"}, []string{"read", "write"})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var profile models.UserProfile
		err = json.Unmarshal(w.Body.Bytes(), &profile)
		require.NoError(t, err)

		assert.Equal(t, userID, profile.ID)
		assert.NotEmpty(t, profile.Username)
		assert.NotEmpty(t, profile.Email)
	})
}
