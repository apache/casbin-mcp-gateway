# mcp-gateway

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/mcp-gateway)](https://goreportcard.com/report/github.com/casbin/mcp-gateway)
[![Build](https://github.com/casbin/mcp-gateway/actions/workflows/ci.yml/badge.svg)](https://github.com/casbin/mcp-gateway/actions/workflows/ci.yml)
[![Coverage Status](https://codecov.io/gh/casbin/mcp-gateway/branch/master/graph/badge.svg)](https://codecov.io/gh/casbin/mcp-gateway)
[![Godoc](https://godoc.org/github.com/casbin/mcp-gateway?status.svg)](https://pkg.go.dev/github.com/casbin/mcp-gateway)
[![Release](https://img.shields.io/github/release/casbin/mcp-gateway.svg)](https://github.com/casbin/mcp-gateway/releases/latest)
[![License](https://img.shields.io/github/license/casbin/mcp-gateway.svg)](https://github.com/casbin/mcp-gateway/blob/master/LICENSE)

A Golang HTTP gateway with OAuth authentication and Casbin authorization middleware, built with Beego web framework and a React frontend.

## Features

- **HTTP Middleware System**: Stackable middleware architecture
- **OAuth Authentication**: JWT-based authentication middleware following MCP recommendations
- **Casbin Authorization**: Role-based access control (RBAC) for fine-grained permission management
- **Modern Frontend**: React + Tailwind CSS + shadcn/ui for a beautiful user interface
- **Dual Serving**: Backend serves both APIs and frontend static files on port 9000
- **CI/CD**: Automated testing and semantic versioning with GitHub Actions

## Architecture

```
┌─────────────────┐
│   Client        │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────┐
│  MCP Gateway (Port 9000)    │
│  ┌───────────────────────┐  │
│  │  OAuth Middleware     │  │
│  │  (JWT Authentication) │  │
│  └──────────┬────────────┘  │
│             ▼               │
│  ┌───────────────────────┐  │
│  │  Casbin Middleware    │  │
│  │  (Authorization)      │  │
│  └──────────┬────────────┘  │
│             ▼               │
│  ┌───────────────────────┐  │
│  │  API Controllers      │  │
│  │  + Static Files       │  │
│  └───────────────────────┘  │
└─────────────────────────────┘
```

## Prerequisites

- Go 1.21.0 or later
- Node.js 20 or later
- npm or yarn

## Quick Start

### Backend Setup

1. Clone the repository:
```bash
git clone https://github.com/casbin/mcp-gateway.git
cd mcp-gateway
```

2. Install Go dependencies:
```bash
go mod download
```

3. Build the backend:
```bash
go build -o mcp-gateway .
```

4. Run the backend:
```bash
./mcp-gateway
```

The backend will start on port 9000.

### Frontend Setup

1. Navigate to the web directory:
```bash
cd web
```

2. Install dependencies:
```bash
npm install
```

3. For development mode (runs on port 8001):
```bash
npm run dev
```

4. For production build:
```bash
npm run build
```

The built files will be in `web/dist` and automatically served by the backend on port 9000.

## Usage

### Demo Accounts

Two demo accounts are available for testing:

- **Admin**: `alice` / `password123` (has admin role with full access)
- **User**: `bob` / `password456` (has user role with limited access)

### API Endpoints

#### Public Endpoints

- `GET /health` - Health check endpoint
- `POST /login` - Login endpoint

#### Protected Endpoints (require authentication)

- `GET /api/users` - List all users (admin only)
- `GET /api/users/:id` - Get specific user (admin only)
- `POST /api/users` - Create new user (admin only)
- `GET /api/profile` - Get current user profile
- `PUT /api/profile` - Update current user profile

### Authentication

All protected endpoints require a JWT token in the Authorization header:

```bash
Authorization: Bearer <token>
```

Example login request:

```bash
curl -X POST http://localhost:9000/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}'
```

Example authenticated request:

```bash
curl http://localhost:9000/api/users \
  -H "Authorization: Bearer <your-token>"
```

## Configuration

### OAuth Configuration

Edit `conf/app.conf` to configure OAuth settings:

```ini
[dev]
oauth.secret = your-secret-key-dev

[prod]
oauth.secret = ${OAUTH_SECRET||your-secret-key}
```

### Casbin Policies

Edit `conf/policy.csv` to modify access control rules:

```csv
p, admin, /api/*, GET
p, admin, /api/*, POST
p, user, /api/profile, GET
g, alice, admin
g, bob, user
```

The format is: `p, role, resource, action` for permissions and `g, user, role` for role assignments.

## Development

### Running Tests

```bash
go test ./... -v
```

### Running with Coverage

```bash
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
```

### Frontend Development

```bash
cd web
npm run dev
```

This will start the development server on port 8001 with hot module replacement.

## Deployment

### Production Build

1. Build the backend:
```bash
go build -ldflags="-s -w" -o mcp-gateway .
```

2. Build the frontend:
```bash
cd web
npm run build
cd ..
```

3. Run the application:
```bash
./mcp-gateway
```

The application will serve both backend APIs and frontend static files on port 9000.

### Docker (Optional)

You can containerize the application:

```dockerfile
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o mcp-gateway .

FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=backend-builder /app/mcp-gateway .
COPY --from=frontend-builder /app/dist ./web/dist
COPY conf ./conf
EXPOSE 9000
CMD ["./mcp-gateway"]
```

## License

Apache-2.0 License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
