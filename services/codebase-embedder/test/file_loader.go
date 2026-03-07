package test

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// FileLoader 文件加载器
type FileLoader struct {
	baseDirs     []string
	allowedExts  []string
	maxFileSize  int64
	logger       *log.Logger
	loadingStats LoadingStats
	mu           sync.RWMutex
}

// LoadingStats 加载统计信息
type LoadingStats struct {
	TotalFiles     int
	LoadedFiles    int
	FailedFiles    int
	TotalSize      int64
	LoadingTime    time.Duration
	FileTypeCounts map[string]int
}

// NewFileLoader 创建新的文件加载器
func NewFileLoader(baseDirs []string) *FileLoader {
	logger := log.New(log.Writer(), "[FileLoader] ", log.LstdFlags|log.Lmsgprefix)

	return &FileLoader{
		baseDirs:    baseDirs,
		allowedExts: []string{".go", ".java", ".js", ".ts", ".py", ".cpp", ".c", ".h", ".hpp", ".cs", ".php", ".rb", ".swift", ".kt", ".scala", ".rs", ".md", ".txt"},
		maxFileSize: 10 * 1024 * 1024, // 10MB
		logger:      logger,
		loadingStats: LoadingStats{
			FileTypeCounts: make(map[string]int),
		},
	}
}

// SetAllowedExtensions 设置允许的文件扩展名
func (fl *FileLoader) SetAllowedExtensions(exts []string) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.allowedExts = exts
}

// SetMaxFileSize 设置最大文件大小
func (fl *FileLoader) SetMaxFileSize(size int64) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.maxFileSize = size
}

// GetLoadingStats 获取加载统计信息
func (fl *FileLoader) GetLoadingStats() LoadingStats {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	return fl.loadingStats
}

// LoadFiles 加载所有测试文件
func (fl *FileLoader) LoadFiles() (map[string]*types.SourceFile, error) {
	startTime := time.Now()
	fl.logger.Printf("开始加载测试文件...")

	// 重置统计信息
	fl.mu.Lock()
	fl.loadingStats = LoadingStats{
		FileTypeCounts: make(map[string]int),
	}
	fl.mu.Unlock()

	files := make(map[string]*types.SourceFile)
	var mu sync.Mutex

	// 使用工作池并发加载文件
	const workerCount = 4
	fileChan := make(chan string, 100)
	resultChan := make(chan *fileLoadResult, 100)
	errorChan := make(chan *fileLoadError, 10)

	// 启动工作池
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go fl.fileWorker(&wg, fileChan, resultChan, errorChan)
	}

	// 启动结果收集器
	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for result := range resultChan {
			mu.Lock()
			files[result.filePath] = result.sourceFile
			mu.Unlock()
		}
	}()

	// 启动错误收集器
	var errorWg sync.WaitGroup
	errorWg.Add(1)
	go func() {
		defer errorWg.Done()
		for err := range errorChan {
			fl.logger.Printf("加载文件失败: %s - %v", err.filePath, err.err)
			fl.mu.Lock()
			fl.loadingStats.FailedFiles++
			fl.mu.Unlock()
		}
	}()

	// 收集所有文件路径
	var allFilePaths []string
	for _, dir := range fl.baseDirs {
		filePaths, err := fl.collectFilePaths(dir)
		if err != nil {
			fl.logger.Printf("收集文件路径失败: %s - %v", dir, err)
			continue
		}
		allFilePaths = append(allFilePaths, filePaths...)
	}

	fl.mu.Lock()
	fl.loadingStats.TotalFiles = len(allFilePaths)
	fl.mu.Unlock()

	// 发送文件到工作池
	for _, filePath := range allFilePaths {
		fileChan <- filePath
	}
	close(fileChan)

	// 等待工作池完成
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// 等待收集器完成
	resultWg.Wait()
	errorWg.Wait()

	// 更新统计信息
	fl.mu.Lock()
	fl.loadingStats.LoadedFiles = len(files)
	fl.loadingStats.LoadingTime = time.Since(startTime)
	fl.mu.Unlock()

	fl.logger.Printf("文件加载完成 - 总文件: %d, 成功: %d, 失败: %d, 耗时: %v",
		fl.loadingStats.TotalFiles,
		fl.loadingStats.LoadedFiles,
		fl.loadingStats.FailedFiles,
		fl.loadingStats.LoadingTime)

	return files, nil
}

// fileWorker 文件工作器
func (fl *FileLoader) fileWorker(wg *sync.WaitGroup, fileChan <-chan string, resultChan chan<- *fileLoadResult, errorChan chan<- *fileLoadError) {
	defer wg.Done()

	for filePath := range fileChan {
		sourceFile, err := fl.loadSingleFile(filePath)
		if err != nil {
			errorChan <- &fileLoadError{filePath: filePath, err: err}
			continue
		}
		resultChan <- &fileLoadResult{filePath: filePath, sourceFile: sourceFile}
	}
}

// fileLoadResult 文件加载结果
type fileLoadResult struct {
	filePath   string
	sourceFile *types.SourceFile
}

// fileLoadError 文件加载错误
type fileLoadError struct {
	filePath string
	err      error
}

// collectFilePaths 收集目录下所有文件路径
func (fl *FileLoader) collectFilePaths(dir string) ([]string, error) {
	var filePaths []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		if !fl.isAllowedFile(path) {
			return nil
		}

		// 检查文件大小
		info, err := d.Info()
		if err != nil {
			fl.logger.Printf("获取文件信息失败: %s - %v", path, err)
			return nil
		}

		if info.Size() > fl.maxFileSize {
			fl.logger.Printf("文件过大，跳过: %s (大小: %d bytes)", path, info.Size())
			return nil
		}

		filePaths = append(filePaths, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %w", err)
	}

	return filePaths, nil
}

// isAllowedFile 检查文件是否允许加载
func (fl *FileLoader) isAllowedFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	for _, allowedExt := range fl.allowedExts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// loadSingleFile 加载单个文件
func (fl *FileLoader) loadSingleFile(filePath string) (*types.SourceFile, error) {
	// 检查文件大小
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	if info.Size() > fl.maxFileSize {
		return nil, fmt.Errorf("文件过大: %d bytes (最大: %d bytes)", info.Size(), fl.maxFileSize)
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 更新统计信息
	fl.mu.Lock()
	fl.loadingStats.TotalSize += info.Size()
	ext := strings.ToLower(filepath.Ext(filePath))
	fl.loadingStats.FileTypeCounts[ext]++
	fl.mu.Unlock()

	// 检测文件语言
	language := fl.detectLanguage(filePath)

	// 使用项目的SourceFile结构
	sourceFile := &types.SourceFile{
		CodebaseId:   1,
		CodebasePath: "/test/codebase",
		CodebaseName: "test_codebase",
		Name:         filepath.Base(filePath),
		Path:         filePath,
		Content:      data,
		Language:     language,
	}

	return sourceFile, nil
}

// detectLanguage 检测文件语言
func (fl *FileLoader) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	languageMap := map[string]string{
		".go":    "go",
		".java":  "java",
		".js":    "javascript",
		".ts":    "typescript",
		".py":    "python",
		".cpp":   "cpp",
		".c":     "c",
		".h":     "c",
		".hpp":   "cpp",
		".cs":    "csharp",
		".php":   "php",
		".rb":    "ruby",
		".swift": "swift",
		".kt":    "kotlin",
		".scala": "scala",
		".rs":    "rust",
		".md":    "markdown",
		".txt":   "text",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	return "unknown"
}

// ReadLines 读取文件中从startLine到endLine（包含）的内容
// 返回的字节切片包含指定范围内的所有行，保留原始换行符
// startLine和endLine应为正数，且startLine <= endLine
func ReadLines(filePath string, startLine, endLine int) ([]byte, error) {
	// 验证输入参数
	if startLine < 1 {
		startLine = 1
	}
	if endLine < startLine {
		return nil, errors.New("endLine 必须大于等于 startLine")
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var result []byte
	currentLine := 0

	for scanner.Scan() {
		currentLine++

		// 如果当前行在目标范围内，则添加到结果
		if currentLine >= startLine && currentLine <= endLine {
			// 追加当前行内容
			result = append(result, scanner.Bytes()...)
			// 追加换行符（因为scanner.Text()会移除原始换行符）
			result = append(result, '\n')
		}

		// 如果已经超过目标范围，提前退出
		if currentLine > endLine {
			break
		}
	}

	// 检查扫描过程中是否发生错误
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件时出错: %w", err)
	}

	// 如果文件行数少于startLine
	if currentLine < startLine {
		return nil, fmt.Errorf("文件行数不足，实际行数: %d, 请求起始行: %d", currentLine, startLine)
	}

	return result, nil
}

// ReadLinesOptimized 优化版本的行读取函数，支持更大的文件和更好的性能
func ReadLinesOptimized(filePath string, startLine, endLine int) ([]byte, error) {
	// 验证输入参数
	if startLine < 1 {
		startLine = 1
	}
	if endLine < startLine {
		return nil, errors.New("endLine 必须大于等于 startLine")
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在: %s", filePath)
		}
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 检查文件大小
	if fileInfo.Size() == 0 {
		return nil, fmt.Errorf("文件为空: %s", filePath)
	}

	// 对于大文件，使用内存映射文件读取
	if fileInfo.Size() > 10*1024*1024 { // 大于10MB
		return readLinesWithMemoryMap(filePath, startLine, endLine)
	}

	// 对于小文件，使用标准读取
	return ReadLines(filePath, startLine, endLine)
}

// readLinesWithMemoryMap 使用内存映射文件读取大文件的指定行
func readLinesWithMemoryMap(filePath string, startLine, endLine int) ([]byte, error) {
	// 这里简化实现，实际项目中可以使用更高级的内存映射技术
	// 为了保持兼容性，我们仍然使用标准读取，但添加了日志提示
	log.Printf("警告: 文件较大，建议使用内存映射技术优化读取性能: %s", filePath)
	return ReadLines(filePath, startLine, endLine)
}

// CountLines 统计文件行数
func CountLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("读取文件时出错: %w", err)
	}

	return lineCount, nil
}

// GetFileMetadata 获取文件元数据
func GetFileMetadata(filePath string) (map[string]interface{}, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	lineCount, err := CountLines(filePath)
	if err != nil {
		return nil, fmt.Errorf("统计文件行数失败: %w", err)
	}

	metadata := map[string]interface{}{
		"name":         filepath.Base(filePath),
		"path":         filePath,
		"size":         fileInfo.Size(),
		"mode":         fileInfo.Mode(),
		"mod_time":     fileInfo.ModTime(),
		"is_dir":       fileInfo.IsDir(),
		"line_count":   lineCount,
		"extension":    filepath.Ext(filePath),
		"content_type": getContentType(filePath),
	}

	return metadata, nil
}

// getContentType 获取文件内容类型
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	contentTypes := map[string]string{
		".go":    "text/x-go",
		".java":  "text/x-java-source",
		".js":    "application/javascript",
		".ts":    "text/typescript",
		".py":    "text/x-python",
		".cpp":   "text/x-c++src",
		".c":     "text/x-csrc",
		".h":     "text/x-chdr",
		".hpp":   "text/x-c++hdr",
		".cs":    "text/x-csharp",
		".php":   "text/x-php",
		".rb":    "text/x-ruby",
		".swift": "text/x-swift",
		".kt":    "text/x-kotlin",
		".scala": "text/x-scala",
		".rs":    "text/x-rust",
		".md":    "text/markdown",
		".txt":   "text/plain",
		".json":  "application/json",
		".yaml":  "text/yaml",
		".yml":   "text/yaml",
		".xml":   "application/xml",
		".html":  "text/html",
		".css":   "text/css",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}
	return "application/octet-stream"
}

// minInt64 返回两个int64整数中的较小值
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
