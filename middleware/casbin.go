package middleware

import (
	"path/filepath"

	"github.com/beego/beego"
	"github.com/beego/beego/context"
	"github.com/casbin/casbin/v2"
)

var enforcer *casbin.Enforcer

// InitCasbin initializes the Casbin enforcer
func InitCasbin() error {
	var err error
	modelPath := filepath.Join("conf", "model.conf")
	policyPath := filepath.Join("conf", "policy.csv")
	
	enforcer, err = casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return err
	}
	return nil
}

// Casbin middleware handles authorization
func Casbin() beego.FilterFunc {
	return func(ctx *context.Context) {
		// Skip authorization for public paths
		path := ctx.Request.URL.Path
		if isPublicPath(path) {
			return
		}

		// Get username from context (set by OAuth middleware)
		username, ok := ctx.Input.GetData("username").(string)
		if !ok || username == "" {
			ctx.Output.SetStatus(403)
			ctx.Output.JSON(map[string]string{"error": "User not authenticated"}, false, false)
			return
		}

		// Get request method and path
		method := ctx.Request.Method
		resource := path

		// Check permission using Casbin
		if enforcer == nil {
			// If enforcer is not initialized, allow all requests
			beego.Warn("Casbin enforcer not initialized, allowing request")
			return
		}

		allowed, err := enforcer.Enforce(username, resource, method)
		if err != nil {
			beego.Error("Casbin enforcement error:", err)
			ctx.Output.SetStatus(500)
			ctx.Output.JSON(map[string]string{"error": "Authorization check failed"}, false, false)
			return
		}

		if !allowed {
			ctx.Output.SetStatus(403)
			ctx.Output.JSON(map[string]string{"error": "Access denied"}, false, false)
			return
		}
	}
}

// GetEnforcer returns the Casbin enforcer instance
func GetEnforcer() *casbin.Enforcer {
	return enforcer
}
