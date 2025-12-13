package main

import (
	"github.com/beego/beego"
	"github.com/casbin/mcp-gateway/controllers"
	"github.com/casbin/mcp-gateway/middleware"
)

func main() {
	// Initialize middleware
	oauthMiddleware := middleware.NewOAuthMiddleware()
	casbinMiddleware := middleware.NewCasbinMiddleware()

	// Register middleware with Beego in specific order
	// OAuth authentication middleware runs first to identify the user
	beego.InsertFilter("*", beego.BeforeRouter, oauthMiddleware.Filter, false)
	// Casbin authorization middleware runs after authentication to check permissions
	// Using the same stage (BeforeRouter) - execution order is guaranteed by registration order
	beego.InsertFilter("*", beego.BeforeRouter, casbinMiddleware.Filter, false)

	// Register routes
	beego.Router("/", &controllers.MainController{})
	beego.Router("/api/resource", &controllers.ResourceController{})
	beego.Router("/api/admin", &controllers.AdminController{})

	// OAuth routes (skip authentication/authorization middleware)
	oauthController := &controllers.OAuthController{}
	beego.Router("/oauth/login", oauthController, "get:Get")
	beego.Router("/oauth/callback", oauthController, "get:Callback")
	beego.Router("/oauth/info", oauthController, "get:Info")
	beego.Router("/oauth/logout", oauthController, "post:Logout")

	// MCP proxy routes (requires authentication and authorization)
	mcpController := &controllers.MCPController{}
	beego.Router("/mcp/tools", mcpController, "get:ListTools")
	beego.Router("/mcp/tools/call", mcpController, "post:CallTool")

	// Run the application
	beego.Run()
}
