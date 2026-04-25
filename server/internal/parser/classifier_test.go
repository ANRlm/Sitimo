package parser

import (
	"testing"

	"mathlib/server/internal/domain"
)

func TestInferTypeMultipleChoice(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "What is 2+2?"},
		{Type: BlockEnvBegin, EnvName: "tasks", EnvArgs: "(4)"},
	}
	result, needsReview := InferType(blocks)
	if result != domain.ProblemTypeMultipleChoice || needsReview {
		t.Errorf("Expected multiple_choice, got %v (needsReview=%v)", result, needsReview)
	}
}

func TestInferTypeFillBlank(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "\\underline{\\hspace{2cm}}"},
	}
	result, needsReview := InferType(blocks)
	if result != domain.ProblemTypeFillBlank || needsReview {
		t.Errorf("Expected fill_blank, got %v", result)
	}
}

func TestInferTypeFillBlankWithFillin(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "\\fillin"},
	}
	result, needsReview := InferType(blocks)
	if result != domain.ProblemTypeFillBlank || needsReview {
		t.Errorf("Expected fill_blank for \\fillin, got %v", result)
	}
}

func TestInferTypeMixed(t *testing.T) {
	blocks := []Block{
		{Type: BlockEnvBegin, EnvName: "tasks"},
		{Type: BlockText, Content: "\\underline{\\hspace{2cm}}"},
	}
	result, needsReview := InferType(blocks)
	if result != domain.ProblemTypeOther || !needsReview {
		t.Errorf("Expected other with needsReview=true, got %v (needsReview=%v)", result, needsReview)
	}
}

func TestInferTypeSolve(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "Solve for x: 2x + 3 = 7"},
	}
	result, _ := InferType(blocks)
	if result != domain.ProblemTypeSolve {
		t.Errorf("Expected solve, got %v", result)
	}
}

func TestInferTypeProof(t *testing.T) {
	blocks := []Block{
		{Type: BlockEnvBegin, EnvName: "proof"},
	}
	result, needsReview := InferType(blocks)
	if result != domain.ProblemTypeProof || needsReview {
		t.Errorf("Expected proof, got %v (needsReview=%v)", result, needsReview)
	}
}

func TestInferTypeProofText(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "\\textbf{证明}"},
	}
	result, needsReview := InferType(blocks)
	if result != domain.ProblemTypeProof || needsReview {
		t.Errorf("Expected proof for \\textbf{证明}, got %v (needsReview=%v)", result, needsReview)
	}
}

func TestHasTasksEnv(t *testing.T) {
	blocks := []Block{
		{Type: BlockEnvBegin, EnvName: "tasks"},
	}
	if !HasTasksEnv(blocks) {
		t.Error("Expected HasTasksEnv to be true")
	}
}

func TestHasTasksEnvFalse(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "hello"},
	}
	if HasTasksEnv(blocks) {
		t.Error("Expected HasTasksEnv to be false")
	}
}

func TestHasUnderline(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "\\underline{hello}"},
	}
	if !HasUnderline(blocks) {
		t.Error("Expected HasUnderline to be true")
	}
}

func TestHasUnderlineWithFillin(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "\\fillin"},
	}
	if !HasUnderline(blocks) {
		t.Error("Expected HasUnderline to be true for fillin")
	}
}

func TestHasUnderlineFalse(t *testing.T) {
	blocks := []Block{
		{Type: BlockText, Content: "plain text"},
	}
	if HasUnderline(blocks) {
		t.Error("Expected HasUnderline to be false")
	}
}
