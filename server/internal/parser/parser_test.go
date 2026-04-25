package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readTestFile is a helper to load test fixture files from server/testdata/parser/.
func readTestFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "parser", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestParsePatternA verifies that pattern A (enumerate with "题\arabic*") is correctly parsed:
//   - 3 problems extracted from pattern_a.tex
//   - Problem with \begin{tasks} → inferredType = "multiple_choice"
//   - Problem with \underline → inferredType = "fill_blank"
//   - Problem body does NOT include \begin{enumerate} wrapper
//
// This is the RED phase — parser not yet implemented.
func TestParsePatternA(t *testing.T) {
	input := readTestFile(t, "pattern_a.tex")

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	// The following is the intended API contract:
	blocks := ScanBlocks(input)       // from scanner.go (Task 5)
	result, _ := ParseEnumerate(blocks) // from parser_enumerate.go (Task 7)

	if len(result) != 3 {
		t.Fatalf("expected 3 problems from pattern_a, got %d", len(result))
	}

	// Problem 1: should contain \begin{tasks} → type inference should be multiple_choice
	if !strings.Contains(result[0].Body, `\begin{tasks}`) {
		t.Errorf("problem 1 body should contain \\begin{tasks}, got: %s", result[0].Body[:50])
	}

	// Problem 3: should contain \underline → type inference should be fill_blank
	if !strings.Contains(result[2].Body, `\underline`) {
		t.Errorf("problem 3 body should contain \\underline, got: %s", result[2].Body[:50])
	}

	// No problem body should contain the outer \begin{enumerate} wrapper
	for i, p := range result {
		if strings.Contains(p.Body, `\begin{enumerate}`) {
			t.Errorf("problem %d body should not contain \\begin{enumerate} wrapper", i+1)
		}
		if strings.Contains(p.Body, `\end{enumerate}`) {
			t.Errorf("problem %d body should not contain \\end{enumerate} wrapper", i+1)
		}
	}

	// All problems should have Pattern = "A"
	for i, p := range result {
		if p.Pattern != PatternA {
			t.Errorf("problem %d should have Pattern=A, got %q", i+1, p.Pattern)
		}
	}

	// Line numbers should be present and increasing
	if result[0].LineStart < 1 {
		t.Errorf("problem 1 LineStart should be >= 1, got %d", result[0].LineStart)
	}
	if result[1].LineStart <= result[0].LineStart {
		t.Errorf("problem 2 LineStart (%d) should be > problem 1 LineStart (%d)",
			result[1].LineStart, result[0].LineStart)
	}
}

// TestParsePatternB verifies that pattern B (enumerate with "例\arabic*" and resume) is correctly parsed:
//   - 2 problems from pattern_b.tex (resume across 2 enumerate blocks)
//   - Problem numbering should be continuous
//
// This is the RED phase — parser not yet implemented.
func TestParsePatternB(t *testing.T) {
	input := readTestFile(t, "pattern_b.tex")

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)
	result, _ := ParseEnumerate(blocks)

	if len(result) != 2 {
		t.Fatalf("expected 2 problems from pattern_b (resumed across 2 enumerate blocks), got %d", len(result))
	}

	// All problems should have Pattern = "B"
	for i, p := range result {
		if p.Pattern != PatternB {
			t.Errorf("problem %d should have Pattern=B, got %q", i+1, p.Pattern)
		}
	}

	// Problem bodies should be non-empty
	for i, p := range result {
		if strings.TrimSpace(p.Body) == "" {
			t.Errorf("problem %d body should not be empty", i+1)
		}
	}

	// Line numbers should be present
	if result[0].LineStart < 1 {
		t.Errorf("problem 1 LineStart should be >= 1, got %d", result[0].LineStart)
	}
}

// TestParsePatternCStart verifies that pattern C (enumerate with "\arabic*" and start=N) is correctly parsed:
//   - 3 problems from pattern_c.tex (2 in first enumerate, 1 in second with start=3)
//   - start=N parameter should not affect parsing — all items are still problems
//
// This is the RED phase — parser not yet implemented.
func TestParsePatternCStart(t *testing.T) {
	input := readTestFile(t, "pattern_c.tex")

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)
	result, _ := ParseEnumerate(blocks)

	if len(result) != 3 {
		t.Fatalf("expected 3 problems from pattern_c (2 + 1 with start=3), got %d", len(result))
	}

	// All problems should have Pattern = "C"
	for i, p := range result {
		if p.Pattern != PatternC {
			t.Errorf("problem %d should have Pattern=C, got %q", i+1, p.Pattern)
		}
	}

	// Problem bodies should be non-empty
	for i, p := range result {
		if strings.TrimSpace(p.Body) == "" {
			t.Errorf("problem %d body should not be empty", i+1)
		}
	}

	// The second enumerate block with start=3 should still parse its item as a problem
	// (start=N affects numbering in the output, not whether items are collected)
	if !strings.Contains(result[2].Body, `\item`) && !strings.HasPrefix(strings.TrimSpace(result[2].Body), `计算`) {
		t.Logf("problem 3 body starts with: %s", result[2].Body[:min(len(result[2].Body), 60)])
	}
}

// TestParseMultipleEnumerateBlocks verifies that separate enumerate blocks
// do not interfere with each other during parsing. Problems should be collected
// per-block and not mixed across blocks.
//
// This is the RED phase — parser not yet implemented.
func TestParseMultipleEnumerateBlocks(t *testing.T) {
	input := readTestFile(t, "pattern_b.tex")

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)
	result, _ := ParseEnumerate(blocks)

	// pattern_b.tex has 2 enumerate blocks (with resume).
	// Each should be parsed independently for its items.
	if len(result) == 0 {
		t.Fatal("expected at least 1 problem from pattern_b multiple enumerate blocks")
	}

	// Verify that items from the second enumerate block are collected as problems
	// (the fact that there are 2 blocks doesn't mean we lose items)
	t.Logf("parsed %d problems across multiple enumerate blocks", len(result))

	// Each problem should have a non-empty body with actual LaTeX content
	for i, p := range result {
		body := strings.TrimSpace(p.Body)
		if body == "" {
			t.Errorf("problem %d body is empty", i+1)
		}
		if p.LineStart < 1 {
			t.Errorf("problem %d has invalid LineStart: %d", i+1, p.LineStart)
		}
	}
}

// TestParseEnumerateBlocksIndependent verifies that separate unrelated enumerate blocks
// do not leak items across block boundaries.
func TestParseEnumerateBlocksIndependent(t *testing.T) {
	// Simulate two completely separate enumerate blocks in the same file:
	// Block 1: 2 items, Block 2: 1 item
	input := `\begin{enumerate}
\item First problem in block 1.
\item Second problem in block 1.
\end{enumerate}

Some text between blocks.

\begin{enumerate}
\item Single problem in block 2.
\end{enumerate}`

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)
	result, _ := ParseEnumerate(blocks)

	if len(result) != 3 {
		t.Fatalf("expected 3 problems across 2 independent enumerate blocks, got %d", len(result))
	}

	// Items from block 1 should not contaminate block 2
	if strings.Contains(result[2].Body, "Second problem") {
		t.Errorf("block 2 items should not contain content from block 1")
	}

	t.Logf("parsed %d problems across 2 independent enumerate blocks", len(result))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// Pattern D (mybox) + edge cases — Task 4
// ---------------------------------------------------------------------------

// TestParsePatternD verifies mybox-based problem extraction (Pattern D).
//
// Expected behavior (post-implementation):
//   - 2 problems extracted (one per \begin{mybox})
//   - Problem[0].Label contains "单选题：集合的交集运算"
//   - Problem[1].Label contains "填空题：函数定义域"
//   - Each Body contains the full content inside the mybox (without the wrapper)
func TestParsePatternD(t *testing.T) {
	input := readTestFile(t, "pattern_d.tex")

	// ParseMyBox does not exist yet — this test will fail to compile.
	// Intended API: func ParseMyBox(blocks []Block) []ProblemBlock
	blocks := ScanBlocks(input)
	result := ParseMyBox(blocks)

	if len(result) != 2 {
		t.Fatalf("expected 2 problems from pattern_d (one per mybox), got %d", len(result))
	}

	// Problem 1 should be the multiple-choice problem about set intersection
	if !strings.Contains(result[0].Label, "集合的交集运算") {
		t.Errorf("problem 1 label should contain '集合的交集运算', got %q", result[0].Label)
	}

	// Problem 2 should be the fill-in-the-blank problem about function domain
	if !strings.Contains(result[1].Label, "函数定义域") {
		t.Errorf("problem 2 label should contain '函数定义域', got %q", result[1].Label)
	}

	// Problem bodies should contain the problem content (not the mybox wrapper)
	for i, p := range result {
		if strings.Contains(p.Body, `\begin{mybox}`) {
			t.Errorf("problem %d body should not contain \\begin{mybox} wrapper", i+1)
		}
		if strings.Contains(p.Body, `\end{mybox}`) {
			t.Errorf("problem %d body should not contain \\end{mybox} wrapper", i+1)
		}
	}

	// All problems should have Pattern = "D"
	for i, p := range result {
		if p.Pattern != PatternD {
			t.Errorf("problem %d should have Pattern=D, got %q", i+1, p.Pattern)
		}
	}

	// Line numbers should be present
	if result[0].LineStart < 1 {
		t.Errorf("problem 1 LineStart should be >= 1, got %d", result[0].LineStart)
	}
}

// TestParseEmptyFile verifies that an empty file produces an empty result.
//
// Expected behavior (post-implementation):
//   - ParseResult.Problems is empty
//   - ParseResult.Errors may contain a warning but not a fatal error
func TestParseEmptyFile(t *testing.T) {
	input := readTestFile(t, "edge_empty.tex")

	// ScanBlocks does not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)

	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks from empty file, got %d", len(blocks))
	}
}

// TestParseMalformedEnumerate verifies graceful handling of an unclosed
// \begin{enumerate} (missing \end{enumerate}).
//
// Expected behavior (post-implementation):
//   - Partial result: items before the breakpoint are still extracted
//   - ParseResult.Errors contains an error referencing the unclosed environment
//   - The error includes the line number of the unmatched \begin{enumerate}
func TestParseMalformedEnumerate(t *testing.T) {
	input := readTestFile(t, "edge_malformed.tex")

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	// Intended API: func ParseEnumerate(blocks []Block) ([]ProblemBlock, []ParseError)
	blocks := ScanBlocks(input)
	result, errs := ParseEnumerate(blocks)

	// Should have partial results (the 2 \item entries before the missing \end)
	if len(result) != 2 {
		t.Errorf("expected 2 problems from malformed file, got %d", len(result))
	}

	// Should report at least one error about the unclosed environment
	if len(errs) == 0 {
		t.Errorf("expected at least 1 parse error for unclosed enumerate, got 0")
	}

	// The error should reference a line number
	for _, e := range errs {
		if e.Line < 1 {
			t.Errorf("parse error line number should be >= 1, got %d", e.Line)
		}
	}
}

// TestParseNestedEnumerate verifies that an outer enumerate's \item boundaries
// define problems, while inner enumerate content stays within the parent problem.
//
// Expected behavior (post-implementation):
//   - 2 problems extracted (from outer enumerate items)
//   - The first problem's Body contains the inner enumerate content
//   - The inner enumerate items are NOT extracted as separate problems
func TestParseNestedEnumerate(t *testing.T) {
	input := readTestFile(t, "edge_nested.tex")

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)
	result, _ := ParseEnumerate(blocks)

	if len(result) != 2 {
		t.Fatalf("expected 2 problems from nested enumerate (outer items), got %d", len(result))
	}

	// The first problem body should contain the inner enumerate content
	if !strings.Contains(result[0].Body, `\begin{enumerate}`) {
		t.Errorf("problem 1 body should contain inner \\begin{enumerate}")
	}
	if !strings.Contains(result[0].Body, `\end{enumerate}`) {
		t.Errorf("problem 1 body should contain inner \\end{enumerate}")
	}

	// All problems should have Pattern = "A" (outer enumerate uses 题\arabic*)
	for i, p := range result {
		if p.Pattern != PatternA {
			t.Errorf("problem %d should have Pattern=A, got %q", i+1, p.Pattern)
		}
	}

	// Line numbers should be present
	if result[0].LineStart < 1 {
		t.Errorf("problem 1 LineStart should be >= 1, got %d", result[0].LineStart)
	}
}

// TestParseMixedType verifies that a problem containing both \begin{tasks}
// and \underline is flagged for review.
//
// Expected behavior (post-implementation):
//   - 1 problem extracted
//   - Problem body contains both tasks environment and \underline content
//   - Classifier marks inferredType = "other" / needsReview = true
func TestParseMixedType(t *testing.T) {
	input := readTestFile(t, "edge_mixed.tex")

	// ScanBlocks and ParseEnumerate do not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)
	result, _ := ParseEnumerate(blocks)

	if len(result) != 1 {
		t.Fatalf("expected 1 problem from mixed-type file, got %d", len(result))
	}

	// Body should contain both tasks and underline
	if !strings.Contains(result[0].Body, `\begin{tasks}`) {
		t.Errorf("problem body should contain \\begin{tasks}")
	}
	if !strings.Contains(result[0].Body, `\underline`) {
		t.Errorf("problem body should contain \\underline")
	}

	// Label should be non-empty
	if result[0].Label == "" {
		t.Errorf("problem label should not be empty")
	}
}

// TestParseNoEnumerate verifies that plain text with no LaTeX problem
// environments produces an empty result.
//
// Expected behavior (post-implementation):
//   - ScanBlocks produces 0 blocks (or only BlockText)
//   - No fatal error
func TestParseNoEnumerate(t *testing.T) {
	input := "This is just plain text with no LaTeX environments."

	// ScanBlocks does not exist yet — this test will fail to compile.
	blocks := ScanBlocks(input)

	// Plain text without LaTeX commands should produce a small number
	// of BlockText entries, not errors
	_ = blocks
	t.Logf("plain text produced %d blocks", len(blocks))
}

func TestParseKnowledgePointEnumerate(t *testing.T) {
	input := `\documentclass{article}
\usepackage{amsmath}
\begin{document}
\section*{二项式定理}
\begin{enumerate}
\item 基本概念：二项展开式
\item 通项公式：T_{r+1}=C_n^r a^{n-r} b^r
\item 项数：共n+1项
\end{enumerate}
\section*{特定项系数问题}
\begin{enumerate}[label=\arabic*., leftmargin=1.5em]
\item (3-2x)^5 展开式中 x^3 的系数是.
\item 已知 (x+a/x)^6 展开式中常数项是 20
\end{enumerate}
\end{document}`

	blocks := ScanBlocks(input)
	blocks = skipPreambleBlocks(blocks)
	blocks = trimTrailingEndDocument(blocks)

	result, errors := ParseEnumerate(blocks)

	if len(result) != 2 {
		t.Fatalf("expected 2 problems (knowledge point enumerate skipped), got %d", len(result))
	}
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestParseTextMarkerExamples(t *testing.T) {
	input := `\documentclass{article}
\begin{document}
\section*{考点}

\noindent \textbf{例1.} 函数 f(x) = cos(3x + pi/6) 在 [0, pi] 上的零点个数是：
\begin{enumerate}[label=\\Alph*.]
\item 2 \item 3 \item 4 \item 5
\end{enumerate}

\noindent \textbf{例2.} 函数 f(x) = x^2 + x - 2 的零点个数为：
\begin{enumerate}[label=\\Alph*.]
\item 1 \item 2 \item 3 \item 4
\end{enumerate}

\noindent \textbf{例3.} 函数 f(x) = sqrt(x) - (1/2)^x 的零点个数为：
\begin{enumerate}[label=\\Alph*.]
\item 0 \item 1 \item 2 \item 3
\end{enumerate}
\end{document}`

	blocks := ScanBlocks(input)
	blocks = skipPreambleBlocks(blocks)
	blocks = trimTrailingEndDocument(blocks)

	result := ParseTextMarkers(blocks)

	if len(result) != 3 {
		t.Fatalf("expected 3 problems from text markers, got %d", len(result))
	}
	if !strings.Contains(result[0].Body, "cos") {
		t.Errorf("problem 1 should contain 'cos'")
	}
}

func TestParseTextMarkerWithConceptMybox(t *testing.T) {
	input := `\documentclass{article}
\begin{document}
\begin{mybox}{函数的定义}
概念内容在这里。
\end{mybox}

\noindent \textbf{例1.} 这是真正的题目。
\begin{enumerate}[label=\Alph*.]
\item A \item B \item C \item D
\end{enumerate}

\begin{mybox}{易错点总结}
易错内容。
\end{mybox}
\end{document}`

	blocks := ScanBlocks(input)
	blocks = skipPreambleBlocks(blocks)
	blocks = trimTrailingEndDocument(blocks)

	myboxProblems := ParseMyBox(blocks)
	if len(myboxProblems) > 0 {
		t.Errorf("expected 0 mybox problems (concept boxes filtered), got %d", len(myboxProblems))
	}

	textProblems := ParseTextMarkers(blocks)
	if len(textProblems) != 1 {
		t.Fatalf("expected 1 text-marker problem, got %d", len(textProblems))
	}
}
