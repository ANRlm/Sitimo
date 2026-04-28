package parser

import (
	"strings"
)

// ParseEnumerate extracts problems from enumerate blocks.
//
// Patterns are detected from the enumerate label:
//   - Pattern A: label contains "题"
//   - Pattern B: label contains "例"
//   - Pattern C: otherwise (bare \arabic*)
//
// Items at the outermost enumerate depth become individual ProblemBlocks.
// Nested enumerate content is preserved inside the parent problem body.
// Malformed (unclosed) enumerates yield partial results with ParseErrors.
func ParseEnumerate(blocks []Block) ([]ProblemBlock, []ParseError) {
	var problems []ProblemBlock
	var errors []ParseError

	enumerateDepth := 0
	var currentProblem *ProblemBlock
	var currentPattern string
	var currentEnvArgs string
	var openStarts []int
	var savedPattern string
	var savedEnvArgs string


	for _, block := range blocks {
		switch block.Type {
		case BlockEnvBegin:
			if block.EnvName == "enumerate" {
				enumerateDepth++
				if enumerateDepth == 1 {
					currentEnvArgs = block.EnvArgs
					currentPattern = detectPattern(block.EnvArgs)
					openStarts = append(openStarts, block.LineStart)
					continue
				}
				if currentProblem != nil {
					currentProblem.Body = appendBody(currentProblem.Body, block.Content)
				}
			} else if block.EnvName == "tasks" {
				enumerateDepth++
				if enumerateDepth == 1 {
					flushProblem(&problems, currentProblem)
					currentProblem = nil
					savedPattern = currentPattern
					savedEnvArgs = currentEnvArgs
					currentEnvArgs = block.EnvArgs
					currentPattern = PatternA
					openStarts = append(openStarts, block.LineStart)
					continue
				}
				if currentProblem != nil {
					currentProblem.Body = appendBody(currentProblem.Body, block.Content)
				}
			} else if currentProblem != nil {
				currentProblem.Body = appendBody(currentProblem.Body, block.Content)
			}

		case BlockEnvEnd:
			if block.EnvName == "enumerate" || block.EnvName == "tasks" {
				if enumerateDepth == 1 {
					flushProblem(&problems, currentProblem)
					currentProblem = nil
					if n := len(openStarts); n > 0 {
						openStarts = openStarts[:n-1]
					}
						if block.EnvName == "tasks" && savedPattern != "" {
						currentPattern = savedPattern
						currentEnvArgs = savedEnvArgs
						savedPattern = ""
						savedEnvArgs = ""
						}
				} else if enumerateDepth > 1 {
					if currentProblem != nil {
						currentProblem.Body = appendBody(currentProblem.Body, block.Content)
					}
				}
				enumerateDepth--
			} else if currentProblem != nil {
				currentProblem.Body = appendBody(currentProblem.Body, block.Content)
			}

		case BlockItem:
			if enumerateDepth == 1 {
				if isKnowledgeContent(block.Content) {
					continue
				}
				flushProblem(&problems, currentProblem)
				label := block.Label
				if label == "" {
					label = extractLabelFromEnvArgs(currentEnvArgs)
				}
				currentProblem = &ProblemBlock{
					Body:      block.Content,
					LineStart: block.LineStart,
					Label:     label,
					Pattern:   currentPattern,
				}
			} else if enumerateDepth > 1 && currentProblem != nil {
				currentProblem.Body = appendBody(currentProblem.Body, block.Content)
			}

		default:
			if enumerateDepth >= 1 && currentProblem != nil {
				currentProblem.Body = appendBody(currentProblem.Body, block.Content)
			}
		}
	}

	flushProblem(&problems, currentProblem)

	for _, line := range openStarts {
		errors = append(errors, ParseError{
			Line:    line,
			Message: "unclosed \\begin{enumerate} (missing \\end{enumerate})",
		})
	}

	return problems, errors
}

func flushProblem(problems *[]ProblemBlock, p *ProblemBlock) {
	if p == nil || strings.TrimSpace(p.Body) == "" {
		return
	}
	*problems = append(*problems, *p)
}

func appendBody(current, next string) string {
	if current == "" {
		return next
	}
	return current + "\n" + next
}

func isKnowledgeContent(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	// Bold label with colon: \textbf{定义：...} or \textbf{定义}：...
	if strings.HasPrefix(trimmed, `\textbf{`) {
		rest := trimmed[len(`\textbf{`):]
		for _, r := range rest {
			if r == '}' {
				return false
			}
			if r == '\uff1a' || r == ':' {
				return true
			}
		}
		return false
	}

	// Non-bold knowledge items: "基本概念：二项展开式" etc.
	// Must be short and start with a knowledge keyword followed by colon.
	runes := []rune(trimmed)
	if len(runes) < 40 {
		for _, kw := range knowledgeKeywords {
			if strings.HasPrefix(trimmed, kw+":") ||
				strings.HasPrefix(trimmed, kw+string('\uff1a')) {
				return true
			}
		}
	}

	return false
}

var knowledgeKeywords = []string{
	"基本概念", "通项公式", "项数", "定义", "定理", "性质",
	"公式", "运算律", "说明", "注意", "补充", "结论",
	"总结", "归纳", "易错", "原理", "表述", "常见变形",
	"核心公式", "使用口诀", "前提条件", "适用范围",
}

// detectPattern inspects enumerate optional arguments to determine the
// pattern type (A/B/C).
func detectPattern(envArgs string) string {
	lower := strings.ToLower(envArgs)
	if strings.Contains(lower, "\u9898") {
		return PatternA
	}
	if strings.Contains(lower, "\u4f8b") {
		return PatternB
	}
	return PatternC
}

// extractLabelFromEnvArgs derives a label text from the enumerate
// environment's optional arguments when the \item itself has no label.
//
// Example: "label=\textbf{题 \arabic*}" → "题"
func extractLabelFromEnvArgs(envArgs string) string {
	prefix := `label=`
	idx := strings.Index(envArgs, prefix)
	if idx < 0 {
		return ""
	}
	rest := envArgs[idx+len(prefix):]
	rest = strings.TrimSpace(rest)

	if idx2 := strings.Index(rest, `\textbf{`); idx2 >= 0 {
		rest = rest[idx2+len(`\textbf{`):]
		// Find the matching closing brace or \arabic*
		depth := 0
		for i, r := range rest {
			switch r {
			case '{':
				depth++
			case '}':
				if depth == 0 {
					return strings.TrimSpace(rest[:i])
				}
				depth--
			case '\\':
				if strings.HasPrefix(rest[i:], `\arabic`) {
					return strings.TrimSpace(rest[:i])
				}
			}
		}
	}
	return ""
}
