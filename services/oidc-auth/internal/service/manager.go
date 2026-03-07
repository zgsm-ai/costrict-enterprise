package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zgsm-ai/oidc-auth/internal/constants"
	"github.com/zgsm-ai/oidc-auth/internal/providers"
)

type MergeRequestPayload struct {
	ReservedUserToken string `json:"reserved_user_token"`
	DeletedUserToken  string `json:"deleted_user_token"`
}

type AuthMethod struct {
	AuthType  string `json:"auth_type"`
	AuthValue string `json:"auth_value"`
}

type MergeResponse struct {
	Status            string       `json:"status"`
	Msg               string       `json:"msg"`
	UniversalID       string       `json:"universal_id"`
	DeletedUserID     string       `json:"deleted_user_id"`
	MergedAuthMethods []AuthMethod `json:"merged_auth_methods"`
}

func MergeByCasdoor(provider providers.OAuthProvider, reservedUserToken, deletedUserToken string, httpClient *http.Client) (*MergeResponse, error) {
	url := provider.GetEndpoint(true) + constants.CasdoorMergeURI
	payload := MergeRequestPayload{
		ReservedUserToken: reservedUserToken,
		DeletedUserToken:  deletedUserToken,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+reservedUserToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API call failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	var mergeResponse MergeResponse
	if err := json.NewDecoder(resp.Body).Decode(&mergeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &mergeResponse, nil
}
