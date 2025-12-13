package controllers

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/beego/beego"
)

// MCPController handles MCP API proxy requests
type MCPController struct {
	beego.Controller
}

// getUpstreamURL returns the configured upstream MCP server URL
func getUpstreamURL() string {
	upstreamURL := beego.AppConfig.DefaultString("mcp.upstream_url", "http://localhost:3000")
	return upstreamURL
}

// proxyRequest forwards the request to the upstream MCP server
func (c *MCPController) proxyRequest(path string) {
	upstreamURL := getUpstreamURL()
	
	// Parse upstream URL
	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		beego.Error("Failed to parse upstream URL:", err)
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "invalid upstream URL"}
		c.ServeJSON()
		return
	}

	// Build the full upstream URL with path
	upstream.Path = path
	if c.Ctx.Request.URL.RawQuery != "" {
		upstream.RawQuery = c.Ctx.Request.URL.RawQuery
	}

	// Create new request to upstream
	var body io.Reader
	if c.Ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Ctx.Request.Body)
		if err != nil {
			beego.Error("Failed to read request body:", err)
			c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			c.Data["json"] = map[string]string{"error": "failed to read request body"}
			c.ServeJSON()
			return
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(c.Ctx.Request.Method, upstream.String(), body)
	if err != nil {
		beego.Error("Failed to create upstream request:", err)
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "failed to create upstream request"}
		c.ServeJSON()
		return
	}

	// Copy headers from original request
	for key, values := range c.Ctx.Request.Header {
		// Skip hop-by-hop headers
		if key == "Connection" || key == "Keep-Alive" || key == "Proxy-Authenticate" ||
			key == "Proxy-Authorization" || key == "Te" || key == "Trailers" ||
			key == "Transfer-Encoding" || key == "Upgrade" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add user information from context as custom headers
	if user := c.Ctx.Input.GetData("user"); user != nil {
		req.Header.Set("X-User", user.(string))
	}
	if userID := c.Ctx.Input.GetData("user_id"); userID != nil {
		req.Header.Set("X-User-ID", userID.(string))
	}
	if userEmail := c.Ctx.Input.GetData("user_email"); userEmail != nil {
		req.Header.Set("X-User-Email", userEmail.(string))
	}

	// Send request to upstream
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		beego.Error("Failed to send request to upstream:", err)
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]string{"error": "failed to reach upstream server"}
		c.ServeJSON()
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		// Skip hop-by-hop headers
		if key == "Connection" || key == "Keep-Alive" || key == "Proxy-Authenticate" ||
			key == "Proxy-Authorization" || key == "Te" || key == "Trailers" ||
			key == "Transfer-Encoding" || key == "Upgrade" {
			continue
		}
		for _, value := range values {
			c.Ctx.ResponseWriter.Header().Add(key, value)
		}
	}

	// Copy status code
	c.Ctx.ResponseWriter.WriteHeader(resp.StatusCode)

	// Copy response body
	io.Copy(c.Ctx.ResponseWriter, resp.Body)
}

// ListTools handles GET /mcp/tools - returns list of available MCP tools
func (c *MCPController) ListTools() {
	c.proxyRequest("/mcp/tools")
}

// CallTool handles POST /mcp/tools/call - executes an MCP tool call
func (c *MCPController) CallTool() {
	c.proxyRequest("/mcp/tools/call")
}
