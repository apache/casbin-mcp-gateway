package controllers

import (
	"encoding/json"
	"time"

	"github.com/beego/beego"
	"github.com/golang-jwt/jwt/v4"
)

type AuthController struct {
	beego.Controller
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  string `json:"user"`
}

func (c *AuthController) Login() {
	var req LoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		// Try form parsing as fallback
		req.Username = c.GetString("username")
		req.Password = c.GetString("password")
		if req.Username == "" || req.Password == "" {
			c.Data["json"] = map[string]string{"error": "Invalid request"}
			c.Ctx.Output.SetStatus(400)
			c.ServeJSON()
			return
		}
	}

	// Simple authentication (in production, verify against a database)
	// For demo purposes, accept predefined users
	validUsers := map[string]string{
		"alice": "password123",
		"bob":   "password456",
	}

	password, exists := validUsers[req.Username]
	if !exists || password != req.Password {
		c.Data["json"] = map[string]string{"error": "Invalid credentials"}
		c.Ctx.Output.SetStatus(401)
		c.ServeJSON()
		return
	}

	// Determine user role
	role := "user"
	if req.Username == "alice" {
		role = "admin"
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  req.Username,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	})

	secret := beego.AppConfig.DefaultString("oauth.secret", "your-secret-key")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.Data["json"] = map[string]string{"error": "Failed to generate token"}
		c.Ctx.Output.SetStatus(500)
		c.ServeJSON()
		return
	}

	c.Data["json"] = LoginResponse{
		Token: tokenString,
		User:  req.Username,
	}
	c.ServeJSON()
}
