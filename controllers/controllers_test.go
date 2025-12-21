package controllers

import (
	"net/http/httptest"
	"testing"

	"github.com/beego/beego"
	"github.com/beego/beego/context"
)

func TestHealthController(t *testing.T) {
	// Create a test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Create Beego context
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Create controller and set context
	controller := HealthController{}
	controller.Init(ctx, "HealthController", "Get", nil)

	// Call the Get method
	controller.Get()

	// Check response
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that response contains expected fields
	if controller.Data["json"] == nil {
		t.Error("Expected JSON data in response")
	}
}

func TestAuthControllerLogin(t *testing.T) {
	// Set up beego config
	beego.AppConfig.Set("oauth.secret", "test-secret")

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "Valid admin login",
			username:       "alice",
			password:       "password123",
			expectedStatus: 200,
		},
		{
			name:           "Valid user login",
			username:       "bob",
			password:       "password456",
			expectedStatus: 200,
		},
		{
			name:           "Invalid credentials",
			username:       "alice",
			password:       "wrongpassword",
			expectedStatus: 401,
		},
		{
			name:           "Non-existent user",
			username:       "charlie",
			password:       "password",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request with form data
			req := httptest.NewRequest("POST", "/login", nil)
			req.Form = map[string][]string{
				"username": {tt.username},
				"password": {tt.password},
			}
			w := httptest.NewRecorder()

			// Create Beego context
			ctx := context.NewContext()
			ctx.Reset(w, req)

			// Create controller and set context
			controller := AuthController{}
			controller.Init(ctx, "AuthController", "Login", nil)

			// Call the Login method
			controller.Login()

			// Check response status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For successful login, check that token is returned
			if tt.expectedStatus == 200 && controller.Data["json"] == nil {
				t.Error("Expected JSON data with token in response")
			}
		})
	}
}
