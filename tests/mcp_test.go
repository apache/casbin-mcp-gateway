package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego"
	beegoCtx "github.com/beego/beego/context"
	"github.com/casbin/mcp-gateway/controllers"
)

func TestMCPProxyListTools(t *testing.T) {
	// Create a mock upstream server
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp/tools" && r.Method == "GET" {
			// Check that user headers are forwarded
			if r.Header.Get("X-User") != "test_user" {
				t.Error("Expected X-User header to be forwarded")
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"tools": ["tool1", "tool2"]}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer upstreamServer.Close()

	// Configure upstream URL
	beego.AppConfig.Set("mcp.upstream_url", upstreamServer.URL)

	// Create MCP controller
	mcpController := &controllers.MCPController{}

	// Create test request
	req := httptest.NewRequest("GET", "/mcp/tools", nil)
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	// Set authenticated user in context
	ctx.Input.SetData("user", "test_user")
	ctx.Input.SetData("user_id", "user123")
	ctx.Input.SetData("user_email", "test@example.com")

	// Set controller context
	mcpController.Ctx = ctx
	mcpController.Data = make(map[interface{}]interface{})

	// Call ListTools
	mcpController.ListTools()

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "tool1") {
		t.Error("Expected response to contain tool1")
	}
}

func TestMCPProxyCallTool(t *testing.T) {
	// Create a mock upstream server
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp/tools/call" && r.Method == "POST" {
			// Check that user headers are forwarded
			if r.Header.Get("X-User") != "test_user" {
				t.Error("Expected X-User header to be forwarded")
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result": "success"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer upstreamServer.Close()

	// Configure upstream URL
	beego.AppConfig.Set("mcp.upstream_url", upstreamServer.URL)

	// Create MCP controller
	mcpController := &controllers.MCPController{}

	// Create test request with body
	body := strings.NewReader(`{"tool": "test_tool", "params": {}}`)
	req := httptest.NewRequest("POST", "/mcp/tools/call", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	// Set authenticated user in context
	ctx.Input.SetData("user", "test_user")
	ctx.Input.SetData("user_id", "user123")

	// Set controller context
	mcpController.Ctx = ctx
	mcpController.Data = make(map[interface{}]interface{})

	// Call CallTool
	mcpController.CallTool()

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "success") {
		t.Error("Expected response to contain success")
	}
}

func TestMCPProxyUpstreamError(t *testing.T) {
	// Configure invalid upstream URL
	beego.AppConfig.Set("mcp.upstream_url", "http://localhost:99999")

	// Create MCP controller
	mcpController := &controllers.MCPController{}

	// Create test request
	req := httptest.NewRequest("GET", "/mcp/tools", nil)
	w := httptest.NewRecorder()

	ctx := &beegoCtx.Context{
		Request:        req,
		ResponseWriter: &beegoCtx.Response{ResponseWriter: w},
		Input:          beegoCtx.NewInput(),
		Output:         beegoCtx.NewOutput(),
	}
	ctx.Reset(w, req)

	// Set authenticated user in context
	ctx.Input.SetData("user", "test_user")

	// Set controller context
	mcpController.Ctx = ctx
	mcpController.Data = make(map[interface{}]interface{})

	// Call ListTools
	mcpController.ListTools()

	// Check that error response is returned
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502 for upstream error, got %d", w.Code)
	}
}
