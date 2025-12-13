package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/beego/beego/context"
	"github.com/casbin/mcp-gateway/middleware"
)

func TestCasbinMiddlewareSkipPaths(t *testing.T) {
	// Initialize Casbin middleware
	casbinMiddleware := middleware.NewCasbinMiddleware()

	// Test paths that should skip authorization
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

		// Should not return forbidden for skip paths
		casbinMiddleware.Filter(ctx)

		if w.Code == http.StatusForbidden {
			t.Errorf("Path %s should skip authorization but got forbidden", path)
		}
	}
}

func TestCasbinMiddlewareMissingUser(t *testing.T) {
	// Initialize Casbin middleware
	casbinMiddleware := middleware.NewCasbinMiddleware()

	// Test request without user in context
	req := httptest.NewRequest("GET", "/api/resource", nil)
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	casbinMiddleware.Filter(ctx)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected forbidden status without user, got %d", w.Code)
	}
}

func TestCasbinMiddlewareAuthorizedUser(t *testing.T) {
	// Initialize Casbin middleware
	casbinMiddleware := middleware.NewCasbinMiddleware()

	// Add policy for demo_user
	casbinMiddleware.AddPolicy("demo_user", "/api/resource", "GET")

	// Test request with authorized user (demo_user has access to /api/resource GET)
	req := httptest.NewRequest("GET", "/api/resource", nil)
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	// Set user in context (simulating OAuth middleware)
	ctx.Input.SetData("user", "demo_user")

	casbinMiddleware.Filter(ctx)

	// Should not return forbidden for authorized user
	if w.Code == http.StatusForbidden {
		t.Error("Expected request to pass for authorized user")
	}
}

func TestCasbinMiddlewareUnauthorizedUser(t *testing.T) {
	// Initialize Casbin middleware
	casbinMiddleware := middleware.NewCasbinMiddleware()

	// Test request with unauthorized user
	req := httptest.NewRequest("GET", "/api/admin", nil)
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	// Set user without admin role
	ctx.Input.SetData("user", "demo_user")

	casbinMiddleware.Filter(ctx)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected forbidden status for unauthorized user, got %d", w.Code)
	}
}

func TestCasbinMiddlewareAdminUser(t *testing.T) {
	// Initialize Casbin middleware
	casbinMiddleware := middleware.NewCasbinMiddleware()

	// Add policies for admin role and assign role to admin_user
	casbinMiddleware.AddPolicy("admin", "/api/admin", "GET")
	// Add role assignment (using AddGroupingPolicy for role membership)
	casbinMiddleware.GetEnforcer().AddGroupingPolicy("admin_user", "admin")

	// Test request with admin user
	req := httptest.NewRequest("GET", "/api/admin", nil)
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	// Set admin user (admin_user has admin role)
	ctx.Input.SetData("user", "admin_user")

	casbinMiddleware.Filter(ctx)

	// Should not return forbidden for admin user
	if w.Code == http.StatusForbidden {
		t.Error("Expected request to pass for admin user")
	}
}

func TestCasbinMiddlewarePolicyManagement(t *testing.T) {
	// Initialize Casbin middleware
	casbinMiddleware := middleware.NewCasbinMiddleware()

	// Test adding a policy
	added, err := casbinMiddleware.AddPolicy("test_user", "/api/test", "GET")
	if err != nil {
		t.Errorf("Failed to add policy: %v", err)
	}
	if !added {
		t.Error("Policy should have been added")
	}

	// Test getting policies
	policies := casbinMiddleware.GetPolicy()
	if len(policies) == 0 {
		t.Error("Should have at least one policy")
	}

	// Test removing a policy
	removed, err := casbinMiddleware.RemovePolicy("test_user", "/api/test", "GET")
	if err != nil {
		t.Errorf("Failed to remove policy: %v", err)
	}
	if !removed {
		t.Error("Policy should have been removed")
	}
}
