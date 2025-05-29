// Package models defines the data structures used throughout the application.
// It includes request/response models, domain entities, and data transfer objects.
package models

import (
	"time"
)

// BaseModel represents common fields for all models
type BaseModel struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Example is an example model
type Example struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// NewExample creates a new example model
func NewExample(id, name, description string) *Example {
	now := time.Now()
	return &Example{
		BaseModel: BaseModel{
			ID:        id,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        name,
		Description: description,
		Status:      "active",
	}
}

// ExampleRequest represents a request to create or update an example
type ExampleRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// ProtectedResource represents a resource that requires authentication
type ProtectedResource struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	OwnerID   string    `json:"ownerId"`
}

// UserProfile represents a user profile
type UserProfile struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	Scopes   []string `json:"scopes"`
}

// ExampleResponse represents an example response
type ExampleResponse struct {
	Example
}
