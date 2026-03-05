package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PerformanceTestResult 性能测试结果
type PerformanceTestResult struct {
	FirstTokenTimes []float64 // 首token延时列表 (秒)
	TotalTimes      []float64 // 总耗时列表 (秒)
	RequestCount    int       // 成功请求数
	ErrorCount      int       // 错误请求数
	Errors          []string  // 错误信息列表
	mu              sync.Mutex
}

// ChatCompletionRequest 聊天完成请求
type ChatCompletionRequest struct {
	Model         string                 `json:"model"`
	Messages      []Message              `json:"messages"`
	Stream        bool                   `json:"stream,omitempty"`
	Temperature   float64                `json:"temperature,omitempty"`
	StreamOptions map[string]interface{} `json:"stream_options,omitempty"`
	ExtraBody     map[string]interface{} `json:"extra_body,omitempty"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatPerformanceTester 聊天性能测试器
type ChatPerformanceTester struct {
	BaseURL    string
	APIPath    string
	APIURL     string
	Result     *PerformanceTestResult
	DefaultReq *ChatCompletionRequest
}

// NewChatPerformanceTester 创建新的性能测试器
func NewChatPerformanceTester(baseURL string) *ChatPerformanceTester {
	apiPath := "/chat-rag/api/v1/chat/completions"
	return &ChatPerformanceTester{
		BaseURL: baseURL,
		APIPath: apiPath,
		APIURL:  baseURL + apiPath,
		Result:  &PerformanceTestResult{},
		DefaultReq: &ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: "user", Content: "请介绍一下Go语言的特点"},
			},
			Stream:      true,
			Temperature: 0.7,
			StreamOptions: map[string]interface{}{
				"include_usage": true,
			},
		},
	}
}

// SingleRequest 发送单个请求并测量性能
func (t *ChatPerformanceTester) SingleRequest(req *ChatCompletionRequest, requestNum int, customRequestFile string) (firstTokenTime, totalTime float64, success bool, responseContent string, errorMsg string) {
	startTime := time.Now()
	firstTokenTime = 0
	totalTime = 0
	success = false
	responseContent = ""
	errorMsg = ""

	// 如果指定了自定义请求文件，每次都随机加载一个请求
	if customRequestFile != "" {
		randomReq, err := loadRandomRequest(customRequestFile)
		if err != nil {
			errorMsg = fmt.Sprintf("加载随机请求失败: %v", err)
			return 0, 0, false, "", errorMsg
		}
		req = randomReq
	}

	// 生成唯一的请求ID
	requestID := generateUUID()

	// 准备请求体
	reqBody, err := json.Marshal(req)
	if err != nil {
		errorMsg = fmt.Sprintf("序列化失败: %v", err)
		return 0, 0, false, "", errorMsg
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", t.APIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		errorMsg = fmt.Sprintf("创建请求失败: %v", err)
		return 0, 0, false, "", errorMsg
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-quota-identity", "system")
	httpReq.Header.Set("X-Request-ID", requestID)
	httpReq.Header.Set("authorization", "Bearer auth认证token")
	httpReq.Header.Set("zgsm-task-id", "82a65f05-ad43-467b-b0e7-b32a96e4b57b")
	httpReq.Header.Set("zgsm-request-id", requestID)
	httpReq.Header.Set("zgsm-client-id", "ee860a4904edf0b5bf21db45a2c0f9515a1c1bbebd0e943a425f0e96c9c39260")
	httpReq.Header.Set("zgsm-project-path", "d:%5Ccodespace%5Cgoproject%5Cchat-rag")

	// 打印请求开始信息和内容
	requestContent := "默认请求"
	if req != nil && len(req.Messages) > 0 {
		content := req.Messages[0].Content
		if len(content) > 100 {
			requestContent = content[:100] + "..."
		} else {
			requestContent = content
		}
	}
	fmt.Printf("请求 #%d 开始 (Request-ID: %s) 内容: %s\n", requestNum, requestID, requestContent)

	// 发送请求
	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		errorMsg = fmt.Sprintf("请求失败: %v", err)
		fmt.Printf("请求 #%d %s\n", requestNum, errorMsg)
		return 0, 0, false, "", errorMsg
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// 读取错误响应内容
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200] + "..."
		}
		errorMsg = fmt.Sprintf("HTTP %d, 响应内容: %s", resp.StatusCode, bodyStr)
		fmt.Printf("请求 #%d 失败: %s\n", requestNum, errorMsg)
		return 0, 0, false, "", errorMsg
	}

	// 处理流式响应
	firstChunkReceived := false
	var responseBuilder strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			dataStr := line[6:] // 去掉 'data: ' 前缀

			if dataStr == "[DONE]" {
				// 流结束
				totalTime = time.Since(startTime).Seconds()
				success = true
				break
			}

			// fmt.Println(dataStr)

			if !firstChunkReceived {
				// 记录首token时间并立即打印
				firstTokenTime = time.Since(startTime).Seconds()
				firstChunkReceived = true
				fmt.Printf("请求 #%d 首token延时: %.3fs (Request-ID: %s)\n", requestNum, firstTokenTime, requestID)
			}

			// 尝试解析JSON并收集响应内容
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
				continue
			}

			// 提取内容
			if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							responseBuilder.WriteString(content)
						}
					}
				}
			}
		}
	}

	// 如果没有收到结束标志，使用当前时间作为总耗时
	if totalTime == 0 {
		totalTime = time.Since(startTime).Seconds()
	}

	responseContent = responseBuilder.String()

	return firstTokenTime, totalTime, success, responseContent, errorMsg
}

// RunConcurrentTest 运行并发测试
func (t *ChatPerformanceTester) RunConcurrentTest(concurrency, totalRequests int, customReq *ChatCompletionRequest, customRequestFile string) *PerformanceTestResult {
	t.Result = &PerformanceTestResult{}

	req := customReq
	if req == nil {
		req = t.DefaultReq
	}

	// 创建带缓冲的通道控制并发
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	requestCounter := 0
	var counterMutex sync.Mutex

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 获取请求编号
			counterMutex.Lock()
			requestCounter++
			requestID := requestCounter
			counterMutex.Unlock()

			// 执行请求
			firstTokenTime, totalTime, success, responseContent, errorMsg := t.SingleRequest(req, requestID, customRequestFile)

			// 请求结束时打印结果
			if success {
				fmt.Printf("请求 #%d 结束: 总耗时=%.3fs\n", requestID, totalTime)
				if len(responseContent) > 200 {
					fmt.Printf("请求 #%d 响应内容: %s...\n", requestID, responseContent[:200])
				} else if len(responseContent) > 0 {
					fmt.Printf("请求 #%d 响应内容: %s\n", requestID, responseContent)
				} else {
					fmt.Printf("请求 #%d 响应内容: [空]\n", requestID)
				}

				t.Result.mu.Lock()
				t.Result.FirstTokenTimes = append(t.Result.FirstTokenTimes, firstTokenTime)
				t.Result.TotalTimes = append(t.Result.TotalTimes, totalTime)
				t.Result.RequestCount++
				t.Result.mu.Unlock()
			} else {
				fmt.Printf("请求 #%d 失败: %s\n", requestID, errorMsg)

				t.Result.mu.Lock()
				t.Result.ErrorCount++
				t.Result.Errors = append(t.Result.Errors, errorMsg)
				t.Result.mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return t.Result
}

// PrintResults 打印测试结果
func (t *ChatPerformanceTester) PrintResults(result *PerformanceTestResult) {
	if result == nil {
		result = t.Result
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("性能测试结果")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("总请求数: %d\n", result.RequestCount+result.ErrorCount)
	fmt.Printf("成功请求数: %d\n", result.RequestCount)
	fmt.Printf("失败请求数: %d\n", result.ErrorCount)
	if result.RequestCount+result.ErrorCount > 0 {
		fmt.Printf("成功率: %.2f%%\n", float64(result.RequestCount)/float64(result.RequestCount+result.ErrorCount)*100)
	}

	if len(result.FirstTokenTimes) > 0 && len(result.TotalTimes) > 0 {
		// 计算统计数据
		firstTokenStats := calculateStats(result.FirstTokenTimes)
		totalTimeStats := calculateStats(result.TotalTimes)

		fmt.Println("\n首token延时 (秒):")
		fmt.Printf("  平均值: %.3f\n", firstTokenStats.Mean)
		fmt.Printf("  中位数: %.3f\n", firstTokenStats.Median)
		fmt.Printf("  最小值: %.3f\n", firstTokenStats.Min)
		fmt.Printf("  最大值: %.3f\n", firstTokenStats.Max)
		fmt.Printf("  P95: %.3f\n", firstTokenStats.P95)
		fmt.Printf("  P99: %.3f\n", firstTokenStats.P99)

		fmt.Println("\n总耗时 (秒):")
		fmt.Printf("  平均值: %.3f\n", totalTimeStats.Mean)
		fmt.Printf("  中位数: %.3f\n", totalTimeStats.Median)
		fmt.Printf("  最小值: %.3f\n", totalTimeStats.Min)
		fmt.Printf("  最大值: %.3f\n", totalTimeStats.Max)
		fmt.Printf("  P95: %.3f\n", totalTimeStats.P95)
		fmt.Printf("  P99: %.3f\n", totalTimeStats.P99)
	}

	if len(result.Errors) > 0 {
		fmt.Println("\n错误信息 (前5个):")
		for i, error := range result.Errors {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %s\n", error)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// Stats 统计数据
type Stats struct {
	Mean   float64
	Median float64
	Min    float64
	Max    float64
	P95    float64
	P99    float64
}

// calculateStats 计算统计数据
func calculateStats(values []float64) Stats {
	if len(values) == 0 {
		return Stats{}
	}

	// 排序
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// 计算均值
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	mean := sum / float64(len(sorted))

	// 计算中位数
	median := sorted[len(sorted)/2]
	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	}

	// 计算百分位数
	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	p95 := sorted[p95Index]

	p99Index := int(float64(len(sorted)) * 0.99)
	if p99Index >= len(sorted) {
		p99Index = len(sorted) - 1
	}
	p99 := sorted[p99Index]

	return Stats{
		Mean:   mean,
		Median: median,
		Min:    sorted[0],
		Max:    sorted[len(sorted)-1],
		P95:    p95,
		P99:    p99,
	}
}

// generateUUID 生成UUID
func generateUUID() string {
	// 使用标准UUID库生成UUID
	u, err := uuid.NewRandom()
	if err != nil {
		// 如果UUID生成失败，使用备用方法
		timestamp := time.Now().UnixNano()
		randBytes := make([]byte, 8)
		rand.Read(randBytes)
		return fmt.Sprintf("%x-%x", timestamp, randBytes)
	}

	// 在UUID基础上添加额外的随机数确保唯一性
	extraRand := make([]byte, 4)
	rand.Read(extraRand)

	return fmt.Sprintf("%s-%x", u.String(), extraRand)
}

// truncateString 截断字符串
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// loadCustomRequest 加载自定义请求
func loadCustomRequest(filename string) (*ChatCompletionRequest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var req ChatCompletionRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

// loadRandomRequest 从请求列表中随机加载一个请求
func loadRandomRequest(filename string) (*ChatCompletionRequest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var requests []ChatCompletionRequest
	err = json.Unmarshal(data, &requests)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return nil, fmt.Errorf("请求列表为空")
	}

	// 随机选择一个请求
	randomIndex := rand.Intn(len(requests))
	return &requests[randomIndex], nil
}

// getActualMessage 获取实际请求内容用于显示
func getActualMessage(req *ChatCompletionRequest) string {
	if len(req.Messages) > 0 {
		content := req.Messages[0].Content
		return truncateString(content, 10)
	}
	return "未知"
}

func main() {
	// 命令行参数
	url := flag.String("url", "http://localhost:8888", "API服务地址")
	concurrency := flag.Int("concurrency", 5, "并发数")
	requests := flag.Int("requests", 10, "总请求数")
	model := flag.String("model", "gpt-4", "模型名称")
	message := flag.String("message", "请介绍一下Go语言的特点", "测试消息")
	stream := flag.Bool("stream", true, "使用流式响应")
	customRequest := flag.String("custom-request", "", "自定义请求JSON文件路径")
	flag.Parse()

	// 创建测试器
	tester := NewChatPerformanceTester(*url)

	// 准备请求数据
	var req *ChatCompletionRequest
	var actualMessage string

	if *customRequest != "" {
		// 如果提供了自定义请求文件，使用随机加载功能
		// 这里不直接加载，而是在每次请求时随机加载
		req = nil // 设置为nil，让SingleRequest函数处理随机加载
		actualMessage = "随机请求"
	} else {
		// 否则使用默认请求和命令行参数
		req = &ChatCompletionRequest{
			Model:       *model,
			Messages:    []Message{{Role: "user", Content: *message}},
			Stream:      *stream,
			Temperature: 0.7,
			StreamOptions: map[string]interface{}{
				"include_usage": true,
			},
		}
		actualMessage = truncateString(*message, 10)
	}

	fmt.Println("开始性能测试...")
	fmt.Printf("API地址: %s\n", *url)
	fmt.Printf("并发数: %d\n", *concurrency)
	fmt.Printf("总请求数: %d\n", *requests)
	fmt.Printf("请求内容: %s\n", actualMessage)
	if req != nil {
		fmt.Printf("使用流式响应: %t\n", req.Stream)
	} else {
		fmt.Printf("使用流式响应: 随机请求\n")
	}

	// 运行测试
	startTime := time.Now()
	result := tester.RunConcurrentTest(*concurrency, *requests, req, *customRequest)
	endTime := time.Now()

	fmt.Printf("\n测试完成，总耗时: %.2f 秒\n", endTime.Sub(startTime).Seconds())

	// 打印结果
	tester.PrintResults(result)
}
