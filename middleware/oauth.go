package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego"
	beegoCtx "github.com/beego/beego/context"
	"golang.org/x/oauth2"
)

// OAuthMiddleware handles OAuth authentication
type OAuthMiddleware struct {
	config *oauth2.Config
	// Skip authentication for these paths
	skipPaths []string
}

// NewOAuthMiddleware creates a new OAuth middleware instance
func NewOAuthMiddleware() *OAuthMiddleware {
	// Configure OAuth2
	// These values should be loaded from configuration
	config := &oauth2.Config{
		ClientID:     beego.AppConfig.DefaultString("oauth2.client_id", ""),
		ClientSecret: beego.AppConfig.DefaultString("oauth2.client_secret", ""),
		RedirectURL:  beego.AppConfig.DefaultString("oauth2.redirect_url", "http://localhost:8080/oauth/callback"),
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  beego.AppConfig.DefaultString("oauth2.auth_url", "https://accounts.google.com/o/oauth2/auth"),
			TokenURL: beego.AppConfig.DefaultString("oauth2.token_url", "https://accounts.google.com/o/oauth2/token"),
		},
	}

	return &OAuthMiddleware{
		config: config,
		skipPaths: []string{
			"/",
			"/oauth/login",
			"/oauth/callback",
			"/health",
		},
	}
}

// Filter is the middleware filter function for Beego
func (m *OAuthMiddleware) Filter(ctx *beegoCtx.Context) {
	// Check if path should skip authentication
	for _, path := range m.skipPaths {
		if ctx.Request.URL.Path == path {
			return
		}
	}

	// Extract token from Authorization header
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		m.unauthorized(ctx, "Missing Authorization header")
		return
	}

	// Parse Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		m.unauthorized(ctx, "Invalid Authorization header format")
		return
	}

	token := parts[1]

	// Verify token and extract user information
	// In a real implementation, this would validate the token with the OAuth provider
	// For demonstration, we'll do a simplified validation
	userInfo, err := m.validateToken(ctx.Request.Context(), token)
	if err != nil {
		m.unauthorized(ctx, fmt.Sprintf("Invalid token: %v", err))
		return
	}

	// Store user information in context for downstream middleware and controllers
	ctx.Input.SetData("user", userInfo.Username)
	ctx.Input.SetData("user_id", userInfo.UserID)
	ctx.Input.SetData("user_email", userInfo.Email)
	ctx.Input.SetData("user_roles", userInfo.Roles)
}

// UserInfo represents authenticated user information
type UserInfo struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// validateToken validates the OAuth token and extracts user information
func (m *OAuthMiddleware) validateToken(ctx context.Context, tokenString string) (*UserInfo, error) {
	// In a real implementation, this would:
	// 1. Validate the token with the OAuth provider
	// 2. Verify token signature and expiry
	// 3. Extract user information from token claims or userinfo endpoint

	// For demonstration purposes, we'll implement a simple mock validation
	// In production, you would use the OAuth2 client to verify the token:
	// token := &oauth2.Token{AccessToken: tokenString}
	// client := m.config.Client(ctx, token)
	// resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	// Mock user info extraction
	// In production, this would come from the OAuth provider's userinfo endpoint
	userInfo := &UserInfo{
		UserID:   "user123",
		Username: "demo_user",
		Email:    "user@example.com",
		Roles:    []string{"user"},
	}

	return userInfo, nil
}

// unauthorized sends an unauthorized response
func (m *OAuthMiddleware) unauthorized(ctx *beegoCtx.Context, message string) {
	ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"error":   "unauthorized",
		"message": message,
	}
	json.NewEncoder(ctx.ResponseWriter).Encode(response)
}

// GetLoginURL returns the OAuth login URL
func (m *OAuthMiddleware) GetLoginURL(state string) string {
	return m.config.AuthCodeURL(state)
}

// ExchangeCode exchanges the authorization code for a token
func (m *OAuthMiddleware) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return m.config.Exchange(ctx, code)
}
