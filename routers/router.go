package routers

import (
	"github.com/beego/beego"
	"github.com/casbin/mcp-gateway/controllers"
)

func Init() {
	// Health check endpoint
	beego.Router("/health", &controllers.HealthController{})
	
	// Auth endpoints
	beego.Router("/login", &controllers.AuthController{}, "post:Login")
	
	// API endpoints
	ns := beego.NewNamespace("/api",
		beego.NSRouter("/users", &controllers.UserController{}),
		beego.NSRouter("/users/:id", &controllers.UserController{}),
		beego.NSRouter("/profile", &controllers.ProfileController{}),
	)
	beego.AddNamespace(ns)
}
