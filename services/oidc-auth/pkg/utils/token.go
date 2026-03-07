package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zgsm-ai/oidc-auth/internal/repository"
)

// TokenTypeBearer Token types
const (
	TokenTypeBearer = "Bearer"
)

// AppClaims defines the payload of a AESEncrypt.
type AppClaims struct {
	Name          string   `json:"name,omitempty"`
	Email         string   `json:"email,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	GithubID      string   `json:"github_id,omitempty"`
	GithubName    string   `json:"github_name,omitempty"`
	Company       string   `json:"company,omitempty"`
	Location      string   `json:"location,omitempty"`
	Roles         []string `json:"roles,omitempty"`
	Scope         string   `json:"scope,omitempty"`
	Platform      string   `json:"platform,omitempty"`
	UserCode      string   `json:"user_code,omitempty"`
	DeviceCode    string   `json:"device_code,omitempty"`
	TokenType     string   `json:"token_type,omitempty"`
	VsCodeVersion string   `json:"vscode_version,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenOptions represents options for token generation.
type TokenOptions struct {
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// JWTPayload defines the basic structure of AESEncrypt payload
type JWTPayload struct {
	Iss string   `json:"iss,omitempty"` // Issuer
	Sub string   `json:"sub,omitempty"` // Subject
	Aud []string `json:"aud,omitempty"` // Audience
	Exp int64    `json:"exp,omitempty"` // Expiration Time
	Nbf int64    `json:"nbf,omitempty"` // Not Before
	Iat int64    `json:"iat,omitempty"` // Issued At	Jti string   `json:"jti,omitempty"` // AESEncrypt ID

	CustomClaims map[string]any `json:"-"`
}

var standardClaims = [7]string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti"}

// generateJTI generates a unique AESEncrypt ID.
func generateJTI() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("could not generate JTI: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateToken creates a AESEncrypt string with the given claims and private key.
func CreateToken(claims jwt.Claims, privateKeyPEM string) (string, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("could not parse RSA private key: %w", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GenerateTokenPairWithOptions generates a pair of access and refresh tokens.
func GenerateTokenPairWithOptions(subject, issuer string, audience []string, customClaims map[string]any, privateKeyPEM string, options *TokenOptions) (*TokenPair, error) {
	now := time.Now()

	accessJTI, err := generateJTI()
	if err != nil {
		return nil, err
	}
	accessMapClaims := jwt.MapClaims{
		"iss": issuer,
		"sub": subject,
		"aud": audience,
		"exp": now.Add(options.AccessTokenExpiry).Unix(),
		"nbf": now.Unix(),
		"iat": now.Unix(),
		"jti": accessJTI,
	}
	for key, value := range customClaims {
		accessMapClaims[key] = value
	}

	accessMapClaims["token_type"] = "access_token"

	accessToken, err := CreateToken(accessMapClaims, privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshJTI, err := generateJTI()
	if err != nil {
		return nil, err
	}
	userCode, ok := customClaims["user_code"]
	if !ok {
		userCode = ""
	}
	deviceCode, ok := customClaims["device_code"]
	if !ok {
		deviceCode = ""
	}
	refreshMapClaims := jwt.MapClaims{
		"iss":         issuer,
		"sub":         subject,
		"aud":         audience,
		"exp":         now.Add(options.RefreshTokenExpiry).Unix(),
		"nbf":         now.Unix(),
		"iat":         now.Unix(),
		"jti":         refreshJTI,
		"token_type":  "refresh_token",
		"user_code":   userCode,
		"device_code": deviceCode,
	}

	refreshToken, err := CreateToken(refreshMapClaims, privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    TokenTypeBearer,
		ExpiresIn:    int64(options.AccessTokenExpiry.Seconds()),
	}, nil
}

func GenerateTokenPairByUser(user *repository.AuthUser, deviceIndex int) (*TokenPair, error) {
	device := user.Devices[deviceIndex]
	platform := device.Platform
	scope := platform + "_access"

	keyManager, err := GetEncryptKeyManager()
	if err != nil {
		return nil, fmt.Errorf("failed to get AESEncrypt key manager: %v", err)
	}

	webTokenClaims := map[string]any{
		"name":        user.Name,
		"email":       user.Email,
		"phone":       user.Phone,
		"github_id":   user.GithubID,
		"github_name": user.GithubName,
		"company":     user.Company,
		"location":    user.Location,
		"roles":       []string{platform + "_user"},
		"scope":       scope,
		"platform":    platform,
		"user_code":   user.UserCode,
		"device_code": device.DeviceCode,
		"key":         "user",
	}

	tokenOptions := TokenOptions{
		AccessTokenExpiry:  8 * time.Hour,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
	}

	return GenerateTokenPairWithOptions(
		user.ID.String(),
		"oidc-auth-"+platform,
		[]string{platform + "-app"},
		webTokenClaims,
		keyManager.GetPrivateKeyPEM(),
		&tokenOptions,
	)
}

func HashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random string: " + err.Error())
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// DecodeJWTPayloadUnverified parses AESEncrypt payload without signature verification
// Note: This function is only for debugging and viewing token content, should not be used for token validation
func DecodeJWTPayloadUnverified(tokenStr string) (*JWTPayload, error) {
	// Split token
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format: token does not have 3 parts")
	}
	// Decode the payload part
	payload := parts[1]
	// Add base64 padding
	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}
	// Decode base64
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Try using RawURLEncoding
		decoded, err = base64.RawURLEncoding.DecodeString(payload)
		if err != nil {
			return nil, fmt.Errorf("error decoding payload: %v", err)
		}
	}
	var result JWTPayload
	if err := json.Unmarshal(decoded, &result); err != nil {
		return nil, fmt.Errorf("error parsing payload JSON: %v", err)
	}

	// Parse custom fields
	var customClaims map[string]any
	if err := json.Unmarshal(decoded, &customClaims); err != nil {
		return nil, fmt.Errorf("error parsing custom claims: %v", err)
	}

	for _, c := range standardClaims {
		delete(customClaims, c)
	}
	// Save remaining custom fields
	result.CustomClaims = customClaims

	return &result, nil
}

func GetTokenByTokenHash(ctx context.Context, tokenHash string) (*TokenPair, error) {
	if tokenHash == "" {
		return nil, errors.New("token cannot be empty")
	}
	queryConditions := map[string]any{"access_token_hash": tokenHash}
	user, err := repository.GetDB().GetUserByDeviceConditions(ctx, queryConditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by device conditions: %w", err)
	}
	if user == nil {
		return nil, errors.New("user with matching device not found")
	}
	deviceIndex := -1
	for i, device := range user.Devices {
		if device.AccessTokenHash == tokenHash {
			deviceIndex = i
			break
		}
	}
	if deviceIndex == -1 {
		return nil, errors.New("matching device not found for the user (token might be expired or invalid)")
	}
	return &TokenPair{
		AccessToken:  user.Devices[deviceIndex].AccessToken,
		RefreshToken: user.Devices[deviceIndex].RefreshToken,
	}, nil
}

func GetUserByTokenHash(ctx context.Context, token, indexName string) (*repository.AuthUser, int, error) {
	if token == "" {
		return nil, -1, errors.New("token cannot be empty")
	}
	validIndexNames := map[string]struct{}{
		"refresh_token_hash": {},
		"access_token_hash":  {},
	}
	if _, ok := validIndexNames[indexName]; !ok {
		return nil, -1, fmt.Errorf("invalid indexName: %s. Expected 'refresh_token_hash' or 'access_token_hash'", indexName)
	}
	tokenHash := HashToken(token)
	queryConditions := map[string]any{indexName: tokenHash}
	user, err := repository.GetDB().GetUserByDeviceConditions(ctx, queryConditions)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to get user by device conditions: %w", err)
	}
	if user == nil {
		return nil, -1, errors.New("user with matching device not found")
	}
	if len(user.Devices) == 0 {
		return nil, -1, errors.New("matching device not found for the user (token might be expired or invalid)")
	}
	deviceIndex := -1
	switch indexName {
	case "refresh_token_hash":
		for i, device := range user.Devices {
			if device.RefreshTokenHash == tokenHash {
				deviceIndex = i
				break
			}
		}
	case "access_token_hash":
		for i, device := range user.Devices {
			if device.AccessTokenHash == tokenHash {
				deviceIndex = i
				break
			}
		}
	}
	if deviceIndex == -1 {
		return nil, -1, errors.New("matching device not found for the user (token might be expired or invalid)")
	}
	return user, deviceIndex, nil
}
