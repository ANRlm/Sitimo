package parser

import "strings"

// ExtractSectionTags extracts section/subsection titles as tag names.
// Scans \section*{}, \subsection*{}, and mybox titles.
// Skips blacklisted generic titles and deduplicates by title text.
func ExtractSectionTags(blocks []Block) []string {
	var tags []string
	seen := make(map[string]bool)

	for _, b := range blocks {
		var title string
		switch {
		case b.Type == BlockCommand && isSectionLike(b.Content):
			title = extractBracedContent(b.Content)
		case b.Type == BlockEnvBegin && b.EnvName == "mybox":
			title = b.EnvArgs
		}
		if title != "" && !defaultBlacklist[title] && !seen[title] {
			seen[title] = true
			tags = append(tags, title)
		}
	}
	return tags
}

// defaultBlacklist contains generic section titles to skip.
var defaultBlacklist = map[string]bool{
	"":     true,
	"练习": true,
	"习题": true,
	"测试": true,
}

// isSectionLike checks if a command is \section* or \subsection*.
func isSectionLike(content string) bool {
	return strings.HasPrefix(content, "\\section*") || strings.HasPrefix(content, "\\subsection*")
}

// extractBracedContent extracts text between the first { and last }.
// For a LaTeX command like \section*{title}, returns "title".
func extractBracedContent(content string) string {
	start := strings.IndexByte(content, '{')
	if start == -1 {
		return ""
	}
	end := strings.LastIndexByte(content, '}')
	if end == -1 || end <= start {
		return ""
	}
	return content[start+1 : end]
}
