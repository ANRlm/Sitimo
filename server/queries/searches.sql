-- name: CreateSearchHistory :one
INSERT INTO search_history (id, query, filters, result_count, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, query, filters, result_count, created_at;

-- name: ListSearchHistory :many
SELECT id, query, filters, result_count, created_at
FROM search_history
ORDER BY created_at DESC
LIMIT 20;

-- name: DeleteSearchHistory :exec
DELETE FROM search_history WHERE id = $1;

-- name: CreateSavedSearch :one
INSERT INTO saved_searches (id, name, query, filters, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, query, filters, created_at;

-- name: ListSavedSearches :many
SELECT id, name, query, filters, created_at
FROM saved_searches
ORDER BY created_at DESC;

-- name: DeleteSavedSearch :exec
DELETE FROM saved_searches WHERE id = $1;
