CREATE TABLE todos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    completed   BOOLEAN NOT NULL DEFAULT FALSE,
    created_by  UUID,
    updated_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER todos_updated_at
    BEFORE UPDATE ON todos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_todos_created_by ON todos(created_by);
