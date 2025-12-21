package controllers

import (
	"github.com/beego/beego"
)

type UserController struct {
	beego.Controller
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Mock user data
var users = []User{
	{ID: "1", Username: "alice", Email: "alice@example.com", Role: "admin"},
	{ID: "2", Username: "bob", Email: "bob@example.com", Role: "user"},
}

func (c *UserController) Get() {
	id := c.Ctx.Input.Param(":id")
	
	if id == "" {
		// Return all users
		c.Data["json"] = users
		c.ServeJSON()
		return
	}

	// Return specific user
	for _, user := range users {
		if user.ID == id {
			c.Data["json"] = user
			c.ServeJSON()
			return
		}
	}

	c.Data["json"] = map[string]string{"error": "User not found"}
	c.Ctx.Output.SetStatus(404)
	c.ServeJSON()
}

func (c *UserController) Post() {
	var user User
	if err := c.ParseForm(&user); err != nil {
		c.Data["json"] = map[string]string{"error": "Invalid request"}
		c.Ctx.Output.SetStatus(400)
		c.ServeJSON()
		return
	}

	// In production, save to database
	users = append(users, user)
	
	c.Data["json"] = user
	c.Ctx.Output.SetStatus(201)
	c.ServeJSON()
}
