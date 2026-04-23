-- name: GetImageByID :one
SELECT id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path,
  description, parent_image_id, created_at, updated_at, deleted_at
FROM images
WHERE id = $1;

-- name: InsertImage :one
INSERT INTO images (
  id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path,
  description, parent_image_id, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8,
  $9, $10, $11, $12
)
RETURNING id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path,
  description, parent_image_id, created_at, updated_at, deleted_at;

-- name: UpdateImage :one
UPDATE images SET
  filename = $2,
  description = $3,
  updated_at = $4
WHERE id = $1
RETURNING id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path,
  description, parent_image_id, created_at, updated_at, deleted_at;

-- name: SoftDeleteImage :exec
UPDATE images SET deleted_at = now() WHERE id = $1;

-- name: RestoreImage :exec
UPDATE images SET deleted_at = NULL WHERE id = $1;

-- name: HardDeleteImage :exec
DELETE FROM images WHERE id = $1;

-- name: ListImages :many
SELECT id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path,
  description, parent_image_id, created_at, updated_at, deleted_at
FROM images
WHERE ($1::boolean = true OR deleted_at IS NULL)
ORDER BY created_at DESC;

-- name: CountImages :one
SELECT count(*) FROM images
WHERE ($1::boolean = true OR deleted_at IS NULL);

-- name: GetImageTagIDs :many
SELECT tag_id FROM image_tags WHERE image_id = $1;

-- name: InsertImageTag :exec
INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteImageTags :exec
DELETE FROM image_tags WHERE image_id = $1;

-- name: GetImageLinkedProblemIDs :many
SELECT problem_id FROM problem_images WHERE image_id = $1;
