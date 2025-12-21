package middleware

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/beego/beego/context"
)

func TestCasbinMiddleware(t *testing.T) {
	// Create temporary test policy files
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "model.conf")
	policyPath := filepath.Join(tmpDir, "policy.csv")

	// Write test model
	modelContent := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	if err := os.WriteFile(modelPath, []byte(modelContent), 0644); err != nil {
		t.Fatalf("Failed to create model file: %v", err)
	}

	// Write test policy
	policyContent := `p, admin, /api/users, GET
p, admin, /api/users, POST
p, user, /api/profile, GET
g, alice, admin
g, bob, user
`
	if err := os.WriteFile(policyPath, []byte(policyContent), 0644); err != nil {
		t.Fatalf("Failed to create policy file: %v", err)
	}

	// Save original paths
	origModelPath := filepath.Join("conf", "model.conf")
	origPolicyPath := filepath.Join("conf", "policy.csv")

	// Create conf directory if it doesn't exist
	os.MkdirAll("conf", 0755)

	// Copy test files to conf directory
	modelData, _ := os.ReadFile(modelPath)
	os.WriteFile(origModelPath, modelData, 0644)
	policyData, _ := os.ReadFile(policyPath)
	os.WriteFile(origPolicyPath, policyData, 0644)

	// Initialize Casbin
	if err := InitCasbin(); err != nil {
		t.Fatalf("Failed to initialize Casbin: %v", err)
	}

	tests := []struct {
		name           string
		username       string
		path           string
		method         string
		expectedStatus int
		expectPass     bool
	}{
		{
			name:       "Admin can GET users",
			username:   "alice",
			path:       "/api/users",
			method:     "GET",
			expectPass: true,
		},
		{
			name:       "Admin can POST users",
			username:   "alice",
			path:       "/api/users",
			method:     "POST",
			expectPass: true,
		},
		{
			name:           "User cannot POST users",
			username:       "bob",
			path:           "/api/users",
			method:         "POST",
			expectedStatus: 403,
			expectPass:     false,
		},
		{
			name:       "User can GET profile",
			username:   "bob",
			path:       "/api/profile",
			method:     "GET",
			expectPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Create a Beego context
			ctx := context.NewContext()
			ctx.Reset(w, req)

			// Set username in context (simulating OAuth middleware)
			ctx.Input.SetData("username", tt.username)

			// Run the middleware
			middleware := Casbin()
			middleware(ctx)

			if !tt.expectPass {
				if w.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				}
			}
		})
	}
}
