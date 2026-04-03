Review the current changes against project rules defined in CLAUDE.md.

Check for:
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

Report violations as a checklist. Block if any MUST/MUST NOT rule is violated.
