# MCP Gateway

A Golang HTTP gateway built with Beego 1.x framework, featuring a middleware system that supports OAuth authentication and Casbin authorization.

## Features

- **Beego 1.x Framework**: Built on the reliable Beego 1.x web framework
- **Middleware System**: Stackable HTTP middleware architecture
- **OAuth Authentication**: Supports OAuth 2.0 for user authentication (recommended by MCP)
- **Casbin Authorization**: Policy-based access control using Casbin
- **Request Flow**: OAuth authentication → User identification → Casbin authorization

## Architecture

The gateway implements a layered middleware approach:

1. **OAuth Middleware**: Authenticates requests using OAuth 2.0 tokens and identifies users
2. **Casbin Middleware**: Enforces authorization policies based on authenticated user identity

## Installation

```bash
go get github.com/casbin/mcp-gateway
```

## Configuration

### Application Configuration

Edit `conf/app.conf` to configure the gateway:

```ini
appname = mcp-gateway
httpport = 8080
runmode = dev

# OAuth2 Configuration
oauth2.client_id = YOUR_CLIENT_ID
oauth2.client_secret = YOUR_CLIENT_SECRET
oauth2.redirect_url = http://localhost:8080/oauth/callback
oauth2.auth_url = https://accounts.google.com/o/oauth2/auth
oauth2.token_url = https://accounts.google.com/o/oauth2/token

# Casbin Configuration
casbin.model = conf/casbin_model.conf
casbin.policy = conf/casbin_policy.csv
```

### Casbin Model

The Casbin model is defined in `conf/casbin_model.conf`. The default model uses RBAC:

```ini
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
```

### Casbin Policy

Define access control policies in `conf/casbin_policy.csv`:

```csv
p, demo_user, /api/resource, GET
p, demo_user, /api/resource, POST
p, admin, /api/resource, GET
p, admin, /api/resource, POST
p, admin, /api/admin, GET

g, admin_user, admin
```

Format: `p, subject, object, action` for policies and `g, user, role` for role assignments.

## Usage

### Running the Gateway

```bash
# Build the application
go build -o mcp-gateway

# Run the application
./mcp-gateway
```

The gateway will start on port 8080 (configurable in `conf/app.conf`).

### Making Authenticated Requests

Requests must include a Bearer token in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
     http://localhost:8080/api/resource
```

### API Endpoints

- `GET /` - Public endpoint, returns gateway information
- `GET /api/resource` - Protected endpoint, requires authentication and authorization
- `POST /api/resource` - Protected endpoint, requires authentication and authorization
- `GET /api/admin` - Admin-only endpoint, requires admin role

### Authentication Flow

1. Client requests a protected resource with an OAuth token
2. OAuth middleware validates the token and extracts user information
3. User identity is stored in the request context
4. Casbin middleware checks if the user has permission to access the resource
5. If authorized, the request proceeds to the controller

## Middleware System

### OAuth Middleware

The OAuth middleware (`middleware/oauth.go`):
- Validates Bearer tokens from the Authorization header
- Extracts user information from OAuth provider
- Stores user data in request context
- Can be configured to skip authentication for specific paths

### Casbin Middleware

The Casbin middleware (`middleware/casbin.go`):
- Enforces access control policies
- Checks permissions based on user, resource path, and HTTP method
- Uses RBAC model for role-based access control
- Supports dynamic policy management

### Adding Custom Middleware

To add custom middleware, use Beego's filter system:

```go
beego.InsertFilter("*", beego.BeforeRouter, yourMiddleware.Filter)
```

Middleware runs in the order they are registered.

## Development

### Project Structure

```
mcp-gateway/
├── main.go              # Application entry point
├── conf/                # Configuration files
│   ├── app.conf        # Application configuration
│   ├── casbin_model.conf  # Casbin model definition
│   └── casbin_policy.csv  # Casbin policies
├── middleware/          # Middleware implementations
│   ├── oauth.go        # OAuth authentication middleware
│   └── casbin.go       # Casbin authorization middleware
├── controllers/         # Request handlers
│   ├── main.go         # Main controller
│   ├── resource.go     # Resource controller
│   └── admin.go        # Admin controller
├── go.mod              # Go module definition
└── README.md           # This file
```

### Building

```bash
go build -o mcp-gateway
```

### Testing

```bash
go test ./...
```

## Security Considerations

1. **OAuth Configuration**: Ensure OAuth client credentials are kept secure and not committed to version control
2. **Token Validation**: In production, implement proper token validation with the OAuth provider
3. **HTTPS**: Always use HTTPS in production to protect tokens in transit
4. **Policy Management**: Regularly review and update Casbin policies
5. **Audit Logging**: Consider adding audit logging for authorization decisions

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.