# Gin Boilerplate

A production-ready REST API boilerplate built with Go and Gin. Designed for teams who need to spin up new projects fast - copy the skeleton, pick the modules you need, and start building.

**20k+ RPS** on a single instance. Zero percent error rate at 10,000 concurrent users.

## Features

- Modular architecture - copy only the modules you need to a new project
- RBAC with permission-based access control (not role-based)
- JWT authentication with device fingerprinting and token revocation
- Raw SQL (no ORM) for full query control
- Generic response helpers with Go generics
- Comprehensive unit tests with go-sqlmock
- Swagger API documentation
- Graceful shutdown
- Hot reload with Air

## Tech Stack

| Component  | Technology                                                  |
| ---------- | ----------------------------------------------------------- |
| Framework  | [Gin](https://github.com/gin-gonic/gin)                     |
| Database   | PostgreSQL ([pgx](https://github.com/jackc/pgx))            |
| Cache      | Redis ([go-redis](https://github.com/redis/go-redis))       |
| Auth       | JWT ([golang-jwt](https://github.com/golang-jwt/jwt))       |
| Migration  | [golang-migrate](https://github.com/golang-migrate/migrate) |
| API Docs   | [Swagger](https://github.com/swaggo/swag)                   |
| Testing    | [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)        |
| Hot Reload | [Air](https://github.com/air-verse/air)                     |

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- [Air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [swag](https://github.com/swaggo/swag) (`go install github.com/swaggo/swag/cmd/swag@latest`)

### Installation

```bash
git clone https://github.com/your-username/gin-boilerplate.git
cd gin-boilerplate
```

Create `.env` from example:

```bash
cp .env.example .env
```

Start infrastructure:

```bash
make docker
```

Run migrations:

```bash
make migrate-up
```

Start the server:

```bash
make run
```

The API is now running at `http://localhost:8080`. Swagger docs at `http://localhost:8080/swagger/index.html`.

### Default Admin

```
Email:    admin@init.com
Password: Abc@1234
```

### Available Commands

```bash
make run            # Start with hot reload
make build          # Build binary to bin/api
make test           # Run all tests
make migrate-up     # Run pending migrations
make migrate-down   # Rollback 1 migration
make swag           # Regenerate Swagger docs
make docker         # Start PostgreSQL + Redis
```

## Project Structure

```
gin-boilerplate/
├── cmd/api/main.go              # Entry point, graceful shutdown
├── app/
│   ├── app.go                   # Bootstrap: DB, Redis, router, middleware
│   └── modules.go               # Module registration (only file importing internal/)
├── configs/
│   └── config.go                # Config from .env, panic if missing
├── pkg/                         # Shared infrastructure (the skeleton)
│   ├── auth/                    # JWT generation/parsing, password hashing
│   ├── cache/                   # Redis client
│   ├── db/                      # PostgreSQL connection pool
│   ├── deps/                    # Shared dependency struct
│   ├── middleware/               # Auth, permission, rate limit, device fingerprint
│   ├── response/                # Generic response helpers, pagination, FieldErr
│   └── validator/               # Custom validation rules
├── internal/                    # Feature modules (copy what you need)
│   ├── iam/                     # Identity & Access Management
│   │   ├── dto/                 # Request/response DTOs
│   │   ├── service/             # Business logic + raw SQL
│   │   ├── handler/             # HTTP handlers
│   │   ├── module.go            # Dependency wiring
│   │   └── routes.go            # Route registration
│   └── todo/                    # Todo module (same structure)
├── migrations/                  # SQL migrations (schema + seed)
├── k6/                          # Load test scripts
├── .claude/                     # AI code review rules + commands
├── CLAUDE.md                    # Project rules for AI assistants
└── docker-compose.yml           # PostgreSQL + Redis
```

### Module Structure

Every module in `internal/` follows the same pattern:

```
internal/<module>/
├── dto/
│   ├── <name>_request.go        # Request DTO with validation tags
│   └── <name>_response.go       # Response, DetailResponse DTOs
├── service/
│   ├── <name>_service.go        # Business logic + raw SQL queries
│   └── <name>_service_test.go   # Unit tests with go-sqlmock
├── handler/
│   └── <name>_handler.go        # Parse request, call service, return response
├── module.go                    # Wire service -> handler
└── routes.go                    # Register routes with middleware
```

## API Endpoints

### Auth

| Method | Path                      | Auth   | Description                                      |
| ------ | ------------------------- | ------ | ------------------------------------------------ |
| POST   | /api/auth/login           | -      | Login, returns access + refresh token            |
| POST   | /api/auth/register        | -      | Register new user (auto-assigned default role)   |
| POST   | /api/auth/refresh         | -      | Refresh tokens                                   |
| PUT    | /api/auth/change-password | Bearer | Change password, optionally logout other devices |

### Users

| Method | Path           | Permission  | Description                              |
| ------ | -------------- | ----------- | ---------------------------------------- |
| GET    | /api/users/me  | auth only   | Get current user profile                 |
| GET    | /api/users     | user:read   | List users (paginated)                   |
| GET    | /api/users/:id | user:read   | Get user detail with roles + permissions |
| POST   | /api/users     | user:create | Create user                              |
| PUT    | /api/users/:id | user:update | Update user                              |
| DELETE | /api/users/:id | user:delete | Delete user                              |

### Roles

| Method | Path           | Permission  | Description                             |
| ------ | -------------- | ----------- | --------------------------------------- |
| GET    | /api/roles     | role:read   | List roles                              |
| GET    | /api/roles/:id | role:read   | Get role detail with permissions        |
| POST   | /api/roles     | role:create | Create role                             |
| PUT    | /api/roles/:id | role:update | Update role (revokes affected users)    |
| DELETE | /api/roles/:id | role:delete | Delete role (only if no users assigned) |

### Permissions

| Method | Path             | Permission      | Description          |
| ------ | ---------------- | --------------- | -------------------- |
| GET    | /api/permissions | permission:read | List all permissions |

### Todos

| Method | Path           | Permission  | Description               |
| ------ | -------------- | ----------- | ------------------------- |
| GET    | /api/todos     | todo:read   | List my todos (paginated) |
| GET    | /api/todos/:id | todo:read   | Get todo (owner only)     |
| POST   | /api/todos     | todo:create | Create todo               |
| PUT    | /api/todos/:id | todo:update | Update todo (owner only)  |
| DELETE | /api/todos/:id | todo:delete | Delete todo (owner only)  |

## Highlights

### Permission-Based RBAC

Roles are just containers for permissions. The system never checks "is this user an admin?" - it checks "does this user have `user:delete` permission?". Adding or changing roles requires zero code changes.

```
Role: admin    (is_superadmin) -> automatically has ALL permissions
Role: user     (is_default)    -> user:read, role:read, todo:*
Role: custom   (created by admin) -> any combination of permissions
```

Users can also receive **direct permissions** beyond what their roles grant, via `user_permissions`.

**Protected system roles** (`is_system = true`) cannot be modified or deleted. Users cannot modify their own roles/permissions (prevents privilege escalation).

### Device Fingerprinting

Every token is bound to a device fingerprint: `SHA-256(X-Device-Id + IP + User-Agent)`. If an attacker steals a token but has a different IP or browser, the request is rejected.

```
Login from Chrome on 192.168.1.1  -> fingerprint: a1b2c3...
Attacker uses token from 10.0.0.1 -> fingerprint: x9y8z7... -> 401 Rejected
```

### Token Revocation

Tokens are revoked without a blacklist. On login/refresh, the server stores a timestamp in Redis. The auth middleware rejects any token issued before that timestamp.

- **Per-device**: `revoke:{userId}:{deviceId}` - logout one device
- **All devices**: `revoke:{userId}:*` - logout everywhere

Triggered automatically on: login (revoke old session), refresh, change password with `logoutOtherDevices: true`, user deletion, role permission changes.

### Superadmin Bypass

Roles with `is_superadmin = true` get all permissions injected into their JWT at token generation. No need to seed permissions for admin when adding new modules.

### Singleflight for Hot Endpoints

`GET /users/me` uses `singleflight` - if 1000 concurrent requests hit the same user ID, only 1 database query runs. Others wait and receive the same result. This doubled RPS from 11k to 21k in load tests.

### Field-Level Error Responses

Services return structured field errors that handlers render consistently:

```go
// In service
return response.NewFieldErr("email", "email already exists")

// Client receives
{
  "message": "validation failed",
  "details": [{"field": "email", "message": "email already exists"}]
}
```

Same format for both validation errors (binding tags) and business logic errors (duplicate email, wrong password).

### Generic Response Helpers

Type-safe response functions using Go generics - no `interface{}` or `any`:

```go
response.Success(c, user)                    // 200: direct data
response.List(c, users, meta)                // 200: {data: [], meta: {page, limit, total}}
response.BadRequest(c, "invalid input")      // 400: {message: "..."}
response.HandleError(c, err)                 // Auto-detects FieldErr vs generic error
```

### Null-Safe Arrays

All slice fields return `[]` instead of `null` when empty - no more `if data != null` checks on the client.

### Audit Columns

Every table has `id (UUID)`, `created_by`, `updated_by`, `created_at`, `updated_at`. Timestamps are managed by PostgreSQL triggers. `created_by`/`updated_by` have no foreign keys to keep modules independent.

### Portable Skeleton

The boilerplate is split into skeleton (always needed) and modules (copy as needed):

```bash
# New project: copy everything except internal/
cp -r gin-boilerplate/ new-project/ && rm -rf new-project/internal/
# Edit modules.go to empty -> server runs with zero routes

# Need auth? Copy the IAM module
cp -r gin-boilerplate/internal/iam/ new-project/internal/iam/
# Add 1 line to modules.go -> done
```

### AI-Assisted Code Review

Includes `.claude/` configuration and `CLAUDE.md` rules for AI code review. Claude (or other AI assistants) can review PRs against 39 checks covering architecture, performance, memory leaks, security, and code quality.

## Load Test Results

Tested with wrk on a real VPS (4 CPU / 8GB RAM), single instance, PostgreSQL + Redis on same server. Endpoint: `GET /api/users/me` (JWT auth + SHA-256 fingerprint + Redis revocation + DB query).

| Connections | RPS | Avg Latency | Errors |
| ----------- | ------ | ----------- | ------ |
| 100 | 11,001 | 9ms | 0 |
| 200 | 12,022 | 17ms | 0 |
| 500 | 12,431 | 41ms | 0 |
| 1,000 | 12,205 | 83ms | 0 |
| 2,000 | 12,374 | 164ms | 0 |
| 5,000 | 11,757 | 430ms | 0 |
| 10,000 | 11,543 | 872ms | 0 |
| 20,000 | 7,420 | 2.3s | 0 |

Run load tests:

```bash
# Login to get token
UA="bench"
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: test" \
  -H "User-Agent: $UA" \
  -d '{"email":"admin@init.com","password":"Abc@1234"}' | jq -r '.accessToken')

# 20k concurrent connections
wrk -t4 -c20000 -d60s --timeout 100s \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Device-Id: test" \
  -H "User-Agent: $UA" \
  http://localhost:8080/api/users/me
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Follow the rules in `CLAUDE.md`
4. Ensure tests pass (`make test`)
5. Create a Pull Request (use the PR template)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
