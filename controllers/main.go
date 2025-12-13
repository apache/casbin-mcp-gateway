package controllers

import (
	"github.com/beego/beego"
)

// MainController handles requests to the root path
type MainController struct {
	beego.Controller
}

// Get handles GET requests to /
func (c *MainController) Get() {
	c.Data["json"] = map[string]interface{}{
		"message": "Welcome to MCP Gateway",
		"version": "1.0.0",
	}
	c.ServeJSON()
}
