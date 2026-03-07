package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	configPath = filepath.Join(baseDir, "test/config.yaml")
	outputDir  = filepath.Join(baseDir, "test/results")
)

// TestReport 测试报告结构
type TestReport struct {
	TestName         string                 `json:"test_name"`
	TestDescription  string                 `json:"test_description"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
	TotalScenarios   int                    `json:"total_scenarios"`
	SuccessScenarios int                    `json:"success_scenarios"`
	FailedScenarios  int                    `json:"failed_scenarios"`
	Results          map[string]interface{} `json:"results"`
	Summary          TestSummary            `json:"summary"`
	Environment      EnvironmentInfo        `json:"environment"`
}

// TestSummary 测试摘要
type TestSummary struct {
	AveragePrecision        float64                           `json:"average_precision"`
	AverageRecall           float64                           `json:"average_recall"`
	AverageF1Score          float64                           `json:"average_f1_score"`
	AverageMAP              float64                           `json:"average_map"`
	AverageNDCG             float64                           `json:"average_ndcg"`
	AverageResponseTime     float64                           `json:"average_response_time"`
	AverageContentPrecision float64                           `json:"average_content_precision"`
	AverageContentRecall    float64                           `json:"average_content_recall"`
	AverageContentF1Score   float64                           `json:"average_content_f1_score"`
	TotalMatchedContents    int                               `json:"total_matched_contents"`
	TotalContents           int                               `json:"total_contents"`
	TotalQueries            int                               `json:"total_queries"`
	BestScenario            string                            `json:"best_scenario"`
	WorstScenario           string                            `json:"worst_scenario"`
	ScenarioDimensionStats  map[string]ScenarioDimensionStats `json:"scenario_dimension_stats"`
}

// ScenarioDimensionStats 场景维度统计
type ScenarioDimensionStats struct {
	Count                   int     `json:"count"`
	AveragePrecision        float64 `json:"average_precision"`
	AverageRecall           float64 `json:"average_recall"`
	AverageF1Score          float64 `json:"average_f1_score"`
	AverageContentPrecision float64 `json:"average_content_precision"`
	AverageContentRecall    float64 `json:"average_content_recall"`
}

// EnvironmentInfo 环境信息
type EnvironmentInfo struct {
	GoVersion     string        `json:"go_version"`
	TestStartTime time.Time     `json:"test_start_time"`
	TestDuration  time.Duration `json:"test_duration"`
	ConfigPath    string        `json:"config_path"`
	OutputDir     string        `json:"output_dir"`
}

func TestEmbedderComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过嵌入模型对比测试")
	}

	testStartTime := time.Now()

	// 加载配置
	testConfig, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建测试运行器
	runner, err := NewTestRunner(testConfig)
	if err != nil {
		t.Fatalf("创建测试运行器失败: %v", err)
	}

	// 运行测试
	t.Log("开始运行嵌入模型对比测试...")
	ctx := context.Background()
	result, err := runner.RunEmbedderComparison(ctx)
	if err != nil {
		t.Fatalf("运行嵌入模型对比测试失败: %v", err)
	}

	// 检查测试结果
	successCount := 0
	failedCount := 0
	for k, r := range result.Results {
		if r.Error != nil {
			t.Errorf("测试 %s 失败: %v", k, r.Error)
			failedCount++
		} else {
			successCount++
		}
	}

	// 生成详细报告
	report := generateTestReport(result, "embedder_comparison", testStartTime)

	// 输出结果
	if err := outputEnhancedResult(result, report, "embedder_comparison", outputDir); err != nil {
		t.Errorf("输出结果失败: %v", err)
	}

	// 输出详细摘要
	printEnhancedTestSummary(report)

	// 输出测试统计
	t.Logf("嵌入模型对比测试完成 - 成功: %d, 失败: %d, 耗时: %v",
		successCount, failedCount, result.EndTime.Sub(result.StartTime))
}

func TestVectorStoreComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过向量数据库对比测试")
	}

	testStartTime := time.Now()

	// 加载配置
	testConfig, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建测试运行器
	runner, err := NewTestRunner(testConfig)
	if err != nil {
		t.Fatalf("创建测试运行器失败: %v", err)
	}

	// 运行测试
	t.Log("开始运行向量数据库对比测试...")
	ctx := context.Background()
	result, err := runner.RunVectorStoreComparison(ctx)
	if err != nil {
		t.Fatalf("运行向量数据库对比测试失败: %v", err)
	}

	// 检查测试结果
	successCount := 0
	failedCount := 0
	for k, r := range result.Results {
		if r.Error != nil {
			t.Errorf("测试 %s 失败: %v", k, r.Error)
			failedCount++
		} else {
			successCount++
		}
	}

	// 生成详细报告
	report := generateTestReport(result, "vector_store_comparison", testStartTime)

	// 输出结果
	if err := outputEnhancedResult(result, report, "vector_store_comparison", outputDir); err != nil {
		t.Errorf("输出结果失败: %v", err)
	}

	// 输出详细摘要
	printEnhancedTestSummary(report)

	// 输出测试统计
	t.Logf("向量数据库对比测试完成 - 成功: %d, 失败: %d, 耗时: %v",
		successCount, failedCount, result.EndTime.Sub(result.StartTime))
}

// loadConfig 加载配置文件
func loadConfig(path string) (TestConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return TestConfig{}, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var testConfig TestConfig
	if err := yaml.Unmarshal(data, &testConfig); err != nil {
		return TestConfig{}, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := validateConfig(testConfig); err != nil {
		return TestConfig{}, fmt.Errorf("配置验证失败: %w", err)
	}

	return testConfig, nil
}

// validateConfig 验证配置
func validateConfig(config TestConfig) error {
	if len(config.Embedders) == 0 {
		return fmt.Errorf("未配置嵌入模型")
	}

	if len(config.VectorStores) == 0 {
		return fmt.Errorf("未配置向量数据库")
	}

	if len(config.Scenarios.EmbedderComparison.Embedders) == 0 {
		return fmt.Errorf("嵌入模型对比测试未配置嵌入模型")
	}

	if len(config.Scenarios.VectorStoreComparison.VectorStores) == 0 {
		return fmt.Errorf("向量数据库对比测试未配置向量数据库")
	}

	return nil
}

// outputEnhancedResult 输出增强的测试结果
func outputEnhancedResult(result *TestResult, report *TestReport, testName, outputDir string) error {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 输出JSON格式 - 原始结果
	jsonFile := filepath.Join(outputDir, fmt.Sprintf("%s_result.json", testName))
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化原始结果失败: %w", err)
	}
	if err := ioutil.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("写入原始结果文件失败: %w", err)
	}

	// 输出JSON格式 - 详细报告
	reportFile := filepath.Join(outputDir, fmt.Sprintf("%s_report.json", testName))
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化测试报告失败: %w", err)
	}
	if err := ioutil.WriteFile(reportFile, reportData, 0644); err != nil {
		return fmt.Errorf("写入测试报告文件失败: %w", err)
	}

	// 输出HTML格式报告
	htmlFile := filepath.Join(outputDir, fmt.Sprintf("%s_report.html", testName))
	if err := generateHTMLReport(report, htmlFile); err != nil {
		return fmt.Errorf("生成HTML报告失败: %w", err)
	}

	// 输出CSV格式摘要
	csvFile := filepath.Join(outputDir, fmt.Sprintf("%s_summary.csv", testName))
	if err := generateCSVSummary(result, csvFile); err != nil {
		return fmt.Errorf("生成CSV摘要失败: %w", err)
	}

	return nil
}

// generateTestReport 生成测试报告
func generateTestReport(result *TestResult, testName string, startTime time.Time) *TestReport {
	report := &TestReport{
		TestName:        result.TestName,
		TestDescription: result.TestDescription,
		StartTime:       result.StartTime,
		EndTime:         result.EndTime,
		Duration:        result.EndTime.Sub(result.StartTime),
		TotalScenarios:  len(result.Results),
		Results:         make(map[string]interface{}),
		Environment: EnvironmentInfo{
			GoVersion:     "unknown", // 可以通过runtime.Version()获取
			TestStartTime: startTime,
			TestDuration:  time.Since(startTime),
			ConfigPath:    configPath,
			OutputDir:     outputDir,
		},
	}

	// 统计成功和失败的场景
	successCount := 0
	failedCount := 0
	var totalPrecision, totalRecall, totalF1, totalMAP, totalNDCG, totalResponseTime float64
	var totalContentPrecision, totalContentRecall, totalContentF1Score float64
	var totalMatchedContents, totalContents int
	var totalQueries int
	var scenarioScores []ScenarioScore

	for name, scenarioResult := range result.Results {
		if scenarioResult.Error != nil {
			failedCount++
			report.Results[name] = map[string]interface{}{
				"status": "failed",
				"error":  scenarioResult.Error.Error(),
			}
			continue
		}

		successCount++
		metrics := scenarioResult.AverageMetrics
		totalPrecision += metrics.Precision
		totalRecall += metrics.Recall
		totalF1 += metrics.F1Score
		totalMAP += metrics.MAP
		totalNDCG += metrics.NDCG
		totalResponseTime += metrics.ResponseTime
		totalContentPrecision += metrics.ContentPrecision
		totalContentRecall += metrics.ContentRecall
		totalContentF1Score += metrics.ContentF1Score
		totalMatchedContents += metrics.MatchedContents
		totalContents += metrics.TotalContents
		totalQueries += len(scenarioResult.QueryResults)

		// 计算综合得分 (F1分数权重最高)
		compositeScore := (metrics.Precision*0.2 + metrics.Recall*0.2 + metrics.F1Score*0.3 + metrics.ContentF1Score*0.3)
		scenarioScores = append(scenarioScores, ScenarioScore{
			Name:           name,
			CompositeScore: compositeScore,
			Precision:      metrics.Precision,
			Recall:         metrics.Recall,
			F1Score:        metrics.F1Score,
			MAP:            metrics.MAP,
			NDCG:           metrics.NDCG,
			ResponseTime:   metrics.ResponseTime,
		})

		report.Results[name] = map[string]interface{}{
			"status":            "success",
			"precision":         metrics.Precision,
			"recall":            metrics.Recall,
			"f1_score":          metrics.F1Score,
			"map":               metrics.MAP,
			"ndcg":              metrics.NDCG,
			"response_time":     metrics.ResponseTime,
			"content_precision": metrics.ContentPrecision,
			"content_recall":    metrics.ContentRecall,
			"content_f1_score":  metrics.ContentF1Score,
			"matched_contents":  metrics.MatchedContents,
			"total_contents":    metrics.TotalContents,
			"query_count":       len(scenarioResult.QueryResults),
			"duration":          scenarioResult.EndTime.Sub(scenarioResult.StartTime),
		}
	}

	report.SuccessScenarios = successCount
	report.FailedScenarios = failedCount

	// 计算平均指标
	if successCount > 0 {
		report.Summary = TestSummary{
			AveragePrecision:        totalPrecision / float64(successCount),
			AverageRecall:           totalRecall / float64(successCount),
			AverageF1Score:          totalF1 / float64(successCount),
			AverageMAP:              totalMAP / float64(successCount),
			AverageNDCG:             totalNDCG / float64(successCount),
			AverageResponseTime:     totalResponseTime / float64(successCount),
			AverageContentPrecision: totalContentPrecision / float64(successCount),
			AverageContentRecall:    totalContentRecall / float64(successCount),
			AverageContentF1Score:   totalContentF1Score / float64(successCount),
			TotalMatchedContents:    totalMatchedContents,
			TotalContents:           totalContents,
			TotalQueries:            totalQueries,
		}

		// 找出最佳和最差场景
		if len(scenarioScores) > 0 {
			sort.Slice(scenarioScores, func(i, j int) bool {
				return scenarioScores[i].CompositeScore > scenarioScores[j].CompositeScore
			})
			report.Summary.BestScenario = scenarioScores[0].Name
			report.Summary.WorstScenario = scenarioScores[len(scenarioScores)-1].Name
		}
	}

	return report
}

// ScenarioScore 场景得分
type ScenarioScore struct {
	Name           string  `json:"name"`
	CompositeScore float64 `json:"composite_score"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	F1Score        float64 `json:"f1_score"`
	MAP            float64 `json:"map"`
	NDCG           float64 `json:"ndcg"`
	ResponseTime   float64 `json:"response_time"`
}

// printEnhancedTestSummary 打印增强的测试摘要
func printEnhancedTestSummary(report *TestReport) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("=== %s ===\n", report.TestName)
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	fmt.Printf("测试描述: %s\n", report.TestDescription)
	fmt.Printf("开始时间: %s\n", report.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("结束时间: %s\n", report.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("总耗时: %v\n", report.Duration)
	fmt.Printf("测试场景: %d (成功: %d, 失败: %d)\n",
		report.TotalScenarios, report.SuccessScenarios, report.FailedScenarios)

	if report.SuccessScenarios > 0 {
		fmt.Printf("\n--- 平均性能指标 ---\n")
		fmt.Printf("准确率: %.4f\n", report.Summary.AveragePrecision)
		fmt.Printf("召回率: %.4f\n", report.Summary.AverageRecall)
		fmt.Printf("F1分数: %.4f\n", report.Summary.AverageF1Score)
		fmt.Printf("MAP:    %.4f\n", report.Summary.AverageMAP)
		fmt.Printf("NDCG:   %.4f\n", report.Summary.AverageNDCG)
		fmt.Printf("响应时间: %.2fms\n", report.Summary.AverageResponseTime)
		fmt.Printf("\n--- 内容片段匹配指标 ---\n")
		fmt.Printf("内容准确率: %.4f\n", report.Summary.AverageContentPrecision)
		fmt.Printf("内容召回率: %.4f\n", report.Summary.AverageContentRecall)
		fmt.Printf("内容F1分数: %.4f\n", report.Summary.AverageContentF1Score)
		fmt.Printf("匹配内容片段: %d/%d\n", report.Summary.TotalMatchedContents, report.Summary.TotalContents)
		fmt.Printf("总查询数: %d\n", report.Summary.TotalQueries)

		fmt.Printf("\n--- 场景排名 ---\n")
		fmt.Printf("最佳场景: %s\n", report.Summary.BestScenario)
		fmt.Printf("最差场景: %s\n", report.Summary.WorstScenario)
	}

	if report.FailedScenarios > 0 {
		fmt.Printf("\n--- 失败场景 ---\n")
		for name, result := range report.Results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if status, exists := resultMap["status"]; exists && status == "failed" {
					fmt.Printf("  %s: %v\n", name, resultMap["error"])
				}
			}
		}
	}

	fmt.Printf(strings.Repeat("=", 60) + "\n")
}

// generateHTMLReport 生成HTML格式报告
func generateHTMLReport(report *TestReport, filePath string) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s - 测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .scenario { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 3px; }
        .success { background-color: #d4edda; }
        .failed { background-color: #f8d7da; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; }
        .metric { background-color: #f8f9fa; padding: 10px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s</h1>
        <p><strong>描述:</strong> %s</p>
        <p><strong>开始时间:</strong> %s</p>
        <p><strong>结束时间:</strong> %s</p>
        <p><strong>总耗时:</strong> %v</p>
        <p><strong>场景统计:</strong> 总计 %d, 成功 %d, 失败 %d</p>
    </div>
    
    <div class="summary">
        <h2>平均性能指标</h2>
        <div class="metrics">
            <div class="metric">准确率: %.4f</div>
            <div class="metric">召回率: %.4f</div>
            <div class="metric">F1分数: %.4f</div>
            <div class="metric">MAP: %.4f</div>
            <div class="metric">NDCG: %.4f</div>
            <div class="metric">响应时间: %.2fms</div>
        </div>
        <h2>内容片段匹配指标</h2>
        <div class="metrics">
            <div class="metric">内容准确率: %.4f</div>
            <div class="metric">内容召回率: %.4f</div>
            <div class="metric">内容F1分数: %.4f</div>
            <div class="metric">匹配内容片段: %d/%d</div>
        </div>
    </div>
    
    <div>
        <h2>场景详情</h2>
`, report.TestName, report.TestName, report.TestDescription,
		report.StartTime.Format("2006-01-02 15:04:05"),
		report.EndTime.Format("2006-01-02 15:04:05"),
		report.Duration, report.TotalScenarios, report.SuccessScenarios, report.FailedScenarios,
		report.Summary.AveragePrecision, report.Summary.AverageRecall,
		report.Summary.AverageF1Score, report.Summary.AverageMAP, report.Summary.AverageNDCG,
		report.Summary.AverageResponseTime, report.Summary.AverageContentPrecision,
		report.Summary.AverageContentRecall, report.Summary.AverageContentF1Score,
		report.Summary.TotalMatchedContents, report.Summary.TotalContents)

	for name, result := range report.Results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			status := resultMap["status"].(string)
			className := "success"
			if status == "failed" {
				className = "failed"
			}

			html += fmt.Sprintf(`        <div class="scenario %s">
            <h3>%s</h3>
`, className, name)

			if status == "failed" {
				html += fmt.Sprintf(`            <p><strong>错误:</strong> %v</p>
`, resultMap["error"])
			} else {
				html += fmt.Sprintf(`            <div class="metrics">
                <div class="metric">准确率: %.4f</div>
                <div class="metric">召回率: %.4f</div>
                <div class="metric">F1分数: %.4f</div>
                <div class="metric">MAP: %.4f</div>
                <div class="metric">NDCG: %.4f</div>
                <div class="metric">响应时间: %.2fms</div>
                <div class="metric">查询数: %d</div>
            </div>
`, resultMap["precision"], resultMap["recall"], resultMap["f1_score"],
					resultMap["map"], resultMap["ndcg"], resultMap["response_time"], resultMap["query_count"])
			}

			html += `        </div>
`
		}
	}

	html += `    </div>
</body>
</html>`

	return ioutil.WriteFile(filePath, []byte(html), 0644)
}

// generateCSVSummary 生成CSV格式摘要
func generateCSVSummary(result *TestResult, filePath string) error {
	var csvContent strings.Builder
	csvContent.WriteString("Scenario,Status,Precision,Recall,F1Score,MAP,NDCG,ContentPrecision,ContentRecall,ContentF1Score,MatchedContents,TotalContents,ResponseTime,QueryCount,Duration\n")

	for name, scenarioResult := range result.Results {
		status := "success"
		if scenarioResult.Error != nil {
			status = "failed"
			csvContent.WriteString(fmt.Sprintf("%s,%s,,,,,,,\n", name, status))
			continue
		}

		metrics := scenarioResult.AverageMetrics
		duration := scenarioResult.EndTime.Sub(scenarioResult.StartTime)
		csvContent.WriteString(fmt.Sprintf("%s,%s,%.4f,%.4f,%.4f,%.4f,%.4f,%.4f,%.4f,%.4f,%d,%d,%.2f,%d,%v\n",
			name, status, metrics.Precision, metrics.Recall, metrics.F1Score,
			metrics.MAP, metrics.NDCG, metrics.ContentPrecision, metrics.ContentRecall,
			metrics.ContentF1Score, metrics.MatchedContents, metrics.TotalContents,
			metrics.ResponseTime, len(scenarioResult.QueryResults), duration))
	}

	return ioutil.WriteFile(filePath, []byte(csvContent.String()), 0644)
}
