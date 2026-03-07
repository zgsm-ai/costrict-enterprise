package parser

import (
	"context"
	"strings"

	treesitter "github.com/tree-sitter/go-tree-sitter"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type ParsedSource struct {
	Path     string
	Package  *Package
	Imports  []*Import
	Language Language
	Elements []CodeElement
}

// CodeElement 定义所有代码元素的接口
type CodeElement interface {
	GetName() string
	GetType() ElementType
	GetRange() []int32
	GetParent() CodeElement
	SetParent(parent CodeElement)
	AddChild(child CodeElement)
	GetChildren() []CodeElement
	Update(ctx context.Context, captureName string,
		capture *treesitter.QueryCapture, source []byte, opts ParseOptions) error
	SetContent(content []byte)
	setRootIndex(rootIndex uint32)
}

// BaseElement 提供接口的基础实现，其他类型嵌入该结构体
type BaseElement struct {
	Name             string
	rootCaptureIndex uint32
	Type             ElementType
	Content          []byte
	Range            []int32
	Parent           CodeElement
	Children         []CodeElement
}

// Import 表示导入语句
type Import struct {
	*BaseElement
	Source    string   // from (xxx)
	Alias     string   // as (xxx)
	FilePaths []string // 相对于项目root的路径（排除标准库/第三方包）
}

// Package 表示代码包
type Package struct {
	*BaseElement
}

// Function 表示函数
type Function struct {
	*BaseElement
	Owner      string
	Parameters []string
	ReturnType string
}

// Method 表示方法
type Method struct {
	*BaseElement
	Owner      string
	Parameters []string
	ReturnType string
}

type Call struct {
	*BaseElement
	Owner     string
	Arguments []string
}

// Class 表示类
type Class struct {
	*BaseElement
	Fields  []*CodeElement
	Methods []*CodeElement
}

type Variable struct {
	*BaseElement
}

func (e *BaseElement) GetName() string              { return e.Name }
func (e *BaseElement) GetType() ElementType         { return e.Type }
func (e *BaseElement) GetRange() []int32            { return e.Range }
func (e *BaseElement) GetParent() CodeElement       { return e.Parent }
func (e *BaseElement) SetParent(parent CodeElement) { e.Parent = parent }
func (e *BaseElement) AddChild(child CodeElement) {
	e.Children = append(e.Children, child)
	child.SetParent(e)
}
func (e *BaseElement) GetChildren() []CodeElement { return e.Children }

func (e *BaseElement) SetContent(content []byte) {
	e.Content = content
}
func (e *BaseElement) setRootIndex(rootCaptureIndex uint32) {
	e.rootCaptureIndex = rootCaptureIndex
}
func (e *BaseElement) Update(ctx context.Context, captureName string, capture *treesitter.QueryCapture,
	source []byte, opts ParseOptions) error {
	node := &capture.Node

	if capture.Index == e.rootCaptureIndex { // root capture: @package @function @class etc
		// rootNode
		rootCaptureNode := node
		e.Range = []int32{
			int32(rootCaptureNode.StartPosition().Row),
			int32(rootCaptureNode.StartPosition().Column),
			int32(rootCaptureNode.StartPosition().Row),
			int32(rootCaptureNode.StartPosition().Column),
		}
		if opts.IncludeContent {
			content := source[node.StartByte():node.EndByte()]
			e.SetContent(content)
		}

	}

	if e.Name == types.EmptyString && isElementNameCapture(e.Type, captureName) {
		// 取root节点的name，比如definition.function.name
		// 获取名称 ,go import 带双引号
		name := strings.ReplaceAll(node.Utf8Text(source), types.SingleDoubleQuote, types.EmptyString)
		if name == types.EmptyString {
			tracer.WithTrace(ctx).Errorf("tree_sitter base_processor name_node %s %v name not found", captureName, e.Range)
		}
		e.Name = name
	}

	return nil
}

func (f *Function) Update(ctx context.Context, captureName string,
	capture *treesitter.QueryCapture, source []byte, opts ParseOptions) error {

	if err := f.BaseElement.Update(ctx, captureName, capture, source, opts); err != nil {
		return err
	}
	node := &capture.Node

	if len(f.Parameters) == 0 && isParametersCapture(captureName) {
		f.Parameters = strings.Split(node.Utf8Text(source), types.Comma)
	}

	if isOwnerCapture(captureName) && f.Owner == types.EmptyString {
		f.Owner = node.Utf8Text(source)
	}

	return nil
}

func (m *Method) Update(ctx context.Context, captureName string,
	capture *treesitter.QueryCapture, source []byte, opts ParseOptions) error {

	if err := m.BaseElement.Update(ctx, captureName, capture, source, opts); err != nil {
		return err
	}

	node := &capture.Node

	if len(m.Parameters) == 0 && isParametersCapture(captureName) {
		m.Parameters = strings.Split(node.Utf8Text(source), types.Comma)
	}

	if isOwnerCapture(captureName) && m.Owner == types.EmptyString {
		m.Owner = node.Utf8Text(source)
	}

	return nil
}

func (c *Call) Update(ctx context.Context, captureName string,
	capture *treesitter.QueryCapture, source []byte, opts ParseOptions) error {

	if err := c.BaseElement.Update(ctx, captureName, capture, source, opts); err != nil {
		return err
	}
	node := &capture.Node

	if len(c.Arguments) == 0 && isArgumentsCapture(captureName) {
		c.Arguments = strings.Split(node.Utf8Text(source), types.Comma)
	}

	if c.Owner == types.EmptyString && isOwnerCapture(captureName) {
		c.Owner = node.Utf8Text(source)
	}

	return nil
}

func (v *Variable) Update(ctx context.Context, captureName string,
	capture *treesitter.QueryCapture, source []byte, opts ParseOptions) error {

	if err := v.BaseElement.Update(ctx, captureName, capture, source, opts); err != nil {
		return err
	}

	node := &capture.Node

	// TODO 局部变量不是很容易区分，存在多层嵌套。找到它的名字不太容易。存在一行返回多个局部变量的情况,当前只取了第一个
	if v.Name == types.EmptyString {
		if nameNode := findIdentifierNode(node); nameNode != nil {
			v.Name = nameNode.Utf8Text(source)
		}
	}
	return nil
}

func (v *Import) Update(ctx context.Context, captureName string,
	capture *treesitter.QueryCapture, source []byte, opts ParseOptions) error {

	if err := v.BaseElement.Update(ctx, captureName, capture, source, opts); err != nil {
		return err
	}

	node := &capture.Node

	// TODO 各个scm 的source、alias full_name 解析。
	if v.Source == types.EmptyString && isSourceCapture(captureName) {
		v.Source = node.Utf8Text(source)
	}

	if v.Alias == types.EmptyString && isAliasCapture(captureName) {
		v.Alias = node.Utf8Text(source)
	}

	return nil
}

func initRootElement(elementTypeValue string) CodeElement {
	elementType := toElementType(elementTypeValue)
	base := &BaseElement{}
	switch elementType {
	case ElementTypePackage:
		base.Type = ElementTypePackage
		return &Package{BaseElement: base}
	case ElementTypeImport:
		base.Type = ElementTypeImport
		return &Import{BaseElement: base}
	case ElementTypeFunction:
		base.Type = ElementTypeFunction
		return &Function{BaseElement: base}
	case ElementTypeClass:
		base.Type = ElementTypeClass
		return &Class{BaseElement: base}
	case ElementTypeMethod:
		base.Type = ElementTypeMethod
		return &Method{BaseElement: base}
	case ElementTypeFunctionCall:
		base.Type = ElementTypeFunctionCall
		return &Call{BaseElement: base}
	case ElementTypeMethodCall:
		base.Type = ElementTypeMethodCall
		return &Call{BaseElement: base}
	default:
		base.Type = ElementTypeUndefined
		return base
	}
}
