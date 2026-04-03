# Gin Boilerplate - Enterprise Directory Structure

## Overview

REST API boilerplate viết bằng Go, sử dụng Gin framework. Thiết kế cho PM/team có thể tái sử dụng skeleton qua nhiều project, chỉ copy feature modules cần thiết.

### Tech Stack

| Component      | Choice                  |
| -------------- | ----------------------- |
| Language       | Go                      |
| Framework      | Gin                     |
| Database       | PostgreSQL (raw SQL)    |
| Cache          | Redis                   |
| Auth           | JWT                     |
| Authorization  | Permission-based (RBAC) |
| Migration      | golang-migrate          |
| API Docs       | Swagger (swag)          |
| Unit Test      | go-sqlmock              |
| Hot Reload     | air                     |

### Quyết định kiến trúc & lý do

| Quyết định                            | Lý do                                                                         |
| ------------------------------------- | ----------------------------------------------------------------------------- |
| Modular/Feature-based                 | Mỗi module tự chứa đầy đủ, copy độc lập sang project khác                    |
| Không dùng repository layer           | Tránh abstraction thừa, viết raw SQL trực tiếp trong service cho đơn giản     |
| Không dùng entity layer               | DTO đã đủ, service scan result SQL thẳng vào DTO                              |
| Không dùng ORM                        | Raw SQL cho toàn quyền kiểm soát query, tránh magic                           |
| Check permission, không check role    | Linh hoạt — thêm/sửa role chỉ cần thay đổi DB, không sửa code               |
| Không dùng ABAC middleware            | Tránh lặp query — ownership check nằm trong service, query 1 lần             |
| Config không có default value         | Explicit — thiếu env thì fail ngay, không chạy ngầm với giá trị không mong muốn |
| Response dùng Go generics `[T]`       | Type-safe tại compile time, không dùng `any`/`interface{}`                    |
| `app/` và `pkg/` ngoài `internal/`   | Skeleton chạy được mà không cần `internal/`, dễ tái sử dụng                  |
| Đăng ký module thủ công              | Explicit — mở `modules.go` là biết project có gì, không magic                 |
| Gin default logger                    | Đủ dùng, thêm structured logger sau khi cần                                  |
| Rate limit theo IP                    | Global, đặt trước auth, không cần đăng ký trong từng module                  |
| CORS config trong `app.go`            | Chỉ vài dòng, không cần tách file riêng                                       |

---

## Rules

### MUST

- Raw SQL trong service — không dùng repository, entity, ORM
- Config đọc từ `.env` — mọi biến bắt buộc, thiếu → panic
- Check permission (`permission("user:delete")`), không check role
- Response helper dùng generic `[T any]`
- Thao tác trên chính mình (GET /me) → chỉ cần `auth()`, không cần permission
- Ownership check (user chỉ sửa todo của mình) → xử lý trong service
- Unit test dùng `go-sqlmock`, file test cạnh source
- Module mới phải đăng ký trong `app/modules.go`
- Migration đặt tên: `{version}_{module}_{type}.{up|down}.sql`
- Migration order: base → schema → seed (per module)
- Bảng chính phải có audit columns: `id (UUID)`, `created_by`, `updated_by`, `created_at`, `updated_at`
- `id` → DB tự generate (`gen_random_uuid()`), `created_at/updated_at` → DB tự set (trigger)
- `created_by/updated_by` → service truyền từ JWT context, kiểu UUID, không FK

### MUST NOT

- Không tạo repository layer hoặc entity layer
- Không dùng ORM (gorm, ent, ...)
- Không dùng default value trong config (`PORT || 8080`)
- Không check role trực tiếp trong code
- Không dùng `any`/`interface{}` cho response helper
- Không dùng ABAC middleware
- Không dùng custom logger middleware
- Không đặt code non-module trong `internal/`
- Không dùng auto-discovery hoặc `init()` cho module registration

---

## Nguyên tắc skeleton/module

```
Skeleton (chạy được ngay)       = mọi thứ NGOÀI internal/
Feature modules (copy chọn lọc) = internal/<module>/
```

> **Bước 1:** Copy tất cả trừ `internal/` → sửa `app/modules.go` thành rỗng → chạy → không lỗi
> **Bước 2:** Copy `internal/iam/` vào → đăng ký trong `modules.go` → chạy → không lỗi
> **Bước 3:** Xóa `iam/`, copy `internal/todo/` → sửa `modules.go` → chạy → không lỗi

`modules.go` là file **duy nhất** cần sửa. Chỉ thêm/xóa 1 dòng import + 1 dòng đăng ký.

---

## Directory Tree

```
gin-boilerplate/
├── cmd/
│   └── api/
│       └── main.go                          # Entry point: load config, start server, graceful shutdown
│
├── app/                                     # ★ SKELETON - luôn có khi copy project
│   ├── app.go                               # Bootstrap: init DB, Redis, router, middleware (CORS ở đây)
│   └── modules.go                           # Đăng ký modules thủ công (file DUY NHẤT import internal/)
│
├── configs/
│   └── config.go                            # Struct định nghĩa config, đọc từ .env, panic nếu thiếu
│
├── pkg/                                     # ★ SKELETON - shared infrastructure
│   ├── cache/
│   │   └── redis.go                         # Redis client: init connection, helper methods
│   ├── db/
│   │   └── postgres.go                      # PostgreSQL: init connection pool, health check
│   ├── middleware/
│   │   ├── auth.go                          # Middleware: xác thực JWT token, lấy user info
│   │   ├── permission.go                    # Middleware: kiểm tra permission cụ thể
│   │   └── ratelimit.go                     # Middleware: giới hạn request rate theo IP (Redis-based)
│   ├── http/
│   │   ├── response.go                      # Helper: Success[T], List[T], BadRequest, Unauthorized, Forbidden, ValidationError
│   │   └── pagination.go                    # Helper: page/limit pagination
│   ├── auth/
│   │   ├── jwt.go                           # JWT: generate, validate, parse token
│   │   └── password.go                      # Bcrypt: hash & verify password
│   └── validator/
│       └── validator.go                     # Custom validator: đăng ký custom rules (vd: vn_phone)
│
├── docs/                                    # ★ Generated bởi swag init — nằm trong .gitignore
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── internal/                                # ★ CHỈ CHỨA feature modules - copy chọn lọc
│   ├── iam/                                 # Module: Identity & Access Management
│   │   ├── dto/
│   │   │   ├── auth_request.go              # DTO: login, register, refresh token request
│   │   │   ├── auth_response.go             # DTO: token response, user info response
│   │   │   ├── user_request.go              # DTO: create/update user request
│   │   │   ├── user_response.go             # DTO: user response
│   │   │   ├── role_request.go              # DTO: create/update role request
│   │   │   └── role_response.go             # DTO: role response
│   │   ├── service/
│   │   │   ├── auth_service.go              # Business logic + raw SQL: login, register, refresh token
│   │   │   ├── auth_service_test.go         # Unit test: auth service
│   │   │   ├── user_service.go              # Business logic + raw SQL: quản lý user
│   │   │   ├── user_service_test.go         # Unit test: user service
│   │   │   ├── role_service.go              # Business logic + raw SQL: quản lý role & permission
│   │   │   └── role_service_test.go         # Unit test: role service
│   │   ├── handler/
│   │   │   ├── auth_handler.go              # HTTP handler: auth endpoints
│   │   │   ├── user_handler.go              # HTTP handler: user CRUD endpoints
│   │   │   └── role_handler.go              # HTTP handler: role CRUD endpoints
│   │   ├── module.go                        # Wire service -> handler
│   │   └── routes.go                        # Đăng ký routes cho module IAM
│   │
│   └── todo/                                # Module: Todo Management
│       ├── dto/
│       │   ├── todo_request.go              # DTO: create/update/filter todo request
│       │   └── todo_response.go             # DTO: todo response, todo list response
│       ├── service/
│       │   ├── todo_service.go              # Business logic + raw SQL: quản lý todo
│       │   └── todo_service_test.go         # Unit test: todo service
│       ├── handler/
│       │   └── todo_handler.go              # HTTP handler: todo CRUD endpoints
│       ├── module.go                        # Wire service -> handler
│       └── routes.go                        # Đăng ký routes cho module Todo
│
├── migrations/
│   ├── 000001_base.up.sql                   # Base: function set_updated_at() trigger
│   ├── 000001_base.down.sql                 # Rollback: xóa function
│   ├── 000002_iam_schema.up.sql             # Schema: tạo bảng users, roles, permissions (có audit columns)
│   ├── 000002_iam_schema.down.sql           # Rollback: xóa bảng users, roles, permissions
│   ├── 000003_iam_seed.up.sql               # Seed: roles, permissions, admin user
│   ├── 000003_iam_seed.down.sql             # Rollback: xóa seed IAM
│   ├── 000004_todo_schema.up.sql            # Schema: tạo bảng todos (có audit columns)
│   ├── 000004_todo_schema.down.sql          # Rollback: xóa bảng todos
│   ├── 000005_todo_seed.up.sql              # Seed: permissions todo + gán vào roles
│   └── 000005_todo_seed.down.sql            # Rollback: xóa seed todo
│
├── .claude/
│   ├── settings.json                        # Claude Code: permissions, allowed/denied tools
│   └── commands/                            # Custom slash commands cho Claude Code
│       ├── review.md                        # /review — review code theo project rules
│       └── new-module.md                    # /new-module — tạo module mới theo chuẩn
│
├── .github/
│   └── PULL_REQUEST_TEMPLATE.md             # PR template: checklist cho member
│
├── CLAUDE.md                                # ★ Rules cho Claude review PR + Claude Code
├── .air.toml                                # Hot reload config (cosmtrek/air)
├── .env.example                             # Mẫu biến môi trường
├── .gitignore                               # Git ignore rules
├── docker-compose.yml                       # Docker Compose: postgres, redis
├── go.mod                                   # Go module definition
├── go.sum                                   # Go module checksums
├── Makefile                                 # Shortcuts: run, build, test, migrate, swag
└── README.md                                # Hướng dẫn setup, chạy project
```

---

## Dependency Flow

Ai import ai — đọc từ trái sang phải.

```
cmd/api/main.go
  └─→ app/app.go
        ├─→ configs/config.go
        ├─→ pkg/db/postgres.go
        ├─→ pkg/cache/redis.go
        ├─→ pkg/middleware/auth.go ──→ pkg/auth/jwt.go
        ├─→ pkg/middleware/permission.go
        ├─→ pkg/middleware/ratelimit.go ──→ pkg/cache/redis.go
        ├─→ pkg/validator/validator.go
        └─→ app/modules.go
              ├─→ internal/iam/module.go
              │     ├─→ iam/service/* ──→ pkg/db, pkg/auth, pkg/cache
              │     └─→ iam/handler/* ──→ pkg/http/response.go
              └─→ internal/todo/module.go
                    ├─→ todo/service/* ──→ pkg/db, pkg/cache
                    └─→ todo/handler/* ──→ pkg/http/response.go
```

**Quy tắc dependency:**
- `cmd/` → `app/` → `pkg/`, `configs/`
- `app/modules.go` → `internal/` (điểm nối duy nhất)
- `internal/<module>/` → `pkg/` (shared infrastructure)
- `internal/<module>/` **KHÔNG** import `internal/<module khác>/` (modules độc lập)
- `pkg/` **KHÔNG** import `internal/` hoặc `app/`

---

## Module Structure

Mỗi module trong `internal/` theo cấu trúc:

```
internal/<module>/
├── dto/
│   ├── <name>_request.go       # Request DTO + validation (binding tags)
│   └── <name>_response.go      # Response DTO
├── service/
│   ├── <name>_service.go       # Business logic + raw SQL
│   └── <name>_service_test.go  # Unit test (go-sqlmock)
├── handler/
│   └── <name>_handler.go       # HTTP handler (parse request → call service → response)
├── module.go                   # Wire: khởi tạo service, handler, inject dependencies
└── routes.go                   # Đăng ký routes cho module
```

**Data flow trong module:**

```
Request → Handler → Service (raw SQL) → Database
            ↓          ↓         ↑
     request DTO   response DTO
            ↓
         Response
```

---

## `app/modules.go` — đăng ký thủ công

```go
// Có cả iam + todo
package app

import (
    "project/internal/iam"
    "project/internal/todo"
)

func (a *App) registerModules() {
    iam.RegisterRoutes(a.router, a.db)
    todo.RegisterRoutes(a.router, a.db)
}
```

```go
// Project mới — chưa có module nào
package app

func (a *App) registerModules() {
}
```

```go
// Chỉ có iam
package app

import "project/internal/iam"

func (a *App) registerModules() {
    iam.RegisterRoutes(a.router, a.db)
}
```

---

## Middleware Order

```
Request → RateLimit (IP) → Auth (JWT) → Permission → Handler
```

- **RateLimit**: global, trước auth — chặn brute force, DDoS
- **Auth**: verify JWT, set user info vào context
- **Permission**: check user có permission cụ thể không
- Route chỉ cần biết user là ai (GET /me): dùng `auth()` only, không cần `permission()`

---

## Response Format

```
// 200 — single item (trả thẳng)
{"id": 1, "title": "Buy milk"}

// 200 — list (có pagination meta)
{"data": [...], "meta": {"page": 1, "limit": 20, "total": 55}}

// 400 — bad request
{"message": "invalid input"}

// 400 — validation error
{"message": "validation failed", "details": [{"field": "title", "message": "required"}]}

// 401
{"message": "token expired"}

// 403
{"message": "permission denied"}
```

---

## Workflow copy project mới

```bash
# Bước 1: Copy skeleton
cp -r gin-boilerplate/ new-project/
rm -rf new-project/internal/
# Sửa go.mod: đổi module name
# Sửa modules.go → để trống
# ✅ Server chạy OK — không route, không lỗi

# Bước 2: Cần IAM
cp -r gin-boilerplate/internal/iam/ new-project/internal/iam/
# Sửa modules.go → thêm 1 import + 1 dòng đăng ký
# ✅ IAM hoạt động

# Bước 3: Đổi ý, chỉ cần Todo
rm -rf new-project/internal/iam/
cp -r gin-boilerplate/internal/todo/ new-project/internal/todo/
# Sửa modules.go → đổi iam thành todo
# ✅ Todo hoạt động
```

---

## Mô tả chi tiết

### `cmd/`

Entry point. Import `app/` để bootstrap — không import `internal/`.

### `app/`

Bootstrap application. Nằm **ngoài** `internal/`.

- `app.go` — init DB, Redis, router, middleware (CORS ở đây), gọi `registerModules()`
- `modules.go` — file **duy nhất** import `internal/`, đăng ký module thủ công

### `configs/`

Quản lý configuration. Đọc từ `.env` (dùng `viper` hoặc `godotenv`).

> **Quy tắc:** KHÔNG dùng default value. Mọi biến env **bắt buộc** phải set trong `.env`. Thiếu biến → panic ngay khi start.

### `pkg/`

Shared infrastructure. Nằm **ngoài** `internal/` — project nào cũng cần.

| Package       | Vai trò                                             |
| ------------- | --------------------------------------------------- |
| `cache/`      | Redis connection & helper methods                   |
| `db/`         | PostgreSQL connection pool & health check            |
| `middleware/` | Gin middleware: auth, permission, rate limit          |
| `http/`       | Response helper (generic `[T]`) & pagination (page/limit) |
| `auth/`       | JWT token + password hashing (bcrypt)                |
| `validator/`  | Custom validation rules (vd: vn_phone)               |

### `internal/` - Feature Modules

**Chỉ chứa feature modules**. Modules độc lập, không import lẫn nhau.

| Layer       | Vai trò                                                                  |
| ----------- | ------------------------------------------------------------------------ |
| `dto/`      | Data Transfer Object — validate input từ client, format output response  |
| `service/`  | Business logic + raw SQL query trực tiếp vào DB, scan result vào DTO     |
| `handler/`  | HTTP handler — parse request, gọi service, trả response                  |
| `module.go` | Wire service → handler, inject dependencies (db, cache)                  |
| `routes.go` | Đăng ký HTTP routes cho module                                           |

### `migrations/`

SQL migration files, quản lý bằng `golang-migrate`. Order per module: schema → seed.

### `CLAUDE.md`

Rules cho Claude — cả khi review PR trên GitHub lẫn khi member dùng Claude Code viết code. Chứa toàn bộ convention: architecture, code rules, naming, "Do NOT" list. Vi phạm → block merge.

### `.claude/`

- `settings.json` — config permissions cho Claude Code
- `commands/` — custom slash commands:
  - `/review` — review code theo project rules
  - `/new-module` — scaffold module mới theo đúng cấu trúc

### `.github/`

- `PULL_REQUEST_TEMPLATE.md` — template PR với checklist, member tự check trước khi submit

### Unit Test

File test đặt cạnh file source: `auth_service.go` → `auth_service_test.go`. Dùng `go-sqlmock`.
