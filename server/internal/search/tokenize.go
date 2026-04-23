package search

import (
	"regexp"
	"slices"
	"strings"
)

var (
	commentRegexp = regexp.MustCompile(`(?m)%.*$`)
	commandRegexp = regexp.MustCompile(`\\([a-zA-Z]+)`)
	envRegexp     = regexp.MustCompile(`\\begin\{([^}]+)\}`)
	numberRegexp  = regexp.MustCompile(`\d+`)
)

var symbolTokens = map[string]string{
	`=`:  "eq",
	`+`:  "plus",
	`-`:  "minus",
	`^`:  "pow",
	`_`:  "sub",
	`\\infty`: "infty",
	`\\pi`:    "pi",
}

func TokenizeLatex(input string) string {
	s := commentRegexp.ReplaceAllString(input, "")
	tokens := make([]string, 0, 16)
	seen := map[string]struct{}{}

	appendToken := func(token string) {
		token = strings.TrimSpace(strings.ToLower(token))
		if token == "" {
			return
		}
		if _, exists := seen[token]; exists {
			return
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}

	for _, match := range envRegexp.FindAllStringSubmatch(s, -1) {
		appendToken("env_" + match[1])
	}

	for _, match := range commandRegexp.FindAllStringSubmatch(s, -1) {
		command := normalizeCommand(match[1])
		appendToken(command)
	}

	for symbol, token := range symbolTokens {
		if strings.Contains(s, symbol) {
			appendToken(token)
		}
	}

	for _, match := range numberRegexp.FindAllString(s, -1) {
		appendToken(match)
	}

	slices.Sort(tokens)
	return strings.Join(tokens, " ")
}

func normalizeCommand(raw string) string {
	switch raw {
	case "int":
		return "int"
	case "sum":
		return "sum"
	case "frac":
		return "frac"
	case "sqrt":
		return "sqrt"
	case "lim":
		return "lim"
	case "sin", "cos", "tan", "ln", "log":
		return raw
	default:
		return raw
	}
}

func LatexWarnings(input string) []string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}

	warnings := make([]string, 0, 2)
	if strings.Count(trimmed, "{") != strings.Count(trimmed, "}") {
		warnings = append(warnings, "花括号数量不匹配，已保存原始 LaTeX。")
	}
	if strings.Count(trimmed, `\begin{`) != strings.Count(trimmed, `\end{`) {
		warnings = append(warnings, "LaTeX 环境命令数量不匹配，建议检查 begin/end。")
	}
	if strings.Count(trimmed, `\(`) != strings.Count(trimmed, `\)`) {
		warnings = append(warnings, "行内公式分隔符 \\( 与 \\) 数量不一致。")
	}
	return warnings
}
