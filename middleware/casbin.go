package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/beego/beego"
	beegoCtx "github.com/beego/beego/context"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// CasbinMiddleware handles authorization using Casbin
type CasbinMiddleware struct {
	enforcer *casbin.Enforcer
	// Skip authorization for these paths
	skipPaths []string
}

// NewCasbinMiddleware creates a new Casbin middleware instance
func NewCasbinMiddleware() *CasbinMiddleware {
	// Load model and policy from configuration
	modelPath := beego.AppConfig.DefaultString("casbin.model", "conf/casbin_model.conf")
	policyPath := beego.AppConfig.DefaultString("casbin.policy", "conf/casbin_policy.csv")

	// Create model
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		// Model file not found - creating a default RBAC model in memory
		// In production, consider failing fast if model file is missing to ensure proper configuration
		beego.Info("Casbin model file not found, creating default RBAC model in memory:", err)
		// Create default model
		m = getDefaultModel()
	}

	var enforcer *casbin.Enforcer

	// Create adapter
	adapter := fileadapter.NewAdapter(policyPath)

	// Create enforcer
	enforcer, err = casbin.NewEnforcer(m, adapter)
	if err != nil {
		beego.Info("Failed to create Casbin enforcer with file adapter, creating enforcer without file persistence:", err)
		// Create enforcer without adapter if it fails (useful for testing)
		// In production, you may want to fail fast instead
		enforcer, err = casbin.NewEnforcer(m)
		if err != nil {
			beego.Error("Failed to create Casbin enforcer:", err)
			panic(err)
		}
	} else {
		// Load policy only if adapter was created successfully
		err = enforcer.LoadPolicy()
		if err != nil {
			beego.Warn("Failed to load policy from file, enforcer will start with no policies:", err)
		}
	}

	return &CasbinMiddleware{
		enforcer: enforcer,
		skipPaths: []string{
			"/",
			"/oauth/login",
			"/oauth/callback",
			"/health",
		},
	}
}

// Filter is the middleware filter function for Beego
func (m *CasbinMiddleware) Filter(ctx *beegoCtx.Context) {
	// Check if path should skip authorization
	for _, path := range m.skipPaths {
		if ctx.Request.URL.Path == path {
			return
		}
	}

	// Get user from context (set by OAuth middleware)
	user := ctx.Input.GetData("user")
	if user == nil {
		m.forbidden(ctx, "User not authenticated")
		return
	}

	username, ok := user.(string)
	if !ok {
		m.forbidden(ctx, "Invalid user data")
		return
	}

	// Get request path and method
	path := ctx.Request.URL.Path
	method := ctx.Request.Method

	// Check permission using Casbin
	allowed, err := m.enforcer.Enforce(username, path, method)
	if err != nil {
		beego.Error("Casbin enforce error:", err)
		m.forbidden(ctx, "Authorization check failed")
		return
	}

	if !allowed {
		m.forbidden(ctx, "Permission denied")
		return
	}

	// User is authorized, continue to controller
}

// forbidden sends a forbidden response
func (m *CasbinMiddleware) forbidden(ctx *beegoCtx.Context, message string) {
	ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"error":   "forbidden",
		"message": message,
	}
	json.NewEncoder(ctx.ResponseWriter).Encode(response)
}

// getDefaultModel returns a default Casbin model
func getDefaultModel() model.Model {
	text := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	m, _ := model.NewModelFromString(text)
	return m
}

// AddPolicy adds a policy rule
func (m *CasbinMiddleware) AddPolicy(params ...string) (bool, error) {
	// Convert []string to []interface{}
	iParams := make([]interface{}, len(params))
	for i, v := range params {
		iParams[i] = v
	}
	return m.enforcer.AddPolicy(iParams...)
}

// RemovePolicy removes a policy rule
func (m *CasbinMiddleware) RemovePolicy(params ...string) (bool, error) {
	// Convert []string to []interface{}
	iParams := make([]interface{}, len(params))
	for i, v := range params {
		iParams[i] = v
	}
	return m.enforcer.RemovePolicy(iParams...)
}

// GetPolicy returns all policy rules
func (m *CasbinMiddleware) GetPolicy() [][]string {
	return m.enforcer.GetPolicy()
}

// SavePolicy saves the current policy to storage
func (m *CasbinMiddleware) SavePolicy() error {
	return m.enforcer.SavePolicy()
}

// GetEnforcer returns the underlying Casbin enforcer
func (m *CasbinMiddleware) GetEnforcer() *casbin.Enforcer {
	return m.enforcer
}
