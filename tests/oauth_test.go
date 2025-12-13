package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/beego/beego/context"
	"github.com/casbin/mcp-gateway/middleware"
)

func TestOAuthMiddlewareSkipPaths(t *testing.T) {
	// Initialize OAuth middleware
	oauthMiddleware := middleware.NewOAuthMiddleware()

	// Test paths that should skip authentication
	skipPaths := []string{"/", "/oauth/login", "/oauth/callback", "/health"}

	for _, path := range skipPaths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()

		ctx := &beegoCtx.Context{
			Request:        req,
			ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
			Input:          beegoCtx.NewInput(),
			Output:         beegoCtx.NewOutput(),
		}
		ctx.Reset(w, req)

		// Should not return unauthorized for skip paths
		oauthMiddleware.Filter(ctx)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Path %s should skip authentication but got unauthorized", path)
		}
	}
}

func TestOAuthMiddlewareMissingAuth(t *testing.T) {
	// Initialize OAuth middleware
	oauthMiddleware := middleware.NewOAuthMiddleware()

	// Test request without Authorization header
	req := httptest.NewRequest("GET", "/api/resource", nil)
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	oauthMiddleware.Filter(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected unauthorized status, got %d", w.Code)
	}
}

func TestOAuthMiddlewareWithToken(t *testing.T) {
	// Initialize OAuth middleware
	oauthMiddleware := middleware.NewOAuthMiddleware()

	// Test request with valid Bearer token
	req := httptest.NewRequest("GET", "/api/resource", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	oauthMiddleware.Filter(ctx)

	// Should not return unauthorized with token
	if w.Code == http.StatusUnauthorized {
		t.Error("Expected request to pass with valid token")
	}

	// Check if user data was set in context
	if ctx.Input.GetData("user") == nil {
		t.Error("User data should be set in context")
	}
}

func TestOAuthMiddlewareInvalidAuthFormat(t *testing.T) {
	// Initialize OAuth middleware
	oauthMiddleware := middleware.NewOAuthMiddleware()

	// Test request with invalid auth format
	testCases := []string{
		"InvalidFormat",
		"Bearer",
		"Basic dXNlcjpwYXNz",
	}

	for _, authHeader := range testCases {
		req := httptest.NewRequest("GET", "/api/resource", nil)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		ctx := &beegoCtx.Context{
			Request:        req,
			ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
			Input:          beegoCtx.NewInput(),
			Output:         beegoCtx.NewOutput(),
		}
		ctx.Reset(w, req)

		oauthMiddleware.Filter(ctx)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected unauthorized for auth header '%s', got %d", authHeader, w.Code)
		}
	}
}
