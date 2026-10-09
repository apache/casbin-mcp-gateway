package middleware

import (
	"fmt"
	"strings"

	"github.com/beego/beego"
	"github.com/beego/beego/context"
	"github.com/golang-jwt/jwt/v4"
)

// OAuth middleware handles OAuth authentication
func OAuth() beego.FilterFunc {
	return func(ctx *context.Context) {
		// Skip authentication for static files and public endpoints
		path := ctx.Request.URL.Path
		if isPublicPath(path) {
			return
		}

		// Get Authorization header
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			ctx.Output.SetStatus(401)
			ctx.Output.JSON(map[string]string{"error": "Missing authorization header"}, false, false)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.Output.SetStatus(401)
			ctx.Output.JSON(map[string]string{"error": "Invalid authorization header format"}, false, false)
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the secret key for validation
			// In production, this should come from configuration
			return []byte(beego.AppConfig.DefaultString("oauth.secret", "your-secret-key")), nil
		})

		if err != nil || !token.Valid {
			ctx.Output.SetStatus(401)
			ctx.Output.JSON(map[string]string{"error": "Invalid or expired token"}, false, false)
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Store user information in context
			if username, ok := claims["sub"].(string); ok {
				ctx.Input.SetData("username", username)
			}
			if role, ok := claims["role"].(string); ok {
				ctx.Input.SetData("role", role)
			}
		}
	}
}

// isPublicPath checks if the path should skip authentication
func isPublicPath(path string) bool {
	return path == "/login" || path == "/health"
}
