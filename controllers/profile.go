package controllers

import (
	"github.com/beego/beego"
)

type ProfileController struct {
	beego.Controller
}

func (c *ProfileController) Get() {
	username := c.Ctx.Input.GetData("username").(string)
	
	// Find user profile
	for _, user := range users {
		if user.Username == username {
			c.Data["json"] = user
			c.ServeJSON()
			return
		}
	}

	c.Data["json"] = map[string]string{"error": "Profile not found"}
	c.Ctx.Output.SetStatus(404)
	c.ServeJSON()
}

func (c *ProfileController) Put() {
	username := c.Ctx.Input.GetData("username").(string)
	
	var updatedProfile User
	if err := c.ParseForm(&updatedProfile); err != nil {
		c.Data["json"] = map[string]string{"error": "Invalid request"}
		c.Ctx.Output.SetStatus(400)
		c.ServeJSON()
		return
	}

	// Update user profile
	for i, user := range users {
		if user.Username == username {
			users[i].Email = updatedProfile.Email
			c.Data["json"] = users[i]
			c.ServeJSON()
			return
		}
	}

	c.Data["json"] = map[string]string{"error": "Profile not found"}
	c.Ctx.Output.SetStatus(404)
	c.ServeJSON()
}
