# Gin Enterprise Boilerplate

Enterprise-ready Golang boilerplate with Gin, PostgreSQL, Redis, Docker, package-by-feature architecture, IAM (auth/user/role/permission), and Product CRUD.

## Stack
- Go + Gin
- PostgreSQL (pgx)
- Redis (go-redis)
- Docker Compose

## Quick Start
1. Start dependencies:
   ```bash
   docker compose up -d
   ```
2. Configure environment:
   - Copy/edit `.env` values.
3. Run SQL migrations:
   - Apply files in `migrations/` using `golang-migrate` or your preferred migration runner.
4. Run API:
   ```bash
   make run
   ```
   For local development with live reload, run:
   ```bash
   make dev
   ```
   This uses [Air](https://github.com/air-verse/air) if it is on your `PATH` or in `$(go env GOPATH)/bin`, otherwise `go run` (no install). Optional: `go install github.com/air-verse/air@latest` for a faster startup. Air rebuilds and restarts when you change `.go` files (see `.air.toml`).

## Implemented APIs
- Auth: `POST /auth/signup`, `POST /auth/signin`, `POST /auth/refresh`
- Users: `GET /users/me`, `GET /users`, `GET /users/:id`, `PUT /users/:id`
- Roles: `POST /roles`, `GET /roles`, `PUT /roles/:id`, `DELETE /roles/:id`
- Permissions: `GET /permissions`, `GET /permissions/:id`
- Products CRUD: `POST /products`, `GET /products`, `GET /products/:id`, `PUT /products/:id`, `DELETE /products/:id`

## IAM Model
- Many-to-many: `users <-> roles`
- Many-to-many: `roles <-> permissions`
- User direct overrides via `user_permissions`
- Effective permission = role permissions UNION user direct permissions

## Useful Commands
- `make run`
- `make dev` — Air live reload (uses `PATH`, `GOPATH/bin/air`, or `go run`; install Air for faster cold starts)
- `make test`
- `make swagger`

## Swagger
- Generate docs: `make swagger` (Swagger/OpenAPI **2.0** via `swag` v1)
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Raw Swagger JSON: `http://localhost:8080/swagger/doc.json`
