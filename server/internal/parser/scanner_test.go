package parser

import (
	"strings"
	"testing"
)

func TestScanSimpleEnumerate(t *testing.T) {
	input := `\begin{enumerate}[label=\textbf{题 \arabic*}]
\item First problem
\item Second problem
\end{enumerate}`
	blocks := ScanBlocks(input)

	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}

	if blocks[0].Type != BlockEnvBegin {
		t.Errorf("block 0: expected BlockEnvBegin, got %v", blocks[0].Type)
	}
	if blocks[0].EnvName != "enumerate" {
		t.Errorf("block 0: expected EnvName=enumerate, got %q", blocks[0].EnvName)
	}
	if blocks[0].EnvArgs != `label=\textbf{题 \arabic*}` {
		t.Errorf("block 0: expected EnvArgs with nested braces, got %q", blocks[0].EnvArgs)
	}
	if blocks[0].LineStart != 1 {
		t.Errorf("block 0: expected LineStart=1, got %d", blocks[0].LineStart)
	}

	if blocks[1].Type != BlockItem {
		t.Errorf("block 1: expected BlockItem, got %v", blocks[1].Type)
	}
	if blocks[1].Content != "First problem" {
		t.Errorf("block 1: expected Content='First problem', got %q", blocks[1].Content)
	}
	if blocks[1].LineStart != 2 {
		t.Errorf("block 1: expected LineStart=2, got %d", blocks[1].LineStart)
	}

	if blocks[2].Type != BlockItem {
		t.Errorf("block 2: expected BlockItem, got %v", blocks[2].Type)
	}
	if blocks[2].Content != "Second problem" {
		t.Errorf("block 2: expected Content='Second problem', got %q", blocks[2].Content)
	}
	if blocks[2].LineStart != 3 {
		t.Errorf("block 2: expected LineStart=3, got %d", blocks[2].LineStart)
	}

	if blocks[3].Type != BlockEnvEnd {
		t.Errorf("block 3: expected BlockEnvEnd, got %v", blocks[3].Type)
	}
	if blocks[3].EnvName != "enumerate" {
		t.Errorf("block 3: expected EnvName=enumerate, got %q", blocks[3].EnvName)
	}
	if blocks[3].LineStart != 4 {
		t.Errorf("block 3: expected LineStart=4, got %d", blocks[3].LineStart)
	}
}

func TestScanItemWithOptionalArg(t *testing.T) {
	input := `\item[Optional] Text content`
	blocks := ScanBlocks(input)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	if blocks[0].Type != BlockItem {
		t.Errorf("expected BlockItem, got %v", blocks[0].Type)
	}
	if blocks[0].Label != "Optional" {
		t.Errorf("expected Label='Optional', got %q", blocks[0].Label)
	}
	if blocks[0].Content != "Text content" {
		t.Errorf("expected Content='Text content', got %q", blocks[0].Content)
	}
}

func TestScanItemNoOptionalArg(t *testing.T) {
	input := `\item Just text`
	blocks := ScanBlocks(input)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	if blocks[0].Type != BlockItem {
		t.Errorf("expected BlockItem, got %v", blocks[0].Type)
	}
	if blocks[0].Label != "" {
		t.Errorf("expected empty Label, got %q", blocks[0].Label)
	}
	if blocks[0].Content != "Just text" {
		t.Errorf("expected Content='Just text', got %q", blocks[0].Content)
	}
}

func TestScanNestedEnvironments(t *testing.T) {
	input := `\begin{enumerate}
\item Outer
\begin{tasks}(2)
\task Option A
\task Option B
\end{tasks}
\end{enumerate}`
	blocks := ScanBlocks(input)

	if len(blocks) != 7 {
		t.Fatalf("expected 7 blocks, got %d", len(blocks))
	}

	if blocks[0].Type != BlockEnvBegin || blocks[0].EnvName != "enumerate" {
		t.Errorf("block 0: expected EnvBegin(enumerate), got Type=%v EnvName=%q", blocks[0].Type, blocks[0].EnvName)
	}

	if blocks[1].Type != BlockItem {
		t.Errorf("block 1: expected BlockItem, got %v", blocks[1].Type)
	}
	if blocks[1].Content != "Outer" {
		t.Errorf("block 1: expected Content='Outer', got %q", blocks[1].Content)
	}

	if blocks[2].Type != BlockEnvBegin || blocks[2].EnvName != "tasks" {
		t.Errorf("block 2: expected EnvBegin(tasks), got Type=%v EnvName=%q", blocks[2].Type, blocks[2].EnvName)
	}

	if blocks[3].Type != BlockText {
		t.Errorf("block 3: expected BlockText (for \\task), got %v", blocks[3].Type)
	}
	if blocks[4].Type != BlockText {
		t.Errorf("block 4: expected BlockText (for \\task), got %v", blocks[4].Type)
	}

	if blocks[5].Type != BlockEnvEnd || blocks[5].EnvName != "tasks" {
		t.Errorf("block 5: expected EnvEnd(tasks), got Type=%v EnvName=%q", blocks[5].Type, blocks[5].EnvName)
	}

	if blocks[6].Type != BlockEnvEnd || blocks[6].EnvName != "enumerate" {
		t.Errorf("block 6: expected EnvEnd(enumerate), got Type=%v EnvName=%q", blocks[6].Type, blocks[6].EnvName)
	}
}

func TestScanBraceNesting(t *testing.T) {
	input := `\textbf{\underline{text}}`
	blocks := ScanBlocks(input)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockText {
		t.Errorf("expected BlockText, got %v", blocks[0].Type)
	}
}

func TestScanBomPrefix(t *testing.T) {
	input := "\xEF\xBB\xBF" + `\begin{enumerate}
\item Test
\end{enumerate}`
	blocks := ScanBlocks(input)

	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks (BOM stripped), got %d", len(blocks))
	}
	if blocks[0].Type != BlockEnvBegin || blocks[0].EnvName != "enumerate" {
		t.Errorf("block 0: expected EnvBegin(enumerate), got Type=%v EnvName=%q", blocks[0].Type, blocks[0].EnvName)
	}
	if blocks[1].Type != BlockItem || blocks[1].Content != "Test" {
		t.Errorf("block 1: expected Item(Test), got Type=%v Content=%q", blocks[1].Type, blocks[1].Content)
	}
	if blocks[2].Type != BlockEnvEnd || blocks[2].EnvName != "enumerate" {
		t.Errorf("block 2: expected EnvEnd(enumerate), got Type=%v EnvName=%q", blocks[2].Type, blocks[2].EnvName)
	}
}

func TestScanLineNumbers(t *testing.T) {
	input := `line1
\begin{enumerate}
\item problem
\end{enumerate}
line4`
	blocks := ScanBlocks(input)

	if len(blocks) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(blocks))
	}

	if blocks[0].Type != BlockText || blocks[0].LineStart != 1 {
		t.Errorf("block 0: expected Text at line 1, got Type=%v Line=%d", blocks[0].Type, blocks[0].LineStart)
	}
	if blocks[1].Type != BlockEnvBegin || blocks[1].LineStart != 2 {
		t.Errorf("block 1: expected EnvBegin at line 2, got Type=%v Line=%d", blocks[1].Type, blocks[1].LineStart)
	}
	if blocks[2].Type != BlockItem || blocks[2].LineStart != 3 {
		t.Errorf("block 2: expected Item at line 3, got Type=%v Line=%d", blocks[2].Type, blocks[2].LineStart)
	}
	if blocks[3].Type != BlockEnvEnd || blocks[3].LineStart != 4 {
		t.Errorf("block 3: expected EnvEnd at line 4, got Type=%v Line=%d", blocks[3].Type, blocks[3].LineStart)
	}
	if blocks[4].Type != BlockText || blocks[4].LineStart != 5 {
		t.Errorf("block 4: expected Text at line 5, got Type=%v Line=%d", blocks[4].Type, blocks[4].LineStart)
	}
}

func TestScanComment(t *testing.T) {
	input := `% This is a comment
\begin{enumerate}
\item First
\end{enumerate}`
	blocks := ScanBlocks(input)

	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}
	if blocks[0].Type != BlockComment {
		t.Errorf("block 0: expected BlockComment, got %v", blocks[0].Type)
	}
	if blocks[0].LineStart != 1 {
		t.Errorf("block 0: expected LineStart=1, got %d", blocks[0].LineStart)
	}
	if blocks[1].Type != BlockEnvBegin || blocks[1].EnvName != "enumerate" {
		t.Errorf("block 1: expected EnvBegin(enumerate), got Type=%v EnvName=%q", blocks[1].Type, blocks[1].EnvName)
	}
}

func TestScanSectionCommand(t *testing.T) {
	input := `\section*{集合与不等式}
Some text
\subsection*{函数}
More text`
	blocks := ScanBlocks(input)

	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}
	if blocks[0].Type != BlockCommand {
		t.Errorf("block 0: expected BlockCommand (section*), got %v", blocks[0].Type)
	}
	if !strings.Contains(blocks[0].Content, `\section*`) {
		t.Errorf("block 0: expected Content to contain \\section*, got %q", blocks[0].Content)
	}
	if blocks[0].LineStart != 1 {
		t.Errorf("block 0: expected LineStart=1, got %d", blocks[0].LineStart)
	}
	if blocks[1].Type != BlockText {
		t.Errorf("block 1: expected BlockText, got %v", blocks[1].Type)
	}
	if blocks[2].Type != BlockCommand {
		t.Errorf("block 2: expected BlockCommand (subsection*), got %v", blocks[2].Type)
	}
}

func TestScanPlainText(t *testing.T) {
	input := `Hello world
This is plain text
No LaTeX commands here`
	blocks := ScanBlocks(input)

	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}
	for i, b := range blocks {
		if b.Type != BlockText {
			t.Errorf("block %d: expected BlockText, got %v", i, b.Type)
		}
		if b.LineStart != i+1 {
			t.Errorf("block %d: expected LineStart=%d, got %d", i, i+1, b.LineStart)
		}
	}
}

func TestScanEmptyInput(t *testing.T) {
	blocks := ScanBlocks("")
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty input, got %d", len(blocks))
	}
}

func TestScanBlankLinesOnly(t *testing.T) {
	blocks := ScanBlocks("\n\n\n")
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for blank lines, got %d", len(blocks))
	}
}

func TestScanWithTrailingWhitespace(t *testing.T) {
	input := `\begin{enumerate}[label=\textbf{题 \arabic*}]
\item First  
\item Second  
\end{enumerate}`
	blocks := ScanBlocks(input)

	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}

	if blocks[1].Content != "First" {
		t.Errorf("block 1 content should have trimmed trailing spaces, got %q", blocks[1].Content)
	}
	if blocks[2].Content != "Second" {
		t.Errorf("block 2 content should have trimmed trailing spaces, got %q", blocks[2].Content)
	}
}

func TestScanCrLfLineEndings(t *testing.T) {
	input := "\\begin{enumerate}\r\n\\item First\r\n\\end{enumerate}"
	blocks := ScanBlocks(input)

	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks with CRLF, got %d", len(blocks))
	}
	if blocks[0].Type != BlockEnvBegin || blocks[0].EnvName != "enumerate" {
		t.Errorf("block 0: expected EnvBegin(enumerate), got Type=%v EnvName=%q", blocks[0].Type, blocks[0].EnvName)
	}
	if blocks[1].Type != BlockItem {
		t.Errorf("block 1: expected BlockItem, got %v", blocks[1].Type)
	}
	if blocks[2].Type != BlockEnvEnd || blocks[2].EnvName != "enumerate" {
		t.Errorf("block 2: expected EnvEnd(enumerate), got Type=%v EnvName=%q", blocks[2].Type, blocks[2].EnvName)
	}
}

func TestScanCrLineEndings(t *testing.T) {
	input := "\r"
	blocks := ScanBlocks(input)

	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for bare CR, got %d", len(blocks))
	}

	input2 := "\\begin{enumerate}\r\\item First\r\\end{enumerate}"
	blocks2 := ScanBlocks(input2)

	if len(blocks2) != 3 {
		t.Fatalf("expected 3 blocks with old-Mac CR, got %d", len(blocks2))
	}
}

func TestScanEnvArgsWithNestedBraces(t *testing.T) {
	input := `\begin{enumerate}[label=\textbf{题 \arabic*}]`
	blocks := ScanBlocks(input)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockEnvBegin {
		t.Fatalf("expected BlockEnvBegin, got %v", blocks[0].Type)
	}
	if blocks[0].EnvName != "enumerate" {
		t.Errorf("expected EnvName=enumerate, got %q", blocks[0].EnvName)
	}
	expectedArgs := `label=\textbf{题 \arabic*}`
	if blocks[0].EnvArgs != expectedArgs {
		t.Errorf("expected EnvArgs=%q, got %q", expectedArgs, blocks[0].EnvArgs)
	}
}

func TestScanItemWithBracketedNestedBraces(t *testing.T) {
	// \item[label=\textbf{题 \arabic*}] text — nested braces in optional arg
	input := `\item[label=\textbf{题 \arabic*}] Some text`
	blocks := ScanBlocks(input)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockItem {
		t.Fatalf("expected BlockItem, got %v", blocks[0].Type)
	}
	expectedLabel := `label=\textbf{题 \arabic*}`
	if blocks[0].Label != expectedLabel {
		t.Errorf("expected Label=%q, got %q", expectedLabel, blocks[0].Label)
	}
	if blocks[0].Content != "Some text" {
		t.Errorf("expected Content='Some text', got %q", blocks[0].Content)
	}
}

func TestScanInlineMathBrackets(t *testing.T) {
	input := `\( (-\infty, 5] \)`
	blocks := ScanBlocks(input)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockText {
		t.Errorf("expected BlockText for math line, got %v", blocks[0].Type)
	}
}

func TestScanBeginDocument(t *testing.T) {
	input := `\begin{document}
content
\end{document}`
	blocks := ScanBlocks(input)

	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}
	if blocks[0].Type != BlockEnvBegin || blocks[0].EnvName != "document" {
		t.Errorf("block 0: expected EnvBegin(document), got Type=%v EnvName=%q", blocks[0].Type, blocks[0].EnvName)
	}
	if blocks[1].Type != BlockText {
		t.Errorf("block 1: expected BlockText, got %v", blocks[1].Type)
	}
	if blocks[2].Type != BlockEnvEnd || blocks[2].EnvName != "document" {
		t.Errorf("block 2: expected EnvEnd(document), got Type=%v EnvName=%q", blocks[2].Type, blocks[2].EnvName)
	}
}

func TestScanPatternA(t *testing.T) {
	input := readTestFile(t, "pattern_a.tex")
	blocks := ScanBlocks(input)

	const minBlocks = 10
	if len(blocks) < minBlocks {
		t.Fatalf("expected at least %d blocks from pattern_a.tex, got %d", minBlocks, len(blocks))
	}

	var hasEnvBegin, hasEnvEnd, hasItem, hasSection bool
	for _, b := range blocks {
		switch b.Type {
		case BlockEnvBegin:
			hasEnvBegin = true
		case BlockEnvEnd:
			hasEnvEnd = true
		case BlockItem:
			hasItem = true
		case BlockCommand:
			hasSection = true
		}
	}
	if !hasEnvBegin {
		t.Errorf("pattern_a.tex should produce at least one BlockEnvBegin")
	}
	if !hasEnvEnd {
		t.Errorf("pattern_a.tex should produce at least one BlockEnvEnd")
	}
	if !hasItem {
		t.Errorf("pattern_a.tex should produce at least one BlockItem")
	}
	if !hasSection {
		t.Errorf("pattern_a.tex should produce at least one BlockCommand (section)")
	}

	t.Logf("pattern_a.tex: %d blocks total", len(blocks))
}
