package test

import (
	"strings"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// MetricsCalculator 指标计算器
// 提供检索系统核心性能指标计算功能
type MetricsCalculator struct {
}

// NewMetricsCalculator 创建新的指标计算器
// 返回初始化完成的MetricsCalculator实例
func NewMetricsCalculator() *MetricsCalculator {
	return &MetricsCalculator{}
}

// Calculate 计算评估指标
// 根据期望结果和实际检索结果计算各项性能指标
// 参数:
//   - expected: 期望匹配的文件名列表
//   - retrieved: 实际检索到的文件列表
//   - queryTime: 查询响应时间
//
// 返回:
//   - Metrics: 包含所有评估指标的结构体
func (mc *MetricsCalculator) Calculate(expected []string, retrieved []*types.SemanticFileItem, queryTime time.Duration) Metrics {
	return mc.CalculateWithContent(expected, nil, retrieved, queryTime)
}

// CalculateWithContent 计算评估指标
// 根据期望结果和实际检索结果计算核心性能指标
// 参数:
//   - expected: 期望匹配的文件名列表
//   - expectedContents: 期望匹配的内容片段列表（已忽略）
//   - retrieved: 实际检索到的文件列表
//   - queryTime: 查询响应时间
//
// 返回:
//   - Metrics: 包含核心评估指标的结构体
func (mc *MetricsCalculator) CalculateWithContent(expected []string, expectedContents []string, retrieved []*types.SemanticFileItem, queryTime time.Duration) Metrics {
	// 处理边界情况
	if len(expected) == 0 && len(retrieved) == 0 {
		return Metrics{
			Precision:    1.0,
			Recall:       1.0,
			F1Score:      1.0,
			ResponseTime: float64(queryTime.Milliseconds()),
		}
	}

	if len(expected) == 0 || len(retrieved) == 0 {
		return Metrics{
			Precision:    0.0,
			Recall:       0.0,
			F1Score:      0.0,
			ResponseTime: float64(queryTime.Milliseconds()),
		}
	}

	// 计算相关性矩阵
	relevanceMatrix := mc.calculateRelevanceMatrix(expected, retrieved)

	// 计算核心指标
	precision := mc.calculatePrecision(relevanceMatrix)
	recall := mc.calculateRecall(relevanceMatrix, len(expected))
	f1Score := mc.calculateF1Score(precision, recall)

	return Metrics{
		Precision:    precision,
		Recall:       recall,
		F1Score:      f1Score,
		ResponseTime: float64(queryTime.Milliseconds()),
	}
}

// calculateRelevanceMatrix 计算相关性矩阵
func (mc *MetricsCalculator) calculateRelevanceMatrix(expected []string, retrieved []*types.SemanticFileItem) [][]bool {
	matrix := make([][]bool, len(retrieved))
	for i := range matrix {
		matrix[i] = make([]bool, len(expected))
	}

	for i, item := range retrieved {
		for j, exp := range expected {
			// 简化的相关性判断：考虑文件路径和内容
			relevant := strings.Contains(item.Content, exp) ||
				strings.Contains(item.FilePath, exp)
			matrix[i][j] = relevant
		}
	}

	return matrix
}

// calculatePrecision 计算准确率
func (mc *MetricsCalculator) calculatePrecision(relevanceMatrix [][]bool) float64 {
	if len(relevanceMatrix) == 0 {
		return 0.0
	}

	relevantCount := 0
	for _, row := range relevanceMatrix {
		for _, relevant := range row {
			if relevant {
				relevantCount++
				break // 只要匹配任何一个期望值就算相关
			}
		}
	}

	return float64(relevantCount) / float64(len(relevanceMatrix))
}

// calculateRecall 计算召回率
func (mc *MetricsCalculator) calculateRecall(relevanceMatrix [][]bool, expectedCount int) float64 {
	if expectedCount == 0 {
		return 0.0
	}

	relevantExpected := make([]bool, expectedCount)
	for _, row := range relevanceMatrix {
		for j, relevant := range row {
			if relevant {
				relevantExpected[j] = true
			}
		}
	}

	relevantCount := 0
	for _, relevant := range relevantExpected {
		if relevant {
			relevantCount++
		}
	}

	return float64(relevantCount) / float64(expectedCount)
}

// calculateF1Score 计算F1分数
func (mc *MetricsCalculator) calculateF1Score(precision, recall float64) float64 {
	if precision+recall == 0 {
		return 0.0
	}
	return 2 * (precision * recall) / (precision + recall)
}

// CalculateAverage 计算平均指标
func (mc *MetricsCalculator) CalculateAverage(queryResults []QueryResult) Metrics {
	if len(queryResults) == 0 {
		return Metrics{}
	}

	var totalPrecision, totalRecall, totalF1, totalResponseTime float64
	validResults := 0

	for _, result := range queryResults {
		// 跳过无效结果
		if result.Metrics.Precision < 0 || result.Metrics.Recall < 0 {
			continue
		}

		totalPrecision += result.Metrics.Precision
		totalRecall += result.Metrics.Recall
		totalF1 += result.Metrics.F1Score
		totalResponseTime += result.Metrics.ResponseTime

		validResults++
	}

	if validResults == 0 {
		return Metrics{}
	}

	count := float64(validResults)
	return Metrics{
		Precision:    totalPrecision / count,
		Recall:       totalRecall / count,
		F1Score:      totalF1 / count,
		ResponseTime: totalResponseTime / count,
	}
}
