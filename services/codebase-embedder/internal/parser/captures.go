package parser

import (
	"github.com/zgsm-ai/codebase-indexer/internal/types"
	"strings"
)

const (
	dotName       = ".name"
	dotArguments  = ".arguments"
	dotParameters = ".parameters"
	dotOwner      = ".owner"
	dotSource     = ".source"
	dotAlias      = ".alias"
)

// ElementType 表示代码元素类型，使用字符串字面量作为枚举值
type ElementType string

const (
	ElementTypeNamespace           ElementType = "namespace"
	ElementTypePackage             ElementType = "package"
	ElementTypeUndefined           ElementType = "undefined"
	ElementTypeImport              ElementType = "import"
	ElementTypeClass               ElementType = "definition.class"
	ElementTypeInterface           ElementType = "definition.interface"
	ElementTypeStruct              ElementType = "definition.struct"
	ElementTypeEnum                ElementType = "definition.enum"
	ElementTypeUnion               ElementType = "definition.union"
	ElementTypeTrait               ElementType = "definition.trait"
	ElementTypeTypeAlias           ElementType = "definition.type_alias"
	ElementTypeFunction            ElementType = "definition.function"
	ElementTypeMethodCall          ElementType = "call.method"
	ElementTypeFunctionCall        ElementType = "call.function"
	ElementTypeFunctionDeclaration ElementType = "declaration.function"
	ElementTypeMethod              ElementType = "definition.method"
	ElementTypeConstructor         ElementType = "definition.constructor"
	ElementTypeDestructor          ElementType = "definition.destructor"
	ElementTypeGlobalVariable      ElementType = "global_variable"
	ElementTypeLocalVariable       ElementType = "local_variable"
	ElementTypeVariable            ElementType = "variable"
	ElementTypeConstant            ElementType = "constant"
	ElementTypeMacro               ElementType = "macro"
	ElementTypeField               ElementType = "definition.field"
	ElementTypeParameter           ElementType = "definition.parameter"
	ElementTypeComment             ElementType = "comment"
	ElementTypeAnnotation          ElementType = "annotation"
)

// 类型映射表 - captureName -> ElementType（使用ElementType字符串值作为键）
var typeMappings = map[string]ElementType{
	string(ElementTypeNamespace):           ElementTypeNamespace,
	string(ElementTypePackage):             ElementTypePackage,
	string(ElementTypeUndefined):           ElementTypeUndefined,
	string(ElementTypeImport):              ElementTypeImport,
	string(ElementTypeClass):               ElementTypeClass,
	string(ElementTypeInterface):           ElementTypeInterface,
	string(ElementTypeStruct):              ElementTypeStruct,
	string(ElementTypeEnum):                ElementTypeEnum,
	string(ElementTypeUnion):               ElementTypeUnion,
	string(ElementTypeTrait):               ElementTypeTrait,
	string(ElementTypeTypeAlias):           ElementTypeTypeAlias,
	string(ElementTypeFunction):            ElementTypeFunction,
	string(ElementTypeFunctionCall):        ElementTypeFunctionCall,
	string(ElementTypeFunctionDeclaration): ElementTypeFunctionDeclaration,
	string(ElementTypeMethod):              ElementTypeMethod,
	string(ElementTypeMethodCall):          ElementTypeMethodCall,
	string(ElementTypeConstructor):         ElementTypeConstructor,
	string(ElementTypeDestructor):          ElementTypeDestructor,
	string(ElementTypeGlobalVariable):      ElementTypeGlobalVariable,
	string(ElementTypeLocalVariable):       ElementTypeLocalVariable,
	string(ElementTypeVariable):            ElementTypeVariable,
	string(ElementTypeConstant):            ElementTypeConstant,
	string(ElementTypeMacro):               ElementTypeMacro,
	string(ElementTypeField):               ElementTypeField,
	string(ElementTypeParameter):           ElementTypeParameter,
	string(ElementTypeComment):             ElementTypeComment,
	string(ElementTypeAnnotation):          ElementTypeAnnotation,
}

//
//// 类型映射表 - captureName -> ElementType
//var typeMappings = map[string]ElementType{
//	"package":                ElementTypePackage,
//	"namespace":              ElementTypeNamespace,
//	"import":                 ElementTypeImport,
//	"declaration.function":   ElementTypeFunctionDeclaration,
//	"definition.method":      ElementTypeMethod,
//	"call.method":            ElementTypeMethodCall,
//	"definition.function":    ElementTypeFunction,
//	"call.function":          ElementTypeFunctionCall,
//	"definition.class":       ElementTypeClass,
//	"definition.interface":   ElementTypeInterface,
//	"definition.struct":      ElementTypeStruct,
//	"definition.enum":        ElementTypeEnum,
//	"definition.union":       ElementTypeUnion,
//	"definition.trait":       ElementTypeTrait,
//	"definition.type_alias":  ElementTypeTypeAlias,
//	"definition.constructor": ElementTypeConstructor,
//	"definition.destructor":  ElementTypeDestructor,
//	"global_variable":        ElementTypeGlobalVariable,
//	"local_variable":         ElementTypeLocalVariable,
//	"variable":               ElementTypeVariable,
//	"constant":               ElementTypeConstant,
//	"macro":                  ElementTypeMacro,
//	"definition.field":       ElementTypeField,
//	"definition.parameter":   ElementTypeParameter,
//	"comment":                ElementTypeComment,
//	"doc_comment":            ElementTypeDocComment,
//	"annotation":             ElementTypeAnnotation,
//	"undefined":              ElementTypeUndefined,
//}

// toElementType 将字符串映射为ElementType
func toElementType(captureName string) ElementType {
	if captureName == types.EmptyString {
		return ElementTypeUndefined
	}
	if et, exists := typeMappings[captureName]; exists {
		return et
	}
	return ElementTypeUndefined
}

// 函数工厂：生成检查字符串是否以特定后缀结尾的函数
func createSuffixChecker(suffix string) func(string) bool {
	return func(captureName string) bool {
		return strings.HasSuffix(captureName, suffix)
	}
}

// 使用工厂函数创建检查器
var (
	isNameCapture       = createSuffixChecker(dotName)
	isParametersCapture = createSuffixChecker(dotParameters)
	isArgumentsCapture  = createSuffixChecker(dotArguments)
	isOwnerCapture      = createSuffixChecker(dotOwner)
	isSourceCapture     = createSuffixChecker(dotSource)
	isAliasCapture      = createSuffixChecker(dotAlias)
)

// 特殊函数（需要额外判断）保留
func isElementNameCapture(elementType ElementType, captureName string) bool {
	return isNameCapture(captureName) &&
		captureName == string(elementType)+dotName
}
