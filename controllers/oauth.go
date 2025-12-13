package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/beego/beego"
	"github.com/casbin/mcp-gateway/middleware"
)

// OAuthController handles OAuth authentication flow
type OAuthController struct {
	beego.Controller
	oauthMiddleware *middleware.OAuthMiddleware
}

// stateStore is a thread-safe in-memory store for OAuth state tokens
// In production, use Redis or a persistent store for scalability and persistence
var (
	stateStore   = make(map[string]time.Time)
	stateStoreMu sync.RWMutex
)

// Get initiates the OAuth login flow
// Route: GET /oauth/login
func (c *OAuthController) Get() {
	// Generate a random state token
	state := generateState()

	// Store state token with expiry (5 minutes)
	stateStoreMu.Lock()
	stateStore[state] = time.Now().Add(5 * time.Minute)
	stateStoreMu.Unlock()

	// Clean up expired states
	cleanupExpiredStates()

	// Initialize OAuth middleware to get login URL
	oauthMiddleware := middleware.NewOAuthMiddleware()

	// Get OAuth authorization URL
	authURL := oauthMiddleware.GetLoginURL(state)

	// Store state in session for verification
	c.SetSession("oauth_state", state)

	// Redirect to OAuth provider
	c.Redirect(authURL, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth callback
// Route: GET /oauth/callback
func (c *OAuthController) Callback() {
	// Get authorization code and state from query parameters
	code := c.GetString("code")
	state := c.GetString("state")
	errorParam := c.GetString("error")

	// Check for OAuth errors
	if errorParam != "" {
		errorDesc := c.GetString("error_description")
		c.Data["json"] = map[string]string{
			"error":       errorParam,
			"description": errorDesc,
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	// Verify state token
	sessionState := c.GetSession("oauth_state")
	if sessionState == nil || sessionState.(string) != state {
		c.Data["json"] = map[string]string{
			"error": "invalid_state",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	// Verify state exists and is not expired
	stateStoreMu.RLock()
	expiry, exists := stateStore[state]
	stateStoreMu.RUnlock()
	
	if !exists || time.Now().After(expiry) {
		c.Data["json"] = map[string]string{
			"error": "expired_state",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	// Clean up used state
	stateStoreMu.Lock()
	delete(stateStore, state)
	stateStoreMu.Unlock()
	c.DelSession("oauth_state")

	// Exchange authorization code for token
	oauthMiddleware := middleware.NewOAuthMiddleware()
	token, err := oauthMiddleware.ExchangeCode(context.Background(), code)
	if err != nil {
		beego.Error("Failed to exchange code for token:", err)
		c.Data["json"] = map[string]string{
			"error": "token_exchange_failed",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.ServeJSON()
		return
	}

	// Return the access token
	// In production, you might want to:
	// 1. Store the token securely
	// 2. Set an HTTP-only cookie with a session ID
	// 3. Store the token in a session store (Redis, database)
	// 4. Return a JWT token that the client can use
	c.Data["json"] = map[string]interface{}{
		"access_token": token.AccessToken,
		"token_type":   token.TokenType,
		"expiry":       token.Expiry,
	}
	c.ServeJSON()
}

// generateState creates a random state token
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// cleanupExpiredStates removes expired state tokens
func cleanupExpiredStates() {
	now := time.Now()
	stateStoreMu.Lock()
	defer stateStoreMu.Unlock()
	
	for state, expiry := range stateStore {
		if now.After(expiry) {
			delete(stateStore, state)
		}
	}
}

// Info returns information about the current authenticated user
// Route: GET /oauth/info
func (c *OAuthController) Info() {
	// Get user information from context (set by OAuth middleware)
	user := c.Ctx.Input.GetData("user")
	userID := c.Ctx.Input.GetData("user_id")
	userEmail := c.Ctx.Input.GetData("user_email")
	userRoles := c.Ctx.Input.GetData("user_roles")

	if user == nil {
		c.Data["json"] = map[string]string{
			"error": "not_authenticated",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"user_id": userID,
		"username": user,
		"email":    userEmail,
		"roles":    userRoles,
	}
	c.ServeJSON()
}

// Logout handles user logout
// Route: POST /oauth/logout
func (c *OAuthController) Logout() {
	// In a production system, you would:
	// 1. Invalidate the session
	// 2. Remove the token from storage
	// 3. Optionally revoke the token with the OAuth provider

	c.DestroySession()

	c.Data["json"] = map[string]string{
		"message": "logged out successfully",
	}
	c.ServeJSON()
}
