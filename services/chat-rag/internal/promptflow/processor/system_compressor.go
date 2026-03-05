package processor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
)

// SYSTEM_SUMMARY_PROMPT defines the template for conversation system prompt summarization
const SYSTEM_SUMMARY_PROMPT = `You are a documentation standardization expert. You will receive a technical specification text. Please compress it to retain key information, operational rules, and core usage principles, while minimizing repetition, verbosity, and secondary descriptions. The goal is to make the content more concise and clear for engineers to quickly understand and implement.

Please strictly follow the requirements below for the compression task:

### Task Objectives:
1. Compress and optimize the technical specification text, extracting key points.
2. Remove redundant or repetitive content and simplify complex sentence structures.
3. Retain key operational rules, tool usage methods, and behavioral constraints.
4. Preserve important restrictions and operational examples completely to avoid missing information.

### Compression Principles:
1. **Information Integrity First**: All necessary usage rules and core constraints must be preserved.
2. **Clear and Concise Expression**: Each sentence should convey a single rule; use bullet points where possible.
3. **Remove Redundant Information**: Eliminate repetitive content, over-explanations, and general knowledge.
4. **Enhance Structural Logic**: Organize content by theme (e.g., tool usage guidelines, editing rules, mode descriptions).

### Output Format:
* Maintain the original Markdown structure and paragraph divisions.
* Divide content into modules using headings (e.g., ## Tool Usage Guidelines).
* Final text length should be 30%-50% of the original to ensure readability, standardization, and structural clarity.
* Output in English only, without additional explanations such as "This is the compressed text.`

// SystemPromptCache is a global singleton cache for system prompt summaries
type SystemPromptCache struct {
	cache map[string]string
	mutex sync.RWMutex
}

var (
	systemPromptCacheInstance *SystemPromptCache
	systemPromptCacheOnce     sync.Once
)

// GetSystemPromptCache returns the singleton instance of SystemPromptCache
func GetSystemPromptCache() *SystemPromptCache {
	systemPromptCacheOnce.Do(func() {
		systemPromptCacheInstance = &SystemPromptCache{
			cache: make(map[string]string),
		}
	})
	return systemPromptCacheInstance
}

// Get retrieves a cached system prompt summary by hash
func (c *SystemPromptCache) Get(hash string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	summary, exists := c.cache[hash]
	return summary, exists
}

// Set stores a system prompt summary in the cache
func (c *SystemPromptCache) Set(hash, summary string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache[hash] = summary
}

// generateHash generates a SHA256 hash for the given content
func generateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

type SystemCompressor struct {
	Recorder
	systemPromptSplitStr string
	llmClient            client.LLMInterface

	next Processor
}

func (s *SystemCompressor) Execute(promptMsg *PromptMsg) {
	logger.Info("starting system prompt compression",
		zap.String("method", "SystemCompressor.Execute"),
	)
	if promptMsg == nil {
		logger.Error("nil prompt message received", zap.String("method", "SystemCompressor.Execute"))
		s.Err = fmt.Errorf("nil prompt message received")
		return
	}

	processedMsg := s.processSystemMessageWithCache(promptMsg.systemMsg)
	promptMsg.systemMsg = processedMsg
	if s.next == nil {
		logger.Warn("system prompt compression completed, but no next processor found",
			zap.String("method", "SystemCompressor.Execute"),
		)
		return
	}

	s.Handled = true
	s.next.Execute(promptMsg)
}

func (s *SystemCompressor) SetNext(next Processor) {
	s.next = next
}

// NewSystemCompressor creates a new system prompt processor with compression logic
func NewSystemCompressor(systemPromptSplitStr string, llmClient client.LLMInterface) *SystemCompressor {
	return &SystemCompressor{
		systemPromptSplitStr: systemPromptSplitStr,
		llmClient:            llmClient,
	}
}

// processSystemMessageWithCache processes system message with caching logic
func (p *SystemCompressor) processSystemMessageWithCache(msg *types.Message) *types.Message {
	var content model.Content

	contents, err := content.ExtractMsgContent(msg)
	if err != nil {
		logger.Warn("failed to extract system content",
			zap.String("method", "processSystemMessageWithCache"),
			zap.Error(err),
		)
		return msg
	}

	if len(contents) != 1 {
		logger.Warn("expected exactly one system content",
			zap.Int("length", len(contents)),
			zap.String("method", "processSystemMessageWithCache"),
		)
		return msg
	}

	systemContent := contents[0].Text
	// Arrange system content with caching
	return p.processContentWithCache(contents, systemContent)
}

// processContentWithCache handles the caching logic for system content
func (p *SystemCompressor) processContentWithCache(content []model.Content, systemContent string) *types.Message {
	// Check if system prompt contains SystemPromptSplitStr
	toolGuidelinesIndex := strings.Index(systemContent, p.systemPromptSplitStr)
	if toolGuidelinesIndex == -1 {
		logger.Warn("No SystemPromptSplitStr found",
			zap.String("method", "processSystemMessageWithCache"),
		)
		return &types.Message{
			Role:    types.RoleSystem,
			Content: content,
		}
	}

	// Split content
	contentBeforeGuidelines := systemContent[:toolGuidelinesIndex]
	contentToCompress := systemContent[toolGuidelinesIndex:]

	// Try to get from cache
	systemHash := generateHash(contentToCompress)
	cache := GetSystemPromptCache()
	if compressedContent, exists := cache.Get(systemHash); exists {
		logger.Info("using cached compressed system prompt",
			zap.String("method", "processSystemMessageWithCache"),
		)
		content[0].Text = contentBeforeGuidelines + compressedContent
		return &types.Message{
			Role:    types.RoleSystem,
			Content: content,
		}
	}

	// Asynchronously compress and cache
	go p.compressAndCache(contentToCompress, systemHash)

	// Return original content
	return &types.Message{
		Role:    types.RoleSystem,
		Content: content,
	}
}

// compressAndCache handles the async compression and caching
func (p *SystemCompressor) compressAndCache(content, hash string) {
	cache := GetSystemPromptCache()
	compressed, err := p.generateSystemPromptSummary(context.Background(), content)
	if err != nil {
		logger.Error("failed to compress system prompt",
			zap.String("method", "processSystemMessageWithCache"),
			zap.Error(err),
		)
		return
	}

	logger.Info("compressed system prompt success",
		zap.String("method", "processSystemMessageWithCache"),
	)
	cache.Set(hash, compressed)
}

// generateSystemPromptSummary generates a system prompt summary of the conversation
func (p *SystemCompressor) generateSystemPromptSummary(ctx context.Context, systemPrompt string) (string, error) {
	logger.Info("generating system prompt summary",
		zap.String("method", "GenerateSystemPromptSummary"),
	)
	// Create a new slice of messages for the summary request
	var summaryMessages []types.Message

	// Add final user instruction
	summaryMessages = append(summaryMessages, types.Message{
		Role:    "user",
		Content: "Please compress the following content:\n\n" + systemPrompt,
	})

	return p.llmClient.GenerateContent(ctx, SYSTEM_SUMMARY_PROMPT, summaryMessages)
}
