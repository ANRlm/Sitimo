-- name: GetExportJobByID :one
SELECT id, paper_id, paper_title, format, variant, status, progress,
  download_path, error_message, created_at, started_at, completed_at, cancel_requested_at
FROM export_jobs
WHERE id = $1;

-- name: ListExportJobs :many
SELECT id, paper_id, paper_title, format, variant, status, progress,
  download_path, error_message, created_at, started_at, completed_at, cancel_requested_at
FROM export_jobs
ORDER BY created_at DESC;

-- name: InsertExportJob :one
INSERT INTO export_jobs (
  id, paper_id, paper_title, format, variant, status, progress, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, paper_id, paper_title, format, variant, status, progress,
  download_path, error_message, created_at, started_at, completed_at, cancel_requested_at;

-- name: UpdateExportJobStatus :exec
UPDATE export_jobs SET
  status = $2,
  progress = $3,
  started_at = $4,
  completed_at = $5,
  error_message = $6,
  download_path = $7,
  cancel_requested_at = $8
WHERE id = $1;

-- name: DeleteExportJob :exec
DELETE FROM export_jobs WHERE id = $1;

-- name: GetExportJobsByPaperID :many
SELECT id, paper_id, paper_title, format, variant, status, progress,
  download_path, error_message, created_at, started_at, completed_at, cancel_requested_at
FROM export_jobs
WHERE paper_id = $1
ORDER BY created_at DESC;

-- name: RequestExportCancellation :exec
UPDATE export_jobs
SET cancel_requested_at = now()
WHERE id = $1
  AND status = 'processing';
