package tokenizer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountTokens(t *testing.T) {
	// 测试有编码器的情况
	t.Run("With encoder", func(t *testing.T) {
		tokenCounter, err := NewTokenCounter()
		assert.NoError(t, err)

		testCases := []struct {
			name     string
			input    string
			expected int
		}{
			{
				name:     "Empty string",
				input:    "",
				expected: 0,
			},
			{
				name:     "English text",
				input:    "Hello world",
				expected: len(tokenCounter.encoder.Encode("Hello world", nil, nil)),
			},
			{
				name:     "Special chars",
				input:    "!@#$%^&*()",
				expected: len(tokenCounter.encoder.Encode("!@#$%^&*()", nil, nil)),
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				count := tokenCounter.CountTokens(testCase.input)
				assert.Equal(t, testCase.expected, count)
			})
		}
	})

	// 测试没有编码器的回退情况
	t.Run("Fallback without encoder", func(t *testing.T) {
		tokenCounter := &TokenCounter{encoder: nil}

		testCases := []struct {
			name     string
			input    string
			expected int
		}{
			{
				name:     "Empty string fallback",
				input:    "",
				expected: 0,
			},
			{
				name:     "English text fallback",
				input:    "Hello world", // 2 words
				expected: 2 * 4 / 3,     // 2 words * 4/3 => 2
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				count := tokenCounter.CountTokens(testCase.input)
				assert.Equal(t, testCase.expected, count)
			})
		}
	})
}

func TestCountTokensCustomer(t *testing.T) {
	t.Run("With encoder", func(t *testing.T) {
		tokenCounter, err := NewTokenCounter()
		assert.NoError(t, err)

		testCases := []struct {
			name     string
			input    string
			expected int
		}{
			{
				name:     "custom string",
				input:    "You are shenma, a highly skilled software engineer with extensive knowledge in many programming languages, frameworks, design patterns, and best practices.\n\n====\n\nMARKDOWN RULES\n\nALL responses MUST show ANY `language construct` OR filename reterence as clickable, exactly as [`filename OR language.declaration()`](relative/file/path.ext:line); line is required for `syntax` and optional for filename links. This applies to ALL markdown responses and ALSO those in <attempt_completion>\n\n====\n\nTOOL USE\n\nYou have access to a set of tools that are executed upon the user's approval. You can use one tool per message, and will receive the result of that tool use in the user's response. You use tools step-by-step to accomplish a given task, with each tool use informed by the result of the previous tool use.\n\n# Tool Use Formatting\n\nTool uses are formatted using XML-style tags. The tool name itself becomes the XML tag name. Each parameter is enclosed within its own set of tags. Here's the structure:\n\n<actual_tool_name>\n<parameter1_name>value1</parameter1_name>\n<parameter2_name>value2</parameter2_name>\n...\n</actual_tool_name>\n\nFor example, to use the read_file tool:\n\n<read_file>\n<path>src/main.js</path>\n</read_file>\n\nAlways use the actual tool name as the XML tag name for proper parsing and execution.\n\n# Tools\n\n## read_file\nDescription: Request to read the contents of a file at the specified path. Use this when you need to examine the contents of an existing file you do not know the contents of, for example to analyze code, review text files, or extract information from configuration files. The output includes line numbers prefixed to each line (e.g. \"1 | const x = 1\"), making it easier to reference specific lines when creating diffs or discussing code. By specifying start_line and end_line parameters, you can efficiently read specific portions of large files without loading the entire file into memory. Automatically extracts raw text from PDF and DOCX files. May not be suitable for other types of binary files, as it returns the raw content as a string.\nParameters:\n- path: (required) The path of the file to read (relative to the current workspace directory d:\\codespace\\tsproject\\zgsm)\n- start_line: (optional) The starting line number to read from (1-based). If not provided, it starts from the beginning of the file.\n- end_line: (optional) The ending line number to read to (1-based, inclusive). If not provided, it reads to the end of the file.\nUsage:\n<read_file>\n<path>File path here</path>\n<start_line>Starting line number (optional)</start_line>\n<end_line>Ending line number (optional)</end_line>\n</read_file>\n\nExamples:\n\n1. Reading an entire file:\n<read_file>\n<path>frontend-config.json</path>\n</read_file>\n\n2. Reading the first 1000 lines of a large log file:\n<read_file>\n<path>logs/application.log</path>\n<end_line>1000</end_line>\n</read_file>\n\n3. Reading lines 500-1000 of a CSV file:\n<read_file>\n<path>data/large-dataset.csv</path>\n<start_line>500</start_line>\n<end_line>1000</end_line>\n</read_file>\n\n4. Reading a specific function in a source file:\n<read_file>\n<path>src/app.ts</path>\n<start_line>46</start_line>\n<end_line>68</end_line>\n</read_file>\n\nNote: When both start_line and end_line are provided, this tool efficiently streams only the requested lines, making it suitable for processing large files like logs, CSV files, and other large datasets without memory issues.\n\n## fetch_instructions\nDescription: Request to fetch instructions to perform a task\nParameters:\n- task: (required) The task to get instructions for.  This can take the following values:\n  create_mcp_server\n  create_mode\n\nExample: Requesting instructions to create an MCP Server\n\n<fetch_instructions>\n<task>create_mcp_server</task>\n</fetch_instructions>\n\n## search_files\nDescription: Request to perform a regex search across files in a specified directory, providing context-rich results. This tool searches for patterns or specific content across multiple files, displaying each match with encapsulating context.\nParameters:\n- path: (required) The path of the directory to search in (relative to the current workspace directory d:\\codespace\\tsproject\\zgsm). This directory will be recursively searched.\n- regex: (required) The regular expression pattern to search for. Uses Rust regex syntax.\n- file_pattern: (optional) Glob pattern to filter files (e.g., '*.ts' for TypeScript files). If not provided, it will search all files (*).\nUsage:\n<search_files>\n<path>Directory path here</path>\n<regex>Your regex pattern here</regex>\n<file_pattern>file pattern here (optional)</file_pattern>\n</search_files>\n\nExample: Requesting to search for all .ts files in the current directory\n<search_files>\n<path>.</path>\n<regex>.*</regex>\n<file_pattern>*.ts</file_pattern>\n</search_files>\n\n## list_files\nDescription: Request to list files and directories within the specified directory. If recursive is true, it will list all files and directories recursively. If recursive is false or not provided, it will only list the top-level contents. Do not use this tool to confirm the existence of files you may have created, as the user will let you know if the files were created successfully or not.\nParameters:\n- path: (required) The path of the directory to list contents for (relative to the current workspace directory d:\\codespace\\tsproject\\zgsm)\n- recursive: (optional) Whether to list files recursively. Use true for recursive listing, false or omit for top-level only.\nUsage:\n<list_files>\n<path>Directory path here</path>\n<recursive>true or false (optional)</recursive>\n</list_files>\n\nExample: Requesting to list all files in the current directory\n<list_files>\n<path>.",
				expected: 1277,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				count := tokenCounter.CountTokens(testCase.input)
				fmt.Println("count", count)
				assert.Equal(t, testCase.expected, count)
			})
		}
	})
}
