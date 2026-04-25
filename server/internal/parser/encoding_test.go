package parser

import (
	"testing"
)

func TestDetectUtf8(t *testing.T) {
	data := []byte("Hello 世界")
	enc := DetectEncoding(data)
	if enc != "UTF-8" {
		t.Errorf("Expected UTF-8, got %s", enc)
	}
}

func TestDetectUtf8BOM(t *testing.T) {
	data := []byte{0xEF, 0xBB, 0xBF, 'a', 'b', 'c'}
	enc := DetectEncoding(data)
	if enc != "UTF-8" {
		t.Errorf("Expected UTF-8 (with BOM), got %s", enc)
	}
}

func TestDecodeUtf8BOM(t *testing.T) {
	data := []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'}
	decoded, enc, err := DecodeContent(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "hello" {
		t.Errorf("Expected 'hello' (BOM stripped), got %q", decoded)
	}
	if enc != "UTF-8" {
		t.Errorf("Expected UTF-8, got %s", enc)
	}
}

func TestDecodeASCII(t *testing.T) {
	data := []byte("pure ascii text with numbers 123")
	decoded, enc, err := DecodeContent(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != string(data) {
		t.Errorf("ASCII round-trip failed")
	}
	if enc != "UTF-8" {
		t.Errorf("Expected UTF-8 for ASCII, got %s", enc)
	}
}

func TestDecodeEmpty(t *testing.T) {
	decoded, enc, err := DecodeContent([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "" || enc != "UTF-8" {
		t.Errorf("Empty input should return empty string, got %q (enc=%s)", decoded, enc)
	}
}

func TestDecodeGBK(t *testing.T) {
	// "你好" in GBK encoding
	gbkData := []byte{0xC4, 0xE3, 0xBA, 0xC3}
	decoded, enc, err := DecodeContent(gbkData)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "你好" {
		t.Errorf("Expected '你好', got %q", decoded)
	}
	if enc != "GBK" {
		t.Errorf("Expected GBK, got %s", enc)
	}
}

func TestDetectGBK(t *testing.T) {
	gbkData := []byte{0xC4, 0xE3, 0xBA, 0xC3}
	enc := DetectEncoding(gbkData)
	if enc != "GBK" {
		t.Errorf("Expected GBK, got %s", enc)
	}
}

func TestDetectEmpty(t *testing.T) {
	enc := DetectEncoding([]byte{})
	if enc != "UTF-8" {
		t.Errorf("Expected UTF-8 for empty input, got %s", enc)
	}
}

func TestDetectBOMWithGBK(t *testing.T) {
	// BOM (EF BB BF) followed by GBK-encoded "你好"
	data := []byte{0xEF, 0xBB, 0xBF, 0xC4, 0xE3, 0xBA, 0xC3}
	enc := DetectEncoding(data)
	if enc != "GBK" {
		t.Errorf("Expected GBK (BOM with GBK content), got %s", enc)
	}
}
