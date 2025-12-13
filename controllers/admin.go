package controllers

import (
	"github.com/beego/beego"
)

// AdminController handles admin-related requests
type AdminController struct {
	beego.Controller
}

// Get handles GET requests to /api/admin
func (c *AdminController) Get() {
	user := c.Ctx.Input.GetData("user")

	c.Data["json"] = map[string]interface{}{
		"message": "Admin panel accessed",
		"user":    user,
		"stats": map[string]interface{}{
			"total_users":     100,
			"active_sessions": 25,
		},
	}
	c.ServeJSON()
}
