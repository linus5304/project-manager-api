-- name: InsertTask :one
INSERT INTO tasks (id, project_id, title, description, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, title, description, status, created_at;

-- name: ListTasks :many
SELECT id, project_id, title, description, status, created_at
FROM tasks
WHERE project_id = $1
ORDER BY created_at DESC, id DESC;

-- name: UpdateTask :one
UPDATE tasks
SET
  title = COALESCE(sqlc.narg('title'), title),
  description = COALESCE(sqlc.narg('description'), description),
  status = COALESCE(sqlc.narg('status'), status)
WHERE project_id = $1 AND id = $2
RETURNING id, project_id, title, description, status, created_at;
