package embedding

import "github.com/zgsm-ai/codebase-indexer/internal/parser"

// languageChunkNodeKind 定义各语言的语义切块节点类型
var languageChunkNodeKind = map[parser.Language][]string{
	parser.Go: {
		"function_declaration", // 函数声明
		"method_declaration",   // 方法声明
		"struct_type",          // 结构体
		"interface_type",       // 接口
		"type_declaration",     // 类型声明
	},
	parser.Java: {
		"method_declaration",    // 方法声明
		"class_declaration",     // 类声明
		"interface_declaration", // 接口声明
		"enum_declaration",      // 枚举声明
	},
	parser.Python: {
		"function_definition", // 函数定义
		"class_definition",    // 类定义
	},
	parser.JavaScript: {
		"function_declaration", // 函数声明
		"class_declaration",    // 类声明
		"arrow_function",       // 箭头函数
		"export_declaration",   // 导出声明（ES模块）
	},
	parser.TypeScript: {
		"function_declaration",  // 函数声明
		"class_declaration",     // 类声明
		"arrow_function",        // 箭头函数
		"interface_declaration", // 接口声明
	},
	parser.TSX: {
		"jsx_element",          // JSX组件
		"function_declaration", // 函数声明
		"class_declaration",    // 类声明
	},
	parser.Rust: {
		"function_item",     // 函数声明
		"struct_definition", // 结构体定义
		"enum_definition",   // 枚举定义
		"module_item",       // 模块声明
		"trait_item",        // 特质声明
	},
	parser.C: {
		"function_definition", // 函数定义
		"struct_declaration",  // 结构体声明
		"union_declaration",   // 联合体声明
		"typedef_declaration", // 类型定义
	},
	parser.CPP: {
		"function_definition",   // 函数定义
		"class_declaration",     // 类声明
		"struct_declaration",    // 结构体声明
		"namespace_declaration", // 命名空间声明
	},
	parser.CSharp: {
		"method_declaration",    // 方法声明
		"class_declaration",     // 类声明
		"interface_declaration", // 接口声明
		"struct_declaration",    // 结构体声明
	},
	parser.Ruby: {
		"def",    // 函数定义
		"class",  // 类定义
		"module", // 模块定义
	},
	parser.PHP: {
		"function_declaration",  // 函数声明
		"class_declaration",     // 类声明
		"interface_declaration", // 接口声明
	},
	parser.Kotlin: {
		"function_declaration",  // 函数声明
		"class_declaration",     // 类声明
		"interface_declaration", // 接口声明
		"companion_object",      // 伴随对象
	},
	parser.Scala: {
		"function_definition", // 函数定义
		"class_declaration",   // 类声明
		"trait_declaration",   // 特质声明
		"object_declaration",  // 对象声明
	},
	parser.Markdown: {
		"section",    // 章节标题（# ## ### 等）
		"code_block", // 代码块
		"paragraph",  // 段落
	},
}
