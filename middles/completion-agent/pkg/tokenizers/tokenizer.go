package tokenizers

import (
	"fmt"
	"os"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
)

// Tokenizer wraps sugarme/tokenizer library, providing a unified interface
type Tokenizer struct {
	tokenizer *tokenizer.Tokenizer
}

// NewTokenizer creates a new tokenizer instance
func NewTokenizer(tokenizerPath string) (*Tokenizer, error) {
	// Use the DeepSeek tokenizer file from bin/deepseek-tokenizer
	if tokenizerPath == "" {
		tokenizerPath = "bin/deepseek-tokenizer/tokenizer.json"
	}

	// Check if the file exists
	if _, err := os.Stat(tokenizerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("deepseek tokenizer file not found: %s", tokenizerPath)
	}

	// Load the tokenizer from the file using pretrained.FromFile
	t, err := pretrained.FromFile(tokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer from file: %s, error: %v", tokenizerPath, err)
	}

	return &Tokenizer{
		tokenizer: t,
	}, nil
}

// Encode encodes text into token IDs
func (t *Tokenizer) Encode(text string) []int {
	// Use EncodeSingle to encode the text
	encoding, err := t.tokenizer.EncodeSingle(text, true)
	if err != nil {
		// Return empty slice on error
		return []int{}
	}

	// Get the token IDs from the encoding
	return encoding.GetIds()
}

// Decode decodes token IDs back to text
func (t *Tokenizer) Decode(ids []int) string {
	return t.tokenizer.Decode(ids, true)
}

// GetTokenCount gets the token count for the given text
func (t *Tokenizer) GetTokenCount(text string) int {
	// Use EncodeSingle to encode the text and get the count
	encoding, err := t.tokenizer.EncodeSingle(text, true)
	if err != nil {
		// Return 0 on error
		return 0
	}

	return len(encoding.GetIds())
}

// GetTokens gets the token list for the given text
func (t *Tokenizer) GetTokens(text string) []int {
	return t.Encode(text)
}

// ConvertNLToLinux converts Windows newlines to Linux newlines
func ConvertNLToLinux(s string) string {
	// Replace Windows CRLF with Linux LF
	result := s
	for i := 0; i < len(result); i++ {
		if result[i] == '\r' && i+1 < len(result) && result[i+1] == '\n' {
			result = result[:i] + result[i+1:]
		}
	}
	return result
}

// ConvertNLToWin converts Linux newlines to Windows newlines
func ConvertNLToWin(s string) string {
	// Replace Linux LF with Windows CRLF
	result := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			// Check if it's not already a CRLF
			if i == 0 || s[i-1] != '\r' {
				result += "\r\n"
			} else {
				result += "\n"
			}
		} else {
			result += string(s[i])
		}
	}
	return result
}

// Close releases resources
func (t *Tokenizer) Close() {
	// sugarme/tokenizer doesn't require explicit resource cleanup
}
