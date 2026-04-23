package service

import (
	"context"
	"slices"
	"strings"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store"
)

func (s *Service) Search(ctx context.Context, keyword, formula string, conditions []domain.SearchCondition) ([]domain.SearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	formula = strings.TrimSpace(formula)

	results, err := s.repo.SearchProblems(ctx, store.SearchOptions{
		Keyword:    keyword,
		Formula:    formula,
		Conditions: conditions,
	})
	if err != nil {
		return nil, err
	}

	if keyword != "" || formula != "" || len(conditions) > 0 {
		_ = s.repo.CreateSearchHistory(ctx, keyword, map[string]any{"formula": formula, "conditions": conditions}, len(results))
	}
	return results, nil
}

func matchesConditions(problem domain.Problem, conditions []domain.SearchCondition) bool {
	for _, condition := range conditions {
		switch condition.Field {
		case "subject":
			if deref(problem.Subject) != condition.Value {
				return false
			}
		case "grade":
			if deref(problem.Grade) != condition.Value {
				return false
			}
		case "difficulty":
			if string(problem.Difficulty) != condition.Value {
				return false
			}
		case "type":
			if string(problem.Type) != condition.Value {
				return false
			}
		case "source":
			if !strings.Contains(strings.ToLower(deref(problem.Source)), strings.ToLower(condition.Value)) {
				return false
			}
		case "hasImage":
			if condition.Value == "yes" && len(problem.ImageIDs) == 0 {
				return false
			}
			if condition.Value == "no" && len(problem.ImageIDs) > 0 {
				return false
			}
		case "tag":
			if !slices.Contains(problem.TagIDs, condition.Value) {
				return false
			}
		case "subjectiveScore":
			if problem.SubjectiveScore == nil {
				return false
			}
			if !matchFloat(*problem.SubjectiveScore, condition) {
				return false
			}
		case "date":
			if !matchTime(problem.CreatedAt, condition) {
				return false
			}
		}
	}
	return true
}

func matchFloat(value float64, condition domain.SearchCondition) bool {
	target := parseFloat(condition.Value)
	switch condition.Operator {
	case "gt":
		return value > target
	case "lt":
		return value < target
	case "between":
		return value >= target && value <= parseFloat(deref(condition.SecondValue))
	default:
		return value == target
	}
}

func matchTime(value time.Time, condition domain.SearchCondition) bool {
	target, err := time.Parse("2006-01-02", condition.Value)
	if err != nil {
		return true
	}
	switch condition.Operator {
	case "gt":
		return value.After(target)
	case "lt":
		return value.Before(target)
	case "between":
		second, err := time.Parse("2006-01-02", deref(condition.SecondValue))
		if err != nil {
			return true
		}
		return !value.Before(target) && !value.After(second)
	default:
		return value.Format("2006-01-02") == target.Format("2006-01-02")
	}
}
