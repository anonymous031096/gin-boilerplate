# Project Rules

## Architecture

- Modular/Feature-based architecture using Gin framework
- `app/` and `pkg/` MUST be outside `internal/` — they are the skeleton
- `internal/` ONLY contains feature modules (iam, todo, ...)
- Each module is self-contained: `dto/`, `service/`, `handler/`, `module.go`, `routes.go`
- Register modules manually in `app/modules.go` — no auto-discovery, no init()

## Code Rules

### Service Layer
- Write raw SQL directly in service — NO repository layer, NO entity layer
- Service receives request DTO, queries DB with raw SQL, returns response DTO
- Ownership checks (e.g. user can only edit their own todo) belong in service, NOT in middleware

### Config
- Read from `.env` only — no YAML config files
- NEVER use default values (e.g. `PORT || 8080` is forbidden)
- All env variables are REQUIRED — panic on startup if any is missing

### Authorization
- Check PERMISSION, never check ROLE
- Role is just a container of permissions in the database
- Middleware: `auth()` for authentication, `permission("resource:action")` for authorization
- Routes that only need to know "who is the user" (e.g. GET /me) use `auth()` only, no permission needed

### Response Format
- Success (200): return data directly, use `Success[T]()` generic helper
- List (200): use `List[T]()` with `{data, meta}` for pagination (page/limit)
- Errors: use specific helpers — `BadRequest()`, `Unauthorized()`, `Forbidden()`, `ValidationError()`
- NO `success` boolean field in response — HTTP status code is sufficient
- Response helpers MUST use Go generics `[T any]`, not `any`/`interface{}`

### Database
- Use `jackc/pgx` driver with `database/sql` interface
- All main tables MUST have audit columns: `id (UUID, gen_random_uuid())`, `created_by (UUID)`, `updated_by (UUID)`, `created_at (TIMESTAMPTZ)`, `updated_at (TIMESTAMPTZ)`
- `created_by/updated_by` are plain UUID, no foreign key — modules must stay independent
- `updated_at` auto-set by database trigger `set_updated_at()`

### Validation
- Use Gin binding tags for standard rules (`required`, `email`, `min`, `max`)
- Custom rules registered in `pkg/validator/validator.go`

### Testing
- Unit tests use `go-sqlmock` — no real database
- Test files sit next to source: `auth_service.go` -> `auth_service_test.go`
- Test only service layer

### Middleware Order
- CORS (in app.go) -> Rate Limit (IP-based, global) -> Auth (JWT) -> Permission -> Handler

### Logging
- Use Gin default logger — no custom logger middleware

## File Naming Convention

- Module files: `<domain>_<type>.go` (e.g. `auth_request.go`, `todo_service.go`)
- Test files: `<name>_test.go` next to source file
- Migration files: `{version}_{module}_{type}.{up|down}.sql` (e.g. `000001_iam_schema.up.sql`)
- Migration order per module: schema first, seed second

## Module Structure

Every new module MUST follow this structure:
```
internal/<module>/
├── dto/
│   ├── <name>_request.go
│   └── <name>_response.go
├── service/
│   ├── <name>_service.go
│   └── <name>_service_test.go
├── handler/
│   └── <name>_handler.go
├── module.go
└── routes.go
```

After creating a module, register it in `app/modules.go`.

## Do NOT

- Do NOT create repository layer or entity layer
- Do NOT use ORM (gorm, ent, etc.) — raw SQL only
- Do NOT add default values to config
- Do NOT check role directly — check permission
- Do NOT use `any`/`interface{}` for response helpers — use generics
- Do NOT add ABAC middleware — handle ownership in service
- Do NOT add custom logger middleware — use Gin default
- Do NOT put non-module code in `internal/`
- Do NOT use auto-discovery or `init()` for module registration
- Do NOT add foreign key for `created_by`/`updated_by` — keep modules independent
