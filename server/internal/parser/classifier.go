package parser

import (
	"strings"

	"mathlib/server/internal/domain"
)

// InferType determines the problem type based on scanned blocks.
// It looks for \begin{tasks} → multiple_choice, \underline → fill_blank, etc.
// If both markers are present, returns domain.ProblemTypeOther with needsReview suggestion.
func InferType(blocks []Block) (domain.ProblemType, bool) {
	hasTasks := HasTasksEnv(blocks)
	hasUnderline := HasUnderline(blocks)
	hasProof := hasProofEnv(blocks)

	if hasTasks && hasUnderline {
		return domain.ProblemTypeOther, true
	}
	if hasTasks {
		return domain.ProblemTypeMultipleChoice, false
	}
	if hasUnderline {
		return domain.ProblemTypeFillBlank, false
	}
	if hasProof {
		return domain.ProblemTypeProof, false
	}
	return domain.ProblemTypeSolve, false
}

// HasTasksEnv checks if blocks contain \begin{tasks}
func HasTasksEnv(blocks []Block) bool {
	for _, b := range blocks {
		if b.Type == BlockEnvBegin && b.EnvName == "tasks" {
			return true
		}
	}
	return false
}

// HasUnderline checks if blocks contain \underline{...} or \fillin pattern.
func HasUnderline(blocks []Block) bool {
	for _, b := range blocks {
		if strings.Contains(b.Content, "\\underline") || strings.Contains(b.Content, "\\fillin") {
			return true
		}
	}
	return false
}

// hasProofEnv checks if blocks contain \begin{proof} or \textbf{证明}.
func hasProofEnv(blocks []Block) bool {
	for _, b := range blocks {
		if b.Type == BlockEnvBegin && b.EnvName == "proof" {
			return true
		}
		if strings.Contains(b.Content, "\\textbf{证明}") {
			return true
		}
	}
	return false
}
