package main

import (
	"github.com/beego/beego"
	"github.com/casbin/mcp-gateway/middleware"
	"github.com/casbin/mcp-gateway/routers"
)

func main() {
	// Initialize Casbin enforcer
	if err := middleware.InitCasbin(); err != nil {
		beego.Error("Failed to initialize Casbin:", err)
		// Continue running even if Casbin fails to initialize
		beego.Warn("Running without Casbin authorization")
	}

	// Set port
	beego.BConfig.Listen.HTTPPort = 9000

	// Enable static file serving
	beego.SetStaticPath("/", "web/dist")

	// Register middleware
	beego.InsertFilter("/*", beego.BeforeRouter, middleware.OAuth())
	beego.InsertFilter("/*", beego.BeforeRouter, middleware.Casbin())

	// Register routes
	routers.Init()

	// Run the application
	beego.Run()
}
