Review the current changes against project rules defined in CLAUDE.md.

## Architecture & Convention
1. Architecture violations (repository layer, entity layer, ORM usage)
2. Config has no default values
3. Permission check (not role check) in middleware and routes
4. Response uses generic helpers (Success[T], List[T], BadRequest, etc.)
5. Audit columns present in new tables (id UUID, created_by, updated_by, created_at, updated_at)
6. No foreign key on created_by/updated_by
7. Raw SQL in service layer, not in handler
8. Unit tests exist for service layer using go-sqlmock
9. Module registered in app/modules.go
10. Migration naming follows convention: {version}_{module}_{type}.{up|down}.sql
11. Service receives context.Context + explicit params, NEVER *gin.Context
12. JSON tags use camelCase, not snake_case

## Performance
13. N+1 query: check for queries inside loops — use JOIN or batch query instead
14. Missing database index on columns used in WHERE, JOIN, ORDER BY
15. Unbounded query: SELECT without LIMIT (list endpoints must use pagination)
16. Singleflight: frequently called endpoints with same params should use singleflight
17. Connection pool: check if DB/Redis connections are properly reused, not created per request
18. Missing context timeout on external calls (DB, Redis, HTTP)

## Memory Leak
19. Unclosed rows: `db.Query()` must have `defer rows.Close()`
20. Unclosed response body: `http.Get()` must have `defer resp.Body.Close()`
21. Goroutine leak: goroutines must have exit condition, context cancellation, or timeout
22. Accumulating data in long-lived structs (maps, slices that grow without bound)
23. Deferred function in loop: `defer` inside loop delays cleanup until function returns

## Security
24. SQL injection: all user input must use parameterized queries ($1, $2), never string concatenation
25. Password stored as bcrypt hash, never plain text
26. JWT secret not hardcoded, read from env
27. Sensitive data (password, token) not logged or returned in response
28. CORS headers properly configured
29. Rate limit enabled on public endpoints (login, register)

## Code Smell
30. Error swallowed: `err` returned but not checked or logged
31. Dead code: unused functions, variables, imports
32. Magic numbers/strings: hardcoded values that should be constants or config
33. God function: function longer than 50 lines — consider splitting
34. Duplicate logic: same code repeated in multiple places — extract helper
35. Handler doing business logic: handler should only parse request, call service, return response
36. Returning generic error to client: internal errors (DB, Redis) should not leak to response

## Transaction Safety
37. Multiple write queries without transaction (INSERT/UPDATE/DELETE that must be atomic)
38. Missing `defer tx.Rollback()` after `BeginTx`
39. Transaction held too long: avoid external calls (HTTP, Redis) inside DB transaction

Report violations as a checklist with file path and line number.
Block if any MUST/MUST NOT rule or Security issue is violated.
Warn for Performance, Memory Leak, Code Smell, and Transaction Safety issues.
