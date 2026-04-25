package parser

import (
	"testing"
)

func TestExtractSectionTags(t *testing.T) {
	blocks := []Block{
		{Type: BlockCommand, Content: "\\section*{一、裂项相消法求和}"},
		{Type: BlockCommand, Content: "\\subsection*{1. 基本公式}"},
		{Type: BlockCommand, Content: "\\section*{二、错位相减法求和}"},
	}
	tags := ExtractSectionTags(blocks)
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d: %v", len(tags), tags)
	}
	expected := []string{"一、裂项相消法求和", "1. 基本公式", "二、错位相减法求和"}
	for i, tag := range tags {
		if tag != expected[i] {
			t.Errorf("Expected tag[%d]=%q, got %q", i, expected[i], tag)
		}
	}
}

func TestExtractSectionTagsWithMybox(t *testing.T) {
	blocks := []Block{
		{Type: BlockEnvBegin, EnvName: "mybox", EnvArgs: "单选题：集合的交集运算"},
	}
	tags := ExtractSectionTags(blocks)
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d: %v", len(tags), tags)
	}
	if tags[0] != "单选题：集合的交集运算" {
		t.Errorf("Expected '单选题：集合的交集运算', got %q", tags[0])
	}
}

func TestExtractSectionTagsBlacklist(t *testing.T) {
	blocks := []Block{
		{Type: BlockCommand, Content: "\\section*{练习}"},
		{Type: BlockCommand, Content: "\\section*{实际内容}"},
	}
	tags := ExtractSectionTags(blocks)
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag (blacklist filtered), got %d: %v", len(tags), tags)
	}
	if tags[0] != "实际内容" {
		t.Errorf("Expected '实际内容', got %q", tags[0])
	}
}

func TestExtractSectionTagsEmpty(t *testing.T) {
	tags := ExtractSectionTags(nil)
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags for nil blocks, got %d", len(tags))
	}
}

func TestExtractSectionTagsDeduplicates(t *testing.T) {
	blocks := []Block{
		{Type: BlockCommand, Content: "\\section*{同标题}"},
		{Type: BlockCommand, Content: "\\section*{同标题}"},
	}
	tags := ExtractSectionTags(blocks)
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag (deduplicated), got %d: %v", len(tags), tags)
	}
}

func TestExtractSectionTagsNoBraces(t *testing.T) {
	blocks := []Block{
		{Type: BlockCommand, Content: "\\section*"},
	}
	tags := ExtractSectionTags(blocks)
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags for command without braces, got %d", len(tags))
	}
}
