package parser

import (
	"fmt"
	"strings"
)

type AnswerEntry struct {
	Index         int
	AnswerLatex   string
	SolutionLatex string
}

const (
	answerMarker         = `\textbf{答案：}`
	solutionMarker       = `\textbf{解析：}`
	solutionMarkerAlt    = `\textbf{解：}`
	answerMarkerBracket  = `\textbf{【答案】}`
	solutionMarkerBracket = `\textbf{【解析】}`
)

// ExtractAnswers extracts answer and solution LaTeX from answer key blocks.
//
// Two strategies are tried in order:
//  1. Enumerate-based: finds \item entries inside \begin{enumerate} blocks,
//     then locates \textbf{答案：} and \textbf{解析：} within each item.
//  2. Direct scanning: linearly scans blocks for \textbf{答案：} markers
//     (fallback for non-enumerate formats like pattern D).
//
// Returns warnings when the extracted count does not match problemCount.
func ExtractAnswers(blocks []Block, problemCount int) ([]AnswerEntry, []string) {
	var warnings []string

	entries := extractFromEnumerate(blocks)

	if len(entries) == 0 {
		entries = extractDirectly(blocks)
	}

	if problemCount > 0 && len(entries) != problemCount {
		warnings = append(warnings,
			fmt.Sprintf("answer count mismatch: expected %d problems, found %d answer entries",
				problemCount, len(entries)))
	}

	if len(entries) == 0 {
		warnings = append(warnings, "no answer entries found in blocks")
	}

	return entries, warnings
}

func extractFromEnumerate(blocks []Block) []AnswerEntry {
	type enumRange struct{ start, end int }
	var ranges []enumRange
	var stack []int

	for i, b := range blocks {
		switch {
		case b.Type == BlockEnvBegin && b.EnvName == "enumerate":
			stack = append(stack, i)
		case b.Type == BlockEnvEnd && b.EnvName == "enumerate" && len(stack) > 0:
			start := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				ranges = append(ranges, enumRange{start, i})
			}
		}
	}

	var entries []AnswerEntry
	for _, r := range ranges {
		items := collectEnumerateItems(blocks[r.start+1 : r.end])
		for _, itemBlocks := range items {
			entry := parseItemContent(itemBlocks)
			entry.Index = len(entries)
			entries = append(entries, entry)
		}
	}
	return entries
}

func collectEnumerateItems(blocks []Block) [][]Block {
	var items [][]Block
	var current []Block
	enumDepth := 0

	for _, b := range blocks {
		if b.Type == BlockEnvBegin && b.EnvName == "enumerate" {
			enumDepth++
		}
		if b.Type == BlockEnvEnd && b.EnvName == "enumerate" {
			enumDepth--
		}

		if b.Type == BlockItem && enumDepth == 0 {
			if len(current) > 0 {
				items = append(items, current)
			}
			current = []Block{b}
			continue
		}
		if len(current) > 0 {
			current = append(current, b)
		}
	}
	if len(current) > 0 {
		items = append(items, current)
	}
	return items
}

func parseItemContent(itemBlocks []Block) AnswerEntry {
	var answerParts, solutionParts []string
	currentSection := ""

	for _, b := range itemBlocks {
		content := b.Content

		answerIdx := indexAfterMarker(content, answerMarker)
		bracketAnswerIdx := indexAfterMarker(content, answerMarkerBracket)
		if bracketAnswerIdx >= 0 && (answerIdx < 0 || bracketAnswerIdx < answerIdx) {
			answerIdx = bracketAnswerIdx
		}
		solutionIdx := indexAfterMarker(content, solutionMarker)
		bracketSolutionIdx := indexAfterMarker(content, solutionMarkerBracket)
		if bracketSolutionIdx >= 0 && (solutionIdx < 0 || bracketSolutionIdx < solutionIdx) {
			solutionIdx = bracketSolutionIdx
		}
		altSolutionIdx := indexAfterMarker(content, solutionMarkerAlt)

		if answerIdx >= 0 && (solutionIdx < 0 || answerIdx < solutionIdx) && (altSolutionIdx < 0 || answerIdx < altSolutionIdx) {
			currentSection = "answer"
			if solutionIdx >= 0 {
				answerParts = append(answerParts, strings.TrimSpace(content[answerIdx:solutionIdx]))
				currentSection = "solution"
				solutionParts = append(solutionParts, strings.TrimSpace(content[solutionIdx:]))
			} else if altSolutionIdx >= 0 {
				answerParts = append(answerParts, strings.TrimSpace(content[answerIdx:altSolutionIdx]))
				currentSection = "solution"
				solutionParts = append(solutionParts, strings.TrimSpace(content[altSolutionIdx:]))
			} else {
				answerParts = append(answerParts, strings.TrimSpace(content[answerIdx:]))
			}
			continue
		}

		if solutionIdx >= 0 && (altSolutionIdx < 0 || solutionIdx < altSolutionIdx) {
			currentSection = "solution"
			solutionParts = append(solutionParts, strings.TrimSpace(content[solutionIdx:]))
			continue
		}
		if altSolutionIdx >= 0 {
			currentSection = "solution"
			solutionParts = append(solutionParts, strings.TrimSpace(content[altSolutionIdx:]))
			continue
		}

		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			continue
		}
		switch currentSection {
		case "answer":
			answerParts = append(answerParts, trimmed)
		case "solution":
			solutionParts = append(solutionParts, trimmed)
		}
	}

	return AnswerEntry{
		AnswerLatex:   joinCompact(answerParts),
		SolutionLatex: joinCompact(solutionParts),
	}
}

func extractDirectly(blocks []Block) []AnswerEntry {
	var entries []AnswerEntry
	var current *AnswerEntry
	currentSection := ""

	flushCurrent := func() {
		if current != nil {
			current.AnswerLatex = joinCompact(strings.Fields(current.AnswerLatex))
			current.SolutionLatex = joinCompact(strings.Fields(current.SolutionLatex))
			entries = append(entries, *current)
			current = nil
		}
	}

	for _, b := range blocks {
		if b.Type != BlockText && b.Type != BlockCommand {
			continue
		}

		content := b.Content

		answerIdx := indexAfterMarker(content, answerMarker)
		bracketAnswerIdx := indexAfterMarker(content, answerMarkerBracket)
		if bracketAnswerIdx >= 0 && (answerIdx < 0 || bracketAnswerIdx < answerIdx) {
			answerIdx = bracketAnswerIdx
		}
		solutionIdx := indexAfterMarker(content, solutionMarker)
		bracketSolutionIdx := indexAfterMarker(content, solutionMarkerBracket)
		if bracketSolutionIdx >= 0 && (solutionIdx < 0 || bracketSolutionIdx < solutionIdx) {
			solutionIdx = bracketSolutionIdx
		}
		altSolutionIdx := indexAfterMarker(content, solutionMarkerAlt)

		if answerIdx >= 0 && (solutionIdx < 0 || answerIdx < solutionIdx) && (altSolutionIdx < 0 || answerIdx < altSolutionIdx) {
			flushCurrent()
			current = &AnswerEntry{Index: len(entries)}
			if solutionIdx >= 0 {
				current.AnswerLatex = strings.TrimSpace(content[answerIdx:solutionIdx])
				currentSection = "solution"
				current.SolutionLatex = strings.TrimSpace(content[solutionIdx:])
			} else if altSolutionIdx >= 0 {
				current.AnswerLatex = strings.TrimSpace(content[answerIdx:altSolutionIdx])
				currentSection = "solution"
				current.SolutionLatex = strings.TrimSpace(content[altSolutionIdx:])
			} else {
				current.AnswerLatex = strings.TrimSpace(content[answerIdx:])
			}
			continue
		}

		if solutionIdx >= 0 && (altSolutionIdx < 0 || solutionIdx < altSolutionIdx) {
			if current == nil {
				current = &AnswerEntry{Index: len(entries)}
			}
			currentSection = "solution"
			if current.SolutionLatex != "" {
				current.SolutionLatex += "\n"
			}
			current.SolutionLatex += strings.TrimSpace(content[solutionIdx:])
			continue
		}
		if altSolutionIdx >= 0 {
			if current == nil {
				current = &AnswerEntry{Index: len(entries)}
			}
			currentSection = "solution"
			if current.SolutionLatex != "" {
				current.SolutionLatex += "\n"
			}
			current.SolutionLatex += strings.TrimSpace(content[altSolutionIdx:])
			continue
		}

		if current == nil {
			continue
		}
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			continue
		}
		switch currentSection {
		case "answer":
			if current.AnswerLatex != "" {
				current.AnswerLatex += "\n"
			}
			current.AnswerLatex += trimmed
		case "solution":
			if current.SolutionLatex != "" {
				current.SolutionLatex += "\n"
			}
			current.SolutionLatex += trimmed
		}
	}

	flushCurrent()
	return entries
}

func indexAfterMarker(s, marker string) int {
	idx := strings.Index(s, marker)
	if idx < 0 {
		return -1
	}
	return idx + len(marker)
}

func joinCompact(parts []string) string {
	var filtered []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}
