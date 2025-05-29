package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

// Standard errors
var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token has expired")
	ErrInsufficientScope    = errors.New("insufficient scope")
	ErrMissingToken         = errors.New("missing token")
	ErrAuthorizationPending = errors.New("authorization is pending")
)

// TokenType represents the type of token
type TokenType string

const (
	// OAuth2Token represents an OAuth2 token
	OAuth2Token TokenType = "oauth2"

	// JWTToken represents a JWT token
	JWTToken TokenType = "jwt"
)

// AuthConfig contains configuration for authentication
type AuthConfig struct {
	// JWT Configuration
	JWTSecret         string          // Secret key for JWT signing (for HMAC algorithms)
	JWTPrivateKey     *rsa.PrivateKey // Private key for JWT signing (for RSA algorithms)
	JWTPublicKey      *rsa.PublicKey  // Public key for JWT verification (for RSA algorithms)
	JWTSigningMethod  string          // Signing method (e.g., "HS256", "RS256")
	JWTExpirationTime time.Duration   // Token expiration time
	JWTIssuer         string          // Token issuer

	// OAuth2 Configuration
	OAuth2ClientID     string   // OAuth2 client ID
	OAuth2ClientSecret string   // OAuth2 client secret
	OAuth2RedirectURL  string   // OAuth2 redirect URL
	OAuth2AuthURL      string   // OAuth2 authorization URL
	OAuth2TokenURL     string   // OAuth2 token URL
	OAuth2Scopes       []string // OAuth2 scopes
}

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID string   `json:"user_id,omitempty"`
	Roles  []string `json:"roles,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
}

// Authenticator handles authentication and authorization
type Authenticator struct {
	jwtSigningMethod jwt.SigningMethod
	jwtSecret        []byte
	jwtPrivateKey    *rsa.PrivateKey
	jwtPublicKey     *rsa.PublicKey
	jwtIssuer        string
	jwtExpiration    time.Duration

	oauth2Config oauth2.Config
	log          logger.Logger
}

// NewAuthenticator creates a new authenticator instance
func NewAuthenticator(config AuthConfig, log logger.Logger) (*Authenticator, error) {
	var signingMethod jwt.SigningMethod

	// Set JWT signing method based on configuration
	switch config.JWTSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	case "RS384":
		signingMethod = jwt.SigningMethodRS384
	case "RS512":
		signingMethod = jwt.SigningMethodRS512
	default:
		signingMethod = jwt.SigningMethodHS256
	}

	// Configure OAuth2
	oauth2Config := oauth2.Config{
		ClientID:     config.OAuth2ClientID,
		ClientSecret: config.OAuth2ClientSecret,
		RedirectURL:  config.OAuth2RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.OAuth2AuthURL,
			TokenURL: config.OAuth2TokenURL,
		},
		Scopes: config.OAuth2Scopes,
	}

	return &Authenticator{
		jwtSigningMethod: signingMethod,
		jwtSecret:        []byte(config.JWTSecret),
		jwtPrivateKey:    config.JWTPrivateKey,
		jwtPublicKey:     config.JWTPublicKey,
		jwtIssuer:        config.JWTIssuer,
		jwtExpiration:    config.JWTExpirationTime,
		oauth2Config:     oauth2Config,
		log:              log,
	}, nil
}

// GenerateJWTToken generates a new JWT token
func (a *Authenticator) GenerateJWTToken(userID string, roles, scopes []string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(a.jwtExpiration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    a.jwtIssuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
		UserID: userID,
		Roles:  roles,
		Scopes: scopes,
	}

	token := jwt.NewWithClaims(a.jwtSigningMethod, claims)

	var tokenString string
	var err error

	// Sign the token based on the signing method
	switch a.jwtSigningMethod {
	case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
		tokenString, err = token.SignedString(a.jwtSecret)
	case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
		tokenString, err = token.SignedString(a.jwtPrivateKey)
	default:
		return "", fmt.Errorf("unsupported signing method: %v", a.jwtSigningMethod)
	}

	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyJWTToken verifies a JWT token and returns the claims
func (a *Authenticator) VerifyJWTToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return a.jwtSecret, nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			return a.jwtPublicKey, nil
		}
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetOAuth2AuthURL generates an OAuth2 authorization URL
func (a *Authenticator) GetOAuth2AuthURL(state string) string {
	return a.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// GetOAuth2Token exchanges an authorization code for an OAuth2 token
func (a *Authenticator) GetOAuth2Token(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.oauth2Config.Exchange(ctx, code)
}

// RefreshOAuth2Token refreshes an OAuth2 token
func (a *Authenticator) RefreshOAuth2Token(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	source := a.oauth2Config.TokenSource(ctx, token)
	newToken, err := source.Token()
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

// GetOAuth2Client returns an HTTP client with the OAuth2 token
func (a *Authenticator) GetOAuth2Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return a.oauth2Config.Client(ctx, token)
}

// ExtractBearerToken extracts a bearer token from the Authorization header
func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

// BasicAuth represents a username and password pair
type BasicAuth struct {
	Username string
	Password string
}

// ExtractBasicAuth extracts basic auth credentials from the Authorization header
func ExtractBasicAuth(r *http.Request) (*BasicAuth, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrMissingToken
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return nil, ErrInvalidToken
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return nil, ErrInvalidToken
	}

	return &BasicAuth{
		Username: pair[0],
		Password: pair[1],
	}, nil
}

// OAuth2Response represents the response from an OAuth2 token endpoint
type OAuth2Response struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}
