package embedding

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zgsm-ai/codebase-indexer/internal/parser"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

func TestSplitOpenAPIFile(t *testing.T) {
	// 创建测试用的 CodeSplitter
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	// 定义测试文件
	testFiles := []struct {
		name        string
		filePath    string
		expectError bool
		expectCount int
		description string
	}{
		{
			name:        "OpenAPI 3.0 JSON 文件",
			filePath:    "../../bin/openapi3.json",
			expectError: false,
			expectCount: 2, // /pets 和 /pets/{petId} 两个路径
			description: "应该成功分割 OpenAPI 3.0 JSON 文件",
		},
		{
			name:        "OpenAPI 3.0 YAML 文件",
			filePath:    "../../bin/openapi3.yaml",
			expectError: false,
			expectCount: 2, // /users 和 /users/{id} 两个路径
			description: "应该成功分割 OpenAPI 3.0 YAML 文件",
		},
		{
			name:        "Swagger 2.0 JSON 文件",
			filePath:    "../../bin/swagger2.json",
			expectError: false,
			expectCount: 14, // 14个不同的路径
			description: "应该成功分割 Swagger 2.0 JSON 文件",
		},
		{
			name:        "Swagger 2.0 YAML 文件",
			filePath:    "../../bin/swagger2.yaml",
			expectError: true, // 目前不支持Swagger 2.0 YAML 文件
			expectCount: 2,    // /users 和 /users/{id} 两个路径
			description: "应该成功分割 Swagger 2.0 YAML 文件",
		},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			// 读取文件内容
			content, err := os.ReadFile(tt.filePath)
			assert.NoError(t, err, "应该能够读取文件 %s", tt.filePath)

			// 创建测试用的 SourceFile
			sourceFile := &types.SourceFile{
				CodebaseId:   1,
				CodebasePath: "/test/path",
				CodebaseName: "test-codebase",
				Path:         filepath.Base(tt.filePath),
				Content:      content,
			}
			if !splitter.splitOptions.EnableOpenAPIParsing {
				assert.Error(t, err, "openapi file parse is close")
			}
			// 执行分割
			chunks, err := splitter.splitOpenAPIFile(sourceFile)
			// 验证结果
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, chunks, "错误时应该返回 nil chunks")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, chunks, "成功时应该返回非 nil chunks")
				assert.Len(t, chunks, tt.expectCount, "应该返回正确数量的 chunks")

				// 验证每个 chunk 的基本属性
				for i, chunk := range chunks {
					assert.Equal(t, "doc", chunk.Language, "chunk %d 的语言应该是 'doc'", i)
					assert.Equal(t, sourceFile.CodebaseId, chunk.CodebaseId, "chunk %d 的 CodebaseId 应该匹配", i)
					assert.Equal(t, sourceFile.CodebasePath, chunk.CodebasePath, "chunk %d 的 CodebasePath 应该匹配", i)
					assert.Equal(t, sourceFile.CodebaseName, chunk.CodebaseName, "chunk %d 的 CodebaseName 应该匹配", i)
					assert.Equal(t, sourceFile.Path, chunk.FilePath, "chunk %d 的 FilePath 应该匹配", i)
					assert.Greater(t, chunk.TokenCount, 0, "chunk %d 的 TokenCount 应该大于 0", i)
					assert.NotEmpty(t, chunk.Content, "chunk %d 的 Content 不应该为空", i)

					// 验证分割后的文档是有效的 JSON
					var doc map[string]interface{}
					err := json.Unmarshal(chunk.Content, &doc)
					assert.NoError(t, err, "chunk %d 的内容应该是有效的 JSON", i)

					// 验证标题包含路径信息
					if info, exists := doc["info"]; exists {
						if infoMap, ok := info.(map[string]interface{}); ok {
							if title, exists := infoMap["title"]; exists {
								titleStr := title.(string)
								assert.Contains(t, titleStr, " - ", "chunk %d 的标题应该包含路径分隔符", i)
							}
						}
					}

					// 验证路径数量
					if paths, exists := doc["paths"]; exists {
						if pathsMap, ok := paths.(map[string]interface{}); ok {
							assert.Len(t, pathsMap, 1, "chunk %d 应该只包含一个路径", i)
						}
					}
				}
			}
		})
	}
}

func TestValidateOpenAPISpec(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		content     []byte
		expectVer   APIVersion
		filePath    string
		expectError bool
	}{
		{
			name:        "OpenAPI 3.0 JSON",
			content:     []byte(`{"openapi": "3.0.3", "info": {"title": "test", "version": "1.0.0"}}`),
			expectVer:   OpenAPI3,
			filePath:    "test.json",
			expectError: false,
		},
		{
			name:        "Swagger 2.0 JSON",
			content:     []byte(`{"swagger": "2.0", "info": {"title": "test", "version": "1.0.0"}}`),
			expectVer:   Swagger2,
			filePath:    "test.yaml",
			expectError: false,
		},
		{
			name:        "无效 JSON",
			content:     []byte(`{ invalid json`),
			expectVer:   Unknown,
			filePath:    "test.json",
			expectError: true,
		},
		{
			name:        "不支持的版本",
			content:     []byte(`{"openapi": "4.0.0", "info": {"title": "test", "version": "1.0.0"}}`),
			expectVer:   Unknown,
			filePath:    "test.yaml",
			expectError: true,
		},
		{
			name:        "缺少版本字段",
			content:     []byte(`{"info": {"title": "test", "version": "1.0.0"}}`),
			expectVer:   Unknown,
			filePath:    "test.json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !splitter.splitOptions.EnableOpenAPIParsing {
				assert.Error(t, err, "openapi file parse is close")
			}
			version, err := splitter.validateOpenAPISpec(tt.content, tt.filePath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectVer, version)
		})
	}
}

// 测试边界情况
func TestSplitOpenAPIFileEdgeCases(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	t.Run("空路径的 OpenAPI 文档", func(t *testing.T) {
		if !splitter.splitOptions.EnableOpenAPIParsing {
			assert.Error(t, err, "openapi file parse is close")
		}
		doc := map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":   "Test API",
				"version": "1.0.0",
			},
			"paths": map[string]interface{}{}, // 空路径
		}

		content, _ := json.Marshal(doc)
		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "test-api.json",
			Content:      content,
		}

		chunks, err := splitter.splitOpenAPIFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 0, "空路径应该返回 0 个 chunks")
	})

	t.Run("单个路径的文档", func(t *testing.T) {
		if !splitter.splitOptions.EnableOpenAPIParsing {
			assert.Error(t, err, "openapi file parse is close")
		}
		doc := map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":   "Test API",
				"version": "1.0.0",
			},
			"paths": map[string]interface{}{
				"/single": map[string]interface{}{
					"get": map[string]interface{}{
						"summary": "Single endpoint",
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "Success",
							},
						},
					},
				},
			},
		}

		content, _ := json.Marshal(doc)
		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "test-api.json",
			Content:      content,
		}

		chunks, err := splitter.splitOpenAPIFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 1, "单个路径应该返回 1 个 chunk")

		// 验证 chunk 内容
		var chunkDoc map[string]interface{}
		err = json.Unmarshal(chunks[0].Content, &chunkDoc)
		assert.NoError(t, err)

		// 验证标题包含路径信息
		if info, exists := chunkDoc["info"]; exists {
			if infoMap, ok := info.(map[string]interface{}); ok {
				if title, exists := infoMap["title"]; exists {
					titleStr := title.(string)
					assert.Contains(t, titleStr, " - /single", "标题应该包含路径信息")
				}
			}
		}
	})
}

// 测试复杂文档的分割结果
func TestComplexOpenAPIDocumentSplitting(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	t.Run("Swagger 2.0 JSON 完整文档分割", func(t *testing.T) {
		if !splitter.splitOptions.EnableOpenAPIParsing {
			assert.Error(t, err, "openapi file parse is close")
		}
		content, err := os.ReadFile("../../bin/swagger2.json")
		assert.NoError(t, err, "应该能够读取 swagger2.json 文件")

		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "petstore-api",
			Path:         "swagger2.json",
			Content:      content,
		}

		chunks, err := splitter.splitOpenAPIFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 14, "Swagger 2.0 JSON 应该有 14 个路径")

		// 验证所有路径都被正确分割
		expectedPaths := []string{
			"/pet", "/pet/findByStatus", "/pet/findByTags", "/pet/{petId}",
			"/pet/{petId}/uploadImage", "/store/inventory", "/store/order",
			"/store/order/{orderId}", "/user", "/user/createWithArray",
			"/user/createWithList", "/user/login", "/user/logout", "/user/{username}",
		}

		for i, chunk := range chunks {
			var chunkDoc map[string]interface{}
			err := json.Unmarshal(chunk.Content, &chunkDoc)
			assert.NoError(t, err, "chunk %d 应该是有效的 JSON", i)

			// 验证每个 chunk 只包含一个路径
			if paths, exists := chunkDoc["paths"]; exists {
				if pathsMap, ok := paths.(map[string]interface{}); ok {
					assert.Len(t, pathsMap, 1, "chunk %d 应该只包含一个路径", i)

					// 验证路径名称
					for path := range pathsMap {
						assert.Contains(t, expectedPaths, path, "chunk %d 包含的路径应该在预期列表中", i)
					}
				}
			}

			// 验证保留了所有必要的字段
			assert.Contains(t, chunkDoc, "swagger", "chunk %d 应该包含 swagger 版本", i)
			assert.Contains(t, chunkDoc, "info", "chunk %d 应该包含 info", i)
			assert.Contains(t, chunkDoc, "definitions", "chunk %d 应该包含 definitions", i)
			assert.Contains(t, chunkDoc, "securityDefinitions", "chunk %d 应该包含 securityDefinitions", i)
		}
	})

	t.Run("OpenAPI 3.0 JSON 文档分割", func(t *testing.T) {
		if !splitter.splitOptions.EnableOpenAPIParsing {
			assert.Error(t, err, "openapi file parse is close")
		}
		content, err := os.ReadFile("../../bin/openapi3.json")
		assert.NoError(t, err, "应该能够读取 openapi3.json 文件")

		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "petstore-extended-api",
			Path:         "openapi3.json",
			Content:      content,
		}

		chunks, err := splitter.splitOpenAPIFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 2, "OpenAPI 3.0 JSON 应该有 2 个路径")

		// 验证所有路径都被正确分割
		expectedPaths := []string{"/pets", "/pets/{petId}"}

		for i, chunk := range chunks {
			var chunkDoc map[string]interface{}
			err := json.Unmarshal(chunk.Content, &chunkDoc)
			assert.NoError(t, err, "chunk %d 应该是有效的 JSON", i)

			// 验证每个 chunk 只包含一个路径
			if paths, exists := chunkDoc["paths"]; exists {
				if pathsMap, ok := paths.(map[string]interface{}); ok {
					assert.Len(t, pathsMap, 1, "chunk %d 应该只包含一个路径", i)

					// 验证路径名称
					for path := range pathsMap {
						assert.Contains(t, expectedPaths, path, "chunk %d 包含的路径应该在预期列表中", i)
					}
				}
			}

			// 验证保留了所有必要的字段
			assert.Contains(t, chunkDoc, "openapi", "chunk %d 应该包含 openapi 版本", i)
			assert.Contains(t, chunkDoc, "info", "chunk %d 应该包含 info", i)
			assert.Contains(t, chunkDoc, "components", "chunk %d 应该包含 components", i)
			assert.Contains(t, chunkDoc, "servers", "chunk %d 应该包含 servers", i)
		}
	})

	t.Run("OpenAPI 3.0 YAML 文档分割", func(t *testing.T) {
		if !splitter.splitOptions.EnableOpenAPIParsing {
			assert.Error(t, err, "openapi file parse is close")
		}
		content, err := os.ReadFile("../../bin/openapi3.yaml")
		assert.NoError(t, err, "应该能够读取 openapi3.yaml 文件")

		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "user-management-api",
			Path:         "openapi3.yaml",
			Content:      content,
		}

		chunks, err := splitter.splitOpenAPIFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 2, "OpenAPI 3.0 YAML 应该有 2 个路径")

		// 验证所有路径都被正确分割
		expectedPaths := []string{"/users", "/users/{id}"}

		for i, chunk := range chunks {
			var chunkDoc map[string]interface{}
			err := json.Unmarshal(chunk.Content, &chunkDoc)
			assert.NoError(t, err, "chunk %d 应该是有效的 JSON", i)

			// 验证每个 chunk 只包含一个路径
			if paths, exists := chunkDoc["paths"]; exists {
				if pathsMap, ok := paths.(map[string]interface{}); ok {
					assert.Len(t, pathsMap, 1, "chunk %d 应该只包含一个路径", i)

					// 验证路径名称
					for path := range pathsMap {
						assert.Contains(t, expectedPaths, path, "chunk %d 包含的路径应该在预期列表中", i)
					}
				}
			}

			// 验证保留了所有必要的字段
			assert.Contains(t, chunkDoc, "openapi", "chunk %d 应该包含 openapi 版本", i)
			assert.Contains(t, chunkDoc, "info", "chunk %d 应该包含 info", i)
			assert.Contains(t, chunkDoc, "components", "chunk %d 应该包含 components", i)
			assert.Contains(t, chunkDoc, "servers", "chunk %d 应该包含 servers", i)
		}
	})

	t.Run("Swagger 2.0 YAML 文档分割", func(t *testing.T) {
		if !splitter.splitOptions.EnableOpenAPIParsing {
			assert.Error(t, err, "openapi file parse is close")
		}
		content, err := os.ReadFile("../../bin/swagger2.yaml")
		assert.NoError(t, err, "应该能够读取 swagger2.yaml 文件")

		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "user-management-api",
			Path:         "swagger2.yaml",
			Content:      content,
		}

		chunks, err := splitter.splitOpenAPIFile(sourceFile)
		assert.IsType(t, err, parser.ErrInvalidOpenAPISpec)
		assert.Len(t, chunks, 0, "Swagger 2.0 YAML 应该有 2 个路径")

		// 验证所有路径都被正确分割
		expectedPaths := []string{"/users", "/users/{id}"}

		for i, chunk := range chunks {
			var chunkDoc map[string]interface{}
			err := json.Unmarshal(chunk.Content, &chunkDoc)
			assert.NoError(t, err, "chunk %d 应该是有效的 JSON", i)

			// 验证每个 chunk 只包含一个路径
			if paths, exists := chunkDoc["paths"]; exists {
				if pathsMap, ok := paths.(map[string]interface{}); ok {
					assert.Len(t, pathsMap, 1, "chunk %d 应该只包含一个路径", i)

					// 验证路径名称
					for path := range pathsMap {
						assert.Contains(t, expectedPaths, path, "chunk %d 包含的路径应该在预期列表中", i)
					}
				}
			}

			// 验证保留了所有必要的字段
			assert.Contains(t, chunkDoc, "swagger", "chunk %d 应该包含 swagger 版本", i)
			assert.Contains(t, chunkDoc, "info", "chunk %d 应该包含 info", i)
			assert.Contains(t, chunkDoc, "definitions", "chunk %d 应该包含 definitions", i)
			assert.Contains(t, chunkDoc, "securityDefinitions", "chunk %d 应该包含 securityDefinitions", i)
		}
	})
}

// 测试错误情况
func TestSplitOpenAPIFileErrorCases(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		content     []byte
		filePath    string
		expectError bool
		description string
	}{
		{
			name:        "无效 JSON",
			content:     []byte(`{ invalid json`),
			filePath:    "test.json",
			expectError: true,
			description: "应该返回 JSON 解析错误",
		},
		{
			name:        "不支持的版本",
			content:     []byte(`{"openapi": "4.0.0", "info": {"title": "test", "version": "1.0.0"}}`),
			filePath:    "test.json",
			expectError: true,
			description: "应该返回不支持的版本错误",
		},
		{
			name:        "缺少必要字段的 OpenAPI 3.0",
			content:     []byte(`{"openapi": "3.0.0", "paths": {}}`),
			filePath:    "test.json",
			expectError: true,
			description: "应该返回验证错误",
		},
		{
			name:        "缺少必要字段的 Swagger 2.0",
			content:     []byte(`{"swagger": "2.0", "info": {"title": "", "version": "1.0.0"}, "paths": {}}`),
			filePath:    "test.json",
			expectError: true,
			description: "应该返回验证错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !splitter.splitOptions.EnableOpenAPIParsing {
				assert.Error(t, err, "openapi file parse is close")
			}
			sourceFile := &types.SourceFile{
				CodebaseId:   1,
				CodebasePath: "/test/path",
				CodebaseName: "test-codebase",
				Path:         tt.filePath,
				Content:      tt.content,
			}

			chunks, err := splitter.splitOpenAPIFile(sourceFile)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, chunks, "错误时应该返回 nil chunks")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, chunks, "成功时应该返回非 nil chunks")
			}
		})
	}
}

// TestSplitMarkdownFile 测试 markdown 文件分割功能
func TestSplitMarkdownFile(t *testing.T) {
	// 创建测试用的 CodeSplitter
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	// 定义测试用例
	testCases := []struct {
		name        string
		content     []byte
		expectError bool
		expectCount int
		description string
	}{
		{
			name:        "简单标题和段落",
			content:     []byte("# 标题1\n\n这是一个段落。\n\n## 标题2\n\n这是另一个段落。"),
			expectError: false,
			expectCount: 2, // 标题1块 + 标题1内容块 + 标题2块（包含内容）
			description: "应该按标题分割 markdown 文件",
		},
		{
			name:        "包含代码块",
			content:     []byte("# 标题1\n\n普通文本内容。\n\n```python\ndef hello():\n    print(\"Hello, World!\")\n```\n\n更多文本内容。\n\n```javascript\nconsole.log(\"Hello, World!\");\n```\n"),
			expectError: false,
			expectCount: 5, // 标题1块 + 标题1内容块 + 第一个代码块 + 更多文本内容块 + 第二个代码块
			description: "应该正确处理代码块分割",
		},
		{
			name:        "只有代码块",
			content:     []byte("```python\ndef hello():\n    print(\"Hello, World!\")\n```\n\n```javascript\nconsole.log(\"Hello, World!\");\n```\n"),
			expectError: false,
			expectCount: 4, // 两个代码块
			description: "应该正确处理只有代码块的文件",
		},
		{
			name:        "空文件",
			content:     []byte(""),
			expectError: false,
			expectCount: 1,
			description: "空文件应该返回 0 个 chunks",
		},
		{
			name:        "只有普通文本",
			content:     []byte("这是第一行文本。\n这是第二行文本。\n这是第三行文本。"),
			expectError: false,
			expectCount: 1, // 一个文本块
			description: "应该正确处理只有普通文本的文件",
		},
		{
			name:        "混合内容",
			content:     []byte("# 主标题\n\n介绍文本。\n\n## 子标题1\n\n更多文本。\n\n```python\ndef example():\n    pass\n```\n\n## 子标题2\n\n结尾文本。"),
			expectError: false,
			expectCount: 5, // 主标题块 + 主标题内容块 + 子标题1块 + 子标题1内容块 + 代码块 + 子标题2块（包含内容）
			description: "应该正确处理混合内容",
		},
		{
			name: "真实文档 - 自定义知识库功能",
			content: []byte(`

#### **Epic 1: 自定义知识库**

- **FR-2.1: 知识库管理界面**
		- **用户故事：** 作为一个开发者，我希望在Costrict的设置界面中有一个"知识库管理"面板，我可以在这里看到当前已导入的文档，并且可以添加或移除它们。
		- **交付物：**
			 1. 在Costrict的VS Code设置页或侧边栏视图中，增加一个"知识库"管理区域。
			 2. 该区域应包含一个"添加文档/文件夹"按钮和一个已导入源的列表。
			 3. 知识库系统页面，要支持树状管理，方便用户组织大量知识库。
			 4. 列表中的每一项都应有关联的"移除"按钮。
		- **验收标准：**
			 1. 用户可以通过点击按钮，打开文件/文件夹选择器。
			 2. 成功选择.md文件或包含.md文件的文件夹后，该源会出现在列表中。
			 3. 点击"移除"按钮后，相应的源会从列表中消失，并且其对应的向量化数据也会从数据库中被删除。
			 4. 点击某个md文档，该文档会出现在vscode编辑窗口，支持编辑该知识
			 5. 每当知识库有变动时，需要增量 向量化该变动，存储起来。

------



#### **Epic 2: AI统一检索接口**

- **FR-3.1: 知识库检索Function Call**
		- **用户故事：** 作为AI应用层的开发人员，我需要一个稳定且高效的内部API（作为Function Call暴露给LLM），通过输入自然语言查询，就能从本地向量数据库中获得最相关的上下文片段。
		- **交付物：**
			 1. 一个名为search_knowledge_base(query: str, top_k: int = 5)的函数或方法。
			 2. 该函数负责将输入的query文本进行向量化，查询本地向量数据库。
			 3. 应用Rerank模型对初步检索结果进行重排序，提升精准度。
			 4. 返回top_k个最相关的文本片段（Chunks）及其元数据（如来源文件、相关性得分）。
		- **验收标准：**
			 1. 该函数接口定义清晰，输入输出符合设计。
			 2. 对于一次典型查询，从接收query到返回结果的总耗时应在1秒以内。
			 3. 在LLM的调用逻辑中，当需要外部知识时，能够正确地构造并调用此函数。

**3. 效果验收方案（量化标准）**

这是本迭代成功的关键，我们需要量化"精准"的定义。

- **3.1. 建立评测基准 (Benchmark)**
		1. **选取/创建一个标准测试项目：** 选择一个功能完整、代码量适中（1-5万行）的开源项目（如 express, fastapi 的某个demo项目）。
		2. **构建"黄金问题-答案对" (Golden Q&A Pairs)：** 手动编写20-30个关于此项目的典型开发问题，并明确指出答案存在于哪个/哪些源文件或文档片段中。
			  - **示例问题：** "如何在项目中添加一个新的认证中间件？"
			  - **黄金答案（源）：** 指向 src/middlewares/auth.ts 文件的特定函数定义。
			  - **示例问题：** "项目的数据库Schema是如何定义的？"
			  - **黄金答案（源）：** 指向 docs/database.md 或 src/models/user.ts。
- **3.2. 验收指标**
		- **指标1：检索准确率 (Retrieval Accuracy - Hit Rate @ K)**
			 - **定义：** 对于一个"黄金问题"，如果其对应的"黄金答案（源）"出现在search_knowledge_base函数返回的前K个结果中，则视为命中。
			 - **验收标准：** **Hit Rate @ 3 >= 90%**。即对于90%以上的测试问题，正确的上下文信息能排在检索结果的前3位。
		- **指标2：平均倒数排名 (Mean Reciprocal Rank - MRR)**
			 - **定义：** 衡量第一个正确答案出现位置的指标。如果第一个正确答案排在第1位，得分为1；排第2，得分为1/2；排第3，得分为1/3，以此类推。MRR是所有问题得分的平均值。
			 - **验收标准：** **MRR >= 0.75**。这代表总体上，正确的答案平均能排在第一或第二位。
		- **指标3：端到端任务成功率 (End-to-End Task Success Rate)**
			 - **定义：** 这是最终的业务指标。对10个预设的、需要背景知识才能完成的编码任务（如"基于现有API文档，为getUserProfile函数编写一个调用示例"），分别在"有知识库"和"无知识库"两种模式下让AI生成代码。
			 - **验收标准：**
			   1. **对比基线：** 首先记录"无知识库"模式下的成功率（例如，10个任务中成功2个，成功率20%）。
			   2. **最终目标：** "有知识库"模式下的**成功率需达到60%以上**，且生成的代码质量（如正确使用内部函数、遵循项目规范）有评测人员可感知的明显提升。
`),
			expectError: false,
			expectCount: 5, // Epic 1标题块 + Epic 1内容块 + Epic 2标题块 + Epic 2内容块 + 效果验收方案块
			description: "应该正确处理真实的自定义知识库功能文档",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试用的 SourceFile
			sourceFile := &types.SourceFile{
				CodebaseId:   1,
				CodebasePath: "/test/path",
				CodebaseName: "test-codebase",
				Path:         "test.md",
				Content:      tt.content,
			}

			// 执行分割
			chunks, err := splitter.splitMarkdownFile(sourceFile)

			// 验证结果
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, chunks, "错误时应该返回 nil chunks")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, chunks, "成功时应该返回非 nil chunks")
				assert.Len(t, chunks, tt.expectCount, "应该返回正确数量的 chunks")

				// 验证每个 chunk 的基本属性
				for i, chunk := range chunks {
					assert.Equal(t, "doc", chunk.Language, "chunk %d 的语言应该是 'doc'", i)
					assert.Equal(t, sourceFile.CodebaseId, chunk.CodebaseId, "chunk %d 的 CodebaseId 应该匹配", i)
					assert.Equal(t, sourceFile.CodebasePath, chunk.CodebasePath, "chunk %d 的 CodebasePath 应该匹配", i)
					assert.Equal(t, sourceFile.CodebaseName, chunk.CodebaseName, "chunk %d 的 CodebaseName 应该匹配", i)
					assert.Equal(t, sourceFile.Path, chunk.FilePath, "chunk %d 的 FilePath 应该匹配", i)
					assert.Greater(t, chunk.TokenCount, 0, "chunk %d 的 TokenCount 应该大于 0", i)
					assert.NotEmpty(t, chunk.Content, "chunk %d 的 Content 不应该为空", i)

					// 验证范围信息
					assert.Len(t, chunk.Range, 4, "chunk %d 的 Range 应该有 4 个元素", i)
					assert.GreaterOrEqual(t, chunk.Range[0], 0, "chunk %d 的起始行应该 >= 0", i)
					assert.GreaterOrEqual(t, chunk.Range[2], chunk.Range[0], "chunk %d 的结束行应该 >= 起始行", i)
				}
			}
		})
	}
}

// TestSplitMarkdownFileEdgeCases 测试 markdown 文件分割的边界情况
func TestSplitMarkdownFileEdgeCases(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	t.Run("未闭合的代码块", func(t *testing.T) {
		content := "# 标题\n\n```python\ndef hello():\n    print(\"Hello, World!\")\n# 没有闭合的代码块\n"
		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "test.md",
			Content:      []byte(content),
		}

		chunks, err := splitter.splitMarkdownFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 2, "应该有 2 个 chunks（标题块 + 标题内容块 + 未闭合的代码块）")
	})

	t.Run("只有标题", func(t *testing.T) {
		content := `# 标题1
## 标题2
### 标题3
`
		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "test.md",
			Content:      []byte(content),
		}

		chunks, err := splitter.splitMarkdownFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 3, "应该有 3 个 chunks（每个标题一个，没有内容块）")
	})

	t.Run("空代码块", func(t *testing.T) {
		content := "# 标题\n\n```\n\n空代码块\n\n```\n"
		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "test.md",
			Content:      []byte(content),
		}

		chunks, err := splitter.splitMarkdownFile(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 3, "应该有 3 个 chunks（标题块 + 标题内容块 + 空代码块）")
	})
}

// TestSplitRealMarkdownFile 测试真实 markdown 文件的分割功能
func TestSplitRealMarkdownFile(t *testing.T) {
	// 创建测试用的 CodeSplitter
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	// 定义测试文件
	testFiles := []struct {
		name        string
		filePath    string
		expectError bool
		expectCount int
		description string
	}{
		{
			name:        "自定义知识库功能文档",
			filePath:    "../../bin/自定义知识库功能.md",
			expectError: false,
			expectCount: 5, // 预期会分成5个主要部分：Epic 1标题块 + Epic 1内容块 + Epic 2标题块 + Epic 2内容块 + 效果验收方案块
			description: "应该成功分割真实的自定义知识库功能文档",
		},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			// 读取文件内容
			content, err := os.ReadFile(tt.filePath)
			assert.NoError(t, err, "应该能够读取文件 %s", tt.filePath)

			// 创建测试用的 SourceFile
			sourceFile := &types.SourceFile{
				CodebaseId:   1,
				CodebasePath: "/test/path",
				CodebaseName: "test-codebase",
				Path:         filepath.Base(tt.filePath),
				Content:      content,
			}

			// 执行分割
			chunks, err := splitter.splitMarkdownFile(sourceFile)

			// 验证结果
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, chunks, "错误时应该返回 nil chunks")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, chunks, "成功时应该返回非 nil chunks")
				assert.GreaterOrEqual(t, len(chunks), tt.expectCount, "应该返回至少 %d 个 chunks", tt.expectCount)

				// 打印所有 chunks 的内容
				fmt.Printf("=== 分割结果：共 %d 个 chunks ===\n", len(chunks))
				for i, chunk := range chunks {
					fmt.Printf("\n--- Chunk %d ---\n", i+1)
					fmt.Printf("Token数量: %d\n", chunk.TokenCount)
					fmt.Printf("字节数: %d\n", len(chunk.Content))
					fmt.Printf("范围: %v\n", chunk.Range)
					fmt.Printf("内容:\n%s\n", string(chunk.Content))
					fmt.Printf("--- Chunk %d 结束 ---\n", i+1)
				}
				fmt.Printf("=== 分割结果结束 ===\n\n")

				// 验证每个 chunk 的基本属性
				for i, chunk := range chunks {
					assert.Equal(t, "doc", chunk.Language, "chunk %d 的语言应该是 'doc'", i)
					assert.Equal(t, sourceFile.CodebaseId, chunk.CodebaseId, "chunk %d 的 CodebaseId 应该匹配", i)
					assert.Equal(t, sourceFile.CodebasePath, chunk.CodebasePath, "chunk %d 的 CodebasePath 应该匹配", i)
					assert.Equal(t, sourceFile.CodebaseName, chunk.CodebaseName, "chunk %d 的 CodebaseName 应该匹配", i)
					assert.Equal(t, sourceFile.Path, chunk.FilePath, "chunk %d 的 FilePath 应该匹配", i)
					assert.Greater(t, chunk.TokenCount, 0, "chunk %d 的 TokenCount 应该大于 0", i)
					assert.NotEmpty(t, chunk.Content, "chunk %d 的 Content 不应该为空", i)

					// 验证范围信息
					assert.Len(t, chunk.Range, 4, "chunk %d 的 Range 应该有 4 个元素", i)
					assert.GreaterOrEqual(t, chunk.Range[0], 0, "chunk %d 的起始行应该 >= 0", i)
					assert.GreaterOrEqual(t, chunk.Range[2], chunk.Range[0], "chunk %d 的结束行应该 >= 起始行", i)

					// 验证 chunk 内容包含预期的 markdown 结构
					// chunkStr := string(chunk.Content)
					// assert.Contains(t, chunkStr, "#", "chunk %d 应该包含 markdown 标题或列表结构", i)
				}

				// 验证特定内容：确保主要章节都被正确分割
				foundEpic1 := false
				foundEpic2 := false
				foundBenchmark := false

				for _, chunk := range chunks {
					chunkStr := string(chunk.Content)
					if strings.Contains(chunkStr, "Epic 1: 自定义知识库") {
						foundEpic1 = true
					}
					if strings.Contains(chunkStr, "Epic 2: AI统一检索接口") {
						foundEpic2 = true
					}
					if strings.Contains(chunkStr, "效果验收方案（量化标准）") {
						foundBenchmark = true
					}
				}

				assert.True(t, foundEpic1, "应该找到包含 Epic 1 的 chunk")
				assert.True(t, foundEpic2, "应该找到包含 Epic 2 的 chunk")
				assert.True(t, foundBenchmark, "应该找到包含效果验收方案的 chunk")
			}
		})
	}
}

// TestSplitMarkdownFileLargeContent 测试大内容 markdown 文件的分割
func TestSplitMarkdownFileLargeContent(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          50, // 设置较小的 token 限制来测试分割
		SlidingWindowOverlapTokens: 10,
		EnableMarkdownParsing:      true,
		EnableOpenAPIParsing:       true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	t.Run("大文本内容分割", func(t *testing.T) {
		// 创建一个长文本
		var longText strings.Builder
		longText.WriteString("# 大标题\n\n")
		for i := 0; i < 100; i++ {
			longText.WriteString(fmt.Sprintf("这是第 %d 行文本，用于测试大内容的分割功能。\n", i))
		}

		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "large.md",
			Content:      []byte(longText.String()),
		}

		chunks, err := splitter.splitMarkdownFile(sourceFile)
		assert.NoError(t, err)
		assert.Greater(t, len(chunks), 1, "大内容应该被分割成多个 chunks")

		// 验证所有 chunks 的 token 数量都不超过限制
		for i, chunk := range chunks {
			assert.LessOrEqual(t, chunk.TokenCount, splitOptions.MaxTokensPerChunk,
				"chunk %d 的 TokenCount 应该 <= MaxTokensPerChunk", i)
		}
	})
}

// TestSplitMarkdownFileBySitter 测试使用 tree-sitter 的 markdown 文件分割功能
func TestSplitMarkdownFileBySitter(t *testing.T) {
	// 创建测试用的 CodeSplitter
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	// 定义测试用例
	testCases := []struct {
		name        string
		content     []byte
		expectError bool
		expectCount int
		description string
	}{
		{
			name:        "简单标题和段落",
			content:     []byte("# 标题1\n\n这是一个段落。\n\n## 标题2\n\n这是另一个段落。"),
			expectError: false,
			expectCount: 2, // 两个标题块，每个包含其内容
			description: "应该按标题分割 markdown 文件",
		},
		{
			name:        "多级标题",
			content:     []byte("# 一级标题\n\n内容1\n\n## 二级标题\n\n内容2\n\n### 三级标题\n\n内容3\n\n# 另一个一级标题\n\n内容4"),
			expectError: false,
			expectCount: 4, // 4个主要部分，每个标题一个块
			description: "应该正确处理多级标题结构",
		},
		{
			name:        "只有标题没有内容",
			content:     []byte("# 标题1\n\n## 标题2\n\n### 标题3"),
			expectError: false,
			expectCount: 3, // 3个标题块，即使没有内容
			description: "应该正确处理只有标题的文档",
		},
		// {
		// 	name:        "空文件",
		// 	content:     []byte(""),
		// 	expectError: false,
		// 	expectCount: 1, // 空文件也会被当作一个块处理
		// 	description: "空文件应该返回 1 个 chunk",
		// },
		{
			name:        "没有标题的纯文本",
			content:     []byte("这是第一行文本。\n这是第二行文本。\n这是第三行文本。"),
			expectError: false,
			expectCount: 1, // 一个文本块
			description: "应该正确处理没有标题的纯文本文件",
		},
		{
			name:        "复杂嵌套标题结构",
			content:     []byte("# 主标题\n\n介绍文本。\n\n## 子标题1\n\n更多文本。\n\n### 子子标题1\n\n详细内容。\n\n## 子标题2\n\n结尾文本。\n\n### 子子标题2\n\n更多详细内容。"),
			expectError: false,
			expectCount: 5, // 主标题 + 子标题1 + 子子标题1 + 子标题2（子子标题2会被包含在子标题2中）
			description: "应该正确处理复杂的嵌套标题结构",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试用的 SourceFile
			sourceFile := &types.SourceFile{
				CodebaseId:   1,
				CodebasePath: "/test/path",
				CodebaseName: "test-codebase",
				Path:         "test.md",
				Content:      tt.content,
			}

			// 执行分割
			chunks, err := splitter.splitMarkdownFileBySitter(sourceFile)

			// 验证结果
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, chunks, "错误时应该返回 nil chunks")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, chunks, "成功时应该返回非 nil chunks")
				assert.Len(t, chunks, tt.expectCount, "应该返回正确数量的 chunks")

				// 验证每个 chunk 的基本属性
				for i, chunk := range chunks {
					assert.Equal(t, "doc", chunk.Language, "chunk %d 的语言应该是 'doc'", i)
					assert.Equal(t, sourceFile.CodebaseId, chunk.CodebaseId, "chunk %d 的 CodebaseId 应该匹配", i)
					assert.Equal(t, sourceFile.CodebasePath, chunk.CodebasePath, "chunk %d 的 CodebasePath 应该匹配", i)
					assert.Equal(t, sourceFile.CodebaseName, chunk.CodebaseName, "chunk %d 的 CodebaseName 应该匹配", i)
					assert.Equal(t, sourceFile.Path, chunk.FilePath, "chunk %d 的 FilePath 应该匹配", i)
					assert.Greater(t, chunk.TokenCount, 0, "chunk %d 的 TokenCount 应该大于 0", i)
					assert.NotEmpty(t, chunk.Content, "chunk %d 的 Content 不应该为空", i)

					// 验证范围信息
					assert.Len(t, chunk.Range, 4, "chunk %d 的 Range 应该有 4 个元素", i)
					assert.GreaterOrEqual(t, chunk.Range[0], 0, "chunk %d 的起始行应该 >= 0", i)
					assert.GreaterOrEqual(t, chunk.Range[2], chunk.Range[0], "chunk %d 的结束行应该 >= 起始行", i)
				}
			}
		})
	}
}

// TestSplitRealMarkdownFileBySitter 测试真实 markdown 文件的 tree-sitter 分割功能
func TestSplitRealMarkdownFileBySitter(t *testing.T) {
	// 创建测试用的 CodeSplitter
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	// 定义测试文件
	testFiles := []struct {
		name        string
		filePath    string
		expectError bool
		expectCount int
		description string
	}{
		{
			name:        "api_documentation文档",
			filePath:    "/Code/Go/zgsm-ai/codebase-embedder/docs/api_documentation.md",
			expectError: false,
			expectCount: 1, // 至少会有1个块，具体数量取决于文档中的标题数量
			description: "应该成功分割真实的 Markdown 文档",
		},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			// 读取文件内容
			content, err := os.ReadFile(tt.filePath)
			if err != nil {
				// 如果文件不存在，跳过这个测试
				t.Skipf("无法读取文件 %s: %v", tt.filePath, err)
			}

			// 创建测试用的 SourceFile
			sourceFile := &types.SourceFile{
				CodebaseId:   1,
				CodebasePath: "/test/path",
				CodebaseName: "test-codebase",
				Path:         filepath.Base(tt.filePath),
				Content:      content,
			}

			// 执行分割
			startTime := time.Now()
			chunks, err := splitter.splitMarkdownFileBySitter(sourceFile)
			duration := time.Since(startTime)

			// 验证结果
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, chunks, "错误时应该返回 nil chunks")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, chunks, "成功时应该返回非 nil chunks")
				assert.GreaterOrEqual(t, len(chunks), tt.expectCount, "应该返回至少 %d 个 chunks", tt.expectCount)

				// 打印所有 chunks 的内容（仅在调试时使用）
				if testing.Verbose() {
					fmt.Printf("=== Tree-sitter 分割结果：共 %d 个 chunks ===\n", len(chunks))
					for i, chunk := range chunks {
						fmt.Printf("\n--- Chunk %d ---\n", i+1)
						fmt.Printf("Token数量: %d\n", chunk.TokenCount)
						fmt.Printf("字节数: %d\n", len(chunk.Content))
						fmt.Printf("范围: %v\n", chunk.Range)
						fmt.Printf("内容:\n%s\n", string(chunk.Content))
						fmt.Printf("--- Chunk %d 结束 ---\n", i+1)
					}
					fmt.Printf("=== 分割结果结束 ===\n\n")
				}

				// 验证每个 chunk 的基本属性
				for i, chunk := range chunks {
					assert.Equal(t, "doc", chunk.Language, "chunk %d 的语言应该是 'doc'", i)
					assert.Equal(t, sourceFile.CodebaseId, chunk.CodebaseId, "chunk %d 的 CodebaseId 应该匹配", i)
					assert.Equal(t, sourceFile.CodebasePath, chunk.CodebasePath, "chunk %d 的 CodebasePath 应该匹配", i)
					assert.Equal(t, sourceFile.CodebaseName, chunk.CodebaseName, "chunk %d 的 CodebaseName 应该匹配", i)
					assert.Equal(t, sourceFile.Path, chunk.FilePath, "chunk %d 的 FilePath 应该匹配", i)
					assert.Greater(t, chunk.TokenCount, 0, "chunk %d 的 TokenCount 应该大于 0", i)
					assert.NotEmpty(t, chunk.Content, "chunk %d 的 Content 不应该为空", i)

					// 验证范围信息
					assert.Len(t, chunk.Range, 4, "chunk %d 的 Range 应该有 4 个元素", i)
					assert.GreaterOrEqual(t, chunk.Range[0], 0, "chunk %d 的起始行应该 >= 0", i)
					assert.GreaterOrEqual(t, chunk.Range[2], chunk.Range[0], "chunk %d 的结束行应该 >= 起始行", i)
				}
				fmt.Printf("Tree-sitter 分割耗时: %v\n", duration)

				// 验证特定内容：确保主要章节都被正确分割
				// foundEpic1 := false
				// foundEpic2 := false
				// foundBenchmark := false

				// for _, chunk := range chunks {
				// 	chunkStr := string(chunk.Content)
				// 	if strings.Contains(chunkStr, "Codebase Embedder API 文档") {
				// 		foundEpic1 = true
				// 	}
				// 	if strings.Contains(chunkStr, "## 1. 服务简介") {
				// 		foundEpic2 = true
				// 	}
				// 	if strings.Contains(chunkStr, "## 2. 认证方法") {
				// 		foundBenchmark = true
				// 	}
				// }

				// assert.True(t, foundEpic1, "应该找到包含 Codebase Embedder API 文档 的 chunk")
				// assert.True(t, foundEpic2, "应该找到包含 ## 1. 服务简介 的 chunk")
				// assert.True(t, foundBenchmark, "应该找到包含 ## 2. 认证方法 的 chunk")
			}
		})
	}
}

// TestSplitMarkdownFileBySitterLargeContent 测试大内容 markdown 文件的 tree-sitter 分割
func TestSplitMarkdownFileBySitterLargeContent(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          50, // 设置较小的 token 限制来测试分割
		SlidingWindowOverlapTokens: 10,
		EnableMarkdownParsing:      true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	t.Run("大文本内容分割", func(t *testing.T) {
		// 创建一个长文本
		var longText strings.Builder
		longText.WriteString("# 大标题\n\n")
		for i := 0; i < 100; i++ {
			longText.WriteString(fmt.Sprintf("这是第 %d 行文本，用于测试大内容的分割功能。\n", i))
		}

		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "large.md",
			Content:      []byte(longText.String()),
		}

		chunks, err := splitter.splitMarkdownFileBySitter(sourceFile)
		assert.NoError(t, err)
		assert.Greater(t, len(chunks), 1, "大内容应该被分割成多个 chunks")

		// 验证所有 chunks 的 token 数量都不超过限制
		for i, chunk := range chunks {
			assert.GreaterOrEqual(t, chunk.TokenCount, splitOptions.MaxTokensPerChunk,
				"chunk %d 的 TokenCount 应该 >= MaxTokensPerChunk", i)
		}
	})
}

// TestSplitMarkdownFileBySitterHeaderPath 测试标题路径的正确性
func TestSplitMarkdownFileBySitterHeaderPath(t *testing.T) {
	splitOptions := SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 100,
		EnableMarkdownParsing:      true,
	}
	splitter, err := NewCodeSplitter(splitOptions)
	assert.NoError(t, err)

	t.Run("标题路径测试", func(t *testing.T) {
		content := []byte(`# 一级标题 A

一级 A 内容。

## 二级标题 A1

二级 A1 内容。

### 三级标题 A1a

三级 A1a 内容。

# 一级标题 B

一级 B 内容。

## 二级标题 B1

二级 B1 内容。
`)

		sourceFile := &types.SourceFile{
			CodebaseId:   1,
			CodebasePath: "/test/path",
			CodebaseName: "test-codebase",
			Path:         "test.md",
			Content:      content,
		}

		chunks, err := splitter.splitMarkdownFileBySitter(sourceFile)
		assert.NoError(t, err)
		assert.Len(t, chunks, 5, "应该有 5 个 chunks（每个标题一个块）")

		// 验证每个 chunk 的标题路径
		for i, chunk := range chunks {
			chunkStr := string(chunk.Content)

			switch i {
			case 0: // 一级标题 A
				assert.Contains(t, chunkStr, "# 一级标题 A", "chunk 0 应该包含一级标题 A")
				assert.Contains(t, chunkStr, "一级 A 内容。", "chunk 0 应该包含一级标题 A 的内容")
				assert.NotContains(t, chunkStr, "## 二级标题 A1", "chunk 0 不应该包含二级标题 A1")
			case 1: // 二级标题 A1
				assert.Contains(t, chunkStr, "## 二级标题 A1", "chunk 1 应该包含二级标题 A1")
				assert.Contains(t, chunkStr, "二级 A1 内容。", "chunk 1 应该包含二级标题 A1 的内容")
				assert.NotContains(t, chunkStr, "### 三级标题 A1a", "chunk 1 不应该包含三级标题 A1a")
			case 2: // 三级标题 A1a
				assert.Contains(t, chunkStr, "### 三级标题 A1a", "chunk 2 应该包含三级标题 A1a")
				assert.Contains(t, chunkStr, "三级 A1a 内容。", "chunk 2 应该包含三级标题 A1a 的内容")
				assert.NotContains(t, chunkStr, "# 一级标题 B", "chunk 2 不应该包含一级标题 B")
			case 3: // 一级标题 B
				assert.Contains(t, chunkStr, "# 一级标题 B", "chunk 3 应该包含一级标题 B")
				assert.Contains(t, chunkStr, "一级 B 内容。", "chunk 3 应该包含一级标题 B 的内容")
				assert.NotContains(t, chunkStr, "## 二级标题 B1", "chunk 3 不应该包含二级标题 B1")
			case 4: // 二级标题 B1
				assert.Contains(t, chunkStr, "## 二级标题 B1", "chunk 4 应该包含二级标题 B1")
				assert.Contains(t, chunkStr, "二级 B1 内容。", "chunk 4 应该包含二级标题 B1 的内容")
				assert.NotContains(t, chunkStr, "# 一级标题 A", "chunk 4 不应该包含一级标题 A")
			}
		}
	})
}
