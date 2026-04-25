package parser

import (
	"regexp"
	"strings"
)

var exampleMarkerRe = regexp.MustCompile(`\\textbf\{例\d+`)

// ParseTextMarkers extracts problems from text-marker-based files
// where problems are indicated by \textbf{例N.} markers.
func ParseTextMarkers(blocks []Block) []ProblemBlock {
	var problems []ProblemBlock
	var currentProblem *ProblemBlock

	for _, block := range blocks {
		if block.Type == BlockEnvBegin || block.Type == BlockEnvEnd {
			if currentProblem != nil {
				currentProblem.Body = appendBody(currentProblem.Body, block.Content)
			}
			continue
		}

		if exampleMarkerRe.MatchString(block.Content) {
			if currentProblem != nil && strings.TrimSpace(currentProblem.Body) != "" {
				problems = append(problems, *currentProblem)
			}
			currentProblem = &ProblemBlock{
				Body:      block.Content,
				LineStart: block.LineStart,
				Pattern:   PatternE,
			}
			continue
		}

		if currentProblem != nil {
			currentProblem.Body = appendBody(currentProblem.Body, block.Content)
		}
	}

	if currentProblem != nil && strings.TrimSpace(currentProblem.Body) != "" {
		problems = append(problems, *currentProblem)
	}

	return problems
}
