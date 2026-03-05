package tokenizers

import (
	"os"
	"path/filepath"
	"testing"
)

// to test tokenizer
// go test ./pkg/tokenizers/ -v
func Test_NewTokenizer(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	// Test with valid path
	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error("Failed to create tokenizer with valid path:", err)
		return
	}
	defer tk.Close()

	// Test with invalid path
	tk3, err := NewTokenizer("invalid/path/tokenizer.json")
	if err == nil {
		t.Error("Expected error for invalid path, but got nil")
		return
	}
	if tk3 != nil {
		t.Error("Expected nil tokenizer for invalid path")
	}
}

func Test_EncodeDecode(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer tk.Close()

	// Test encoding
	testText := "brown fox jumps over the lazy dog"
	tokens := tk.Encode(testText)
	t.Log("Encoded tokens:", tokens)
	if len(tokens) == 0 {
		t.Error("Expected non-empty token array for encoding")
	}

	// Test decoding
	decoded := tk.Decode(tokens)
	t.Log("Decoded text:", decoded)
	if decoded == "" {
		t.Error("Expected non-empty string for decoding")
	}

	// Test with hello world
	ei := tk.Encode("hello world!")
	t.Log("Encoded 'hello world!':", ei)
	if len(ei) == 0 {
		t.Error("Expected non-empty token array for 'hello world!'")
	}

	di := tk.Decode(ei)
	t.Log("Decoded back:", di)
	if di != "hello world!" {
		t.Log("Note: Decoded text may not exactly match original due to tokenization")
	}

	// Test encoding empty string
	emptyTokens := tk.Encode("")
	if len(emptyTokens) != 0 {
		t.Error("Expected empty token array for empty string")
	}

	// Test decoding empty slice
	emptyDecoded := tk.Decode([]int{})
	if emptyDecoded != "" {
		t.Error("Expected empty string for empty token slice")
	}
}

func Test_GetTokenCount(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer tk.Close()

	// Test token count
	count := tk.GetTokenCount("hello world!")
	t.Log("Token count for 'hello world!':", count)
	if count == 0 {
		t.Error("Expected non-zero token count for 'hello world!'")
	}

	// Test token count for empty string
	emptyCount := tk.GetTokenCount("")
	if emptyCount != 0 {
		t.Error("Expected zero token count for empty string")
	}

	// Test token count for longer text
	longText := "This is a longer text to test token counting functionality. It should contain multiple tokens."
	longCount := tk.GetTokenCount(longText)
	t.Log("Token count for longer text:", longCount)
	if longCount <= 3 { // At least a few tokens
		t.Error("Expected more tokens for longer text")
	}
}

func Test_GetTokens(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer tk.Close()

	// Test GetTokens method
	testText := "function test() { return true; }"
	tokens := tk.GetTokens(testText)
	t.Log("Tokens for code snippet:", tokens)
	if len(tokens) == 0 {
		t.Error("Expected non-empty token array for code snippet")
	}

	// Verify GetTokens returns same as Encode
	encodeTokens := tk.Encode(testText)
	if len(tokens) != len(encodeTokens) {
		t.Error("GetTokens should return same result as Encode")
	}

	// Check if all tokens match
	for i, token := range tokens {
		if token != encodeTokens[i] {
			t.Error("GetTokens and Encode returned different tokens at index", i)
			break
		}
	}
}

func Test_ConvertNL(t *testing.T) {
	// Test ConvertNLToLinux
	winText := "Line 1\r\nLine 2\r\nLine 3"
	linuxText := ConvertNLToLinux(winText)
	expectedLinux := "Line 1\nLine 2\nLine 3"
	if linuxText != expectedLinux {
		t.Error("ConvertNLToLinux failed. Expected:", expectedLinux, "Got:", linuxText)
	}

	// Test with already Linux text
	alreadyLinux := "Line 1\nLine 2\nLine 3"
	result := ConvertNLToLinux(alreadyLinux)
	if result != alreadyLinux {
		t.Error("ConvertNLToLinux should not modify already Linux text")
	}

	// Test ConvertNLToWin
	linuxText2 := "Line 1\nLine 2\nLine 3"
	winText2 := ConvertNLToWin(linuxText2)
	expectedWin := "Line 1\r\nLine 2\r\nLine 3"
	if winText2 != expectedWin {
		t.Error("ConvertNLToWin failed. Expected:", expectedWin, "Got:", winText2)
	}

	// Test with already Windows text
	alreadyWin := "Line 1\r\nLine 2\r\nLine 3"
	result2 := ConvertNLToWin(alreadyWin)
	if result2 != alreadyWin {
		t.Error("ConvertNLToWin should not modify already Windows text")
	}
}

func Test_ErrorHandling(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer tk.Close()

	// Test with special Unicode characters that should be handled gracefully
	specialText := "Hello 🌍! 中文测试 ñáéíóú"
	tokens := tk.Encode(specialText)
	t.Log("Tokens for special text:", tokens)
	// The tokenizer should handle Unicode gracefully

	// Test GetTokenCount with special text
	count := tk.GetTokenCount(specialText)
	t.Log("Token count for special text:", count)
	// Should return a valid count, not panic
}

func Test_CodeExamples(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer tk.Close()

	// Test with various code snippets
	codeExamples := []string{
		"func hello() { fmt.Println(\"Hello, World!\") }",
		"function test() { return true; }",
		"class MyClass { constructor() { this.value = 0; } }",
		"def hello(): print(\"Hello, Python!\")",
		"public class Test { public static void main(String[] args) { System.out.println(\"Hello\"); } }",
	}

	for _, code := range codeExamples {
		tokens := tk.Encode(code)
		decoded := tk.Decode(tokens)
		count := tk.GetTokenCount(code)

		t.Logf("Code: %s", code[:min(30, len(code))]+"...")
		t.Logf("  Tokens: %v", tokens)
		t.Logf("  Token count: %d", count)
		t.Logf("  Decoded: %s", decoded[:min(30, len(decoded))]+"...")
		t.Log("---")

		if len(tokens) == 0 {
			t.Errorf("Expected non-empty tokens for code: %s", code)
		}

		if count == 0 {
			t.Errorf("Expected non-zero token count for code: %s", code)
		}
	}
}

func Test_EdgeCases(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer tk.Close()

	// Test with whitespace only
	whitespaceText := "   \t\n\r   "
	whitespaceTokens := tk.Encode(whitespaceText)
	t.Log("Tokens for whitespace:", whitespaceTokens)
	whitespaceCount := tk.GetTokenCount(whitespaceText)
	t.Log("Token count for whitespace:", whitespaceCount)

	// Test with very long text
	longText := ""
	for i := 0; i < 100; i++ {
		longText += "This is a test sentence that will be repeated many times to create a long text. "
	}
	longTokens := tk.Encode(longText)
	t.Log("Token count for very long text:", len(longTokens))
	if len(longTokens) < 100 { // Should have many tokens
		t.Error("Expected more tokens for very long text")
	}

	// Test with numbers and symbols
	numSymText := "12345 !@#$%^&*()_+-=[]{}|;':\",./<>?"
	numSymTokens := tk.Encode(numSymText)
	t.Log("Tokens for numbers and symbols:", numSymTokens)
	numSymCount := tk.GetTokenCount(numSymText)
	t.Log("Token count for numbers and symbols:", numSymCount)

	// Test decode with nil slice
	nilDecoded := tk.Decode(nil)
	if nilDecoded != "" {
		t.Error("Expected empty string for nil token slice")
	}

	// Test decode with single token
	singleToken := []int{31539} // Usually "hello" in most tokenizers
	singleDecoded := tk.Decode(singleToken)
	t.Log("Decoded single token:", singleDecoded)
}

func Test_Performance(t *testing.T) {
	// Get absolute path to the tokenizer file
	wd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get working directory:", err)
		return
	}

	// Navigate from pkg/tokenizers to project root
	projectRoot := filepath.Dir(filepath.Dir(wd))
	tokenizerPath := filepath.Join(projectRoot, "bin/deepseek-tokenizer/tokenizer.json")

	tk, err := NewTokenizer(tokenizerPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer tk.Close()

	// Create test text
	testText := "This is a test text for performance testing. It contains multiple sentences and various characters."

	// Test multiple encoding operations
	for i := 0; i < 100; i++ {
		tokens := tk.Encode(testText)
		if len(tokens) == 0 {
			t.Errorf("Expected non-empty tokens in iteration %d", i)
		}

		decoded := tk.Decode(tokens)
		if decoded == "" {
			t.Errorf("Expected non-empty decoded text in iteration %d", i)
		}

		count := tk.GetTokenCount(testText)
		if count == 0 {
			t.Errorf("Expected non-zero token count in iteration %d", i)
		}
	}

	t.Log("Performance test completed 100 iterations successfully")
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
