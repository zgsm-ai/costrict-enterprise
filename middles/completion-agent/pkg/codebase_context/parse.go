package codebase_context

import (
	"fmt"
	"sort"
)

// ParsedSemanticResult 解析后的语义搜索结果
type ParsedSemanticResult struct {
	FilePath string
	Content  string
	Score    float64
}

// ParsedDefinitionResult 解析后的定义搜索结果
type ParsedDefinitionResult struct {
	Name     string
	FilePath string
	Content  string
}

// ParsedRelationResult 解析后的关系搜索结果
type ParsedRelationResult struct {
	FilePath string
	Content  string
	Score    float64
}

// parseSemantic 解析语义检索结果
func parseSemantic(data []*ResponseData) []ParsedSemanticResult {
	if len(data) == 0 {
		return nil
	}

	var result []ParsedSemanticResult
	contextSet := make(map[string]bool)

	for _, item := range data {
		if item == nil {
			continue
		}

		semanticList := item.Data.List
		for _, semantic := range semanticList {
			if semantic == nil {
				continue
			}

			// 获取相似代码信息
			content := getStringValue(semantic, "content")
			if content == "" || contextSet[content] {
				continue
			}

			contextSet[content] = true
			filePath := getStringValue(semantic, "filePath")
			score := getFloat64Value(semantic, "score")

			result = append(result, ParsedSemanticResult{
				FilePath: filePath,
				Content:  content,
				Score:    score,
			})
		}
	}

	// 按分数从高到低排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	// // 转换为字符串数组格式
	// var finalResult [][]string
	// for _, item := range result {
	// 	finalResult = append(finalResult, []string{item.FilePath, item.Content})
	// }

	return result
}

// parseDefinition 解析定义检索结果
func parseDefinition(data []*ResponseData) []ParsedDefinitionResult {
	if len(data) == 0 {
		return nil
	}

	var result []ParsedDefinitionResult
	contextSet := make(map[string]bool)

	for _, item := range data {
		if item == nil {
			continue
		}

		defList := item.Data.List
		for _, defItem := range defList {
			if defItem == nil {
				continue
			}

			// 提取filePath, name, content
			filePath := getStringValue(defItem, "filePath")
			name := getStringValue(defItem, "name")
			content := getStringValue(defItem, "content")
			defType := getStringValue(defItem, "type")

			// 根据类型处理内容
			if content != "" {
				switch defType {
				case "definition.method", "definition.function", "declaration.method", "declaration.function":
					content = sliceBeforeNthInstance(content, "\n", 20)
				case "definition.class", "definition.struct", "declaration.struct", "declaration.class":
					content = sliceBeforeNthInstance(content, "\n", 50)
				default:
					content = sliceBeforeNthInstance(content, "\n", 10)
				}
			}

			// 去重
			key := fmt.Sprintf("%s:%s", filePath, name)
			if contextSet[key] {
				continue
			}

			contextSet[key] = true
			result = append(result, ParsedDefinitionResult{
				Name:     name,
				FilePath: filePath,
				Content:  content,
			})
		}
	}

	// // 转换为字符串数组格式
	// var finalResult [][]string
	// for _, item := range result {
	// 	finalResult = append(finalResult, []string{item.Name, item.FilePath, item.Content})
	// }

	return result
}

// parseRelation 解析关系检索结果
func parseRelation(data []*ResponseData) []ParsedRelationResult {
	if len(data) == 0 {
		return nil
	}

	var result []ParsedRelationResult

	for _, item := range data {
		if item == nil {
			continue
		}

		relationList := item.Data.List
		for _, relation := range relationList {
			if relation == nil {
				continue
			}

			filePath := getStringValue(relation, "filePath")
			content := getStringValue(relation, "content")
			score := getFloat64Value(relation, "score")

			result = append(result, ParsedRelationResult{
				FilePath: filePath,
				Content:  content,
				Score:    score,
			})
		}
	}

	// 按分数排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	// // 转换为字符串数组格式
	// var finalResult [][]string
	// for _, item := range result {
	// 	finalResult = append(finalResult, []string{item.FilePath, item.Content})
	// }

	return result
}

// 辅助函数
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloat64Value(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}
