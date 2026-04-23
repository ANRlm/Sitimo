package service

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"mathlib/server/internal/domain"

	"github.com/oklog/ulid/v2"
)

func paginate[T any](items []T, page, pageSize int) domain.Paginated[T] {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 24
	}
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return domain.Paginated[T]{
		Items:    items[start:end],
		Total:    len(items),
		Page:     page,
		PageSize: pageSize,
	}
}

func sortProblems(items []domain.Problem, sortBy, sortOrder string) {
	switch sortBy {
	case "code":
		slices.SortFunc(items, func(a, b domain.Problem) int {
			return strings.Compare(a.Code, b.Code)
		})
	case "created_at":
		slices.SortFunc(items, func(a, b domain.Problem) int {
			if a.CreatedAt.Equal(b.CreatedAt) {
				return 0
			}
			if a.CreatedAt.After(b.CreatedAt) {
				return -1
			}
			return 1
		})
	default:
		slices.SortFunc(items, func(a, b domain.Problem) int {
			if a.UpdatedAt.Equal(b.UpdatedAt) {
				return 0
			}
			if a.UpdatedAt.After(b.UpdatedAt) {
				return -1
			}
			return 1
		})
	}
	if strings.ToLower(sortOrder) == "asc" {
		slices.Reverse(items)
	}
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func parseFloat(raw string) float64 {
	value, _ := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	return value
}

func mapStringPtr(input map[string]any, key string) *string {
	raw, ok := input[key]
	if !ok {
		return nil
	}
	value := strings.TrimSpace(fmt.Sprintf("%v", raw))
	if value == "" || value == "<nil>" {
		return nil
	}
	return &value
}

func parseDefaultDifficulty(input map[string]any) domain.Difficulty {
	if raw, ok := input["difficulty"]; ok {
		value := fmt.Sprintf("%v", raw)
		switch value {
		case "easy", "medium", "hard", "olympiad":
			return domain.Difficulty(value)
		}
	}
	return domain.DifficultyMedium
}

func parseTagNames(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			text := strings.TrimSpace(fmt.Sprintf("%v", item))
			if text != "" {
				out = append(out, text)
			}
		}
		return out
	case string:
		if value == "" {
			return nil
		}
		parts := strings.Split(value, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
		return out
	default:
		return nil
	}
}

func newImportID(index int) string {
	return fmt.Sprintf("draft-%d", index)
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func safeImageExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		return ext
	default:
		return ".png"
	}
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func makeID() string {
	return ulid.Make().String()
}
