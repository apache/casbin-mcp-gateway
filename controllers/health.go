package controllers

import (
	"github.com/beego/beego"
)

type HealthController struct {
	beego.Controller
}

func (c *HealthController) Get() {
	c.Data["json"] = map[string]interface{}{
		"status":  "ok",
		"service": "mcp-gateway",
		"version": "1.0.0",
	}
	c.ServeJSON()
}
