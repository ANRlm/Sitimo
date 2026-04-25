package parser

import (
	"bufio"
	"regexp"
	"strings"
)

var (
	envBeginRe = regexp.MustCompile(`^\\begin\{([^}]*)\}`)
	envEndRe   = regexp.MustCompile(`^\\end\{([^}]*)\}`)
	sectionRe  = regexp.MustCompile(`^\\(sub)?section\*?\{`)
)

// ScanBlocks tokenizes LaTeX source into a flat list of structural blocks.
func ScanBlocks(input string) []Block {
	input = normalizeLineEndings(input)
	input = stripStringBOM(input)

	var blocks []Block
	lineNum := 0
	sc := bufio.NewScanner(strings.NewReader(input))

	for sc.Scan() {
		line := sc.Text()
		lineNum++

		block := classifyLine(line, lineNum)
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	return blocks
}

func stripStringBOM(s string) string {
	if len(s) >= 3 && s[0] == 0xEF && s[1] == 0xBB && s[2] == 0xBF {
		return s[3:]
	}
	return s
}

func normalizeLineEndings(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")
	return input
}

func classifyLine(line string, lineNum int) *Block {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}

	if strings.HasPrefix(trimmed, "%") {
		return &Block{Type: BlockComment, Content: line, LineStart: lineNum}
	}

	if m := envBeginRe.FindStringSubmatch(trimmed); m != nil {
		envName := m[1]
		rest := trimmed[len(m[0]):]
		envArgs := extractEnvArgs(rest)
		return &Block{
			Type:      BlockEnvBegin,
			EnvName:   envName,
			EnvArgs:   envArgs,
			Content:   line,
			LineStart: lineNum,
		}
	}

	if m := envEndRe.FindStringSubmatch(trimmed); m != nil {
		return &Block{
			Type:      BlockEnvEnd,
			EnvName:   m[1],
			Content:   line,
			LineStart: lineNum,
		}
	}

	if strings.HasPrefix(trimmed, "\\item") {
		rest := trimmed[5:]
		label := extractBracketedContent(&rest)
		return &Block{
			Type:      BlockItem,
			Label:     label,
			Content:   strings.TrimSpace(rest),
			LineStart: lineNum,
		}
	}

	if sectionRe.MatchString(trimmed) {
		return &Block{
			Type:      BlockCommand,
			Content:   line,
			LineStart: lineNum,
		}
	}

	return &Block{
		Type:      BlockText,
		Content:   line,
		LineStart: lineNum,
	}
}

// extractEnvArgs extracts optional [args] following \begin{env} with brace matching.
func extractEnvArgs(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") {
		return ""
	}
	depth := 0
	for i, r := range s[1:] {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
		case ']':
			if depth == 0 {
				return s[1 : i+1]
			}
		}
	}
	return ""
}

// extractBracketedContent extracts content from leading [...] with brace awareness,
// and updates s to point past the bracket group.
func extractBracketedContent(s *string) string {
	*s = strings.TrimSpace(*s)
	if !strings.HasPrefix(*s, "[") {
		return ""
	}
	depth := 0
	for i, r := range (*s)[1:] {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
		case ']':
			if depth == 0 {
				content := (*s)[1 : i+1]
				*s = strings.TrimSpace((*s)[i+2:])
				return content
			}
		}
	}
	return ""
}
