package service

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

type OrphanSweepResult struct {
	Deleted    int      `json:"deleted"`
	BytesFreed int64    `json:"bytesFreed"`
	ImageIDs   []string `json:"imageIds"`
}

func (s *Service) SweepOrphanImages(ctx context.Context) (OrphanSweepResult, error) {
	candidates, err := s.repo.ListOrphanImagesForSweep(ctx, time.Now().Add(-7*24*time.Hour))
	if err != nil {
		return OrphanSweepResult{}, err
	}

	result := OrphanSweepResult{
		ImageIDs: make([]string, 0, len(candidates)),
	}
	for _, candidate := range candidates {
		for _, relativePath := range []string{candidate.StoragePath, candidate.ThumbnailPath} {
			if relativePath == "" {
				continue
			}
			absolutePath := filepath.Join(s.repo.StorageRoot(), relativePath)
			if info, err := os.Stat(absolutePath); err == nil {
				result.BytesFreed += info.Size()
			}
			if err := os.Remove(absolutePath); err != nil && !os.IsNotExist(err) {
				return OrphanSweepResult{}, err
			}
		}

		if err := s.repo.HardDeleteImage(ctx, candidate.ID); err != nil {
			return OrphanSweepResult{}, err
		}
		result.Deleted++
		result.ImageIDs = append(result.ImageIDs, candidate.ID)
	}

	return result, nil
}
