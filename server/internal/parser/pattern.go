package parser

// Pattern type constants for identifying LaTeX problem bank structure styles.
const (
	PatternA = "A" // enumerate with "题\arabic*"
	PatternB = "B" // enumerate with "例\arabic*"
	PatternC = "C" // enumerate with "\arabic*"
	PatternD = "D" // mybox-based
	PatternE = "E" // text-marker-based (\textbf{例N.})
)
