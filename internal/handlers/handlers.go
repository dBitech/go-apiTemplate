package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/dBiTech/go-apiTemplate/internal/models"
	"github.com/dBiTech/go-apiTemplate/internal/repository"
	"github.com/dBiTech/go-apiTemplate/internal/service"
	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

// Handler provides HTTP handlers
type Handler struct {
	log     logger.Logger
	service *service.Service
}

// NewHandler creates a new handler instance
func NewHandler(log logger.Logger, service *service.Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":500,"message":"Internal Server Error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, status int, message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	response := ErrorResponse{
		Status:  status,
		Message: message,
		Error:   errorMsg,
	}

	RespondJSON(w, status, response)
}

// HelloHandler is a simple example handler
// @Summary Hello world endpoint
// @Description Returns a friendly greeting
// @Tags general
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Successfully returned hello message"
// @Router /hello [get]
func (h *Handler) HelloHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get logger from context
		log := logger.FromContext(r.Context())

		log.Info("handling hello request")

		// Get span from context and add attributes
		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(attribute.String("handler", "hello"))

		response := map[string]string{
			"message": "Hello, World!",
		}

		RespondJSON(w, http.StatusOK, response)
	}
}

// GetExampleHandler handles GET /examples/{id}
// @Summary Get example by ID
// @Description Retrieves a single example by its ID
// @Tags examples
// @Accept json
// @Produce json
// @Param id path string true "Example ID"
// @Success 200 {object} models.Example "Successfully retrieved example"
// @Failure 404 {object} ErrorResponse "Example not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /examples/{id} [get]
func (h *Handler) GetExampleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "getExample"))

		// Get ID from URL
		id := chi.URLParam(r, "id")
		span.SetAttributes(attribute.String("example.id", id))

		// Get example from service
		example, err := h.service.GetExample(ctx, id)
		if err != nil {
			log.Error("failed to get example", logger.String("id", id), logger.Error(err))

			if err == repository.ErrNotFound {
				RespondError(w, http.StatusNotFound, "Example not found", nil)
			} else {
				RespondError(w, http.StatusInternalServerError, "Failed to get example", nil)
			}
			return
		}

		// Respond with example
		RespondJSON(w, http.StatusOK, example)
	}
}

// ListExamplesHandler handles GET /examples
// @Summary List examples
// @Description Returns a list of examples with optional pagination
// @Tags examples
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of results to return" default(10)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {array} models.Example "Successfully retrieved examples"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /examples [get]
func (h *Handler) ListExamplesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "listExamples"))

		// Parse query parameters
		limit := 10
		offset := 0

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		span.SetAttributes(
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		)

		// Get examples from service
		examples, err := h.service.ListExamples(ctx, limit, offset)
		if err != nil {
			log.Error("failed to list examples", logger.Error(err))
			RespondError(w, http.StatusInternalServerError, "Failed to list examples", nil)
			return
		}

		// Respond with examples
		RespondJSON(w, http.StatusOK, examples)
	}
}

// CreateExampleHandler handles POST /examples
// @Summary Create new example
// @Description Creates a new example resource
// @Tags examples
// @Accept json
// @Produce json
// @Param example body models.ExampleRequest true "Example data"
// @Success 201 {object} models.Example "Successfully created example"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "Example already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /examples [post]
func (h *Handler) CreateExampleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "createExample"))

		// Parse request body
		var req models.ExampleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request", logger.Error(err))
			RespondError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		// Validate request
		if req.Name == "" {
			RespondError(w, http.StatusBadRequest, "Name is required", nil)
			return
		}

		// Create example
		example, err := h.service.CreateExample(ctx, &req)
		if err != nil {
			log.Error("failed to create example", logger.Error(err))

			if err == repository.ErrAlreadyExists {
				RespondError(w, http.StatusConflict, "Example already exists", nil)
			} else {
				RespondError(w, http.StatusInternalServerError, "Failed to create example", nil)
			}
			return
		}

		// Respond with created example
		RespondJSON(w, http.StatusCreated, example)
	}
}

// UpdateExampleHandler handles PUT /examples/{id}
// @Summary Update example
// @Description Updates an existing example by ID
// @Tags examples
// @Accept json
// @Produce json
// @Param id path string true "Example ID"
// @Param example body models.ExampleRequest true "Example data"
// @Success 200 {object} models.Example "Successfully updated example"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Example not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /examples/{id} [put]
func (h *Handler) UpdateExampleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "updateExample"))

		// Get ID from URL
		id := chi.URLParam(r, "id")
		span.SetAttributes(attribute.String("example.id", id))

		// Parse request body
		var req models.ExampleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request", logger.Error(err))
			RespondError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		// Validate request
		if req.Name == "" {
			RespondError(w, http.StatusBadRequest, "Name is required", nil)
			return
		}

		// Update example
		example, err := h.service.UpdateExample(ctx, id, &req)
		if err != nil {
			log.Error("failed to update example", logger.String("id", id), logger.Error(err))

			if err == repository.ErrNotFound {
				RespondError(w, http.StatusNotFound, "Example not found", nil)
			} else {
				RespondError(w, http.StatusInternalServerError, "Failed to update example", nil)
			}
			return
		}

		// Respond with updated example
		RespondJSON(w, http.StatusOK, example)
	}
}

// DeleteExampleHandler handles DELETE /examples/{id}
// @Summary Delete example
// @Description Deletes an example by ID
// @Tags examples
// @Accept json
// @Produce json
// @Param id path string true "Example ID"
// @Success 204 "Successfully deleted example"
// @Failure 404 {object} ErrorResponse "Example not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /examples/{id} [delete]
func (h *Handler) DeleteExampleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "deleteExample"))

		// Get ID from URL
		id := chi.URLParam(r, "id")
		span.SetAttributes(attribute.String("example.id", id))

		// Delete example
		err := h.service.DeleteExample(ctx, id)
		if err != nil {
			log.Error("failed to delete example", logger.String("id", id), logger.Error(err))

			if err == repository.ErrNotFound {
				RespondError(w, http.StatusNotFound, "Example not found", nil)
			} else {
				RespondError(w, http.StatusInternalServerError, "Failed to delete example", nil)
			}
			return
		}

		// Respond with no content
		w.WriteHeader(http.StatusNoContent)
	}
}

// JWTProtectedResourceHandler handles GET /protected/jwt
// @Summary Get JWT protected resources
// @Description Returns a list of resources that require JWT authentication
// @Tags protected
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ProtectedResource "Successfully retrieved protected resources"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden: insufficient scope"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /protected/jwt [get]
func (h *Handler) JWTProtectedResourceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		log.Info("handling JWT protected resource request")

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "jwtProtectedResource"))

		// Get resources
		resources, err := h.service.ListProtectedResources(ctx)
		if err != nil {
			log.Error("failed to list protected resources", logger.Error(err))
			RespondError(w, http.StatusInternalServerError, "Failed to list protected resources", nil)
			return
		}

		// Respond with resources
		RespondJSON(w, http.StatusOK, resources)
	}
}

// OAuth2ProtectedResourceHandler handles GET /protected/oauth2
// @Summary Get OAuth2 protected resources
// @Description Returns a list of resources that require OAuth2 authentication
// @Tags protected
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ProtectedResource "Successfully retrieved protected resources"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden: insufficient scope"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /protected/oauth2 [get]
func (h *Handler) OAuth2ProtectedResourceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		log.Info("handling OAuth2 protected resource request")

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "oauth2ProtectedResource"))

		// Get resources
		resources, err := h.service.ListProtectedResources(ctx)
		if err != nil {
			log.Error("failed to list protected resources", logger.Error(err))
			RespondError(w, http.StatusInternalServerError, "Failed to list protected resources", nil)
			return
		}

		// Respond with resources
		RespondJSON(w, http.StatusOK, resources)
	}
}

// UserProfileHandler handles GET /me
// @Summary Get user profile
// @Description Returns the authenticated user's profile
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserProfile "Successfully retrieved user profile"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /me [get]
func (h *Handler) UserProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		log.Info("handling user profile request")

		// Get span and add attributes
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("handler", "userProfile"))

		// Get user ID from context (set by auth middleware)
		userID, ok := ctx.Value("user_id").(string)
		if !ok {
			log.Error("user ID not found in context")
			RespondError(w, http.StatusInternalServerError, "User ID not found", nil)
			return
		}

		// Get user profile
		profile, err := h.service.GetUserProfile(ctx, userID)
		if err != nil {
			log.Error("failed to get user profile", logger.String("userID", userID), logger.Error(err))
			RespondError(w, http.StatusInternalServerError, "Failed to get user profile", nil)
			return
		}

		// Respond with profile
		RespondJSON(w, http.StatusOK, profile)
	}
}
