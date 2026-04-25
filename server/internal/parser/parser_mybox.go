package parser

import (
	"strings"
)

// ParseMyBox extracts problems from mybox environments.
//
// Each \begin{mybox}{title} → \end{mybox} block is treated as one problem.
// The title (inside the first brace group after \begin{mybox}) is extracted
// as both a section tag and stored in the Label field.
//
// Nested mybox environments are handled: inner \begin{mybox} increments
// the depth counter, and only the outermost matching \end{mybox} finalizes
// a problem. Content from inner mybox environments is preserved in the body.
//
// Blocks that are inside an active mybox are accumulated into the problem body.
// The \begin{mybox}{title} and \end{mybox} wrapper lines themselves are
// excluded from the body.
func ParseMyBox(blocks []Block) []ProblemBlock {
	var problems []ProblemBlock
	var currentBody strings.Builder
	var currentTitle string
	var currentLineStart int
	myboxDepth := 0
	startedBody := false

	for _, block := range blocks {
		switch block.Type {
		case BlockEnvBegin:
			if block.EnvName == "mybox" {
				myboxDepth++
				if myboxDepth == 1 {
					// Start of new problem.
					currentTitle = extractMyboxTitle(block.Content)
					currentLineStart = block.LineStart
					currentBody.Reset()
					startedBody = false
					// Do not include the \begin{mybox} line in body
					continue
				}
				// Nested mybox: include in body
			}

			// Non-mybox envbegin while inside mybox: accumulate
			if myboxDepth > 0 {
				if startedBody {
					currentBody.WriteString("\n")
				}
				currentBody.WriteString(block.Content)
				startedBody = true
			}

		case BlockEnvEnd:
			if block.EnvName == "mybox" {
				myboxDepth--
				if myboxDepth == 0 {
					// Finalize problem — but skip concept boxes
					if !isConceptBox(currentTitle) {
						problems = append(problems, ProblemBlock{
							Body:        currentBody.String(),
							LineStart:   currentLineStart,
							Label:       currentTitle,
							Pattern:     PatternD,
							SectionTags: []string{currentTitle},
						})
					}
				}
				// Do not include the \end{mybox} line in body
				continue
			}

			// Non-mybox envend while inside mybox: accumulate
			if myboxDepth > 0 {
				if startedBody {
					currentBody.WriteString("\n")
				}
				currentBody.WriteString(block.Content)
				startedBody = true
			}

		default:
			// BlockItem, BlockCommand, BlockText, BlockComment
			if myboxDepth > 0 {
				if startedBody {
					currentBody.WriteString("\n")
				}
				currentBody.WriteString(block.Content)
				startedBody = true
			}
		}
	}

	return problems
}

var conceptKeywords = []string{
	"定义", "定理", "结论", "补充", "易错",
	"注意", "性质", "原理", "总结", "归纳",
}

var problemTypeIndicators = []string{
	"单选题", "多选题", "填空题", "解答题",
	"简答题", "证明题", "计算题", "判断题",
	"综合题", "选择题",
}

func isConceptBox(title string) bool {
	lower := strings.ToLower(title)
	for _, pt := range problemTypeIndicators {
		if strings.Contains(lower, pt) {
			return false
		}
	}
	for _, kw := range conceptKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// extractMyboxTitle extracts the title from the first brace group
// following \begin{mybox} in the raw content line.
//
// For example, given "\begin{mybox}{单选题：集合的交集运算}",
// it returns "单选题：集合的交集运算".
//
// Brace matching is used so nested braces within the title are
// handled correctly.
func extractMyboxTitle(content string) string {
	// Find "{title}" part after \begin{mybox}
	idx := strings.Index(content, `\begin{mybox}`)
	if idx < 0 {
		return ""
	}
	rest := content[idx+len(`\begin{mybox}`):]
	rest = strings.TrimSpace(rest)

	if !strings.HasPrefix(rest, "{") {
		return ""
	}

	// Brace-aware extraction: track depth to handle nested braces.
	depth := 0
	for i, r := range rest[1:] {
		switch r {
		case '{':
			depth++
		case '}':
			if depth == 0 {
				return rest[1 : i+1]
			}
			depth--
		}
	}
	return ""
}
