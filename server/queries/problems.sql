-- name: GetProblemByID :one
SELECT id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, search_tsv::text AS search_tsv, formula_tokens, formula_tsv::text AS formula_tsv,
  version, created_at, updated_at, deleted_at
FROM problems
WHERE id = $1;

-- name: GetProblemsByIDs :many
SELECT id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, search_tsv::text AS search_tsv, formula_tokens, formula_tsv::text AS formula_tsv,
  version, created_at, updated_at, deleted_at
FROM problems
WHERE id = ANY($1::text[])
ORDER BY updated_at DESC;

-- name: InsertProblem :one
INSERT INTO problems (
  id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, formula_tokens, version,
  created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11, $12, $13, $14,
  $15, $16
)
RETURNING id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, search_tsv::text AS search_tsv, formula_tokens, formula_tsv::text AS formula_tsv,
  version, created_at, updated_at, deleted_at;

-- name: UpdateProblem :one
UPDATE problems SET
  latex = $2,
  answer_latex = $3,
  solution_latex = $4,
  problem_type = $5,
  difficulty = $6,
  subjective_score = $7,
  subject = $8,
  grade = $9,
  source = $10,
  notes = $11,
  formula_tokens = $12,
  version = $13,
  updated_at = $14
WHERE id = $1
RETURNING id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, search_tsv::text AS search_tsv, formula_tokens, formula_tsv::text AS formula_tsv,
  version, created_at, updated_at, deleted_at;

-- name: SoftDeleteProblem :exec
UPDATE problems SET deleted_at = now() WHERE id = $1;

-- name: RestoreProblem :exec
UPDATE problems SET deleted_at = NULL WHERE id = $1;

-- name: HardDeleteProblem :exec
DELETE FROM problems WHERE id = $1;

-- name: ListProblems :many
SELECT id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, search_tsv::text AS search_tsv, formula_tokens, formula_tsv::text AS formula_tsv,
  version, created_at, updated_at, deleted_at
FROM problems
WHERE ($1::boolean = true OR deleted_at IS NULL)
ORDER BY updated_at DESC;

-- name: CountProblems :one
SELECT count(*) FROM problems
WHERE ($1::boolean = true OR deleted_at IS NULL);

-- name: IncrementProblemVersion :exec
UPDATE problems SET version = version + 1, updated_at = now() WHERE id = $1;

-- name: GetProblemTagIDs :many
SELECT tag_id FROM problem_tags WHERE problem_id = $1;

-- name: GetProblemImageIDs :many
SELECT image_id FROM problem_images WHERE problem_id = $1 ORDER BY order_index;

-- name: InsertProblemTag :exec
INSERT INTO problem_tags (problem_id, tag_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteProblemTags :exec
DELETE FROM problem_tags WHERE problem_id = $1;

-- name: InsertProblemImage :exec
INSERT INTO problem_images (problem_id, image_id, order_index) VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: DeleteProblemImages :exec
DELETE FROM problem_images WHERE problem_id = $1;

-- name: InsertProblemVersion :exec
INSERT INTO problem_versions (id, problem_id, version, snapshot, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetProblemVersions :many
SELECT id, problem_id, version, snapshot, created_at
FROM problem_versions
WHERE problem_id = $1
ORDER BY version DESC;

-- name: GetProblemVersion :one
SELECT id, problem_id, version, snapshot, created_at
FROM problem_versions
WHERE problem_id = $1 AND version = $2;

-- name: IncrementCodeCounter :one
INSERT INTO problem_code_counters (year, current_serial, updated_at)
VALUES ($1, 1, now())
ON CONFLICT (year) DO UPDATE
  SET current_serial = problem_code_counters.current_serial + 1,
      updated_at = now()
RETURNING current_serial;
