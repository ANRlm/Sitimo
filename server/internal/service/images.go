package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store"

	"github.com/disintegration/imaging"
)

func (s *Service) ListImages(ctx context.Context, params ImageListParams) (domain.Paginated[domain.ImageAsset], error) {
	return s.repo.ListImages(ctx, store.ImageListOptions{
		Keyword:        params.Keyword,
		TagIDs:         params.TagIDs,
		MIME:           params.MIME,
		Page:           params.Page,
		PageSize:       params.PageSize,
		IncludeDeleted: params.Deleted,
	})
}

func (s *Service) SaveUploadedImage(ctx context.Context, filename string, body []byte, description *string, tagIDs []string) (*domain.ImageAsset, error) {
	if len(body) > 20*1024*1024 {
		return nil, fmt.Errorf("单文件大小不能超过 20MB")
	}

	mime := http.DetectContentType(body)
	if mime == "image/heic" || mime == "image/heif" {
		converted, err := convertWithImageMagick(body)
		if err != nil {
			return nil, err
		}
		body = converted
		mime = "image/jpeg"
		if !strings.HasSuffix(strings.ToLower(filename), ".jpg") && !strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
			filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".jpg"
		}
	}

	switch mime {
	case "image/png", "image/jpeg", "image/webp":
	default:
		return nil, fmt.Errorf("不支持的图像类型：%s", mime)
	}

	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("读取图像失败：%w", err)
	}

	id := makeID()
	ext := safeImageExt(filename)
	originalRel := filepath.Join("original", id[:2], id+ext)
	thumbRel := filepath.Join("thumbnails", id[:2], id+".png")
	originalAbs := filepath.Join(s.repo.StorageRoot(), originalRel)
	thumbAbs := filepath.Join(s.repo.StorageRoot(), thumbRel)

	if err := os.MkdirAll(filepath.Dir(originalAbs), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(thumbAbs), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(originalAbs, body, 0o644); err != nil {
		return nil, err
	}
	thumb := imaging.Fit(img, 400, 400, imaging.Lanczos)
	if err := imaging.Save(thumb, thumbAbs); err != nil {
		return nil, err
	}

	return s.repo.CreateImageRecord(ctx, filename, mime, int64(len(body)), img.Bounds().Dx(), img.Bounds().Dy(), originalRel, thumbRel, description, tagIDs, nil)
}

func (s *Service) EditImage(ctx context.Context, id string, input domain.ImageEditInput, problemID *string) (*domain.ImageAsset, error) {
	info, err := s.repo.GetImageStorageInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	sourceAbs := filepath.Join(s.repo.StorageRoot(), info.StoragePath)
	img, err := imaging.Open(sourceAbs)
	if err != nil {
		return nil, err
	}

	if input.Crop != nil {
		rect := image.Rect(input.Crop.X, input.Crop.Y, input.Crop.X+input.Crop.W, input.Crop.Y+input.Crop.H)
		img = imaging.Crop(img, rect)
	}
	if input.Rotate != nil && *input.Rotate != 0 {
		img = imaging.Rotate(img, float64(*input.Rotate), image.Transparent)
	}
	if input.Resize != nil && input.Resize.W > 0 && input.Resize.H > 0 {
		img = imaging.Resize(img, input.Resize.W, input.Resize.H, imaging.Lanczos)
	}

	if problemID != nil && *problemID != "" {
		nextID := makeID()
		originalRel := filepath.Join("original", nextID[:2], nextID+".png")
		thumbRel := filepath.Join("thumbnails", nextID[:2], nextID+".png")
		originalAbs := filepath.Join(s.repo.StorageRoot(), originalRel)
		thumbAbs := filepath.Join(s.repo.StorageRoot(), thumbRel)
		if err := os.MkdirAll(filepath.Dir(originalAbs), 0o755); err != nil {
			return nil, err
		}
		if err := imaging.Save(img, originalAbs); err != nil {
			return nil, err
		}
		if err := imaging.Save(imaging.Fit(img, 400, 400, imaging.Lanczos), thumbAbs); err != nil {
			return nil, err
		}
		tagIDs := []string{}
		current, err := s.repo.GetImage(ctx, id, true)
		if err == nil {
			tagIDs = current.TagIDs
		}
		created, err := s.repo.CreateImageRecord(ctx, info.Filename, info.MIME, fileSize(originalAbs), img.Bounds().Dx(), img.Bounds().Dy(), originalRel, thumbRel, info.Description, tagIDs, &id)
		if err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceProblemImage(ctx, *problemID, id, created.ID); err != nil {
			return nil, err
		}
		return created, nil
	}

	if err := imaging.Save(img, sourceAbs); err != nil {
		return nil, err
	}
	thumbAbs := filepath.Join(s.repo.StorageRoot(), info.ThumbnailPath)
	if err := imaging.Save(imaging.Fit(img, 400, 400, imaging.Lanczos), thumbAbs); err != nil {
		return nil, err
	}
	return s.repo.OverwriteImageBinary(ctx, id, fileSize(sourceAbs), img.Bounds().Dx(), img.Bounds().Dy(), info.StoragePath, info.ThumbnailPath)
}

func convertWithImageMagick(body []byte) ([]byte, error) {
	source, err := os.CreateTemp("", "mathlib-heic-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(source.Name())
	if _, err := source.Write(body); err != nil {
		return nil, err
	}
	source.Close()

	target := source.Name() + ".jpg"
	cmdName := "magick"
	args := []string{"convert", source.Name(), target}
	if _, err := exec.LookPath("magick"); err != nil {
		cmdName = "convert"
		args = []string{source.Name(), target}
	}
	cmd := exec.Command(cmdName, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("HEIC 转换失败：%s", strings.TrimSpace(string(output)))
	}
	defer os.Remove(target)
	return os.ReadFile(target)
}
