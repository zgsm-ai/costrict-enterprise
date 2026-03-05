package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"go.uber.org/zap"
)

/*
const systemClassificationPrompt = `Classify the LAST USER QUESTION in this conversation into ONE of the following EXACT categories based on the user's intention (respond ONLY with the exact category name, no extra text):

- CodeWriting: Writing or generating code to implement functionality
- BugFixing: Fixing errors, bugs, or unexpected behavior in existing code
- CodeUnderstanding: Understanding how code works or asking about programming concepts
- CodeRefactoring: Improving code readability, structure, or maintainability without changing its functionality
- DesignDiscussion: Discussing software design, architecture, or best practices
- DocumentationHelp: Asking about writing or understanding documentation, comments, or code explanations
- EnvironmentHelp: Setting up or troubleshooting the development environment, dependencies, or tools
- ToolUsage: Questions about using development tools, IDEs, debuggers, or plugins
- GeneralQuestion: Any question unrelated to code or development tasks`

const userClassificationPrompt = `
Respond ONLY with one of these exact category names:
- "CodeWriting"
- "BugFixing"
- "CodeUnderstanding"
- "CodeRefactoring"
- "DesignDiscussion"
- "DocumentationHelp"
- "EnvironmentHelp"
- "ToolUsage"
- "GeneralQuestion"

Do not include any extra text, just the exact matching category name.`

// validCategories is a documentation string listing all accepted log categories
const validCategoriesStr = "CodeWriting,BugFixing,CodeUnderstanding,CodeRefactoring,DesignDiscussion,DocumentationHelp,EnvironmentHelp,ToolUsage,GeneralQuestion"
*/

// LogRecordInterface defines the interface for the logger service
type LogRecordInterface interface {
	// Start starts the logger service
	Start() error
	// Stop stops the logger service
	Stop()
	// LogAsync logs a chat completion asynchronously
	LogAsync(logs *model.ChatLog, headers *http.Header)
	// LogSync logs a chat completion synchronously
	SetMetricsService(metricsService MetricsInterface)
}

// LoggerRecordService handles logging operations
type LoggerRecordService struct {
	// tempLogFilePath      string // Temporary log file path - no longer needed
	// scanInterval         time.Duration
	// llmConfig      config.LLMConfig
	// classifyModel  string
	// llmClient            client.LLMInterface

	logFilePath    string // Permanent storage log directory path
	metricsService MetricsInterface
	deptClient     client.DepartmentInterface
	instanceID     string
	// enableClassification bool

	logChan         chan *model.ChatLog
	stopChan        chan struct{}
	wg              sync.WaitGroup
	mu              sync.Mutex
	metricsReporter *ChatMetricsReporter

	// processorStarted bool
}

// NewLogRecordService creates a new logger service
func NewLogRecordService(config config.Config) LogRecordInterface {
	// Create temp directory under logFilePath for temporary log files
	// tempLogDir := filepath.Join(config.Log.LogFilePath, "temp") // No longer needed

	instanceID := os.Getenv("HOSTNAME")
	if instanceID == "" {
		instanceID = fmt.Sprintf("instance-%d", rand.Intn(10000))
	}

	var deptClient client.DepartmentInterface
	if config.DepartmentApiEndpoint != "" {
		deptClient = client.NewDepartmentClient(config.DepartmentApiEndpoint)
	}
	var metricsReporter *ChatMetricsReporter = nil
	if config.ChatMetrics.Enabled {
		metricsReporter = NewChatMetricsReporter(config.ChatMetrics.Url, config.ChatMetrics.Method)
	}

	return &LoggerRecordService{
		// tempLogFilePath:      tempLogDir,             // Temporary logs directory - no longer needed
		// scanInterval:         time.Duration(config.Log.LogScanIntervalSec) * time.Second,
		// llmConfig:            config.LLM,
		// classifyModel:        config.Log.ClassifyModel,
		// enableClassification: config.Log.EnableClassification,

		logFilePath:     config.Log.LogFilePath, // Permanent storage directory
		logChan:         make(chan *model.ChatLog, 1000),
		stopChan:        make(chan struct{}),
		instanceID:      instanceID,
		deptClient:      deptClient,
		metricsReporter: metricsReporter,
	}
}

// SetMetricsService sets the metrics service for the logger
func (ls *LoggerRecordService) SetMetricsService(metricsService MetricsInterface) {
	ls.metricsService = metricsService
}

// Start starts the logger service
func (ls *LoggerRecordService) Start() error {
	logger.Info("==> Start logger")
	// Ensure permanent log directory exists
	if err := os.MkdirAll(ls.logFilePath, 0755); err != nil {
		return fmt.Errorf("failed to create permanent log directory: %w", err)
	}

	// Temp directory creation is no longer needed since we write directly to permanent storage
	/*
		// Ensure temp log directory exists
		if err := os.MkdirAll(filepath.Dir(ls.tempLogFilePath), 0755); err != nil {
			return fmt.Errorf("failed to create temp log directory: %w", err)
		}
	*/

	// Start log writer goroutine
	ls.wg.Add(1)
	go ls.logWriter()

	return nil
}

// Stop stops the logger service
func (ls *LoggerRecordService) Stop() {
	close(ls.stopChan)
	close(ls.logChan)
	ls.wg.Wait()
}

/*
// copyAndSetQuotaIdentity
func copyAndSetQuotaIdentity(headers *http.Header) *http.Header {
	headersCopy := make(http.Header)
	for k, v := range *headers {
		headersCopy[k] = v
	}
	headersCopy.Set(types.HeaderQuotaIdentity, "system")
	return &headersCopy
}
*/

// LogAsync logs a chat completion asynchronously
func (ls *LoggerRecordService) LogAsync(logs *model.ChatLog, headers *http.Header) {
	/*
		// Use default timeout config for log classification
		timeoutCfg := config.LLMTimeoutConfig{
			IdleTimeoutMs:      30000,
			TotalIdleTimeoutMs: 30000,
		}
		llmClient, err := client.NewLLMClient(ls.llmConfig, timeoutCfg, ls.classifyModel, copyAndSetQuotaIdentity(headers))
		if err != nil {
			logger.Error("Failed to create LLM client",
				zap.String("operation", "LogAsync"),
				zap.Error(err),
			)
			return
		}

		ls.llmClient = llmClient
	*/

	select {
	case ls.logChan <- logs:
	default:
		// Channel is full, log directly to storage to avoid blocking
		ls.logDirectToStorage(logs)
		// Original code: ls.logSync(logs)
	}

	// No longer need to start the logProcessor since we're writing directly to storage
	/*
		if !ls.processorStarted {
			ls.mu.Lock()
			defer ls.mu.Unlock()
			if !ls.processorStarted {
				ls.processorStarted = true
				ls.wg.Add(1)
				go ls.logProcessor()
			}
		}
	*/
}

// logWriter writes logs to file
func (ls *LoggerRecordService) logWriter() {
	defer ls.wg.Done()

	for {
		select {
		case log := <-ls.logChan:
			if log != nil {
				// ls.logSync(log)
				ls.logDirectToStorage(log)
			}
		case <-ls.stopChan:
			// Arrange remaining logs
			for len(ls.logChan) > 0 {
				log := <-ls.logChan
				if log != nil {
					// ls.logSync(log)
					ls.logDirectToStorage(log)
				}
			}
			return
		}
	}
}

// writeLogToFile writes log content to specified file path
func (ls *LoggerRecordService) writeLogToFile(filePath string, content string, mode int) error {
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file with specified mode
	file, err := os.OpenFile(filePath, mode, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Convert to raw bytes to avoid any string escaping
	contentBytes := []byte(content)
	contentBytes = append(contentBytes, '\n') // Add newline as raw byte

	// Write content as raw bytes
	if _, err := file.Write(contentBytes); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateRandomNumber creates a 6-digit random number from 100000 to 999999
func (ls *LoggerRecordService) generateRandomNumber() int {
	return rand.Intn(900000) + 100000
}

/* Deprecated
// logSync writes a log entry to temp file synchronously - kept for backward compatibility
func (ls *LoggerRecordService) logSync(logs *model.ChatLog) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Since tempLogFilePath is no longer available, we'll use a temp directory under logFilePath
	tempDir := filepath.Join(ls.logFilePath, "temp")

	// Create timestamped filename
	datePart := logs.Timestamp.Format("20060102")
	timePart := logs.Timestamp.Format("150405")
	username := ls.sanitizeFilename(logs.Identity.UserName, "unknown")
	randNum := ls.generateRandomNumber()
	filename := fmt.Sprintf("%s-%s-%s-%d-%s.log", datePart, timePart, username, randNum, ls.instanceID)
	filePath := filepath.Join(tempDir, filename)

	logJSON, err := logs.ToCompressedJSON()
	if err != nil {
		logger.Error("Failed to marshal log",
			zap.Error(err),
		)
		return
	}

	if err := ls.writeLogToFile(filePath, logJSON, os.O_CREATE|os.O_WRONLY); err != nil {
		logger.Error("Failed to write temp log",
			zap.Error(err),
		)
	}
}
*/

// logDirectToStorage processes and writes a log entry directly to permanent storage
func (ls *LoggerRecordService) logDirectToStorage(logs *model.ChatLog) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if logs == nil {
		logger.Error("Invalid log entry")
		return
	}

	// Get department info
	ls.getDepartment(logs)

	// Record metrics if available
	if ls.metricsService != nil {
		ls.metricsService.RecordChatLog(logs)
	}

	// Save directly to permanent storage
	ls.saveLogToPermanentStorage(logs)
}

/*
// logProcessor processes logs periodically - no longer needed since we write directly to permanent storage
func (ls *LoggerRecordService) logProcessor() {
	logger.Info("==> start logProcessor")
	defer ls.wg.Done()

	ticker := time.NewTicker(ls.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ls.processLogs()
		case <-ls.stopChan:
			// Arrange logs one last time before stopping
			ls.processLogs()
			return
		}
	}
}

func (ls *LoggerRecordService) processLogs() {
	files, err := ls.getLogFiles()
	if err != nil || len(files) == 0 {
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 2) // Limit concurrent goroutines

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(f os.DirEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			ls.processSingleFile(f)
		}(file)
	}

	wg.Wait()
}

// getLogFiles retrieves log files from the temporary log directory - no longer needed
func (ls *LoggerRecordService) getLogFiles() ([]os.DirEntry, error) {
	files, err := os.ReadDir(ls.tempLogFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		logger.Error("Failed to list log files", zap.Error(err))
		return nil, err
	}

	// Filter out non-log files
	var validFiles []os.DirEntry
	for _, file := range files {
		name := file.Name()
		if (strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".json")) &&
			strings.Contains(name, ls.instanceID) {
			validFiles = append(validFiles, file)
		}
	}

	return validFiles, nil
}

// processSingleFile processes a single log file - no longer needed
func (ls *LoggerRecordService) processSingleFile(file os.DirEntry) {
	filePath := filepath.Join(ls.tempLogFilePath, file.Name())

	// 1. Read log file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("Log file not found",
				zap.String("filename", file.Name()),
				zap.Error(err),
			)
		} else {
			logger.Error("Failed to read log file",
				zap.String("filename", file.Name()),
				zap.Error(err),
			)
		}
		return
	}

	// 2. Parse log content
	chatLog, err := model.FromJSON(string(content))
	if err != nil {
		logger.Error("Failed to parse log file",
			zap.String("filename", file.Name()),
			zap.Error(err),
		)
		return
	}

	// 3. Arrange classification
	if ls.enableClassification {
		if err := ls.processClassification(chatLog, filePath); err != nil {
			logger.Error("Failed to process classification",
				zap.String("filename", file.Name()),
				zap.Error(err),
			)
			return
		}
	}

	// 4. Get department info
	ls.getDepartment(chatLog)

	// 5. Upload and process log
	if err := ls.uploadAndProcessLog(chatLog, file); err != nil {
		logger.Error("Failed to upload and process log",
			zap.String("filename", file.Name()),
			zap.Error(err),
		)
	}
}
*/

func (ls *LoggerRecordService) getDepartment(chatLog *model.ChatLog) {
	if chatLog.Identity.UserInfo.EmployeeNumber == "" {
		return
	}

	if ls.deptClient == nil {
		return
	}

	deptInfo, err := ls.deptClient.GetDepartment(chatLog.Identity.UserInfo.EmployeeNumber)
	if err != nil {
		logger.Error("Failed to get department info",
			zap.String("employeeNumber", chatLog.Identity.UserInfo.EmployeeNumber),
			zap.Error(err),
		)

		return
	}

	chatLog.Identity.UserInfo.Department = deptInfo
}

/*
// processClassification processes the classification of a single log entry - no longer needed
func (ls *LoggerRecordService) processClassification(chatLog *model.ChatLog, filePath string) error {
	if chatLog.Identity.Caller == "review-checker" {
		chatLog.Category = "CodeReview"
	}

	if chatLog.Category != "" {
		return nil
	}

	chatLog.Category = ls.classifyLog(chatLog)
	logJSON, err := chatLog.ToCompressedJSON()
	if err != nil {
		return fmt.Errorf("marshal updated log: %w", err)
	}

	if err := ls.writeLogToFile(filePath, logJSON, os.O_WRONLY|os.O_TRUNC); err != nil {
		return fmt.Errorf("update temp log file: %w", err)
	}

	return nil
}

// uploadAndProcessLog uploads a single log to Loki and saves it to permanent storage - no longer needed
func (ls *LoggerRecordService) uploadAndProcessLog(chatLog *model.ChatLog, file os.DirEntry) error {
	if ls.metricsService != nil {
		ls.metricsService.RecordChatLog(chatLog)
	}

	ls.saveLogToPermanentStorage(chatLog)
	ls.deleteTempLogFile(filepath.Join(ls.tempLogFilePath, file.Name()))

	return nil
}
*/

/*
// classifyLog classifies a single log entry
func (ls *LoggerRecordService) classifyLog(logs *model.ChatLog) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// classify uses the recent 2 user messages
	userMessages := utils.GetRecentUserMsgsWithNum(logs.ProcessedPrompt, 2)
	userMessages = append(userMessages, types.Message{
		Role:    types.RoleUser,
		Content: userClassificationPrompt,
	})

	category, err := ls.llmClient.GenerateContent(ctx, systemClassificationPrompt, userMessages)
	if err != nil {
		logger.Error("Failed to classify log",
			zap.Error(err),
		)
		return "unknown"
	}

	validatedCategory := ls.validateCategory(category)
	logger.Info("Log classification result",
		zap.String("category", validatedCategory),
		zap.String("model", ls.llmClient.GetModelName()),
	)

	return validatedCategory
}

// validateCategory checks if the LLM generated category is valid, returns "extra" if not
func (ls *LoggerRecordService) validateCategory(category string) string {
	valid := strings.Split(validCategoriesStr, ",")
	for _, v := range valid {
		if category == v {
			return category
		}
	}

	logger.Debug("Invalid category detected",
		zap.String("category", category),
	)
	return "extra"
}
*/

// saveLogToPermanentStorage saves a single log to permanent storage
func (ls *LoggerRecordService) saveLogToPermanentStorage(chatLog *model.ChatLog) {
	if chatLog == nil {
		logger.Error("Invalid log or missing required identity fields")
		return
	}

	// Directory structure: year-month/day/username
	yearMonth := chatLog.Timestamp.Format("2006-01")
	day := chatLog.Timestamp.Format("02")
	// Get and sanitize username for filesystem usage
	username := ls.sanitizeFilename(chatLog.Identity.UserName, "unknown")

	// Create hierarchical directory path
	dateDir := filepath.Join(ls.logFilePath, yearMonth, day, username)

	// Timestamp for filename: yyyymmdd-HHMMSS_requestID.json
	timestamp := chatLog.Timestamp.Format("20060102-150405")
	requestId := chatLog.Identity.RequestID
	if requestId == "" {
		requestId = "null"
	}
	filename := fmt.Sprintf("%s_%s_%d.json", timestamp, requestId, ls.generateRandomNumber())

	// Full file path
	logFile := filepath.Join(dateDir, filename)

	// Convert to pretty JSON using the extracted method
	jsonStr, err := chatLog.ToPrettyJSON()
	if err != nil {
		logger.Error("Failed to marshal log for permanent storage",
			zap.Error(err),
		)
		return
	}
	if ls.metricsReporter != nil {
		var e string = ""
		if len(chatLog.Error) > 0 {
			// first item's first key
			for key, _ := range chatLog.Error[0] {
				e = string(key)
				break
			}
		}
		go ls.metricsReporter.ReportMetrics(chatLog, e) // async report metrics
	}

	// Create new file instead of appending
	if err := ls.writeLogToFile(logFile, jsonStr, os.O_CREATE|os.O_WRONLY); err != nil {
		logger.Error("Failed to write log to permanent storage",
			zap.Error(err),
		)
		return
	}

	logger.Info("Log saved in storage", zap.String("fileName", logFile))
}

/*
// deleteTempLogFile deletes a single temp log file - no longer needed
func (ls *LoggerRecordService) deleteTempLogFile(filePath string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if err := os.Remove(filePath); err != nil {
		logger.Error("Failed to remove temp log file",
			zap.String("filename", filepath.Base(filePath)),
			zap.Error(err),
		)
	}
}
*/

// sanitizeFilename cleans a string to make it safe for use in file/folder names
func (ls *LoggerRecordService) sanitizeFilename(name string, defaultName string) string {
	if name == "" {
		return defaultName
	}

	// Remove invalid characters for both Windows and Linux
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|", "\x00", "\n", "\r", "\t"}
	// Also replace any non-printable ASCII characters
	for i := 0; i < 32; i++ {
		invalidChars = append(invalidChars, string(rune(i)))
	}
	for _, c := range invalidChars {
		name = strings.ReplaceAll(name, c, "")
	}

	// Limit length to 255 bytes for Linux compatibility
	if len(name) > 255 {
		name = name[:255]
	}

	return name
}
