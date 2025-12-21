package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beego/beego"
	"github.com/beego/beego/context"
	"github.com/golang-jwt/jwt/v4"
)

func TestOAuthMiddleware(t *testing.T) {
	// Set up beego config
	beego.AppConfig.Set("oauth.secret", "test-secret")

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectPass     bool
	}{
		{
			name:           "No Authorization Header",
			authHeader:     "",
			expectedStatus: 401,
			expectPass:     false,
		},
		{
			name:           "Invalid Authorization Format",
			authHeader:     "InvalidToken",
			expectedStatus: 401,
			expectPass:     false,
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: 401,
			expectPass:     false,
		},
		{
			name:           "Valid Token",
			authHeader:     generateValidToken(t, "alice", "admin"),
			expectedStatus: 200,
			expectPass:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/api/users", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a test response recorder
			w := httptest.NewRecorder()

			// Create a Beego context
			ctx := context.NewContext()
			ctx.Reset(w, req)

			// Run the middleware
			middleware := OAuth()
			middleware(ctx)

			if tt.expectPass {
				// Check that username was set in context
				username := ctx.Input.GetData("username")
				if username == nil {
					t.Error("Expected username to be set in context")
				}
			} else {
				if w.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				}
			}
		})
	}
}

func TestPublicPaths(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/login", true},
		{"/health", true},
		{"/static/css/app.css", true},
		{"/favicon.ico", true},
		{"/index.html", true},
		{"/app.css", true},
		{"/app.js", true},
		{"/api/users", false},
		{"/api/profile", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isPublicPath(tt.path)
			if result != tt.expected {
				t.Errorf("isPublicPath(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func generateValidToken(t *testing.T, username, role string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  username,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	return "Bearer " + tokenString
}
