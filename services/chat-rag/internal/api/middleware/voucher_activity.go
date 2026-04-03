package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/api/helper"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/service"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// VoucherActivityMiddleware handles voucher activity logic
func VoucherActivityMiddleware(svcCtx *bootstrap.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1. Check if voucher activity is enabled
		if svcCtx.Config.VoucherActivityConfig == nil || !svcCtx.Config.VoucherActivityConfig.Enabled {
			c.Next()
			return
		}

		voucherConfig := svcCtx.Config.VoucherActivityConfig
		logger.InfoC(ctx, "voucher activity is enable, start to process voucher activity",
			zap.Int("activities count", len(svcCtx.Config.VoucherActivityConfig.Activities)))

		// 2. Get identity from context
		identity, exists := model.GetIdentityFromContext(ctx)
		if !exists || identity == nil || identity.UserInfo == nil {
			logger.WarnC(ctx, "Failed to get identity from context in voucher activity middleware")
			c.Next()
			return
		}

		// 3. Parse request body to get messages
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.ErrorC(ctx, "Failed to read request body", zap.Error(err))
			c.Next()
			return
		}
		// Restore body for subsequent handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		var req struct {
			Messages []types.Message `json:"messages"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			logger.ErrorC(ctx, "Failed to parse request body", zap.Error(err))
			c.Next()
			return
		}

		// 4. Check message role
		if len(req.Messages) == 0 {
			c.Next()
			return
		}

		// Excluding cases where CLI automatically retrieves the title
		if len(req.Messages) == 3 {
			secondMessage := fmt.Sprintf("%v", req.Messages[1].Content)
			if strings.Contains(secondMessage, "Generate a title for this conversation") {
				logger.InfoC(ctx, "Cli generating titile request, skip")
				c.Next()
				return
			}
		}

		lastMessage := req.Messages[len(req.Messages)-1]
		userMessage := fmt.Sprintf("%v", lastMessage.Content)

		// 5. Find matching activity
		var matchedActivity *config.VoucherActivity
		for i := range voucherConfig.Activities {
			keyword := voucherConfig.Activities[i].Keyword
			matchedPattern := "\n" + keyword + "\n"
			if strings.Contains(userMessage, matchedPattern) || userMessage == keyword {
				matchedActivity = &voucherConfig.Activities[i]
				break
			}
		}
		if matchedActivity == nil {
			// No matching activity found, continue processing
			c.Next()
			return
		}
		logger.InfoC(ctx, "Matched activity", zap.String("keyword", matchedActivity.Keyword))

		// 6. Check activity time validity
		currentTime := time.Now()
		if currentTime.Before(matchedActivity.StartTime) {
			logger.WarnC(ctx, "Activity not started yet")
			c.Next()
			return
		}
		if currentTime.After(matchedActivity.EndTime) {
			logger.InfoC(ctx, "Activity has ended")
			helper.SendSSEResponseMessage(c, identity.ClientIDE, matchedActivity.ExpiredMessage, map[string]interface{}{
				"Config":      matchedActivity,
				"CurrentTime": currentTime,
			})
			c.Abort()
			return
		}

		// 7. Extract user unique identifier
		userID := identity.UserInfo.UUID
		if userID == "" {
			logger.WarnC(ctx, "Empty user UUID in voucher activity middleware")
			c.Next()
			return
		}

		// 8. Check if user has already redeemed
		usersKey := fmt.Sprintf("voucher:activity:%s:users", matchedActivity.Keyword)
		redeemedRecord, err := svcCtx.RedisClient.GetHashField(ctx, usersKey, userID)
		if err != nil {
			logger.WarnC(ctx, "Failed to get user redemption status from Redis", zap.Error(err))
		}
		if err == nil && redeemedRecord != "" {
			logger.InfoC(ctx, "User has already redeemed this activity", zap.String("user", identity.UserName))
			var record config.VoucherRedemptionRecord
			if err := json.Unmarshal([]byte(redeemedRecord), &record); err != nil {
				logger.WarnC(ctx, "Failed to unmarshal redemption record", zap.Error(err))
			}
			helper.SendSSEResponseMessage(c, identity.ClientIDE, matchedActivity.AlreadyRedeemedMessage, map[string]interface{}{
				"Config":      matchedActivity,
				"CurrentTime": currentTime,
				"VoucherCode": record.VoucherCode,
			})
			c.Abort()
			return
		}

		// 9. Check activity quota using HashLen
		userCount, err := svcCtx.RedisClient.HashLen(ctx, usersKey)
		if err != nil {
			logger.WarnC(ctx, "Failed to get user count from Redis", zap.Error(err))
			userCount = 0
		}

		logger.InfoC(ctx, "Got users from resdis", zap.Int64("userCount", userCount),
			zap.Int("TotalQuota", matchedActivity.TotalQuota))
		if userCount >= int64(matchedActivity.TotalQuota) {
			helper.SendSSEResponseMessage(c, identity.ClientIDE, matchedActivity.QuotaExhaustedMessage, map[string]interface{}{
				"Config":      matchedActivity,
				"CurrentTime": currentTime,
			})
			c.Abort()
			return
		}

		// 10. Generate voucher code
		voucherData := &service.VoucherData{
			GiverID:    fmt.Sprintf("《%s》活动", matchedActivity.Keyword),
			GiverName:  "admin",
			ReceiverID: identity.UserInfo.UUID,
			QuotaList: []service.VoucherQuotaItem{
				{
					Amount:     matchedActivity.CreditAmount,
					ExpiryDate: currentTime.AddDate(0, 0, matchedActivity.VoucherExpiryDays),
				},
			},
		}
		voucherCode, err := svcCtx.VoucherService.GenerateVoucher(voucherData)
		if err != nil {
			logger.ErrorC(ctx, "Failed to generate voucher code", zap.Error(err))
			c.Next()
			return
		}
		logger.InfoC(ctx, "voucher code gengrated", zap.String("voucherCode", voucherCode))

		// 11. Store redemption record in Redis
		redemptionRecord := config.VoucherRedemptionRecord{
			UserID:         userID,
			UserName:       identity.UserInfo.Name,
			VoucherCode:    voucherCode,
			RedemptionTime: currentTime,
		}
		recordJSON, err := json.Marshal(redemptionRecord)
		if err != nil {
			logger.ErrorC(ctx, "Failed to marshal redemption record", zap.Error(err))
			c.Next()
			return
		}

		// Extend expiration by 15 days to preserve activity data
		expiration := matchedActivity.EndTime.Sub(currentTime) + 15*24*time.Hour
		if err := svcCtx.RedisClient.SetHashField(ctx, usersKey, userID, string(recordJSON), expiration); err != nil {
			logger.ErrorC(ctx, "Failed to store redemption record", zap.Error(err))
		}
		logger.InfoC(ctx, "redemption record setted in redis")

		// 13-14. Prepare template data and render using Go template engine
		templateData := map[string]interface{}{
			"VoucherCode": voucherCode,
			"Config":      matchedActivity,
			"CurrentTime": currentTime,
		}

		helper.SendSSEResponseMessage(c, identity.ClientIDE, matchedActivity.SuccessTemplate, templateData)
		c.Abort()
	}
}
