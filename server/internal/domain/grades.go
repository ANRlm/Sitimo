package domain

var StandardGrades = []string{
	"初一",
	"初二",
	"初三",
	"高一",
	"高二",
	"高三",
}

func BuildGradeOptions(values []string) []string {
	ordered := make([]string, 0, len(StandardGrades)+len(values))
	seen := make(map[string]struct{}, len(StandardGrades)+len(values))

	for _, grade := range StandardGrades {
		ordered = append(ordered, grade)
		seen[grade] = struct{}{}
	}

	for _, grade := range values {
		if grade == "" {
			continue
		}
		if _, exists := seen[grade]; exists {
			continue
		}
		ordered = append(ordered, grade)
		seen[grade] = struct{}{}
	}

	return ordered
}
