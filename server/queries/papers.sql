-- name: GetPaperByID :one
SELECT id, title, subtitle, school_name, exam_name, subject, duration_min,
  total_score, description, status, instructions, footer_text, header_json, layout_json,
  created_at, updated_at, deleted_at
FROM papers
WHERE id = $1;

-- name: ListPapers :many
SELECT id, title, subtitle, school_name, exam_name, subject, duration_min,
  total_score, description, status, instructions, footer_text, header_json, layout_json,
  created_at, updated_at, deleted_at
FROM papers
WHERE ($1::boolean = true OR deleted_at IS NULL)
ORDER BY updated_at DESC;

-- name: CountPapers :one
SELECT count(*) FROM papers
WHERE ($1::boolean = true OR deleted_at IS NULL);

-- name: InsertPaper :one
INSERT INTO papers (
  id, title, subtitle, school_name, exam_name, subject, duration_min,
  total_score, description, status, instructions, footer_text, header_json, layout_json,
  created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11, $12, $13, $14,
  $15, $16
)
RETURNING id, title, subtitle, school_name, exam_name, subject, duration_min,
  total_score, description, status, instructions, footer_text, header_json, layout_json,
  created_at, updated_at, deleted_at;

-- name: UpdatePaper :one
UPDATE papers SET
  title = $2,
  subtitle = $3,
  school_name = $4,
  exam_name = $5,
  subject = $6,
  duration_min = $7,
  total_score = $8,
  description = $9,
  status = $10,
  instructions = $11,
  footer_text = $12,
  header_json = $13,
  layout_json = $14,
  updated_at = $15
WHERE id = $1
RETURNING id, title, subtitle, school_name, exam_name, subject, duration_min,
  total_score, description, status, instructions, footer_text, header_json, layout_json,
  created_at, updated_at, deleted_at;

-- name: SoftDeletePaper :exec
UPDATE papers SET deleted_at = now() WHERE id = $1;

-- name: RestorePaper :exec
UPDATE papers SET deleted_at = NULL WHERE id = $1;

-- name: HardDeletePaper :exec
DELETE FROM papers WHERE id = $1;

-- name: GetPaperItems :many
SELECT id, paper_id, problem_id, order_index, score, image_position, blank_lines
FROM paper_items
WHERE paper_id = $1
ORDER BY order_index;

-- name: DeletePaperItems :exec
DELETE FROM paper_items WHERE paper_id = $1;

-- name: InsertPaperItem :exec
INSERT INTO paper_items (id, paper_id, problem_id, order_index, score, image_position, blank_lines)
VALUES ($1, $2, $3, $4, $5, $6, $7);
