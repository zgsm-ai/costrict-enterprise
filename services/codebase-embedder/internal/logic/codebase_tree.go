package logic

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/zgsm-ai/codebase-indexer/internal/errs"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type CodebaseTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCodebaseTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CodebaseTreeLogic {
	return &CodebaseTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CodebaseTreeLogic) GetCodebaseTree(req *types.CodebaseTreeRequest) (*types.CodebaseTreeResponse, error) {
	// 参数验证
	if err := l.validateRequest(req); err != nil {
		log.Printf("[DEBUG] 参数验证失败: %v", err)
		return nil, errs.FileNotFound
	}

	// 构建目录树
	log.Printf("[DEBUG] 开始构建目录树...")
	tree, err := l.buildDirectoryTree(req.ClientId, req)
	if err != nil {
		log.Printf("[DEBUG] 构建目录树失败: %v", err)
		return nil, fmt.Errorf("构建目录树失败: %w", err)
	}

	log.Printf("[DEBUG] 目录树构建完成，最终结果:")
	if tree != nil {
		// 调用独立的树结构打印函数
		l.printTreeStructure(tree)
	} else {
		log.Printf("[DEBUG] 警告: 构建的树为空")
	}

	log.Printf("[DEBUG] ===== GetCodebaseTree 执行完成 =====")
	return &types.CodebaseTreeResponse{
		Code:    0,
		Message: "ok",
		Success: true,
		Data:    tree,
	}, nil
}

func (l *CodebaseTreeLogic) validateRequest(req *types.CodebaseTreeRequest) error {
	if req.ClientId == "" {
		return fmt.Errorf("缺少必需参数: clientId")
	}
	if req.CodebasePath == "" {
		return fmt.Errorf("缺少必需参数: codebasePath")
	}
	if req.CodebaseName == "" {
		return fmt.Errorf("缺少必需参数: codebaseName")
	}
	return nil
}

// printTreeStructure 递归打印树结构
func (l *CodebaseTreeLogic) printTreeStructure(tree *types.TreeNode) {
	// 递归打印树结构
	var printTree func(node *types.TreeNode, indent string)
	printTree = func(node *types.TreeNode, indent string) {
		log.Printf("[DEBUG] %s├── %s (%s) - 子节点数: %d", indent, node.Name, node.Type, len(node.Children))
		for i := range node.Children {
			newIndent := indent + "│  "
			if i == len(node.Children)-1 {
				newIndent = indent + "   "
			}
			printTree(node.Children[i], newIndent)
		}
	}
	printTree(tree, "")
}

func (l *CodebaseTreeLogic) buildDirectoryTree(clientId string, req *types.CodebaseTreeRequest) (*types.TreeNode, error) {
	// 从向量存储中获取文件路径
	records, err := l.getRecordsFromVectorStore(clientId, req.CodebasePath)
	if err != nil {
		return nil, err
	}

	// 分析记录并提取文件路径
	filePaths, err := l.analyzeRecordsAndExtractPaths(records)
	if err != nil {
		return nil, err
	}

	// 设置构建参数
	maxDepth, includeFiles := l.buildTreeParameters(req)

	result, err := BuildDirectoryTree(filePaths, maxDepth, includeFiles)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// checkCodebaseInDatabase 检查数据库中是否存在该 codebaseId
func (l *CodebaseTreeLogic) checkCodebaseInDatabase(codebaseId int32) {
	log.Printf("[DEBUG] 检查数据库中是否存在 codebaseId: %d", codebaseId)
	codebase, err := l.svcCtx.Querier.Codebase.WithContext(l.ctx).Where(l.svcCtx.Querier.Codebase.ID.Eq(codebaseId)).First()
	if err != nil {
		log.Printf("[DEBUG] 数据库中未找到 codebaseId %d: %v", codebaseId, err)
	} else {
		log.Printf("[DEBUG] 数据库中找到 codebase 记录 - ID: %d, Name: %s, Path: %s, Status: %s",
			codebase.ID, codebase.Name, codebase.Path, codebase.Status)
	}
}

// getRecordsFromVectorStore 从向量存储中获取文件记录
func (l *CodebaseTreeLogic) getRecordsFromVectorStore(clientId string, codebasePath string) ([]*types.CodebaseRecord, error) {
	if l.svcCtx.VectorStore == nil {
		return nil, fmt.Errorf("VectorStore 未初始化")
	}

	records, err := l.svcCtx.VectorStore.GetCodebaseRecords(l.ctx, clientId, codebasePath)
	if err != nil {
		return nil, fmt.Errorf("查询文件路径失败: %w", err)
	}

	// 合并相同文件路径的记录
	mergedRecords, _ := l.mergeRecordsByFilePath(records)

	return mergedRecords, nil
}

// mergeRecordsByFilePath 合并相同文件路径的记录
func (l *CodebaseTreeLogic) mergeRecordsByFilePath(records []*types.CodebaseRecord) ([]*types.CodebaseRecord, int) {
	// 使用 map 按文件路径分组
	filePathMap := make(map[string][]*types.CodebaseRecord)

	for _, record := range records {
		filePathMap[record.FilePath] = append(filePathMap[record.FilePath], record)
	}

	// 合并重复路径的记录
	var mergedRecords []*types.CodebaseRecord
	mergeCount := 0

	for _, fileRecords := range filePathMap {
		if len(fileRecords) == 1 {
			// 没有重复，直接添加
			mergedRecords = append(mergedRecords, fileRecords[0])
		} else {
			// 有重复，合并记录
			mergedRecord := l.mergeSingleFileRecords(fileRecords)
			mergedRecords = append(mergedRecords, mergedRecord)
			mergeCount += len(fileRecords) - 1
		}
	}

	return mergedRecords, mergeCount
}

// mergeSingleFileRecords 合并单个文件的多条记录
func (l *CodebaseTreeLogic) mergeSingleFileRecords(records []*types.CodebaseRecord) *types.CodebaseRecord {
	if len(records) == 0 {
		return nil
	}

	// 以第一条记录为基础
	baseRecord := records[0]

	// 合并内容
	var mergedContent strings.Builder
	var totalTokens int
	var allRanges []int

	for _, record := range records {
		mergedContent.WriteString(record.Content)
		totalTokens += record.TokenCount
		allRanges = append(allRanges, record.Range...)
	}

	// 创建合并后的记录
	mergedRecord := &types.CodebaseRecord{
		Id:          baseRecord.Id,
		FilePath:    baseRecord.FilePath,
		Language:    baseRecord.Language,
		Content:     mergedContent.String(),
		TokenCount:  totalTokens,
		LastUpdated: baseRecord.LastUpdated,
	}

	// 合并范围信息（简单连接，可能需要更复杂的逻辑）
	if len(allRanges) > 0 {
		mergedRecord.Range = allRanges
	}

	return mergedRecord
}

// analyzeRecordsAndExtractPaths 分析记录并提取文件路径
func (l *CodebaseTreeLogic) analyzeRecordsAndExtractPaths(records []*types.CodebaseRecord) ([]string, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("没有记录可供分析")
	}

	// 提取文件路径
	var filePaths []string
	for _, record := range records {
		filePaths = append(filePaths, record.FilePath)
	}

	// 添加调试：检查是否有重复的文件路径
	pathCount := make(map[string]int)
	for _, path := range filePaths {
		pathCount[path]++
	}

	return filePaths, nil
}

// buildTreeParameters 设置构建参数
func (l *CodebaseTreeLogic) buildTreeParameters(req *types.CodebaseTreeRequest) (int, bool) {
	// 设置默认值
	maxDepth := 10 // 默认最大深度
	if req.MaxDepth != nil {
		maxDepth = *req.MaxDepth
	}

	includeFiles := true // 默认包含文件
	if req.IncludeFiles != nil {
		includeFiles = *req.IncludeFiles
	}

	log.Printf("[DEBUG] 目录树构建参数:")
	log.Printf("[DEBUG]   maxDepth: %d (请求值: %v)", maxDepth, req.MaxDepth)
	log.Printf("[DEBUG]   includeFiles: %v (请求值: %v)", includeFiles, req.IncludeFiles)

	return maxDepth, includeFiles
}

// BuildDirectoryTree 构建目录树
func BuildDirectoryTree(filePaths []string, maxDepth int, includeFiles bool) (*types.TreeNode, error) {

	if len(filePaths) == 0 {
		log.Printf("[DEBUG] ❌ 文件路径列表为空，这是问题的直接原因！")
		return nil, fmt.Errorf("文件路径列表为空")
	}

	// 🔧 修复：在开始处理前对所有路径进行规范化
	normalizedPaths := make([]string, len(filePaths))
	for i, path := range filePaths {
		normalizedPaths[i] = normalizePath(path)
	}
	filePaths = normalizedPaths
	log.Printf("[DEBUG] 🔍 关键诊断：多级路径规范化处理完成")

	// 对文件路径进行去重处理
	uniquePaths := make([]string, 0)
	pathSet := make(map[string]bool)
	duplicateCount := 0

	for _, path := range filePaths {
		if !pathSet[path] {
			pathSet[path] = true
			uniquePaths = append(uniquePaths, path)
		} else {
			duplicateCount++
		}
	}

	// 使用去重后的路径列表
	filePaths = uniquePaths

	// 提取根路径
	rootPath := extractRootPath(filePaths)

	// 🔧 修复：确保根路径也被规范化
	rootPath = normalizePath(rootPath)

	// 处理根路径为空的情况
	if rootPath == "" {
		rootPath = "."
	}

	// 🔧 修复：确保根路径规范化
	rootPath = normalizePath(rootPath)

	root := &types.TreeNode{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		Type:     "directory",
		Children: make([]*types.TreeNode, 0),
	}

	pathMap := make(map[string]*types.TreeNode)
	pathMap[root.Path] = root

	// 添加调试：跟踪文件处理过程
	processedFiles := make(map[string]int)
	skippedFiles := 0
	processedFilesCount := 0

	for _, filePath := range filePaths {
		// 添加调试：跟踪每个文件路径的处理
		processedFiles[filePath]++

		if !includeFiles && !isDirectory(filePath) {
			skippedFiles++
			continue
		}

		// 🔧 修复：改进的相对路径计算逻辑，支持多级路径
		var relativePath string
		if rootPath == "." {
			// 当根路径为 "." 时，不应该去掉任何字符
			relativePath = filePath
		} else {
			// 🔧 修复：确保根路径匹配后再进行截取
			if strings.HasPrefix(filePath, rootPath) {
				// 原有逻辑：去掉根路径部分
				relativePath = filePath[len(rootPath):]
			} else {
				// 🔧 修复：如果文件路径不以根路径开头，可能是路径规范化问题
				// 尝试使用规范化后的路径进行比较
				normalizedFilePath := normalizePath(filePath)
				normalizedRootPath := normalizePath(rootPath)

				if strings.HasPrefix(normalizedFilePath, normalizedRootPath) {
					relativePath = normalizedFilePath[len(normalizedRootPath):]
				} else {
					relativePath = filePath
				}
			}
		}

		// 🔧 修复：更安全地移除开头的分隔符
		if len(relativePath) > 0 {
			firstChar := relativePath[0]
			if firstChar == '/' || firstChar == '\\' {
				relativePath = relativePath[1:]
			}
		}

		currentDepth := strings.Count(relativePath, string(filepath.Separator))

		if maxDepth > 0 && currentDepth > maxDepth {
			skippedFiles++
			continue
		}

		// 🔧 修复：确保所有路径都使用规范化格式
		// 构建路径节点
		dir := normalizePath(filepath.Dir(filePath))

		{
			// 路径组件分析
			pathComponents := strings.Split(filePath, string(filepath.Separator))

			// 给文件创建目录
			mountPath := ""
			currentPath := ""
			for idx, pathComponent := range pathComponents {
				if idx+1 == len(pathComponents) {
					break
				}

				if currentPath == "" {
					currentPath = pathComponent
				} else {
					currentPath = currentPath + "\\" + pathComponent
					currentPath = normalizePath(currentPath)
				}

				// 存在当前路径，则跳过，不创建
				if _, exists := pathMap[currentPath]; exists {
					if mountPath == "" {
						mountPath = pathComponent
					} else {
						mountPath = mountPath + "\\" + pathComponent
						mountPath = normalizePath(mountPath)
					}
					continue
				}

				// 创建目录
				node := &types.TreeNode{
					Name:     filepath.Base(pathComponent),
					Path:     currentPath,
					Type:     "directory",
					Children: make([]*types.TreeNode, 0),
				}
				pathMap[currentPath] = node

				// 挂载目录
				if _, exists := pathMap[mountPath]; exists {
					pathMap[mountPath].Children = append(pathMap[mountPath].Children, node)
				} else {
					pathMap[rootPath].Children = append(pathMap[rootPath].Children, node)
				}
				if mountPath == "" {
					mountPath = pathComponent
				} else {
					mountPath = mountPath + "\\" + pathComponent
					mountPath = normalizePath(mountPath)
				}
			}

		}

		// 添加文件节点
		if includeFiles && !isDirectory(filePath) {
			processedFilesCount++

			fileNode, err := createFileNode(filePath)
			if err != nil {
				continue
			}

			parentFound := false
			var foundParentNode *types.TreeNode
			normalizedDir := normalizePath(dir)

			for path, parentNode := range pathMap {
				if path == normalizedDir {
					foundParentNode = parentNode
					parentFound = true
					break
				}
			}
			if parentFound && foundParentNode != nil {
				foundParentNode.Children = append(foundParentNode.Children, fileNode)
			} else {
				if dir == rootPath {
					root.Children = append(root.Children, fileNode)
				}
			}
		}
	}

	return root, nil
}

// extractRootPath 提取根路径
func extractRootPath(filePaths []string) string {
	if len(filePaths) == 0 {
		return ""
	}

	// 分析路径深度分布（使用规范化后的路径）
	depthAnalysis := make(map[int]int)
	for _, path := range filePaths {
		depth := strings.Count(path, string(filepath.Separator))
		depthAnalysis[depth]++
	}

	if len(filePaths) == 0 {
		return ""
	}

	// 首先分析所有路径的深度，确保找到正确的公共前缀
	minDepth := int(^uint(0) >> 1) // 最大int值
	for _, path := range filePaths {
		depth := strings.Count(path, string(filepath.Separator))
		if depth < minDepth {
			minDepth = depth
		}
	}

	// 使用改进的算法，考虑路径组件的匹配
	commonPrefix := filePaths[0]

	for _, path := range filePaths[1:] {
		newPrefix := findCommonPrefix(commonPrefix, path)

		commonPrefix = newPrefix
		if commonPrefix == "" {
			break
		}
	}

	// 🔧 修复：如果公共前缀不以目录分隔符结尾，找到最后一个分隔符
	lastSeparator := strings.LastIndexAny(commonPrefix, string(filepath.Separator))

	if lastSeparator == -1 {
		// 🔧 修复：对于多级路径，如果没有共同的目录前缀，尝试找到父目录
		// 检查是否所有路径都有相同的第一级目录
		firstComponents := make([]string, len(filePaths))
		allHaveSameFirstComponent := true
		var firstComponent string

		for i, path := range filePaths {
			components := strings.Split(path, string(filepath.Separator))
			if len(components) > 0 {
				if i == 0 {
					firstComponent = components[0]
				} else if components[0] != firstComponent {
					allHaveSameFirstComponent = false
					break
				}
				firstComponents[i] = components[0]
			}
		}

		if allHaveSameFirstComponent && firstComponent != "" {
			return firstComponent
		} else {
			return "."
		}
	}

	rootPath := commonPrefix[:lastSeparator+1]

	// 🔧 修复：确保根路径也被规范化
	rootPath = normalizePath(rootPath)

	return rootPath
}

// findCommonPrefix 找到两个路径的公共前缀
func findCommonPrefix(path1, path2 string) string {
	parts1 := strings.Split(path1, string(filepath.Separator))
	parts2 := strings.Split(path2, string(filepath.Separator))

	var commonParts []string
	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] == parts2[i] {
			commonParts = append(commonParts, parts1[i])
		} else {
			break
		}
	}

	return strings.Join(commonParts, string(filepath.Separator))
}

// isDirectory 判断路径是否为目录
func isDirectory(path string) bool {
	// 简单实现：根据路径末尾是否有分隔符判断
	return strings.HasSuffix(path, string(filepath.Separator)) || strings.HasSuffix(path, "/")
}

// normalizePath 统一路径格式，确保所有路径都使用系统标准分隔符
func normalizePath(path string) string {
	if path == "" {
		return ""
	}

	// 🔧 修复：处理多级路径的特殊情况
	// 首先统一使用 / 作为分隔符进行处理
	unifiedPath := strings.ReplaceAll(path, "\\", "/")

	// 使用 filepath.Clean 进行基本规范化
	cleaned := filepath.Clean(unifiedPath)

	// 再次确保路径使用系统标准的分隔符
	// 在 Windows 上，这会将 / 转换为 \
	// 在 Unix 上，这会将 \ 转换为 /
	normalized := filepath.FromSlash(cleaned)

	// 🔧 修复：确保多级路径的格式一致性
	// 如果路径以分隔符结尾，移除它（除非是根目录）
	if len(normalized) > 1 && (strings.HasSuffix(normalized, "\\") || strings.HasSuffix(normalized, "/")) {
		normalized = normalized[:len(normalized)-1]
	}

	return normalized
}

// createFileNode 创建文件节点
func createFileNode(filePath string) (*types.TreeNode, error) {
	normalizedPath := normalizePath(filePath)

	node := &types.TreeNode{
		Name: filepath.Base(normalizedPath),
		Path: normalizedPath, // 🔧 修复：使用规范化后的路径
		Type: "file",
	}
	return node, nil
}
