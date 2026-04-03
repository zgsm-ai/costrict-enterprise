package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// VoucherActivityQueryResponse represents the response for voucher activity query
type VoucherActivityQueryResponse struct {
	Keyword            string                           `json:"keyword"`
	StartTime          time.Time                        `json:"start_time"`
	EndTime            time.Time                        `json:"end_time"`
	TotalQuota         int                              `json:"total_quota"`
	TotalRedeemed      int                              `json:"total_redeemed"`
	RemainingQuota     int                              `json:"remaining_quota"`
	CreditAmount       float64                          `json:"credit_amount"`
	RedemptionRecords  []config.VoucherRedemptionRecord `json:"redemption_records"`
	TotalRedeemedUsers int                              `json:"total_redeemed_users"`
}

// VoucherActivitiesQueryResponse represents the response for multiple voucher activities
type VoucherActivitiesQueryResponse struct {
	Activities          []VoucherActivityQueryResponse `json:"activities"`
	TotalActivities     int                            `json:"total_activities"`
	TotalRedeemed       int                            `json:"total_redeemed"`
	TotalRemainingQuota int                            `json:"total_remaining_quota"`
}

// VoucherActivityQueryHandler handles voucher activity query requests
func VoucherActivityQueryHandler(svcCtx *bootstrap.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get keyword from query parameter (optional)
		keyword := c.Query("keyword")

		// Get activity configuration
		voucherConfig := svcCtx.Config.VoucherActivityConfig
		if voucherConfig == nil {
			logger.Warn("Voucher activity config is not available")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "Voucher activity is not configured",
					"type":    "config_error",
				},
			})
			return
		}

		// If keyword is not provided, return summary of all activities
		if keyword == "" {
			if len(voucherConfig.Activities) == 0 {
				c.JSON(http.StatusOK, VoucherActivitiesQueryResponse{
					Activities:          []VoucherActivityQueryResponse{},
					TotalActivities:     0,
					TotalRedeemed:       0,
					TotalRemainingQuota: 0,
				})
				return
			}

			responses := make([]VoucherActivityQueryResponse, 0, len(voucherConfig.Activities))
			totalRedeemed := 0
			totalRemainingQuota := 0

			for _, activity := range voucherConfig.Activities {
				response, err := getActivityQueryResponse(c, svcCtx, &activity)
				if err != nil {
					logger.Error("Failed to get activity query response",
						zap.String("keyword", activity.Keyword),
						zap.Error(err))
					continue
				}
				responses = append(responses, response)
				totalRedeemed += response.TotalRedeemed
				totalRemainingQuota += response.RemainingQuota
			}

			c.JSON(http.StatusOK, VoucherActivitiesQueryResponse{
				Activities:          responses,
				TotalActivities:     len(responses),
				TotalRedeemed:       totalRedeemed,
				TotalRemainingQuota: totalRemainingQuota,
			})
			return
		}

		// Find the specified activity
		var matchedActivity *config.VoucherActivity
		for i := range voucherConfig.Activities {
			if voucherConfig.Activities[i].Keyword == keyword {
				matchedActivity = &voucherConfig.Activities[i]
				break
			}
		}

		if matchedActivity == nil {
			logger.Warn("Voucher activity not found",
				zap.String("keyword", keyword))
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"message": fmt.Sprintf("Voucher activity not found with keyword: %s", keyword),
					"type":    "not_found",
				},
			})
			return
		}

		// Get query response for the specified activity
		response, err := getActivityQueryResponse(c, svcCtx, matchedActivity)
		if err != nil {
			logger.Error("Failed to get activity query response",
				zap.String("keyword", keyword),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": "Failed to retrieve activity information",
					"type":    "redis_error",
				},
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// getActivityQueryResponse retrieves query response for a specific activity
func getActivityQueryResponse(c *gin.Context, svcCtx *bootstrap.ServiceContext, activity *config.VoucherActivity) (VoucherActivityQueryResponse, error) {
	// Get user count from Redis using HashLen
	usersKey := fmt.Sprintf("voucher:activity:%s:users", activity.Keyword)
	userCount, err := svcCtx.RedisClient.HashLen(c.Request.Context(), usersKey)
	if err != nil {
		logger.Error("Failed to get user count from Redis",
			zap.String("key", usersKey),
			zap.Error(err))
		userCount = 0
	}

	// Calculate remaining quota
	totalRedeemed := int(userCount)
	remainingQuota := activity.TotalQuota - totalRedeemed
	if remainingQuota < 0 {
		remainingQuota = 0
	}

	// Read redemption records from Redis
	usersData, err := svcCtx.RedisClient.GetHash(c.Request.Context(), usersKey)
	if err != nil {
		logger.Error("Failed to get redemption records from Redis",
			zap.String("key", usersKey),
			zap.Error(err))
		return VoucherActivityQueryResponse{}, err
	}

	// Parse redemption records
	redemptionRecords := make([]config.VoucherRedemptionRecord, 0, len(usersData))
	for _, recordStr := range usersData {
		var record config.VoucherRedemptionRecord
		if err := json.Unmarshal([]byte(recordStr), &record); err != nil {
			logger.Warn("Failed to parse redemption record",
				zap.String("record", recordStr),
				zap.Error(err))
			continue
		}
		redemptionRecords = append(redemptionRecords, record)
	}

	// Build response
	response := VoucherActivityQueryResponse{
		Keyword:            activity.Keyword,
		StartTime:          activity.StartTime,
		EndTime:            activity.EndTime,
		TotalQuota:         activity.TotalQuota,
		TotalRedeemed:      totalRedeemed,
		RemainingQuota:     remainingQuota,
		CreditAmount:       activity.CreditAmount,
		RedemptionRecords:  redemptionRecords,
		TotalRedeemedUsers: len(redemptionRecords),
	}

	logger.Info("Voucher activity query successful",
		zap.String("keyword", activity.Keyword),
		zap.Int("total_redeemed", totalRedeemed),
		zap.Int("remaining_quota", remainingQuota),
		zap.Int("total_users", len(redemptionRecords)))

	return response, nil
}
