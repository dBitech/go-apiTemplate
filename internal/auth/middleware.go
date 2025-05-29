package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

// ContextKey is a key for storing authentication context
type ContextKey string

const (
	// ClaimsContextKey is the context key for claims
	ClaimsContextKey ContextKey = "claims"

	// ScopesContextKey is the context key for scopes
	ScopesContextKey ContextKey = "scopes"

	// UserIDContextKey is the context key for user ID
	UserIDContextKey ContextKey = "user_id"
)

// JWTAuthMiddleware creates a middleware that requires a valid JWT token
func (a *Authenticator) JWTAuthMiddleware(requiredScopes []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			token, err := ExtractBearerToken(r)
			if err != nil {
				a.log.Debug("JWT auth failed", logger.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := a.VerifyJWTToken(token)
			if err != nil {
				a.log.Debug("JWT verification failed", logger.Error(err))

				if err == ErrExpiredToken {
					http.Error(w, "Token expired", http.StatusUnauthorized)
				} else {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				}
				return
			}

			// Check scopes if required
			if len(requiredScopes) > 0 {
				hasScope := false
				for _, requiredScope := range requiredScopes {
					for _, tokenScope := range claims.Scopes {
						if tokenScope == requiredScope || tokenScope == "admin" {
							hasScope = true
							break
						}
					}
					if hasScope {
						break
					}
				}

				if !hasScope {
					a.log.Debug("Insufficient scope",
						logger.String("required", strings.Join(requiredScopes, ",")),
						logger.String("provided", strings.Join(claims.Scopes, ",")),
					)
					http.Error(w, "Forbidden: insufficient scope", http.StatusForbidden)
					return
				}
			}

			// Store claims in request context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			ctx = context.WithValue(ctx, ScopesContextKey, claims.Scopes)
			ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)

			// Proceed with the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OAuth2AuthMiddleware creates a middleware that requires a valid OAuth2 token
func (a *Authenticator) OAuth2AuthMiddleware(requiredScopes []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			_, err := ExtractBearerToken(r)
			if err != nil {
				a.log.Debug("OAuth2 auth failed", logger.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// In a real application, you would validate the token with the OAuth2 provider
			// For this template, we'll use a simple introspection simulation

			// Get the request context
			ctx := r.Context()

			// Here you would make a request to the OAuth2 introspection endpoint using a client
			// For example: httpClient := &http.Client{}
			// response, err := httpClient.Post(introspectionURL, "application/x-www-form-urlencoded", formData)

			// For the template, we'll assume the token is valid and has certain scopes

			// Example scopes (in a real app, these would come from the introspection response)
			scopes := []string{"read", "write"} // Example scopes
			userID := "oauth2-user-123"         // Example user ID

			// Check required scopes
			if len(requiredScopes) > 0 {
				hasScope := false
				for _, requiredScope := range requiredScopes {
					for _, tokenScope := range scopes {
						if tokenScope == requiredScope || tokenScope == "admin" {
							hasScope = true
							break
						}
					}
					if hasScope {
						break
					}
				}

				if !hasScope {
					a.log.Debug("Insufficient OAuth2 scope",
						logger.String("required", strings.Join(requiredScopes, ",")),
						logger.String("provided", strings.Join(scopes, ",")),
					)
					http.Error(w, "Forbidden: insufficient scope", http.StatusForbidden)
					return
				}
			}

			// Store scopes and user ID in request context
			ctx = context.WithValue(ctx, ScopesContextKey, scopes)
			ctx = context.WithValue(ctx, UserIDContextKey, userID)

			// Proceed with the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID returns the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}

// GetScopes returns the scopes from the context
func GetScopes(ctx context.Context) ([]string, bool) {
	scopes, ok := ctx.Value(ScopesContextKey).([]string)
	return scopes, ok
}

// GetClaims returns the JWT claims from the context
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}
