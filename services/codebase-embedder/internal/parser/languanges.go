package parser

import (
	"path/filepath"

	sitterkotlin "github.com/tree-sitter-grammars/tree-sitter-kotlin/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
	sittercsharp "github.com/tree-sitter/tree-sitter-c-sharp/bindings/go"
	sitterc "github.com/tree-sitter/tree-sitter-c/bindings/go"
	sittercpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	sittergo "github.com/tree-sitter/tree-sitter-go/bindings/go"
	sitterjava "github.com/tree-sitter/tree-sitter-java/bindings/go"
	sitterjavascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	sitterphp "github.com/tree-sitter/tree-sitter-php/bindings/go"
	sitterpython "github.com/tree-sitter/tree-sitter-python/bindings/go"
	sitterruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	sitterrust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	sitterscala "github.com/tree-sitter/tree-sitter-scala/bindings/go"
	sittertypescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

// Language represents a programming language.
type Language string

const (
	Java       Language = "java"
	Python     Language = "python"
	Go         Language = "go"
	JavaScript Language = "javascript"
	TypeScript Language = "typescript"
	TSX        Language = "tsx"
	Rust       Language = "rust"
	C          Language = "c"
	CPP        Language = "cpp"
	CSharp     Language = "csharp"
	Ruby       Language = "ruby"
	PHP        Language = "php"
	Kotlin     Language = "kotlin"
	Scala      Language = "scala"
	Markdown   Language = "markdown"
	OpenAPI    Language = "openapi"
	Swagger    Language = "swagger"
)

// LanguageConfig holds the configuration for a language
type LanguageConfig struct {
	Language       Language
	SitterLanguage func() *sitter.Language
	SupportedExts  []string
}

// languageConfigs 定义了所有支持的语言配置
var languageConfigs = []*LanguageConfig{
	{
		Language: Go,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sittergo.Language())
		},
		SupportedExts: []string{".go"},
	},
	{
		Language: Java,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterjava.Language())
		},
		SupportedExts: []string{".java"},
	},
	{
		Language: Python,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterpython.Language())
		},
		SupportedExts: []string{".py"},
	},
	{
		Language: JavaScript,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterjavascript.Language())
		},
		SupportedExts: []string{".js", ".jsx"},
	},
	{
		Language: TypeScript,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sittertypescript.LanguageTypescript())
		},
		SupportedExts: []string{".ts"},
	},
	{
		Language: TSX,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sittertypescript.LanguageTSX())
		},
		SupportedExts: []string{".tsx"},
	},
	{
		Language: Rust,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterrust.Language())
		},
		SupportedExts: []string{".rs"},
	},
	{
		Language: C,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterc.Language())
		},
		SupportedExts: []string{".c", ".h"},
	},
	{
		Language: CPP,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sittercpp.Language())
		},
		SupportedExts: []string{".cpp", ".cc", ".cxx", ".hpp"},
	},
	{
		Language: CSharp,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sittercsharp.Language())
		},
		SupportedExts: []string{".cs"},
	},
	{
		Language: Ruby,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterruby.Language())
		},
		SupportedExts: []string{".rb"},
	},
	{
		Language: PHP,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterphp.LanguagePHP())
		},
		SupportedExts: []string{".php", ".phtml"},
	},
	{
		Language: Kotlin,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterkotlin.Language())
		},
		SupportedExts: []string{".kt", ".kts"},
	},
	{
		Language: Scala,
		SitterLanguage: func() *sitter.Language {
			return sitter.NewLanguage(sitterscala.Language())
		},
		SupportedExts: []string{".scala"},
	},
	{
		Language: Markdown,
		SitterLanguage: func() *sitter.Language {
			// Markdown 没有对应的 tree-sitter 解析器，返回 nil
			return nil
			// return sitter.NewLanguage(sittermarkdown.Language())
		},
		SupportedExts: []string{".md", ".mdx"},
	},
}

// GetLanguageConfigs 获取所有语言配置
func GetLanguageConfigs() []*LanguageConfig {
	return languageConfigs
}

// getLanguageConfigByExt 根据文件扩展名获取语言配置
func getLanguageConfigByExt(ext string) *LanguageConfig {
	for _, config := range languageConfigs {
		for _, supportedExt := range config.SupportedExts {
			if supportedExt == ext {
				return config
			}
		}
	}
	return nil
}

func GetLangConfigByFilePath(path string) (*LanguageConfig, error) {
	ext := filepath.Ext(path)
	if ext == "" {
		return nil, ErrFileExtNotFound
	}
	langConf := getLanguageConfigByExt(ext)
	if langConf == nil {
		return nil, ErrLangConfNotFound
	}
	return langConf, nil
}
