CREATE TABLE
    IF NOT EXISTS projects (
        id UUID PRIMARY KEY,
        name TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        CONSTRAINT projects_name_nonempty CHECK (length (btrim (name)) > 0)
    );

CREATE TABLE
    IF NOT EXISTS tasks (
        id UUID PRIMARY KEY,
        project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
        title TEXT NOT NULL,
        description TEXT NOT NULL DEFAULT '',
        status TEXT NOT NULL DEFAULT 'todo',
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        CONSTRAINT tasks_title_nonempty CHECK (length (btrim (title)) > 0),
        CONSTRAINT tasks_status_valid CHECK (status IN ('todo', 'doing', 'done'))
    );

-- Deterministic newest-first per project
CREATE INDEX IF NOT EXISTS tasks_project_newest_idx ON tasks (project_id, created_at DESC, id DESC);

-- Helpful for listing projects newest-first
CREATE INDEX IF NOT EXISTS projects_newest_idx ON projects (created_at DESC, id DESC);