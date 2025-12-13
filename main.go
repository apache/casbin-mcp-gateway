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

	// Register middleware with Beego
	// OAuth authentication middleware runs first
	beego.InsertFilter("*", beego.BeforeRouter, oauthMiddleware.Filter)
	// Casbin authorization middleware runs after authentication
	beego.InsertFilter("*", beego.BeforeRouter, casbinMiddleware.Filter)

	// Register routes
	beego.Router("/", &controllers.MainController{})
	beego.Router("/api/resource", &controllers.ResourceController{})
	beego.Router("/api/admin", &controllers.AdminController{})

	// Run the application
	beego.Run()
}
