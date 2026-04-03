## What
<!-- What does this PR do? -->

## Why
<!-- Why is this change needed? -->

## Module affected
<!-- e.g. iam, todo, pkg/auth, etc. -->

## Checklist
- [ ] Raw SQL in service (no repository, no entity, no ORM)
- [ ] Config has no default values
- [ ] Permission check (not role check)
- [ ] Response uses generic helpers (Success[T], List[T], BadRequest, etc.)
- [ ] Audit columns on new tables (id, created_by, updated_by, created_at, updated_at)
- [ ] No FK on created_by/updated_by
- [ ] Unit test with go-sqlmock for service layer
- [ ] Migration follows naming: `{version}_{module}_{type}.{up|down}.sql`
- [ ] New module registered in `app/modules.go`
