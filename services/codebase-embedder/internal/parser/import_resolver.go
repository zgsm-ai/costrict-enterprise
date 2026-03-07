package parser

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
	"github.com/zgsm-ai/codebase-indexer/pkg/utils"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"strings"
)

// 项目基础配置信息
type ProjectConfig struct {
	language   Language            // 项目语言
	SourceRoot string              // 源码根路径（如 java 的 src/main/java）
	Dirs       []string            // 源文件目录（相对于 SourceRoot）
	dirToFiles map[string][]string // 目录到文件列表的索引（完整路径）
	fileSet    map[string]struct{} // 文件路径集合（完整路径）
}

func NewProjectConfig(language Language, sourceRoot string, files []string) *ProjectConfig {
	pc := &ProjectConfig{
		language:   language,
		SourceRoot: sourceRoot,
	}
	pc.buildIndex(files)
	return pc
}

// 构建目录和文件索引
func (c *ProjectConfig) buildIndex(files []string) {
	c.dirToFiles = make(map[string][]string)
	c.fileSet = make(map[string]struct{})
	dirSet := make(map[string]struct{})
	if files == nil {
		return
	}
	for _, f := range files {
		dir := utils.ToUnixPath(filepath.Dir(f))
		c.dirToFiles[dir] = append(c.dirToFiles[dir], f)
		c.fileSet[f] = struct{}{}
		dirSet[dir] = struct{}{}
	}

	// 提取相对于 SourceRoot 的目录
	c.Dirs = make([]string, 0, len(dirSet))
	for dir := range dirSet {
		// 计算相对于 SourceRoot 的路径
		c.Dirs = append(c.Dirs, dir)
	}
}

// 导入解析器接口
type ImportResolver interface {
	Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error
}

// 解析器管理器
type ResolverManager struct {
	resolvers map[Language]ImportResolver
}

// 新建解析器管理器
func NewResolverManager() *ResolverManager {
	manager := &ResolverManager{
		resolvers: make(map[Language]ImportResolver),
	}

	manager.register(Java, &JavaResolver{})
	manager.register(Python, &PythonResolver{})
	manager.register(Go, &GoResolver{})
	manager.register(C, &CppResolver{})
	manager.register(CPP, &CppResolver{})
	manager.register(JavaScript, &JavaScriptResolver{})
	manager.register(TypeScript, &JavaScriptResolver{})
	manager.register(Ruby, &RubyResolver{})
	manager.register(Kotlin, &KotlinResolver{})
	manager.register(PHP, &PHPResolver{})
	manager.register(Scala, &ScalaResolver{})
	manager.register(Rust, &RustResolver{})

	return manager

}

// 注册解析器
func (rm *ResolverManager) register(language Language, resolver ImportResolver) {
	rm.resolvers[language] = resolver
}

// 解析导入语句
func (rm *ResolverManager) ResolveImport(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	resolver, exists := rm.resolvers[config.language]

	if !exists {
		return fmt.Errorf("import resolver unsupported language: %s", config.language)
	}

	return resolver.Resolve(importStmt, currentFilePath, config)
}

// Java解析器
type JavaResolver struct {
}

func (r *JavaResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	// 处理类导入
	classPath := strings.ReplaceAll(importName, ".", "/") + ".java"
	fullPath := utils.ToUnixPath(filepath.Join(config.SourceRoot, classPath))

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{fullPath}
		return nil
	}

	// 处理静态导入
	if strings.HasPrefix(importName, "static ") {
		importName = strings.TrimPrefix(importName, "static ")
	}

	// 处理包导入
	if strings.HasSuffix(importName, ".*") {
		pkgPath := strings.ReplaceAll(strings.TrimSuffix(importName, ".*"), ".", "/")
		fullPkgPath := utils.ToUnixPath(filepath.Join(config.SourceRoot, pkgPath))
		files := findFilesInDirIndex(config, fullPkgPath, ".java")
		importStmt.FilePaths = files
		if len(importStmt.FilePaths) == 0 {
			return fmt.Errorf("cannot find file which package belongs to: %s", importName)
		}
		return nil
	}

	importStmt.FilePaths = findMatchingFiles(config, fullPath)

	if len(importStmt.FilePaths) == 0 {
		return fmt.Errorf("cannot find file which import belongs to: %s", importName)
	}

	return nil
}

// Python解析器
type PythonResolver struct {
}

func (r *PythonResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	importPath := strings.ReplaceAll(importName, ".", "/")
	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{importPath}
		return nil
	}

	// 处理相对导入
	if strings.HasPrefix(importName, ".") {
		// 计算当前文件相对于 SourceRoot 的路径
		currentRelPath, _ := filepath.Rel(config.SourceRoot, currentFilePath)
		currentDir := utils.ToUnixPath(filepath.Dir(currentRelPath))
		dots := strings.Count(importName, ".")
		modulePath := strings.TrimPrefix(importName, strings.Repeat(".", dots))

		// 向上移动目录层级
		dir := currentDir
		for i := 0; i < dots-1; i++ {
			dir = utils.ToUnixPath(filepath.Dir(dir))
		}

		// 构建完整路径
		if modulePath != "" {
			modulePath = strings.ReplaceAll(modulePath, ".", "/")
			dir = utils.ToUnixPath(filepath.Join(dir, modulePath))
		}

		// 检查是否为包或模块
		for _, ext := range []string{"__init__.py", ".py"} {
			fullPath := utils.ToUnixPath(filepath.Join(config.SourceRoot, dir, ext))
			if containsFileIndex(config, fullPath) {
				importStmt.FilePaths = append(importStmt.FilePaths, fullPath)
			}
		}

		if len(importStmt.FilePaths) > 0 {
			return nil
		}

		return fmt.Errorf("cannot find file which relative import belongs to: %s", importName)
	}

	// 处理绝对导入
	foundPaths := []string{}

	// 检查是否为包或模块
	for _, ext := range []string{"__init__.py", ".py"} {
		fullPath := utils.ToUnixPath(filepath.Join(importPath, ext))
		if containsFileIndex(config, fullPath) {
			foundPaths = append(foundPaths, fullPath)
		}
		fullPath = utils.ToUnixPath(filepath.Join(importPath + ext))
		if containsFileIndex(config, fullPath) {
			foundPaths = append(foundPaths, fullPath)
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which abs import belongs to: %s", importName)
}

// Go解析器（简化版）
type GoResolver struct {
}

func (r *GoResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	// 标准库，直接排除
	if yes, _ := r.isStandardLibrary(importName); yes {
		logx.Debugf("import_resolver import %s is stantdard lib, skip", importName)
		return nil
	}
	// 移除mod，如果有
	relPath := importName
	if strings.HasPrefix(importName, config.SourceRoot) {
		relPath = strings.TrimPrefix(importName, config.SourceRoot+"/")
	}

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{relPath}
		return nil
	}

	// 尝试匹配 .go 文件
	relPathWithExt := relPath + ".go"
	if containsFileIndex(config, relPathWithExt) {
		importStmt.FilePaths = []string{relPathWithExt}
		return nil
	}

	// 匹配包目录下所有 .go 文件

	filesInDir := findFilesInDirIndex(config, relPath, ".go")
	if len(filesInDir) > 0 {
		importStmt.FilePaths = append(importStmt.FilePaths, filesInDir...)
	}

	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

func (g *GoResolver) isStandardLibrary(pkgPath string) (bool, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName,
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return false, fmt.Errorf("import_resolver load package: %v", err)
	}

	if len(pkgs) == 0 {
		return false, fmt.Errorf("import_resolver package not found: %s", pkgPath)
	}

	// 标准库包的PkgPath以"internal/"或非模块路径开头
	return !strings.Contains(pkgs[0].PkgPath, "."), nil
}

// C/C++解析器
type CppResolver struct {
}

func (r *CppResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	// 处理系统头文件
	if strings.HasPrefix(importName, "<") && strings.HasSuffix(importName, ">") {
		return nil // 系统头文件，不映射到项目文件
	}

	// 移除引号
	headerFile := strings.Trim(importName, "\"")

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{headerFile}
		return nil
	}

	foundPaths := []string{}

	// 相对路径导入
	if strings.HasPrefix(headerFile, ".") {
		// 计算当前文件相对于 SourceRoot 的路径
		currentRelPath, _ := filepath.Rel(config.SourceRoot, currentFilePath)
		currentDir := utils.ToUnixPath(filepath.Dir(currentRelPath))
		relPath := utils.ToUnixPath(filepath.Join(currentDir, headerFile))
		fullPath := utils.ToUnixPath(filepath.Join(config.SourceRoot, relPath))
		if containsFileIndex(config, fullPath) {
			foundPaths = append(foundPaths, fullPath)
		}
	}

	// 在源目录中查找
	for _, relDir := range config.Dirs {
		fullPath := utils.ToUnixPath(filepath.Join(relDir, headerFile))
		if containsFileIndex(config, fullPath) {
			foundPaths = append(foundPaths, fullPath)
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

// JavaScript/TypeScript解析器
type JavaScriptResolver struct {
}

func (r *JavaScriptResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		cleanedPath := strings.ReplaceAll(strings.ReplaceAll(importName, "./", ""), "../", "")
		importStmt.FilePaths = []string{cleanedPath}
		return nil
	}

	// 处理相对路径
	if strings.HasPrefix(importName, "./") || strings.HasPrefix(importName, "../") {
		// 计算当前文件相对于 SourceRoot 的路径
		currentRelPath, _ := filepath.Rel(config.SourceRoot, currentFilePath)
		currentDir := utils.ToUnixPath(filepath.Dir(currentRelPath))
		targetPath := utils.ToUnixPath(filepath.Join(currentDir, importName))
		foundPaths := []string{}

		// 尝试不同的文件扩展名
		for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", "/index.ts", "/index.tsx", "/index.js", "/index.jsx"} {
			fullPath := utils.ToUnixPath(filepath.Join(config.SourceRoot, targetPath+ext))
			if containsFileIndex(config, fullPath) {
				foundPaths = append(foundPaths, fullPath)
			}
		}

		importStmt.FilePaths = foundPaths
		if len(importStmt.FilePaths) > 0 {
			return nil
		}

		return fmt.Errorf("cannot find file which relative import belongs to: %s", importName)
	}

	// 处理项目内绝对路径导入
	foundPaths := []string{}
	for _, relDir := range config.Dirs {
		for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", "/index.ts", "/index.tsx", "/index.js", "/index.jsx"} {
			fullPath := utils.ToUnixPath(filepath.Join(relDir, importName+ext))
			if containsFileIndex(config, fullPath) {
				foundPaths = append(foundPaths, fullPath)
			}
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

// Rust解析器
type RustResolver struct {
}

func (r *RustResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	// 处理crate根路径
	if strings.HasPrefix(importName, "crate::") {
		importName = strings.TrimPrefix(importName, "crate::")
	}

	// 将::转换为路径分隔符
	modulePath := strings.ReplaceAll(importName, "::", "/")

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{modulePath}
		return nil
	}

	foundPaths := []string{}

	// 尝试查找.rs文件或模块目录
	for _, relDir := range config.Dirs {
		relPath := utils.ToUnixPath(filepath.Join(relDir, modulePath+".rs"))
		if containsFileIndex(config, relPath) {
			foundPaths = append(foundPaths, relPath)
		}
		modPath := utils.ToUnixPath(filepath.Join(relDir, modulePath, "mod.rs"))
		if containsFileIndex(config, modPath) {
			foundPaths = append(foundPaths, modPath)
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

// Ruby解析器
type RubyResolver struct {
}

func (r *RubyResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		cleanedPath := strings.ReplaceAll(strings.ReplaceAll(importName, "./", ""), "../", "")
		importStmt.FilePaths = []string{cleanedPath}
		return nil
	}

	// 处理相对导入
	if strings.HasPrefix(importName, ".") {
		// 计算当前文件相对于 SourceRoot 的路径
		currentRelPath, _ := filepath.Rel(config.SourceRoot, currentFilePath)
		currentDir := utils.ToUnixPath(filepath.Dir(currentRelPath))
		relPath := strings.TrimPrefix(importName, ".")
		if relPath == "" {
			return fmt.Errorf("invalid relative import: %s", importName)
		}

		// 添加.rb扩展名
		if !strings.HasSuffix(relPath, ".rb") {
			relPath += ".rb"
		}

		fullPath := utils.ToUnixPath(filepath.Join(config.SourceRoot, currentDir, relPath))
		if containsFileIndex(config, fullPath) {
			importStmt.FilePaths = []string{fullPath}
			return nil
		}

		return fmt.Errorf("canot find file which relative import belongs to: %s", importName)
	}

	// 处理项目内导入
	foundPaths := []string{}
	for _, relDir := range config.Dirs {
		relPath := utils.ToUnixPath(filepath.Join(relDir, importName+".rb"))
		if containsFileIndex(config, relPath) {
			foundPaths = append(foundPaths, relPath)
		}
		relPath = utils.ToUnixPath(filepath.Join(relDir, importName))
		if containsFileIndex(config, relPath) {
			foundPaths = append(foundPaths, relPath)
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

// Kotlin解析器
type KotlinResolver struct {
}

func (r *KotlinResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	// 处理包导入
	if strings.HasSuffix(importName, ".*") {
		return nil // 包导入不映射到具体文件
	}

	// 处理类导入
	classPath := strings.ReplaceAll(importName, ".", "/")

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{classPath}
		return nil
	}

	foundPaths := []string{}

	// 尝试Kotlin文件
	for _, relDir := range config.Dirs {
		relPath := utils.ToUnixPath(filepath.Join(relDir, classPath+".kt"))
		if containsFileIndex(config, relPath) {
			foundPaths = append(foundPaths, relPath)
		}
		// 尝试Java文件
		relPath = utils.ToUnixPath(filepath.Join(relDir, classPath+".java"))
		if containsFileIndex(config, relPath) {
			foundPaths = append(foundPaths, relPath)
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

// PHP解析器（简化版）
type PHPResolver struct {
}

func (r *PHPResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	// 处理命名空间导入
	if strings.HasPrefix(importName, "\\") {
		importName = strings.TrimPrefix(importName, "\\")
	}

	// 将命名空间分隔符转换为路径分隔符
	namespacePath := strings.ReplaceAll(importName, "\\", "/")

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{namespacePath}
		return nil
	}

	foundPaths := []string{}

	// 在源目录中查找
	for _, relDir := range config.Dirs {
		fullPath := utils.ToUnixPath(filepath.Join(relDir, namespacePath+".php"))
		if containsFileIndex(config, fullPath) {
			foundPaths = append(foundPaths, fullPath)
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

// Scala解析器
type ScalaResolver struct {
}

func (r *ScalaResolver) Resolve(importStmt *Import, currentFilePath string, config *ProjectConfig) error {
	if importStmt.Name == types.EmptyString {
		return fmt.Errorf("import is empty")
	}

	importStmt.FilePaths = []string{}
	importName := importStmt.Name

	// 处理包导入
	if strings.HasSuffix(importName, "._") {
		return nil // 包导入不映射到具体文件
	}

	// 处理类导入
	classPath := strings.ReplaceAll(importName, ".", "/")

	if len(config.fileSet) == 0 {
		logx.Debugf("not support project file list, use default resolve")
		importStmt.FilePaths = []string{classPath}
		return nil
	}

	foundPaths := []string{}

	// 尝试Scala文件
	for _, relDir := range config.Dirs {
		relPath := utils.ToUnixPath(filepath.Join(relDir, classPath+".scala"))
		if containsFileIndex(config, relPath) {
			foundPaths = append(foundPaths, relPath)
		}
		// 尝试Java文件
		relPath = utils.ToUnixPath(filepath.Join(relDir, classPath+".java"))
		if containsFileIndex(config, relPath) {
			foundPaths = append(foundPaths, relPath)
		}
	}

	importStmt.FilePaths = foundPaths
	if len(importStmt.FilePaths) > 0 {
		return nil
	}

	return fmt.Errorf("cannot find file which import belongs to: %s", importName)
}

// 辅助函数：查找匹配的文件路径
func findMatchingFiles(config *ProjectConfig, targetPath string) []string {
	var result []string
	if containsFileIndex(config, targetPath) {
		result = append(result, targetPath)
	}
	return result
}

// 辅助函数：查找目录下所有指定扩展名的文件
func findFilesInDirIndex(config *ProjectConfig, dir string, ext string) []string {
	var result []string
	files, ok := config.dirToFiles[dir]
	if !ok {
		return result
	}
	for _, f := range files {
		if strings.HasSuffix(f, ext) {
			result = append(result, f)
		}
	}
	return result
}

// 辅助函数：检查文件是否存在于项目文件集合中
func containsFileIndex(config *ProjectConfig, path string) bool {
	_, ok := config.fileSet[path]
	return ok
}
