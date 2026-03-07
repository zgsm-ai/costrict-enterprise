package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/zgsm-ai/oidc-auth/internal/constants"
	"github.com/zgsm-ai/oidc-auth/internal/repository"
	"github.com/zgsm-ai/oidc-auth/pkg/utils"
)

type CasdoorFactory struct{}

type casdoorConfig struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	InternalURL  string
	Scopes       []string
}

type CasdoorProvider struct {
	config     *casdoorConfig
	httpClient *http.Client
}

func NewCasdoorFactory() *CasdoorFactory {
	return &CasdoorFactory{}
}

func (s *CasdoorFactory) GetName() string {
	return "casdoor"
}

func (f *CasdoorFactory) CreateProvider(config *ProviderConfig) OAuthProvider {
	return NewCasdoorProvider(config)
}

func NewCasdoorProvider(config *ProviderConfig) *CasdoorProvider {
	return &CasdoorProvider{
		httpClient: config.Client,
		config: &casdoorConfig{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			BaseURL:      config.BaseURL,
			InternalURL:  config.InternalURL,
		},
	}
}

func (s *CasdoorProvider) GetName() string {
	return "casdoor"
}

func (s *CasdoorProvider) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("client_secret", s.config.ClientSecret)
	data.Set("client_id", s.config.ClientID)

	getTokenURL := s.GetEndpoint(true) + constants.CasdoorTokenURI
	req, err := http.NewRequest(http.MethodPost, getTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpClient := s.httpClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get token, status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	return &tokenResp, nil
}

func (s *CasdoorProvider) Update(ctx context.Context, data *repository.AuthUser) error {
	if len(data.Devices) != 1 {
		return fmt.Errorf("invalid input: data must contain exactly one device")
	}
	var existingUser *repository.AuthUser
	var err error

	// data.ID must not be nil
	existingUser, err = repository.GetDB().GetUserByField(ctx, "id", data.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if existingUser == nil {
		if data.UserCode == "" {
			data.UserCode, err = utils.GenerateRandomString(16)
			if err != nil {
				return err
			}
		}
		if data.Devices[0].DeviceCode == "" {
			data.Devices[0].DeviceCode, err = utils.GenerateRandomString(16)
			if err != nil {
				return err
			}
		}
		// 使用 id 作为幂等键进行 Upsert，确保手机号变更不会导致重复创建
		err = repository.GetDB().Upsert(ctx, data, "id", data.ID)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		return nil
	}
	existingUser.GithubName = data.GithubName
	existingUser.Name = data.Name
	existingUser.Email = data.Email
	existingUser.Location = data.Location
	existingUser.Company = data.Company
	existingUser.Phone = data.Phone
	existingUser.Vip = data.Vip
	existingUser.EmployeeNumber = data.EmployeeNumber
	existingUser.ID = data.ID

	newDevice := data.Devices[0]
	newDevice.UpdatedAt = time.Now()

	if existingUser.ID == uuid.Nil {
		existingUser.ID = uuid.New()
		existingUser.CreatedAt = time.Now()
		existingUser.UpdatedAt = time.Now()
	}

	if newDevice.ID == uuid.Nil {
		newDevice.ID = uuid.New()
		newDevice.CreatedAt = time.Now()
		newDevice.UpdatedAt = time.Now()
	}

	deviceFound := false
	for i, device := range existingUser.Devices {
		if device.MachineCode == newDevice.MachineCode && device.VSCodeVersion == newDevice.VSCodeVersion {
			newDevice.CreatedAt = device.CreatedAt
			if newDevice.DeviceCode == "" {
				newDevice.DeviceCode = existingUser.Devices[i].DeviceCode
			}
			if existingUser.Devices[i].ID.String() != "" {
				newDevice.ID = existingUser.Devices[i].ID
			}
			existingUser.Devices[i] = newDevice
			deviceFound = true
			break
		}
	}
	if !deviceFound {
		newDevice.DeviceCode, err = utils.GenerateRandomString(16)
		if err != nil {
			return err
		}
		existingUser.Devices = append(existingUser.Devices, newDevice)
	}

	// 统一通过 id 执行 Upsert，更新手机号等资料到同一条记录
	err = repository.GetDB().Upsert(ctx, *existingUser, "id", existingUser.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	return nil
}

func (s *CasdoorProvider) GetUserInfo(ctx context.Context, accessToken string) (*repository.AuthUser, error) {
	payload, err := utils.DecodeJWTPayloadUnverified(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode AESEncrypt payload: %w", err)
	}
	/*
		Meeting the following conditions indicates that custom login is used
		casdoor -> /providers/admin/customAuth -> User mapping
			{
			   "id": "employee_number",
			   "username": "username",
			   "displayName": "username",
			   "email": "phone_number",
			   "avatarUrl": ""
			  }
	*/
	name, _ := payload.CustomClaims["name"].(string)
	phone, _ := payload.CustomClaims["phone"].(string)
	universalID, _ := payload.CustomClaims["universal_id"].(string)

	var githubID, githubName, customName, employeeNum, email string

	id, err := uuid.Parse(universalID)
	if err != nil {
		return nil, err
	}

	customPayload, ok := payload.CustomClaims["properties"].(map[string]any)
	if ok && len(customPayload) != 0 {
		githubID, _ = customPayload["oauth_GitHub_id"].(string)
		githubName, _ = customPayload["oauth_GitHub_username"].(string)
		customName, _ = customPayload["oauth_Custom_username"].(string)
		employeeNum, _ = customPayload["oauth_Custom_id"].(string)
		customPhone, _ := customPayload["oauth_Custom_email"].(string)
		if githubID != "" {
			email, _ = payload.CustomClaims["email"].(string)
		}
		if customPhone != "" {
			phone = customPhone
		}
		if githubName != "" {
			name = githubName
		} else if customName != "" {
			name = customName + employeeNum
		}
	} else {
		name = phone
	}

	if strings.Contains(phone, "+86") {
		phone = strings.ReplaceAll(phone, "+86", "")
	}
	user := &repository.AuthUser{
		ID:             id,
		Phone:          phone,
		GithubID:       githubID,
		Email:          email,
		Name:           name,
		GithubName:     githubName,
		EmployeeNumber: employeeNum,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return user, nil
}

func (s *CasdoorProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("client_secret", s.config.ClientSecret)
	data.Set("client_id", s.config.ClientID)
	refreshTokenURL := s.GetEndpoint(true) + constants.CasdoorRefreshTokenURI
	req, err := http.NewRequest(http.MethodPost, refreshTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get token, status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	return &tokenResp, nil
}

func (s *CasdoorProvider) GetAuthURL(state, redirectURL string) string {
	return s.config.BaseURL + constants.CasdoorAuthURI + "?client_id=" +
		s.config.ClientID + "&state=" + state + "&redirect_uri=" + redirectURL + "&response_type=code"
}

func (s *CasdoorProvider) GetEndpoint(isInternal bool) string {
	if isInternal {
		return s.config.InternalURL
	}
	return s.config.BaseURL
}

func (s *CasdoorProvider) ValidateToken(ctx context.Context, accessToken string) error {
	return fmt.Errorf("oauth oauth does not support refresh token")
}
