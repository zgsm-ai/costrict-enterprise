package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/zgsm-ai/oidc-auth/internal/config"
	"github.com/zgsm-ai/oidc-auth/internal/constants"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

type QuotaMergeRequest struct {
	MainUserID  string `json:"main_user_id" validate:"required,uuid"`  // The user who will receive the merged quota
	OtherUserID string `json:"other_user_id" validate:"required,uuid"` // The user who willbe merged into the main user
}

// QuotaMergeData represents the data structure for quota merge response
type QuotaMergeData struct {
	MainUserID  string `json:"main_user_id"`
	OtherUserID string `json:"other_user_id"`
	Amount      int64  `json:"amount"`
	Operation   string `json:"operation"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// QuotaMergeResponse represents the complete quota merge response
type QuotaMergeResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Success bool           `json:"success"`
	Data    QuotaMergeData `json:"data"`
}

var (
	quotaConfig *config.QuotaConfig
	quotaOnce   sync.Once
)

func InitQuotaService(cfg *config.QuotaConfig) {
	quotaOnce.Do(func() {
		quotaConfig = cfg
		if quotaConfig != nil {
			log.Info(nil, "Quota service initialized successfully")
		} else {
			log.Info(nil, "Quota service is not configured")
		}
	})
}

func MergeUserQuota(MainUserID, OtherUserID, userToken string) error {
	if quotaConfig == nil {
		log.Info(nil, "Quota service is not configured, skipping quota merge")
		return nil
	}

	if quotaConfig.BaseURL == "" {
		log.Warn(nil, "Quota service base URL is not configured, skipping quota merge")
		return nil
	}

	url := quotaConfig.BaseURL + constants.QuotaMergeURI
	log.Info(nil, "start merged quota from other user %s to main user %s", OtherUserID, MainUserID)

	request := QuotaMergeRequest{
		MainUserID:  MainUserID,
		OtherUserID: OtherUserID,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal quota merge request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create quota merge request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if userToken != "" {
		req.Header.Set("Authorization", "Bearer "+userToken)
	}

	resp, err := quotaConfig.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call quota merge API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Error(nil, "quota merge API returned status: %s, body: %s", resp.Status, string(bodyBytes))
		return fmt.Errorf("quota merge API returned status: %s, message: %s", resp.Status, string(bodyBytes))
	}

	var response QuotaMergeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode quota merge response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("quota merge failed: %s", response.Message)
	}

	log.Info(nil, "Successfully merged quota from other user %s to main user %s", OtherUserID, MainUserID)
	return nil
}
