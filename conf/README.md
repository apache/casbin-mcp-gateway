# Example Configuration

This directory contains example configuration files to help you get started.

## Quick Start

1. Copy `app.example.conf` to `conf/app.conf`
2. Update OAuth credentials in `app.conf`
3. Customize Casbin policies in `casbin_policy.csv` as needed
4. Run the gateway: `./mcp-gateway`

## Configuration Files

### app.example.conf
Example application configuration with placeholder OAuth credentials.
Copy this file to `conf/app.conf` and update with your actual credentials.

### Development vs Production

**Development:**
```ini
runmode = dev
httpport = 8080
```

**Production:**
```ini
runmode = prod
EnableHTTPS = true
HTTPSPort = 8443
HTTPSCertFile = "/path/to/cert.pem"
HTTPSKeyFile = "/path/to/key.pem"
```

## OAuth Provider Setup

### Google OAuth

1. Go to: https://console.cloud.google.com/
2. Create project
3. Enable OAuth 2.0
4. Create credentials
5. Add redirect URI: `http://localhost:8080/oauth/callback`

### GitHub OAuth

1. Go to: https://github.com/settings/developers
2. Register new OAuth application
3. Set callback URL: `http://localhost:8080/oauth/callback`

### Microsoft Azure AD

1. Go to: https://portal.azure.com/
2. Register application
3. Configure redirect URI: `http://localhost:8080/oauth/callback`

## Casbin Policy Examples

### Allow all authenticated users to read resources
```csv
p, user, /api/resource, GET
```

### Allow admins full access
```csv
p, admin, /api/*, GET
p, admin, /api/*, POST
p, admin, /api/*, PUT
p, admin, /api/*, DELETE
```

### Role-based policies
```csv
# Define permissions for roles
p, viewer, /api/resource, GET
p, editor, /api/resource, GET
p, editor, /api/resource, POST
p, admin, /api/*, *

# Assign roles to users
g, alice@example.com, viewer
g, bob@example.com, editor
g, admin@example.com, admin
```

## Environment Variables

You can also use environment variables (recommended for production):

```bash
export OAUTH_CLIENT_ID="your-client-id"
export OAUTH_CLIENT_SECRET="your-client-secret"
export CASBIN_MODEL_PATH="/etc/mcp-gateway/casbin_model.conf"
export CASBIN_POLICY_PATH="/etc/mcp-gateway/casbin_policy.csv"
```

Then reference them in app.conf:
```ini
oauth2.client_id = ${OAUTH_CLIENT_ID}
oauth2.client_secret = ${OAUTH_CLIENT_SECRET}
```
