# MCP Gateway - Production Setup Guide

This guide provides instructions for setting up the MCP Gateway in a production environment with real OAuth authentication.

## Prerequisites

- Go 1.16 or later
- OAuth 2.0 provider credentials (Google, GitHub, Auth0, etc.)
- PostgreSQL or MySQL (optional, for Casbin policy persistence)

## Step 1: Configure OAuth Provider

### Using Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to "APIs & Services" > "Credentials"
4. Create OAuth 2.0 Client ID
5. Add authorized redirect URI: `http://your-domain.com/oauth/callback`
6. Save the Client ID and Client Secret

### Using GitHub OAuth

1. Go to GitHub Settings > Developer settings > OAuth Apps
2. Create a new OAuth App
3. Set Authorization callback URL: `http://your-domain.com/oauth/callback`
4. Save the Client ID and Client Secret

## Step 2: Configure Application

Update `conf/app.conf` with your OAuth credentials:

```ini
appname = mcp-gateway
httpport = 8080
runmode = prod

# Google OAuth Configuration
oauth2.client_id = YOUR_GOOGLE_CLIENT_ID
oauth2.client_secret = YOUR_GOOGLE_CLIENT_SECRET
oauth2.redirect_url = https://your-domain.com/oauth/callback
oauth2.auth_url = https://accounts.google.com/o/oauth2/auth
oauth2.token_url = https://accounts.google.com/o/oauth2/token

# GitHub OAuth Configuration (alternative)
# oauth2.client_id = YOUR_GITHUB_CLIENT_ID
# oauth2.client_secret = YOUR_GITHUB_CLIENT_SECRET
# oauth2.redirect_url = https://your-domain.com/oauth/callback
# oauth2.auth_url = https://github.com/login/oauth/authorize
# oauth2.token_url = https://github.com/login/oauth/access_token

# Casbin Configuration
casbin.model = conf/casbin_model.conf
casbin.policy = conf/casbin_policy.csv
```

## Step 3: Update OAuth Middleware for Production

Replace the mock validation in `middleware/oauth.go` with real validation:

```go
func (m *OAuthMiddleware) validateToken(ctx context.Context, tokenString string) (*UserInfo, error) {
    // Create OAuth2 token
    token := &oauth2.Token{AccessToken: tokenString}
    
    // Create HTTP client with token
    client := m.config.Client(ctx, token)
    
    // Call userinfo endpoint (example for Google)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode)
    }
    
    // Parse response
    var userInfo UserInfo
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return nil, fmt.Errorf("failed to decode user info: %w", err)
    }
    
    return &userInfo, nil
}
```

For GitHub:
```go
// GitHub userinfo endpoint
resp, err := client.Get("https://api.github.com/user")
```

## Step 4: Configure Casbin Policies

Edit `conf/casbin_policy.csv` to define your access control rules:

```csv
# Format: p, subject, resource, action

# Basic user permissions
p, user, /api/resource, GET
p, user, /api/resource, POST

# Admin permissions
p, admin, /api/resource, GET
p, admin, /api/resource, POST
p, admin, /api/resource, DELETE
p, admin, /api/admin, GET
p, admin, /api/admin, POST

# Role assignments
# Format: g, user, role
g, alice@example.com, user
g, bob@example.com, user
g, admin@example.com, admin
```

## Step 5: Enable HTTPS

### Using Let's Encrypt with Certbot

```bash
# Install certbot
sudo apt-get install certbot

# Get SSL certificate
sudo certbot certonly --standalone -d your-domain.com

# Update app.conf
EnableHTTPS = true
HTTPSPort = 8443
HTTPSCertFile = "/etc/letsencrypt/live/your-domain.com/fullchain.pem"
HTTPSKeyFile = "/etc/letsencrypt/live/your-domain.com/privkey.pem"
```

### Using Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Step 6: Deploy

### Using Systemd

Create `/etc/systemd/system/mcp-gateway.service`:

```ini
[Unit]
Description=MCP Gateway
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/mcp-gateway
ExecStart=/opt/mcp-gateway/mcp-gateway
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable mcp-gateway
sudo systemctl start mcp-gateway
sudo systemctl status mcp-gateway
```

### Using Docker

Create `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-gateway

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mcp-gateway .
COPY --from=builder /app/conf ./conf
EXPOSE 8080
CMD ["./mcp-gateway"]
```

Build and run:
```bash
docker build -t mcp-gateway .
docker run -d -p 8080:8080 -v $(pwd)/conf:/root/conf mcp-gateway
```

## Step 7: Testing

### Get OAuth Token

1. Navigate to OAuth provider's authorization URL
2. Authorize the application
3. Exchange authorization code for access token
4. Use the access token in API requests

### Test API Endpoints

```bash
# Test public endpoint
curl https://your-domain.com/

# Test protected endpoint with token
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
     https://your-domain.com/api/resource

# Test admin endpoint (requires admin role)
curl -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
     https://your-domain.com/api/admin
```

## Step 8: Monitor and Maintain

### Logging

Add logging middleware to track requests:
```go
beego.InsertFilter("*", beego.BeforeRouter, func(ctx *context.Context) {
    beego.Info("Request:", ctx.Request.Method, ctx.Request.URL.Path)
})
```

### Metrics

Consider adding Prometheus metrics:
```bash
go get github.com/prometheus/client_golang/prometheus
```

### Policy Updates

Reload policies without restarting:
```go
// In your admin controller
func (c *AdminController) ReloadPolicies() {
    casbinMiddleware.GetEnforcer().LoadPolicy()
    c.Data["json"] = map[string]string{"status": "policies reloaded"}
    c.ServeJSON()
}
```

## Troubleshooting

### Common Issues

1. **Token validation fails**: Check OAuth credentials and redirect URI
2. **Permission denied**: Verify Casbin policies are correctly configured
3. **Port already in use**: Check if another service is using the port
4. **SSL certificate errors**: Ensure certificates are valid and properly configured

### Debug Mode

Enable debug logging:
```ini
runmode = dev
```

Check logs:
```bash
tail -f logs/mcp-gateway.log
```

## Security Checklist

- [ ] OAuth credentials stored securely (environment variables or secrets manager)
- [ ] HTTPS enabled for all production traffic
- [ ] Token validation implemented properly
- [ ] Casbin policies reviewed and tested
- [ ] Rate limiting configured
- [ ] Logging and monitoring enabled
- [ ] Regular security updates applied
- [ ] Backup strategy for policies
