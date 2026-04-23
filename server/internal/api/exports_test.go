package api

import (
	"strings"
	"testing"

	"mathlib/server/internal/domain"
)

func TestExportDownloadFilename(t *testing.T) {
	job := &domain.ExportJob{
		PaperTitle: "2024 年高三数学期中考试",
		Format:     domain.ExportFormatLatex,
		Variant:    domain.ExportVariantAnswer,
	}

	if got := exportDownloadFilename(job); got != "2024 年高三数学期中考试-answer.zip" {
		t.Fatalf("unexpected download filename: %q", got)
	}
}

func TestExportAttachmentDisposition(t *testing.T) {
	disposition := exportAttachmentDisposition("2024 年高三数学期中考试-answer.zip")

	if !strings.HasPrefix(disposition, "attachment; ") {
		t.Fatalf("expected attachment disposition, got %q", disposition)
	}
	if !strings.Contains(disposition, `filename="2024-answer.zip"`) {
		t.Fatalf("expected ascii fallback filename, got %q", disposition)
	}
	if !strings.Contains(disposition, "filename*=UTF-8''2024%20%E5%B9%B4%E9%AB%98%E4%B8%89%E6%95%B0%E5%AD%A6%E6%9C%9F%E4%B8%AD%E8%80%83%E8%AF%95-answer.zip") {
		t.Fatalf("expected utf-8 filename, got %q", disposition)
	}
}

func TestExportDownloadFilenameFallback(t *testing.T) {
	job := &domain.ExportJob{
		PaperTitle: "   ",
		Format:     domain.ExportFormatPDF,
		Variant:    domain.ExportVariantStudent,
	}

	if got := exportDownloadFilename(job); got != "mathlib-export-student.pdf" {
		t.Fatalf("unexpected fallback filename: %q", got)
	}
}
