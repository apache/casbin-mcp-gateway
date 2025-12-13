package controllers

import (
	"github.com/beego/beego"
)

// ResourceController handles resource-related requests
type ResourceController struct {
	beego.Controller
}

// Get handles GET requests to /api/resource
func (c *ResourceController) Get() {
	// Get user information from context (set by OAuth middleware)
	user := c.Ctx.Input.GetData("user")
	userEmail := c.Ctx.Input.GetData("user_email")

	c.Data["json"] = map[string]interface{}{
		"message": "Resource data accessed successfully",
		"user":    user,
		"email":   userEmail,
		"data": []map[string]string{
			{"id": "1", "name": "Resource 1"},
			{"id": "2", "name": "Resource 2"},
		},
	}
	c.ServeJSON()
}

// Post handles POST requests to /api/resource
func (c *ResourceController) Post() {
	user := c.Ctx.Input.GetData("user")

	c.Data["json"] = map[string]interface{}{
		"message": "Resource created successfully",
		"user":    user,
	}
	c.ServeJSON()
}
