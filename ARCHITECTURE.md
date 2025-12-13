# MCP Gateway Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         MCP Gateway                              │
│                       (Beego 1.x Framework)                      │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow

```
┌──────────┐
│  Client  │
└────┬─────┘
     │
     │ HTTP Request with Bearer Token
     │
     ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Beego Router                               │
└────┬────────────────────────────────────────────────────────┬───┘
     │                                                         │
     │ Public Routes                                           │ Protected Routes
     │ (/,/oauth/login,/oauth/callback)                       │ (/api/*)
     │                                                         │
     ▼                                                         ▼
┌─────────────────┐                            ┌──────────────────────────────┐
│  Controllers    │                            │   Middleware Stack           │
│                 │                            │                              │
│  - MainCtrl     │                            │  ┌────────────────────────┐  │
│  - OAuthCtrl    │                            │  │ 1. OAuth Middleware    │  │
└─────────────────┘                            │  │    - Extract token     │  │
                                               │  │    - Validate token    │  │
                                               │  │    - Get user info     │  │
                                               │  │    - Set context       │  │
                                               │  └──────────┬─────────────┘  │
                                               │             │                 │
                                               │             ▼                 │
                                               │  ┌────────────────────────┐  │
                                               │  │ 2. Casbin Middleware   │  │
                                               │  │    - Get user from ctx │  │
                                               │  │    - Check policies    │  │
                                               │  │    - Enforce RBAC      │  │
                                               │  │    - Allow/Deny        │  │
                                               │  └──────────┬─────────────┘  │
                                               └─────────────┼─────────────────┘
                                                             │
                                                             ▼
                                               ┌──────────────────────────────┐
                                               │      Controllers             │
                                               │                              │
                                               │  - ResourceController        │
                                               │  - AdminController           │
                                               └──────────────────────────────┘
```

## Component Interaction

```
┌──────────────────────┐         ┌──────────────────────┐
│  OAuth Provider      │         │  Casbin Engine       │
│  (Google/GitHub/etc) │         │                      │
│                      │         │  ┌────────────────┐  │
│  - Authenticate user │         │  │ Model          │  │
│  - Issue token       │◄────────┼──│ (RBAC)         │  │
│  - Userinfo endpoint │         │  └────────────────┘  │
└──────────────────────┘         │                      │
                                 │  ┌────────────────┐  │
                                 │  │ Policy         │  │
                                 │  │ (CSV/Database) │  │
                                 │  └────────────────┘  │
                                 └──────────────────────┘
```

## Middleware Execution Order

```
Request → Router → [OAuth Auth] → [Casbin Authz] → Controller → Response
                         │               │
                         │               └─ Fails: 403 Forbidden
                         └─ Fails: 401 Unauthorized
```

## OAuth Authentication Flow

```
1. User Login
   ┌──────┐                    ┌──────────┐                  ┌──────────────┐
   │Client│                    │ Gateway  │                  │OAuth Provider│
   └──┬───┘                    └────┬─────┘                  └──────┬───────┘
      │                             │                                │
      │  GET /oauth/login           │                                │
      ├────────────────────────────►│                                │
      │                             │                                │
      │  Redirect to OAuth provider │                                │
      │◄────────────────────────────┤                                │
      │                             │                                │
      │  Authorize App              │                                │
      ├────────────────────────────────────────────────────────────► │
      │                             │                                │
      │  Authorization Code         │                                │
      │◄────────────────────────────────────────────────────────────┤
      │                             │                                │
      │  GET /oauth/callback?code=X │                                │
      ├────────────────────────────►│                                │
      │                             │  Exchange code for token       │
      │                             ├───────────────────────────────►│
      │                             │                                │
      │                             │  Access Token                  │
      │                             │◄───────────────────────────────┤
      │  Access Token               │                                │
      │◄────────────────────────────┤                                │
      │                             │                                │

2. API Request with Token
   ┌──────┐                    ┌──────────┐
   │Client│                    │ Gateway  │
   └──┬───┘                    └────┬─────┘
      │                             │
      │  GET /api/resource          │
      │  Authorization: Bearer XXX  │
      ├────────────────────────────►│
      │                             │
      │                             │ [OAuth Middleware]
      │                             │ - Validate token
      │                             │ - Extract user: "alice"
      │                             │
      │                             │ [Casbin Middleware]
      │                             │ - Check: alice, /api/resource, GET
      │                             │ - Policy match: ✓ ALLOW
      │                             │
      │  Resource Data              │
      │◄────────────────────────────┤
      │                             │
```

## Casbin RBAC Model

```
┌─────────────────────────────────────────────────────────────────┐
│                    Casbin Authorization                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Users                  Roles              Resources             │
│  ┌──────┐              ┌──────┐          ┌──────────────┐       │
│  │alice │─────────────►│ user │─────────►│ /api/resource│ GET   │
│  └──────┘              └──────┘          └──────────────┘       │
│                                                                  │
│  ┌──────┐              ┌──────┐          ┌──────────────┐       │
│  │ bob  │─────────────►│ user │─────────►│ /api/resource│ POST  │
│  └──────┘              └──────┘          └──────────────┘       │
│                                                                  │
│  ┌──────┐              ┌──────┐          ┌──────────────┐       │
│  │admin │─────────────►│admin │─────────►│ /api/admin   │ GET   │
│  └──────┘              └──────┘          └──────────────┘       │
│                                          ┌──────────────┐        │
│                                          │ /api/resource│ *      │
│                                          └──────────────┘        │
│                                                                  │
│  Request: (alice, /api/resource, GET)                           │
│  ↓                                                               │
│  Check: alice → user role?  ✓                                   │
│  Check: user has /api/resource GET? ✓                           │
│  Result: ALLOW                                                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration Files

```
conf/
├── app.conf              # Main application config
│   ├── Server settings   (port, runmode)
│   ├── OAuth settings    (client_id, client_secret, endpoints)
│   └── Casbin settings   (model path, policy path)
│
├── casbin_model.conf     # Casbin RBAC model definition
│   ├── Request definition  (subject, object, action)
│   ├── Policy definition   (permission rules)
│   ├── Role definition     (role hierarchy)
│   ├── Policy effect       (allow/deny logic)
│   └── Matchers           (matching rules)
│
└── casbin_policy.csv     # Casbin policies
    ├── Permissions       (p, subject, object, action)
    └── Role assignments  (g, user, role)
```

## Security Layers

```
┌────────────────────────────────────────────────┐
│                                                │
│  1. Transport Security (HTTPS)                │
│     └─ TLS encryption for all traffic         │
│                                                │
├────────────────────────────────────────────────┤
│                                                │
│  2. Authentication (OAuth Middleware)          │
│     ├─ Token validation                       │
│     ├─ User identification                    │
│     └─ Session management                     │
│                                                │
├────────────────────────────────────────────────┤
│                                                │
│  3. Authorization (Casbin Middleware)          │
│     ├─ RBAC enforcement                       │
│     ├─ Resource-level permissions             │
│     └─ Action-based control                   │
│                                                │
├────────────────────────────────────────────────┤
│                                                │
│  4. Application Logic (Controllers)            │
│     └─ Business logic execution               │
│                                                │
└────────────────────────────────────────────────┘
```

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Internet                             │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌───────────────────────────────────────────────────────────────┐
│                      Load Balancer                            │
│                    (HTTPS Termination)                        │
└───────────────────────────┬───────────────────────────────────┘
                            │
           ┌────────────────┴────────────────┐
           │                                  │
           ▼                                  ▼
┌─────────────────────┐          ┌─────────────────────┐
│  MCP Gateway (1)    │          │  MCP Gateway (2)    │
│  - OAuth Middleware │          │  - OAuth Middleware │
│  - Casbin Middleware│          │  - Casbin Middleware│
└──────────┬──────────┘          └──────────┬──────────┘
           │                                  │
           └────────────────┬─────────────────┘
                            │
           ┌────────────────┴────────────────┐
           │                                  │
           ▼                                  ▼
┌─────────────────────┐          ┌─────────────────────┐
│  OAuth Provider     │          │  Policy Store       │
│  (External)         │          │  (File/Database)    │
└─────────────────────┘          └─────────────────────┘
```
