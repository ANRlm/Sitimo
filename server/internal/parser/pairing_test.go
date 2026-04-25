package parser

import (
	"testing"
)

func TestPairAnswerExactMatch(t *testing.T) {
	files := []string{
		"数列求和 韩靖劼 .tex",
		"数列求和 配套解析 韩靖劼.tex",
		"边缘文件.tex",
	}
	result, found, warnings := PairAnswerFile("数列求和 韩靖劼 .tex", files)
	if !found {
		t.Fatal("expected match, got not found")
	}
	if result != "数列求和 配套解析 韩靖劼.tex" {
		t.Errorf("expected %q, got %q", "数列求和 配套解析 韩靖劼.tex", result)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestPairAnswerNoMatch(t *testing.T) {
	files := []string{"向量数量积.tex", "其它文件.tex"}
	_, found, _ := PairAnswerFile("向量数量积.tex", files)
	if found {
		t.Error("expected no match for unpaired file")
	}
}

func TestPairAnswerFuzzyMatch(t *testing.T) {
	files := []string{
		"数列求和 配套解析 韩靖劼.tex",
	}
	result, found, _ := PairAnswerFile("数列求和 韩靖劼 .tex", files)
	if !found {
		t.Fatal("expected fuzzy match, got not found")
	}
	if result != "数列求和 配套解析 韩靖劼.tex" {
		t.Errorf("expected %q, got %q", "数列求和 配套解析 韩靖劼.tex", result)
	}
}

func TestPairAnswerMultipleCandidates(t *testing.T) {
	files := []string{
		"数列求和 配套解析 韩靖劼.tex",
		"数列求和 配套解析 修订版.tex",
	}
	_, found, warnings := PairAnswerFile("数列求和 韩靖劼 .tex", files)
	if !found {
		t.Fatal("expected at least one match")
	}
	if len(warnings) == 0 {
		t.Error("expected warning about multiple candidates")
	}
}

func TestPairAnswerEmptyFileList(t *testing.T) {
	_, found, _ := PairAnswerFile("test.tex", []string{})
	if found {
		t.Error("expected no match with empty list")
	}
}

func TestLevenshtein(t *testing.T) {
	if d := levenshtein("kitten", "sitting"); d != 3 {
		t.Errorf("kitten/sitting: expected 3, got %d", d)
	}
	if d := levenshtein("abc", "abc"); d != 0 {
		t.Errorf("abc/abc: expected 0, got %d", d)
	}
	if d := levenshtein("", "abc"); d != 3 {
		t.Errorf("empty/abc: expected 3, got %d", d)
	}
	if d := levenshtein("abc", ""); d != 3 {
		t.Errorf("abc/empty: expected 3, got %d", d)
	}
	if d := levenshtein("", ""); d != 0 {
		t.Errorf("empty/empty: expected 0, got %d", d)
	}
}

func TestPairAnswerNoTopicAuthor(t *testing.T) {
	files := []string{
		"向量数量积 配套解析 韩靖劼.tex",
	}
	_, found, _ := PairAnswerFile("纯题目.tex", files)
	if found {
		t.Error("expected no match when problem filename lacks author marker")
	}
}

func TestPairAnswerUniqueLoneCandidate(t *testing.T) {
	// Problem topic differs enough that neither exact nor fuzzy match succeeds,
	// so the unique-candidate fallback must kick in.
	files := []string{
		"特殊题型 配套解析 韩靖劼.tex",
	}
	_, found, warnings := PairAnswerFile("完全不同话题 韩靖劼.tex", files)
	if !found {
		t.Fatal("expected unique lone candidate to match")
	}
	if len(warnings) == 0 {
		t.Error("expected warning for lone candidate match")
	}
}

func TestPairAnswerExactSpaceVariant(t *testing.T) {
	files := []string{
		"测验 配套解析 韩靖劼 .tex",
	}
	result, found, _ := PairAnswerFile("测验 韩靖劼.tex", files)
	if !found || result != "测验 配套解析 韩靖劼 .tex" {
		t.Errorf("expected space-variant match, got found=%v result=%q", found, result)
	}
}

func TestExtractTopic(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"数列求和 韩靖劼 .tex", "数列求和"},
		{"等比数列的性质 韩靖劼.tex", "等比数列的性质"},
		{"向量数量积.tex", ""},
		{"韩靖劼 数列求和.tex", ""},
		{"trapezoid area 韩靖劼.tex", "trapezoid area"},
		{"  数列求和 韩靖劼 .tex", "数列求和"},
	}
	for _, tc := range tests {
		got := extractTopic(tc.filename)
		if got != tc.expected {
			t.Errorf("extractTopic(%q): expected %q, got %q", tc.filename, tc.expected, got)
		}
	}
}

func TestExtractTopicFromAnswer(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"数列求和 配套解析 韩靖劼.tex", "数列求和"},
		{"等比数列的性质 配套解析与答案 韩靖劼 .tex", "等比数列的性质"},
		{"孤立文件.tex", ""},
		{"配套解析 韩靖劼.tex", ""},
	}
	for _, tc := range tests {
		got := extractTopicFromAnswer(tc.filename)
		if got != tc.expected {
			t.Errorf("extractTopicFromAnswer(%q): expected %q, got %q", tc.filename, tc.expected, got)
		}
	}
}
