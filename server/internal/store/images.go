package store

import (
	"context"
	"time"

	"mathlib/server/internal/domain"
)

type StoredImageInfo struct {
	ID            string
	Filename      string
	MIME          string
	StoragePath   string
	ThumbnailPath string
	Width         int
	Height        int
	Size          int64
	Description   *string
}

func (r *Repository) CreateImageRecord(
	ctx context.Context,
	filename string,
	mime string,
	size int64,
	width int,
	height int,
	storagePath string,
	thumbnailPath string,
	description *string,
	tagIDs []string,
	parentImageID *string,
) (*domain.ImageAsset, error) {
	id := newID()
	_, err := r.db.Exec(ctx, `INSERT INTO images
		(id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path, description, parent_image_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),now())`,
		id, filename, mime, size, width, height, storagePath, thumbnailPath, description, parentImageID,
	)
	if err != nil {
		return nil, err
	}
	if err := r.replaceImageTags(ctx, id, tagIDs); err != nil {
		return nil, err
	}
	return r.GetImage(ctx, id, true)
}

func (r *Repository) UpdateImageMetadata(ctx context.Context, id string, tagIDs []string, description *string) (*domain.ImageAsset, error) {
	_, err := r.db.Exec(ctx, `UPDATE images SET description = $2, updated_at = now() WHERE id = $1`, id, description)
	if err != nil {
		return nil, err
	}
	if err := r.replaceImageTags(ctx, id, tagIDs); err != nil {
		return nil, err
	}
	return r.GetImage(ctx, id, true)
}

func (r *Repository) OverwriteImageBinary(ctx context.Context, id string, size int64, width, height int, storagePath, thumbnailPath string) (*domain.ImageAsset, error) {
	_, err := r.db.Exec(ctx, `UPDATE images SET size_bytes = $2, width = $3, height = $4, storage_path = $5, thumbnail_path = $6, updated_at = now() WHERE id = $1`,
		id, size, width, height, storagePath, thumbnailPath,
	)
	if err != nil {
		return nil, err
	}
	return r.GetImage(ctx, id, true)
}

func (r *Repository) replaceImageTags(ctx context.Context, imageID string, tagIDs []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM image_tags WHERE image_id = $1`, imageID); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2)`, imageID, tagID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) ReplaceProblemImage(ctx context.Context, problemID, oldImageID, newImageID string) error {
	_, err := r.db.Exec(ctx, `UPDATE problem_images SET image_id = $3 WHERE problem_id = $1 AND image_id = $2`, problemID, oldImageID, newImageID)
	return err
}

func (r *Repository) SoftDeleteImage(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE images SET deleted_at = now(), updated_at = now() WHERE id = $1`, id)
	return err
}

func (r *Repository) RestoreImage(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE images SET deleted_at = NULL, updated_at = now() WHERE id = $1`, id)
	return err
}

func (r *Repository) HardDeleteImage(ctx context.Context, id string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM problem_images WHERE image_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM images WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetImageStorageInfo(ctx context.Context, id string) (*StoredImageInfo, error) {
	var info StoredImageInfo
	var description *string
	err := r.db.QueryRow(ctx, `SELECT id, filename, mime, storage_path, thumbnail_path, width, height, size_bytes, description FROM images WHERE id = $1`, id).
		Scan(&info.ID, &info.Filename, &info.MIME, &info.StoragePath, &info.ThumbnailPath, &info.Width, &info.Height, &info.Size, &description)
	if err != nil {
		return nil, err
	}
	info.Description = description
	return &info, nil
}

func (r *Repository) BatchDeleteImages(ctx context.Context, imageIDs []string) (int, error) {
	if len(imageIDs) == 0 {
		return 0, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		UPDATE images
		SET deleted_at = now(), updated_at = now()
		WHERE id = ANY($1) AND deleted_at IS NULL
	`, imageIDs)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	affected := result.RowsAffected()
	return int(affected), nil
}

func (r *Repository) ListOrphanImagesForSweep(ctx context.Context, deletedBefore time.Time) ([]StoredImageInfo, error) {
	rows, err := r.db.Query(ctx, `SELECT id, filename, mime, storage_path, thumbnail_path, width, height, size_bytes, description
FROM images i
WHERE i.deleted_at IS NOT NULL
  AND i.deleted_at <= $1
  AND NOT EXISTS (
    SELECT 1
    FROM problem_images pi
    WHERE pi.image_id = i.id
  )`, deletedBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StoredImageInfo, 0)
	for rows.Next() {
		var (
			item        StoredImageInfo
			description *string
		)
		if err := rows.Scan(&item.ID, &item.Filename, &item.MIME, &item.StoragePath, &item.ThumbnailPath, &item.Width, &item.Height, &item.Size, &description); err != nil {
			return nil, err
		}
		item.Description = description
		items = append(items, item)
	}
	return items, rows.Err()
}
