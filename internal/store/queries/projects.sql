-- name: InsertProject :one
INSERT INTO projects (id, name, created_at)
VALUES ($1, $2, $3)
RETURNING id, name, created_at;

-- name: GetProject :one
SELECT id, name, created_at
FROM projects
WHERE id = $1;

-- name: ListProjects :many
SELECT id, name, created_at
FROM projects
ORDER BY created_at DESC, id DESC;
