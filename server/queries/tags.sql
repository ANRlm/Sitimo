-- name: GetTagByID :one
SELECT id, name, category, color, description, created_at, updated_at
FROM tags
WHERE id = $1;

-- name: ListTags :many
SELECT id, name, category, color, description, created_at, updated_at
FROM tags
ORDER BY name ASC;

-- name: InsertTag :one
INSERT INTO tags (id, name, category, color, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, name, category, color, description, created_at, updated_at;

-- name: UpdateTag :one
UPDATE tags SET
  name = $2,
  category = $3,
  color = $4,
  description = $5,
  updated_at = $6
WHERE id = $1
RETURNING id, name, category, color, description, created_at, updated_at;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: CountTagProblems :one
SELECT count(*) FROM problem_tags WHERE tag_id = $1;

-- name: UpdateProblemTagsForMerge :exec
UPDATE problem_tags pt SET tag_id = $1
WHERE pt.tag_id = $2
  AND pt.problem_id NOT IN (
    SELECT pt2.problem_id FROM problem_tags pt2 WHERE pt2.tag_id = $1
  );

-- name: UpdateImageTagsForMerge :exec
UPDATE image_tags it SET tag_id = $1
WHERE it.tag_id = $2
  AND it.image_id NOT IN (
    SELECT it2.image_id FROM image_tags it2 WHERE it2.tag_id = $1
  );
