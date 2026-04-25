package parser

import (
	"fmt"
	"strings"
)

// PairAnswerFile finds the answer key file matching a problem file.
// Returns: matched filename, whether found, any warnings.
func PairAnswerFile(problemFilename string, allFilenames []string) (string, bool, []string) {
	var warnings []string

	topic := extractTopic(problemFilename)
	if topic == "" {
		return "", false, nil
	}

	var candidates []string
	seen := make(map[string]bool)

	// 1. Exact match: construct candidate and check all filenames
	for _, suffix := range []string{" 配套解析 韩靖劼.tex", " 配套解析 韩靖劼 .tex"} {
		exact := topic + suffix
		for _, f := range allFilenames {
			if f == exact && !seen[f] {
				candidates = append(candidates, f)
				seen[f] = true
			}
		}
	}

	// 2. Fuzzy match: files containing "配套解析" where the topic part
	//    has Levenshtein distance ≤ 3 from the problem's topic.
	for _, f := range allFilenames {
		if !strings.Contains(f, "配套解析") || seen[f] {
			continue
		}
		answerTopic := extractTopicFromAnswer(f)
		if answerTopic == "" {
			continue
		}
		if levenshtein(topic, answerTopic) <= 3 {
			candidates = append(candidates, f)
			seen[f] = true
		}
	}

	// 3. Unique candidate fallback: only one file in the list contains "配套解析"
	if len(candidates) == 0 {
		var lone []string
		for _, f := range allFilenames {
			if strings.Contains(f, "配套解析") {
				lone = append(lone, f)
			}
		}
		if len(lone) == 1 {
			candidates = append(candidates, lone[0])
			warnings = append(warnings, "唯一包含配套解析的文件，假定为答案文件: "+lone[0])
		}
	}

	if len(candidates) == 0 {
		return "", false, nil
	}

	if len(candidates) > 1 {
		warnings = append(warnings,
			fmt.Sprintf("存在多个候选答案文件(%d个)，使用第一个: %s", len(candidates), candidates[0]))
	}

	return candidates[0], true, warnings
}

// extractTopic extracts the topic part from a problem filename
// by finding "韩靖劼" and taking everything before it.
func extractTopic(filename string) string {
	name := strings.TrimSuffix(filename, ".tex")
	name = strings.TrimSpace(name)

	idx := strings.Index(name, "韩靖劼")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(name[:idx])
}

// extractTopicFromAnswer extracts the topic part from an answer filename
// by finding "配套解析" and taking everything before it.
func extractTopicFromAnswer(filename string) string {
	name := strings.TrimSuffix(filename, ".tex")
	name = strings.TrimSpace(name)

	idx := strings.Index(name, "配套解析")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(name[:idx])
}

// levenshtein computes the Levenshtein distance between two strings,
// comparing Unicode characters (runes) rather than bytes.
func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	alen, blen := len(ar), len(br)

	if alen == 0 {
		return blen
	}
	if blen == 0 {
		return alen
	}

	// Use two-row optimisation to keep memory O(min(n,m)).
	prev := make([]int, blen+1)
	curr := make([]int, blen+1)

	for j := 0; j <= blen; j++ {
		prev[j] = j
	}

	for i := 1; i <= alen; i++ {
		curr[0] = i
		for j := 1; j <= blen; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = prev[j] + 1               // deletion
			if v := curr[j-1] + 1; v < curr[j] { // insertion
				curr[j] = v
			}
			if v := prev[j-1] + cost; v < curr[j] { // substitution
				curr[j] = v
			}
		}
		prev, curr = curr, prev
	}

	return prev[blen]
}
