package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// CountLines  is a helper to count lines in a byte slice.
// Reuses the existing logic.
func CountLines(data []byte) int {
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	// Add one for the last line if it doesn't end with a newline
	if len(data) > 0 && data[len(data)-1] != '\n' {
		lines++
	}
	// If the content is empty, there are 0 lines. If it's not empty but has no newline, it's 1 line.
	if len(data) == 0 {
		return 0
	}
	if lines == 0 && len(data) > 0 {
		return 1
	} // Fix: handle non-empty single line
	if lines == 0 && len(data) == 0 {
		return 0
	} // Explicitly handle empty
	return lines
}

// SplitLines splits a string into lines (preserving line endings)
func SplitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// JoinLines joins lines into a single string with \n
func JoinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// DecodeBase64 decode Base64
func DecodeBase64(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded, fmt.Errorf("decode Base64 error: %w", err)
	}
	return string(decoded), nil
}
