package embedding

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/yaml"
	"github.com/tiktoken-go/tokenizer"
	tree_sitter_markdown "github.com/tree-sitter-grammars/tree-sitter-markdown/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"github.com/zgsm-ai/codebase-indexer/internal/parser"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

const (
	LanguageTypeCode = "code"
	LanguageTypeDoc  = "doc"
)

type CodeSplitter struct {
	tokenizer    tokenizer.Codec
	splitOptions SplitOptions
}

type SplitOptions struct {
	MaxTokensPerChunk          int
	SlidingWindowOverlapTokens int
	EnableMarkdownParsing      bool // 是否启用markdown文件解析
	EnableOpenAPIParsing       bool // 是否启用OpenAPI文档解析
}

// NewCodeSplitter 创建代码分割器
func NewCodeSplitter(splitOptions SplitOptions) (*CodeSplitter, error) {
	codec, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokenizer: %w", err)
	}

	return &CodeSplitter{
		tokenizer:    codec,
		splitOptions: splitOptions,
	}, nil
}

// Split 将代码文件分割成多个代码块
func (p *CodeSplitter) Split(codeFile *types.SourceFile) ([]*types.CodeChunk, error) {

	language, err := parser.GetLangConfigByFilePath(codeFile.Path)
	if err != nil {
		return nil, err
	}

	if language.Language == parser.Markdown && !p.splitOptions.EnableMarkdownParsing {
		return nil, fmt.Errorf("mardownfile parse is close")
	}

	// 特殊处理 markdown 文件 - 只有在配置开启时才解析markdown
	if language.Language == parser.Markdown && p.splitOptions.EnableMarkdownParsing {
		return p.splitMarkdownFileBySitter(codeFile)
	}
	if (language.Language == parser.OpenAPI || language.Language == parser.Swagger)  {
		if !p.splitOptions.EnableOpenAPIParsing{
			return nil,fmt.Errorf("openapi file parse is close")
		}
		return p.splitOpenAPIFile(codeFile)
	}

	sitterParser := sitter.NewParser()
	defer sitterParser.Close()
	// 设置解析器语言（复用已创建的Parser）
	if err := sitterParser.SetLanguage(language.SitterLanguage()); err != nil {
		return nil, fmt.Errorf("failed to set parser language: %w", err)
	}

	// 解析代码
	tree := sitterParser.Parse(codeFile.Content, nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse code: %s", codeFile.Path)
	}
	defer tree.Close()

	// 获取要提取的节点类型
	nodeKinds, ok := languageChunkNodeKind[language.Language]
	if !ok {
		return nil, fmt.Errorf("missing chunk config for language %s", language.Language)
	}

	// 预分配切片，减少内存重新分配
	estimatedChunks := 10 // 预估每个文件约10个代码块
	allChunks := make([]*types.CodeChunk, 0, estimatedChunks)

	// 遍历语法树
	cursor := tree.RootNode().Walk()
	defer cursor.Close()

	// 使用更简洁的遍历逻辑
	for {
		currentNode := cursor.Node()
		kind := currentNode.Kind()
		// 处理目标节点类型
		if slices.Contains(nodeKinds, kind) {
			// 提取节点信息
			startPos := currentNode.StartPosition()
			endPos := currentNode.EndPosition()
			content := codeFile.Content[currentNode.StartByte():currentNode.EndByte()]
			tokenCount := p.countToken(content)

			// 处理代码切块
			if tokenCount > p.splitOptions.MaxTokensPerChunk {
				subChunks := p.splitFuncWithSlidingWindow(string(content), codeFile, int(startPos.Row), LanguageTypeCode)
				allChunks = append(allChunks, subChunks...)
			} else {
				allChunks = append(allChunks, &types.CodeChunk{
					Language:     LanguageTypeCode,
					CodebaseId:   codeFile.CodebaseId,
					CodebasePath: codeFile.CodebasePath,
					CodebaseName: codeFile.CodebaseName,
					Content:      content,
					FilePath:     codeFile.Path,
					Range:        []int{int(startPos.Row), int(startPos.Column), int(endPos.Row), int(endPos.Column)},
					TokenCount:   tokenCount,
				})
			}

			// 跳过子节点，直接移动到兄弟节点
			if !cursor.GotoNextSibling() {
				// 没有兄弟节点，回溯到父节点的兄弟节点
				for {
					if !cursor.GotoParent() {
						return allChunks, nil // 遍历完成
					}
					if cursor.GotoNextSibling() {
						break
					}
				}
			}
			continue
		}

		// 非目标节点，继续深度优先遍历
		if cursor.GotoFirstChild() {
			continue
		}

		// 无子节点，尝试兄弟节点
		for {
			if cursor.GotoNextSibling() {
				break
			}

			// 无兄弟节点，回溯父节点
			if !cursor.GotoParent() {
				return allChunks, nil // 遍历完成
			}
		}
	}
}

// countToken 计算内容的token数量
func (p *CodeSplitter) countToken(content []byte) int {
	// 避免不必要的字符串转换
	contentStr := string(content)
	tokenCount, err := p.tokenizer.Count(contentStr)
	if err != nil {
		// 回退到简单的长度计算
		return len(contentStr) / 4 // 粗略估计：1token≈4字符
	}
	return tokenCount
}

// splitFuncWithSlidingWindow 使用滑动窗口将大函数分割成多个小块
func (p *CodeSplitter) splitFuncWithSlidingWindow(content string, codeFile *types.SourceFile, funcStartLine int, languageType string) []*types.CodeChunk {
	filePath := codeFile.Path
	maxTokens := p.splitOptions.MaxTokensPerChunk
	overlapTokens := p.splitOptions.SlidingWindowOverlapTokens

	if maxTokens <= 0 || overlapTokens < 0 || overlapTokens >= maxTokens {
		return nil
	}

	// 编码内容获取tokens和字节偏移量
	_, tokens, err := p.tokenizer.Encode(content)
	if err != nil {
		return nil
	}

	totalTokens := len(tokens)
	if totalTokens == 0 {
		return nil
	}

	// 计算每个token的字节偏移量
	byteOffsets := make([]int, len(tokens)+1)
	currentOffset := 0
	for i, token := range tokens {
		byteOffsets[i] = currentOffset
		currentOffset += len(token)
	}
	byteOffsets[len(tokens)] = currentOffset

	// 预分配切片
	estimatedChunks := (totalTokens + maxTokens - 1) / maxTokens
	chunks := make([]*types.CodeChunk, 0, estimatedChunks)

	startTokenIdx := 0

	for startTokenIdx < totalTokens {
		// 计算当前块的结束位置
		endTokenIdx := startTokenIdx + maxTokens
		if endTokenIdx > totalTokens {
			endTokenIdx = totalTokens
		}

		// 提取代码块
		startByte := byteOffsets[startTokenIdx]
		endByte := byteOffsets[endTokenIdx] - 1
		if endByte >= len(content) {
			endByte = len(content) - 1
		}

		chunkContent := content[startByte : endByte+1]

		// 计算起始行和列
		startLine := funcStartLine + countLines(content[:startByte])
		startColumn := calculateColumn(content, startByte)

		// 计算结束行和列
		endLine := startLine + countLines(chunkContent) - 1
		endColumn := calculateColumn(chunkContent, endByte-startByte)

		chunks = append(chunks, &types.CodeChunk{
			Language:     languageType,
			CodebaseId:   codeFile.CodebaseId,
			CodebasePath: codeFile.CodebasePath,
			CodebaseName: codeFile.CodebaseName,
			Content:      []byte(chunkContent),
			FilePath:     filePath,
			Range:        []int{startLine, startColumn, endLine, endColumn},
			TokenCount:   endTokenIdx - startTokenIdx,
		})

		if endTokenIdx >= totalTokens {
			break
		}

		// 计算下一个块的起始位置（应用滑动窗口）
		if remaining := totalTokens - endTokenIdx; remaining < maxTokens {
			// 最后一块，调整重叠量
			startTokenIdx = endTokenIdx - (maxTokens - remaining)
		} else {
			// 正常情况，使用固定重叠
			startTokenIdx = endTokenIdx - overlapTokens
		}

		// 防止索引越界
		if startTokenIdx < 0 {
			startTokenIdx = 0
		}
	}

	return chunks
}

// splitTextWithSlidingWindow 使用滑动窗口将大文本分割成多个小块（基于字节数而非token数）
func (p *CodeSplitter) splitTextWithSlidingWindow(content string, codeFile *types.SourceFile, funcStartLine int, languageType string) []*types.CodeChunk {
	filePath := codeFile.Path
	maxBytes := p.splitOptions.MaxTokensPerChunk * 4              // 将token转换为字节数
	overlapBytes := p.splitOptions.SlidingWindowOverlapTokens * 4 // 将token转换为字节数

	if maxBytes <= 0 || overlapBytes < 0 || overlapBytes >= maxBytes {
		return nil
	}

	totalBytes := len(content)
	if totalBytes == 0 {
		return nil
	}

	// 预分配切片
	estimatedChunks := (totalBytes + maxBytes - 1) / maxBytes
	chunks := make([]*types.CodeChunk, 0, estimatedChunks)

	startByteIdx := 0

	for startByteIdx < totalBytes {
		// 计算当前块的结束位置
		endByteIdx := startByteIdx + maxBytes
		if endByteIdx > totalBytes {
			endByteIdx = totalBytes
		}

		// 提取代码块
		chunkContent := content[startByteIdx:endByteIdx]

		// 计算起始行和列
		startLine := funcStartLine + countLines(content[:startByteIdx])
		startColumn := calculateColumn(content, startByteIdx)

		// 计算结束行和列
		endLine := startLine + countLines(chunkContent) - 1
		endColumn := calculateColumn(chunkContent, endByteIdx-startByteIdx)

		// 计算token数量（用于兼容性，但不再用于分割逻辑）
		tokenCount := (endByteIdx - startByteIdx) / 4 // 粗略估计：1token≈4字节

		chunks = append(chunks, &types.CodeChunk{
			Language:     languageType,
			CodebaseId:   codeFile.CodebaseId,
			CodebasePath: codeFile.CodebasePath,
			CodebaseName: codeFile.CodebaseName,
			Content:      []byte(chunkContent),
			FilePath:     filePath,
			Range:        []int{startLine, startColumn, endLine, endColumn},
			TokenCount:   tokenCount,
		})

		if endByteIdx >= totalBytes {
			break
		}

		// 计算下一个块的起始位置（应用滑动窗口）
		if remaining := totalBytes - endByteIdx; remaining < maxBytes {
			// 最后一块，调整重叠量
			startByteIdx = endByteIdx - (maxBytes - remaining)
		} else {
			// 正常情况，使用固定重叠
			startByteIdx = endByteIdx - overlapBytes
		}

		// 防止索引越界
		if startByteIdx < 0 {
			startByteIdx = 0
		}
	}

	return chunks
}

// calculateColumn 根据字节偏移量计算在当前行的列位置
func calculateColumn(content string, byteOffset int) int {
	if byteOffset >= len(content) {
		byteOffset = len(content) - 1
	}
	if byteOffset < 0 {
		return 0
	}

	// 从字节偏移量向前查找最后一个换行符
	column := 0
	for i := byteOffset; i >= 0; i-- {
		if content[i] == '\n' {
			break
		}
		column++
	}
	return column
}

// countLines 计算字符串中的行数
func countLines(s string) int {
	if len(s) == 0 {
		return 0
	}

	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}

	// 最后一行可能没有换行符
	if len(s) > 0 && s[len(s)-1] != '\n' {
		count++
	}

	return count
}

// splitMarkdownFile 将 markdown 文件分割成多个代码块
func (p *CodeSplitter) splitMarkdownFile(codeFile *types.SourceFile) ([]*types.CodeChunk, error) {
	content := string(codeFile.Content)
	lines := strings.Split(content, "\n")

	var chunks []*types.CodeChunk
	var currentChunk strings.Builder
	var currentLine int
	var inCodeBlock bool

	for i, line := range lines {
		// 检查是否是代码块开始或结束
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// 代码块结束
				currentChunk.WriteString(line + "\n")
				chunkContent := currentChunk.String()
				tokenCount := p.countToken([]byte(chunkContent))

				chunks = append(chunks, &types.CodeChunk{
					Language:     LanguageTypeDoc,
					CodebaseId:   codeFile.CodebaseId,
					CodebasePath: codeFile.CodebasePath,
					CodebaseName: codeFile.CodebaseName,
					Content:      []byte(chunkContent),
					FilePath:     codeFile.Path,
					Range:        []int{currentLine, 0, i, len(line)},
					TokenCount:   tokenCount,
				})

				currentChunk.Reset()
				inCodeBlock = false
			} else {
				// 代码块开始，先保存之前的内容
				if currentChunk.Len() > 0 {
					chunkContent := currentChunk.String()
					tokenCount := p.countToken([]byte(chunkContent))

					chunks = append(chunks, &types.CodeChunk{
						Language:     LanguageTypeDoc,
						CodebaseId:   codeFile.CodebaseId,
						CodebasePath: codeFile.CodebasePath,
						CodebaseName: codeFile.CodebaseName,
						Content:      []byte(chunkContent),
						FilePath:     codeFile.Path,
						Range:        []int{currentLine, 0, i - 1, len(lines[i-1])},
						TokenCount:   tokenCount,
					})

					currentChunk.Reset()
				}

				currentChunk.WriteString(line + "\n")
				currentLine = i
				inCodeBlock = true
			}
			continue
		}

		// 检查是否是标题（# ## ### 等）
		if !inCodeBlock && strings.HasPrefix(line, "#") {
			// 保存之前的内容
			if currentChunk.Len() > 0 {
				chunkContent := currentChunk.String()
				tokenCount := p.countToken([]byte(chunkContent))

				chunks = append(chunks, &types.CodeChunk{
					Language:     LanguageTypeDoc,
					CodebaseId:   codeFile.CodebaseId,
					CodebasePath: codeFile.CodebasePath,
					CodebaseName: codeFile.CodebaseName,
					Content:      []byte(chunkContent),
					FilePath:     codeFile.Path,
					Range:        []int{currentLine, 0, i - 1, len(lines[i-1])},
					TokenCount:   tokenCount,
				})

				currentChunk.Reset()
			}

			currentChunk.WriteString(line + "\n")
			currentLine = i
			continue
		}

		// 普通内容
		currentChunk.WriteString(line + "\n")

		// 检查当前块是否超过最大 token 数量
		if currentChunk.Len() > 0 {
			tokenCount := p.countToken([]byte(currentChunk.String()))
			if tokenCount > p.splitOptions.MaxTokensPerChunk {
				chunkContent := currentChunk.String()
				subChunks := p.splitFuncWithSlidingWindow(chunkContent, codeFile, currentLine, LanguageTypeDoc)
				chunks = append(chunks, subChunks...)
				currentChunk.Reset()
				currentLine = i + 1
			}
		}
	}

	// 添加最后一块内容
	if currentChunk.Len() > 0 {
		chunkContent := currentChunk.String()
		tokenCount := p.countToken([]byte(chunkContent))

		chunks = append(chunks, &types.CodeChunk{
			Language:     LanguageTypeDoc,
			CodebaseId:   codeFile.CodebaseId,
			CodebasePath: codeFile.CodebasePath,
			CodebaseName: codeFile.CodebaseName,
			Content:      []byte(chunkContent),
			FilePath:     codeFile.Path,
			Range:        []int{currentLine, 0, len(lines) - 1, len(lines[len(lines)-1])},
			TokenCount:   tokenCount,
		})
	}

	return chunks, nil
}

// splitMarkdownFileBySitter 使用 tree-sitter 解析 markdown 文件并分割成多个代码块
func (p *CodeSplitter) splitMarkdownFileBySitter(codeFile *types.SourceFile) ([]*types.CodeChunk, error) {
	source := string(codeFile.Content)

	// 创建 markdown 解析器
	parser := sitter.NewParser()
	defer parser.Close()

	// 设置 markdown 语言
	markdownLanguage := tree_sitter_markdown.Language()
	if err := parser.SetLanguage(sitter.NewLanguage(markdownLanguage)); err != nil {
		return nil, fmt.Errorf("failed to set markdown language: %w", err)
	}

	// 解析 markdown 内容
	tree := parser.Parse(codeFile.Content, nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse markdown: %s", codeFile.Path)
	}
	defer tree.Close()

	rootNode := tree.RootNode()

	// 收集所有标题节点并按位置排序
	var allHeaders []*sitter.Node
	collectAllHeaders(rootNode, &allHeaders)
	allHeaders = sortHeadersByPosition(allHeaders)

	var chunks []*types.CodeChunk

	// 如果没有标题节点，返回空切片
	if len(allHeaders) == 0 {
		byteCount := len(codeFile.Content)
		if byteCount > p.splitOptions.MaxTokensPerChunk*4 {
			subChunks := p.splitTextWithSlidingWindow(source, codeFile, 0, LanguageTypeDoc)
			chunks = append(chunks, subChunks...)
		} else {
			chunk := &types.CodeChunk{
				Language:     LanguageTypeDoc,
				CodebaseId:   codeFile.CodebaseId,
				CodebasePath: codeFile.CodebasePath,
				CodebaseName: codeFile.CodebaseName,
				Content:      codeFile.Content,
				FilePath:     codeFile.Path,
				Range:        []int{0, 0, countLines(source) - 1, calculateColumn(source, len(source)-1)},
				TokenCount:   byteCount / 4,
			}
			chunks = append(chunks, chunk)
		}
		return chunks, nil
	}

	// 为每个标题节点直接创建代码块
	for i, header := range allHeaders {
		// 获取当前标题的完整路径
		headerPath := getHeaderPath(header, source, allHeaders)

		// 查找下一个标题节点
		var nextHeader *sitter.Node
		if i < len(allHeaders)-1 {
			nextHeader = allHeaders[i+1]
		}

		// 提取内容（不包含标题路径）
		content := extractContentBetweenHeaders(header, nextHeader, source)

		// 计算位置信息
		startLine := int(header.EndPosition().Row)
		startCol := int(header.EndPosition().Column)
		var endLine, endCol int
		if nextHeader != nil {
			endLine = int(nextHeader.StartPosition().Row)
			endCol = int(nextHeader.StartPosition().Column)
		} else {
			endLine = int(rootNode.EndPosition().Row)
			endCol = int(rootNode.EndPosition().Column)
		}

		// 计算内容的token数量
		tokenCount := p.countToken([]byte(content))

		// 处理超过最大 token 数量的情况
		if tokenCount > p.splitOptions.MaxTokensPerChunk {
			// 使用滑动窗口分割大块内容
			subChunks := p.splitTextWithSlidingWindow(content, codeFile, startLine, LanguageTypeDoc)

			// 为每个子块添加标题路径
			for _, subChunk := range subChunks {
				// 构建完整的内容（标题路径 + 子块内容）
				var fullContent strings.Builder

				// 添加标题路径
				for _, header := range headerPath {
					fullContent.WriteString(header + "\n")
				}

				// 添加子块内容
				fullContent.WriteString(string(subChunk.Content))

				// 更新子块的内容和token数量
				subChunk.Content = []byte(fullContent.String())
				subChunk.TokenCount = p.countToken(subChunk.Content)

				chunks = append(chunks, subChunk)
			}
		} else {
			// 构建完整的内容（标题路径 + 内容）
			var fullContent strings.Builder

			// 添加标题路径
			for _, header := range headerPath {
				fullContent.WriteString(header + "\n")
			}

			// 添加内容
			if content != "" {
				fullContent.WriteString(content)
			}

			contentStr := fullContent.String()
			finalTokenCount := p.countToken([]byte(contentStr))

			// 创建代码块
			chunk := &types.CodeChunk{
				Language:     LanguageTypeDoc,
				CodebaseId:   codeFile.CodebaseId,
				CodebasePath: codeFile.CodebasePath,
				CodebaseName: codeFile.CodebaseName,
				Content:      []byte(contentStr),
				FilePath:     codeFile.Path,
				Range:        []int{startLine, startCol, endLine, endCol},
				TokenCount:   finalTokenCount,
			}
			chunks = append(chunks, chunk)
		}
	}

	return chunks, nil
}

// collectAllHeaders 收集所有标题节点
func collectAllHeaders(node *sitter.Node, headers *[]*sitter.Node) {
	// 如果是标题节点，添加到列表中
	if node.Kind() == "atx_heading" {
		*headers = append(*headers, node)
	}

	// 递归处理子节点
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		collectAllHeaders(child, headers)
	}
}

// sortHeadersByPosition 按位置排序标题节点
func sortHeadersByPosition(headers []*sitter.Node) []*sitter.Node {
	for i := 0; i < len(headers)-1; i++ {
		for j := i + 1; j < len(headers); j++ {
			if headers[i].StartByte() > headers[j].StartByte() {
				headers[i], headers[j] = headers[j], headers[i]
			}
		}
	}
	return headers
}

// getHeaderPath 获取标题的完整路径
func getHeaderPath(headerNode *sitter.Node, source string, allHeaders []*sitter.Node) []string {
	var path []string

	// 获取当前标题在文档中的位置
	currentStartByte := headerNode.StartByte()

	// 找到当前标题在排序后列表中的位置
	var currentIndex int
	for i, header := range allHeaders {
		if header.StartByte() == currentStartByte {
			currentIndex = i
			break
		}
	}

	// 构建标题路径
	var currentLevel int
	currentTitle := extractTitleText(headerNode, source)
	if currentTitle != "" {
		// 获取当前标题的级别
		currentLevel = getHeaderLevel(headerNode)
		// 当前标题路径记录在path中，在后续content中不包含
		path = append(path, currentTitle)
	}

	// 向前查找父级标题
	for i := currentIndex - 1; i >= 0; i-- {
		prevHeader := allHeaders[i]
		prevLevel := getHeaderLevel(prevHeader)

		// 如果前一个标题的级别小于当前标题的级别，则是父级标题
		if prevLevel < currentLevel {
			prevTitle := extractTitleText(prevHeader, source)
			if prevTitle != "" {
				path = append([]string{prevTitle}, path...)
				currentLevel = prevLevel
			}
		}
	}

	return path
}

// extractTitleText 提取标题文本
func extractTitleText(node *sitter.Node, source string) string {
	if node.Kind() != "atx_heading" {
		return ""
	}

	// 直接获取 atx_heading 节点的文本内容，保留 # 符号
	return strings.TrimSpace(source[node.StartByte():node.EndByte()])
}

// getHeaderLevel 获取标题级别
func getHeaderLevel(headerNode *sitter.Node) int {
	if headerNode.Kind() != "atx_heading" {
		return 0
	}

	// 查找 atx_h?_marker 子节点来确定级别
	for i := uint(0); i < headerNode.ChildCount(); i++ {
		child := headerNode.Child(i)
		if strings.HasPrefix(child.Kind(), "atx_h") && strings.HasSuffix(child.Kind(), "_marker") {
			// 从 atx_h1_marker 这样的字符串中提取数字
			levelStr := strings.TrimPrefix(child.Kind(), "atx_h")
			levelStr = strings.TrimSuffix(levelStr, "_marker")
			if level, err := strconv.Atoi(levelStr); err == nil {
				return level
			}
		}
	}

	return 0
}

// extractContentBetweenHeaders 提取两个标题之间的内容
func extractContentBetweenHeaders(currentHeader, nextHeader *sitter.Node, source string) string {
	startPos := int(currentHeader.EndByte())

	var endPos int
	if nextHeader != nil {
		endPos = int(nextHeader.StartByte())
	} else {
		// 如果没有下一个标题，返回文档末尾
		endPos = len(source)
	}

	if startPos >= endPos {
		return ""
	}

	// 提取内容
	content := source[startPos:endPos]

	return content
}

type APIVersion string

const (
	OpenAPI3 APIVersion = "openapi3"
	Swagger2 APIVersion = "swagger2"
	Unknown  APIVersion = "unknown"
)

// splitOpenAPIFile 将 OpenAPI 文件的Paths分割成多个新的OpenAPI文件
func (p *CodeSplitter) splitOpenAPIFile(codeFile *types.SourceFile) ([]*types.CodeChunk, error) {
	// 1. 验证 OpenAPI 规范版本
	version, err := p.validateOpenAPISpec(codeFile.Content, codeFile.Path)
	if err != nil {
		return nil, parser.ErrInvalidOpenAPISpec
	}

	// 2. 根据版本解析文档
	var chunks []*types.CodeChunk

	switch version {
	case OpenAPI3:
		chunks, err = p.splitOpenAPI3File(codeFile)
	case Swagger2:
		chunks, err = p.splitSwagger2File(codeFile)
	default:
		return nil, parser.ErrInvalidOpenAPISpec
	}
	if err != nil {
		return nil, parser.ErrInvalidOpenAPISpec
	}

	return chunks, nil
}

// validateOpenAPISpec 验证 OpenAPI 规范版本
func (p *CodeSplitter) validateOpenAPISpec(data []byte, filePath string) (APIVersion, error) {
	// 根据文件后缀选择解析方式
	var m map[string]any

	// 获取文件后缀
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		// YAML文件
		if err := yaml.Unmarshal(data, &m); err != nil {
			return Unknown, fmt.Errorf("YAML解析失败: %v", err)
		}
	} else if strings.HasSuffix(filePath, ".json") {
		// JSON文件
		if err := json.Unmarshal(data, &m); err != nil {
			return Unknown, fmt.Errorf("JSON解析失败: %v", err)
		}
	} else {
		return Unknown, fmt.Errorf("不是合法的 YAML/JSON: %v", filePath)
	}

	switch {
	case m["openapi"] != nil:
		openapiVersion, ok := m["openapi"].(string)
		if !ok {
			return Unknown, fmt.Errorf("openapi版本字段格式错误")
		}
		if strings.HasPrefix(openapiVersion, "3") {
			return OpenAPI3, nil
		}
		return Unknown, fmt.Errorf("不支持的 OpenAPI 版本: %s", openapiVersion)

	case m["swagger"] != nil:
		swaggerVersion, ok := m["swagger"].(string)
		if !ok {
			return Unknown, fmt.Errorf("swagger版本字段格式错误")
		}
		if strings.HasPrefix(swaggerVersion, "2") {
			return Swagger2, nil
		}
		return Unknown, fmt.Errorf("不支持的 Swagger 版本: %s", swaggerVersion)

	default:
		return Unknown, fmt.Errorf("既不是 openapi 3.x 也不是 swagger 2.0")
	}
}

// splitOpenAPI3File 分割 OpenAPI 3.x 文件
func (p *CodeSplitter) splitOpenAPI3File(codeFile *types.SourceFile) ([]*types.CodeChunk, error) {
	// 解析 OpenAPI 3.x 文档
	loader := openapi3.NewLoader()
	// TODO 是否处理外部引用呢
	loader.IsExternalRefsAllowed = false // 默认状态
	doc, err := loader.LoadFromData(codeFile.Content)
	if err != nil {
		return nil, fmt.Errorf("openapi3 解析失败: %v", err)
	}

	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("openapi3 验证失败: %v", err)
	}

	var chunks []*types.CodeChunk

	// 为每个路径创建单独的文档
	for _, path := range doc.Paths.InMatchingOrder() {
		pathItem := doc.Paths.Find(path)

		// 创建新的文档副本
		newDoc := &openapi3.T{
			OpenAPI:      doc.OpenAPI,
			Info:         doc.Info,
			Servers:      doc.Servers,
			Paths:        openapi3.NewPaths(openapi3.WithPath(path, pathItem)),
			Components:   doc.Components, // 保留所有组件
			Security:     doc.Security,
			Tags:         doc.Tags,
			ExternalDocs: doc.ExternalDocs,
		}

		// 更新文档标题以包含路径信息
		newDoc.Info.Title = fmt.Sprintf("%s - %s", doc.Info.Title, path)

		// 序列化新的文档
		docBytes, err := json.Marshal(newDoc)
		if err != nil {
			return nil, fmt.Errorf("序列化 OpenAPI 3.x 文档失败: %v", err)
		}

		// 计算 token 数量
		tokenCount := p.countToken(docBytes)

		// 创建代码块
		chunk := &types.CodeChunk{
			Language:     LanguageTypeDoc,
			CodebaseId:   codeFile.CodebaseId,
			CodebasePath: codeFile.CodebasePath,
			CodebaseName: codeFile.CodebaseName,
			Content:      docBytes,
			FilePath:     codeFile.Path,
			Range:        []int{0, 0, 0, 0}, // OpenAPI 分割不涉及行号
			TokenCount:   tokenCount,
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// splitSwagger2File 分割 Swagger 2.0 文件
func (p *CodeSplitter) splitSwagger2File(codeFile *types.SourceFile) ([]*types.CodeChunk, error) {
	// 解析 Swagger 2.0 文档
	var doc openapi2.T
	if err := json.Unmarshal(codeFile.Content, &doc); err != nil {
		return nil, fmt.Errorf("swagger2 解析失败: %v", err)
	}

	// 验证 Swagger 2.0 文档
	if err := p.validateSwagger2Doc(&doc); err != nil {
		return nil, fmt.Errorf("swagger2 验证失败: %v", err)
	}

	var chunks []*types.CodeChunk

	// 为每个路径创建单独的文档
	for path, pathItem := range doc.Paths {
		// 创建新的文档副本
		newDoc := &openapi2.T{
			Swagger:             doc.Swagger,
			Info:                doc.Info,
			Host:                doc.Host,
			BasePath:            doc.BasePath,
			Schemes:             doc.Schemes,
			Consumes:            doc.Consumes,
			Produces:            doc.Produces,
			Paths:               make(map[string]*openapi2.PathItem),
			Definitions:         doc.Definitions, // 保留所有定义
			Parameters:          doc.Parameters,
			Responses:           doc.Responses,
			Security:            doc.Security,
			SecurityDefinitions: doc.SecurityDefinitions,
			Tags:                doc.Tags,
			ExternalDocs:        doc.ExternalDocs,
		}

		// 只添加当前路径
		newDoc.Paths[path] = pathItem
		// 更新文档标题以包含路径信息
		newDoc.Info.Title = fmt.Sprintf("%s - %s", doc.Info.Title, path)

		// 序列化新的文档
		docBytes, err := json.Marshal(newDoc)
		if err != nil {
			return nil, fmt.Errorf("序列化 Swagger 2.0 文档失败: %v", err)
		}

		// 计算 token 数量
		tokenCount := p.countToken(docBytes)

		// 创建代码块
		chunk := &types.CodeChunk{
			Language:     LanguageTypeDoc,
			CodebaseId:   codeFile.CodebaseId,
			CodebasePath: codeFile.CodebasePath,
			CodebaseName: codeFile.CodebaseName,
			Content:      docBytes,
			FilePath:     codeFile.Path,
			Range:        []int{0, 0, 0, 0}, // Swagger 分割不涉及行号
			TokenCount:   tokenCount,
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// validateSwagger2Doc 验证 Swagger 2.0 文档
func (p *CodeSplitter) validateSwagger2Doc(doc *openapi2.T) error {
	// 检查必要字段
	if doc.Info.Title == "" {
		return fmt.Errorf("info.title 不能为空")
	}
	if doc.Info.Version == "" {
		return fmt.Errorf("info.version 不能为空")
	}

	// 检查路径
	if doc.Paths == nil {
		return fmt.Errorf("缺少 paths 字段")
	}
	if len(doc.Paths) == 0 {
		return fmt.Errorf("paths 不能为空")
	}

	// 检查每个路径的操作
	for path, pathItem := range doc.Paths {
		if pathItem == nil {
			return fmt.Errorf("路径 %s 的 pathItem 不能为空", path)
		}

		// 检查是否有至少一个操作
		hasOperation := pathItem.Get != nil || pathItem.Post != nil ||
			pathItem.Put != nil || pathItem.Delete != nil ||
			pathItem.Patch != nil || pathItem.Head != nil ||
			pathItem.Options != nil

		if !hasOperation {
			return fmt.Errorf("路径 %s 必须包含至少一个操作", path)
		}
	}

	// 检查定义
	if doc.Definitions != nil {
		for name, schema := range doc.Definitions {
			if schema == nil {
				return fmt.Errorf("定义 %s 不能为空", name)
			}
		}
	}

	return nil
}
