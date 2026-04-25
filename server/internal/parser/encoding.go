package parser

import (
	"bytes"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// DetectEncoding identifies the character encoding of the given byte data.
// Returns "UTF-8", "GBK", or "unknown".
func DetectEncoding(data []byte) string {
	if len(data) == 0 {
		return "UTF-8"
	}

	if hasBOM(data) {
		stripped := stripBOM(data)
		if utf8.Valid(stripped) {
			return "UTF-8"
		}
		if isGBK(stripped) {
			return "GBK"
		}
		return "unknown"
	}

	if utf8.Valid(data) {
		return "UTF-8"
	}

	if isGBK(data) {
		return "GBK"
	}

	return "unknown"
}

// DecodeContent decodes byte content to a UTF-8 string.
// Returns: (decoded string, encoding name, error).
//   - If the data is valid UTF-8 (with or without BOM), the BOM is stripped
//     and the data is returned as a UTF-8 string.
//   - If the data is not valid UTF-8, a GBK decode is attempted.
//   - An empty slice returns ("", "UTF-8", nil).
func DecodeContent(data []byte) (string, string, error) {
	if len(data) == 0 {
		return "", "UTF-8", nil
	}

	clean := stripBOM(data)

	if utf8.Valid(clean) {
		return string(clean), "UTF-8", nil
	}

	decoded, err := decodeGBK(clean)
	if err != nil {
		return "", "unknown", err
	}
	return decoded, "GBK", nil
}

func hasBOM(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) ||
		bytes.HasPrefix(data, []byte{0xFE, 0xFF}) ||
		bytes.HasPrefix(data, []byte{0xFF, 0xFE})
}

func stripBOM(data []byte) []byte {
	switch {
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}):
		return data[3:]
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		return data[2:]
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}):
		return data[2:]
	default:
		return data
	}
}

func isGBK(data []byte) bool {
	_, err := decodeGBK(data)
	return err == nil
}

func decodeGBK(data []byte) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder()
	decoded, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
