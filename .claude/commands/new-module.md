Create a new feature module following the project structure.

Arguments: $ARGUMENTS (module name, e.g. "product")

Steps:
1. Create directory structure:
   - internal/<module>/dto/
   - internal/<module>/service/
   - internal/<module>/handler/
   - internal/<module>/module.go
   - internal/<module>/routes.go

2. Create DTO files with request/response structs (use binding tags for validation)

3. Create service with raw SQL (no repository, no entity)
   - Include audit columns (created_by, updated_by from context)
   - Include ownership check where applicable

4. Create handler using pkg/http response helpers (Success[T], List[T], etc.)

5. Create module.go to wire service -> handler

6. Create routes.go with RegisterRoutes function
   - Use middleware.Auth for authenticated routes
   - Use middleware.Permission for permission-protected routes

7. Register module in app/modules.go

8. Create migrations:
   - Schema migration with audit columns (id UUID, created_by, updated_by, created_at, updated_at)
   - Seed migration with permissions + assign to roles
   - Add updated_at trigger

9. Create unit test for service using go-sqlmock
