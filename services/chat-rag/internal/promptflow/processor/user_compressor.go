package processor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/tokenizer"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"github.com/zgsm-ai/chat-rag/internal/utils"
	"go.uber.org/zap"
)

// USER_SUMMARY_PROMPT defines the template for conversation user prompt summarization
const USER_SUMMARY_PROMPT = `Your task is to create a detailed summary of the conversation so far, paying close attention to the user's explicit requests and your previous actions.
This summary should be thorough in capturing technical details, code patterns, and architectural decisions that would be essential for continuing with the conversation and supporting any continuing tasks.

Your summary should be structured as follows:
Context: The context to continue the conversation with. If applicable based on the current task, this should include:
  1. Previous Conversation: High level details about what was discussed throughout the entire conversation with the user. This should be written to allow someone to be able to follow the general overarching conversation flow.
  2. Current Work: Describe in detail what was being worked on prior to this request to summarize the conversation. Pay special attention to the more recent messages in the conversation.
  3. Key Technical Concepts: List all important technical concepts, technologies, coding conventions, and frameworks discussed, which might be relevant for continuing with this work.
  4. Relevant Files and Code: If applicable, enumerate specific files and code sections examined, modified, or created for the task continuation. Pay special attention to the most recent messages and changes.
  5. Problem Solving: Document problems solved thus far and any ongoing troubleshooting efforts.
  6. Pending Tasks and Next Steps: Outline all pending tasks that you have explicitly been asked to work on, as well as list the next steps you will take for all outstanding work, if applicable. Include code snippets where they add clarity. For any next steps, include direct quotes from the most recent conversation showing exactly what task you were working on and where you left off. This should be verbatim to ensure there's no information loss in context between tasks.
  7. Language: Emphasize the language mentioned by system.

Example summary structure:
1. Previous Conversation:
  [Detailed description]
2. Current Work:
  [Detailed description]
3. Key Technical Concepts:
  - [Concept 1]
  - [Concept 2]
  - [...]
4. Relevant Files and Code:
  - [File Name 1]
    - [Summary of why this file is important]
    - [Summary of the changes made to this file, if any]
    - [Important Code Snippet]
  - [File Name 2]
    - [Important Code Snippet]
  - [...]
5. Problem Solving:
  [Detailed description]
6. Pending Tasks and Next Steps:
  - [Task 1 details & next steps]
  - [Task 2 details & next steps]
  - [...]
7. Langeuage:
	[Always answer in the language]

Output only the summary of the conversation so far, without any additional commentary or explanation.`

// Deprecated
type UserCompressor struct {
	Recorder
	ctx          context.Context
	config       config.Config
	llmClient    client.LLMInterface
	tokenCounter *tokenizer.TokenCounter

	next Processor
}

func NewUserCompressor(
	ctx context.Context,
	config config.Config,
	llmClient client.LLMInterface,
	tokenCounter *tokenizer.TokenCounter,
) *UserCompressor {
	return &UserCompressor{
		ctx:          ctx,
		config:       config,
		llmClient:    llmClient,
		tokenCounter: tokenCounter,
	}
}

func (u *UserCompressor) Execute(promptMsg *PromptMsg) {
	const method = "UserCompressor.Execute"

	if promptMsg == nil {
		logger.Error("nil prompt message received", zap.String("method", method))
		u.Err = fmt.Errorf("nil prompt message received")
		return
	}

	startTime := time.Now()
	defer func() {
		u.Latency = time.Since(startTime).Milliseconds()
	}()

	// Check if user message needs to be compressed
	userMsgList := append(promptMsg.olderUserMsgList, *promptMsg.lastUserMsg)
	userMessageTokens := u.tokenCounter.CountMessagesTokens(userMsgList)
	needsCompressUserMsg := u.config.ContextCompressConfig.EnableCompress &&
		userMessageTokens > u.config.ContextCompressConfig.TokenThreshold
	logger.Info("user message tokens",
		zap.Int("tokens", userMessageTokens),
		zap.Bool("needsCompression", needsCompressUserMsg),
		zap.String("method", method),
	)

	if !needsCompressUserMsg {
		logger.Info("no need to compress user message", zap.String("method", method))
		u.passToNext(promptMsg)
		return
	}

	// Split out the messages that need to be summarized from olderUserMsgList according to the threshold
	messagesToSummarize, retainedMessages := u.trimMessagesToTokenThreshold(promptMsg.olderUserMsgList)
	if len(messagesToSummarize) == 0 {
		logger.Info("no messages to summarize", zap.String("method", method))
		u.passToNext(promptMsg)
		return
	}

	summary, err := u.compressMessages(messagesToSummarize)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(u.ctx.Err(), context.Canceled) {
			logger.Warn("Context canceled during message compression",
				zap.Error(err),
				zap.String("method", method),
			)
		} else {
			logger.Error("failed to compress messages",
				zap.Error(err),
				zap.String("method", method),
			)
		}
		u.Err = err
		u.passToNext(promptMsg)
		return
	}

	u.updatePromptMessages(promptMsg, summary, retainedMessages)
	u.Handled = true
	u.passToNext(promptMsg)
}

func (u *UserCompressor) SetNext(next Processor) {
	u.next = next
}

func (u *UserCompressor) passToNext(promptMsg *PromptMsg) {
	if u.next == nil {
		logger.Warn("user compression completed but no next processor configured",
			zap.String("method", "UserCompressor.Execute"),
		)
		return
	}
	u.next.Execute(promptMsg)
}

func (u *UserCompressor) compressMessages(messages []types.Message) (string, error) {
	// Add final user instruction
	messagesToSummarize := make([]types.Message, len(messages), len(messages)+1)
	copy(messagesToSummarize, messages)
	messagesToSummarize = append(messagesToSummarize, types.Message{
		Role:    types.RoleUser,
		Content: "Summarize the conversation so far, as described in the prompt instructions.",
	})

	summary, err := u.llmClient.GenerateContent(
		u.ctx,
		USER_SUMMARY_PROMPT,
		messagesToSummarize,
	)
	if err != nil {
		return "", fmt.Errorf("LLM generate content failed in UserCompressor: %w", err)
	}
	return summary, nil
}

func (u *UserCompressor) updatePromptMessages(promptMsg *PromptMsg, summary string, retained []types.Message) {
	var compressedMessages []types.Message
	compressedMessages = append(compressedMessages, types.Message{
		Role:    types.RoleAssistant,
		Content: summary,
	})
	compressedMessages = append(compressedMessages, retained...)
	promptMsg.olderUserMsgList = compressedMessages
}

func (u *UserCompressor) trimMessagesToTokenThreshold(messages []types.Message) ([]types.Message, []types.Message) {
	const method = "UserCompressor.trimMessagesToTokenThreshold"

	if len(messages) <= u.config.ContextCompressConfig.RecentUserMsgUsedNums {
		logger.Warn("no enough messages to trim",
			zap.Int("messages length", len(messages)),
			zap.Int("RecentUserMsgUsedNums", u.config.ContextCompressConfig.RecentUserMsgUsedNums),
		)
		return []types.Message{}, messages
	}

	messagesToSummarize := utils.GetOldUserMsgsWithNum(messages, u.config.ContextCompressConfig.RecentUserMsgUsedNums)
	retainedMessages := utils.GetRecentUserMsgsWithNum(messages, u.config.ContextCompressConfig.RecentUserMsgUsedNums)

	currentTokens := u.tokenCounter.CountMessagesTokens(messagesToSummarize)
	bufferTokens := 5000 // buffer for summary tokens
	totalTokens := currentTokens + bufferTokens

	var removedCount int
	for totalTokens > u.config.ContextCompressConfig.SummaryModelTokenThreshold && len(messagesToSummarize) > 0 {
		removedTokens := u.tokenCounter.CountOneMessageTokens(messagesToSummarize[0])
		totalTokens -= removedTokens
		messagesToSummarize = messagesToSummarize[1:]
		removedCount++
	}

	logger.Info("message token statistics",
		zap.Int("totalTokens", totalTokens),
		zap.Int("remainingMessages", len(messagesToSummarize)),
		zap.Int("removedMessages", removedCount),
		zap.String("method", method),
	)

	return messagesToSummarize, retainedMessages
}
