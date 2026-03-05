package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	IdentityContextKey ContextKey = "identity"
)

type Identity struct {
	TaskID        string    `json:"task_id"`
	RequestID     string    `json:"request_id"`
	ClientID      string    `json:"client_id"`
	ClientIDE     string    `json:"client_ide"`
	ClientVersion string    `json:"client_version"`
	ClientOS      string    `json:"client_os"`
	UserName      string    `json:"user_name"`
	ProjectPath   string    `json:"project_path"`
	AuthToken     string    `json:"auth_token"`
	LoginFrom     string    `json:"login_from"`
	Caller        string    `json:"caller"` // ide, code-review, ...
	Sender        string    `json:"sender"` // user, system, ...
	Language      string    `json:"language"`
	UserInfo      *UserInfo `json:"user_info"`
}

// UserInfo defines the user information structure
type UserInfo struct {
	UUID           string          `json:"uuid"`
	Phone          string          `json:"phone"`
	GithubID       string          `json:"github_id"`
	Email          string          `json:"email"`
	Name           string          `json:"name"`
	GithubName     string          `json:"github_name"`
	EmployeeNumber string          `json:"employee_number"`
	Department     *DepartmentInfo `json:"department"`
	Vip            int             `json:"vip"`
	VipExpire      *time.Time      `json:"vip_expire"`
}

// DepartmentInfo department information structure
type DepartmentInfo struct {
	Level1Dept string `json:"dept_1"`
	Level2Dept string `json:"dept_2"`
	Level3Dept string `json:"dept_3"`
	Level4Dept string `json:"dept_4"`
}

// JWTClaims defines the JWT claims structure
type JWTClaims struct {
	Phone       string                 `json:"phone"`
	UniversalID string                 `json:"universal_id"`
	Email       string                 `json:"email,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Vip         int                    `json:"vip,omitempty"`
	VipExpire   *time.Time             `json:"vip_expire,omitempty"`
}

// CustomProperties defines the custom properties structure
type CustomProperties struct {
	GithubID       string `json:"oauth_GitHub_id,omitempty"`
	GithubName     string `json:"oauth_GitHub_username,omitempty"`
	CustomName     string `json:"oauth_Custom_username,omitempty"`
	EmployeeNumber string `json:"oauth_Custom_id,omitempty"`
	CustomPhone    string `json:"oauth_Custom_email,omitempty"`
	Vip            int       `json:"vip,omitempty"`
	VipExpire      *time.Time `json:"vip_expire,omitempty"`
}

// NewUserInfo creates user info from JWT token
func NewUserInfo(jwtToken string) *UserInfo {
	claims, err := parseJWT(jwtToken)
	if err != nil {
		logger.Error("Failed to parse JWT:", zap.Error(err))
		return &UserInfo{}
	}

	userInfo, err := extractUserInfo(claims)
	if err != nil {
		logger.Error("Failed to extract user info:", zap.Error(err))
		return &UserInfo{}
	}

	return userInfo
}

// parseJWT parses JWT token
func parseJWT(tokenString string) (*JWTClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("invalid JWT format: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims structure")
	}

	return mapClaimsToJWTClaims(claims)
}

// mapClaimsToJWTClaims maps jwt.MapClaims to JWTClaims structure
func mapClaimsToJWTClaims(claims jwt.MapClaims) (*JWTClaims, error) {
	var jwtClaims JWTClaims

	if phone, ok := claims["phone"].(string); ok {
		jwtClaims.Phone = phone
	}
	if universalID, ok := claims["universal_id"].(string); ok {
		jwtClaims.UniversalID = universalID
	}
	if email, ok := claims["email"].(string); ok {
		jwtClaims.Email = email
	}
	if properties, ok := claims["properties"].(map[string]interface{}); ok {
		jwtClaims.Properties = properties
	}

	if vipFloat, ok := claims["vip"].(float64); ok {
		jwtClaims.Vip = int(vipFloat)
	}

	if vipExpire, ok := claims["vip_expire"].(string); ok {
		if vipExpireTime, err := time.Parse(time.RFC3339, vipExpire); err == nil {
			jwtClaims.VipExpire = &vipExpireTime
		} else {
			logger.Warn("Failed to parse vip_expire time from jwt claims",
				zap.String("vip_expire", vipExpire),
				zap.Error(err))
		}
	}

	return &jwtClaims, nil
}

// extractUserInfo extracts user info from JWT claims
func extractUserInfo(claims *JWTClaims) (*UserInfo, error) {
	id, err := uuid.Parse(claims.UniversalID)
	if err != nil {
		logger.Warn("Failed to parse universal_id:", zap.Error(err))
	}

	customProps := parseCustomProperties(claims.Properties)
	user := buildUserInfo(claims, customProps, id.String())

	return user, nil
}

// parseCustomProperties parses custom properties
func parseCustomProperties(properties map[string]interface{}) CustomProperties {
	var props CustomProperties

	if properties == nil {
		return props
	}

	if id, ok := properties["oauth_GitHub_id"].(string); ok {
		props.GithubID = id
	}
	if name, ok := properties["oauth_GitHub_username"].(string); ok {
		props.GithubName = name
	}
	if customName, ok := properties["oauth_Custom_username"].(string); ok {
		props.CustomName = customName
	}
	if empNum, ok := properties["oauth_Custom_id"].(string); ok {
		props.EmployeeNumber = empNum
	}
	if phone, ok := properties["oauth_Custom_email"].(string); ok {
		props.CustomPhone = phone
	}

	return props
}

// buildUserInfo constructs user info from claims and properties
func buildUserInfo(claims *JWTClaims, props CustomProperties, id string) *UserInfo {
	user := &UserInfo{
		UUID:           id,
		Phone:          normalizePhone(claims.Phone),
		GithubID:       props.GithubID,
		Email:          claims.Email,
		GithubName:     props.GithubName,
		EmployeeNumber: props.EmployeeNumber,
		Vip:            claims.Vip,
		VipExpire:      claims.VipExpire,
	}

	applyCustomProperties(user, props)
	determineUserName(user, props)

	return user
}

// applyCustomProperties applies custom properties to user info
func applyCustomProperties(user *UserInfo, props CustomProperties) {
	if props.CustomPhone != "" {
		user.Phone = normalizePhone(props.CustomPhone)
	}
}

// determineUserName determines the final user name
func determineUserName(user *UserInfo, props CustomProperties) {
	switch {
	case props.GithubName != "":
		user.Name = props.GithubName
	case props.CustomName != "":
		user.Name = props.CustomName + props.EmployeeNumber
	case user.Name == "":
		user.Name = user.Phone
	}
}

// normalizePhone normalizes phone number format
func normalizePhone(phone string) string {
	return strings.ReplaceAll(phone, "+86", "")
}

// ExtractLoginFromToken parses JWT token to extract login type.
func (u *UserInfo) ExtractLoginFromToken() string {
	if u.GithubName != "" {
		return "github"
	}

	if u.EmployeeNumber != "" {
		return "sangfor"
	}

	if u.Phone != "" {
		return "phone"
	}

	return "unknown"
}

// GetIdentityFromContext retrieves identity from context
func GetIdentityFromContext(ctx context.Context) (*Identity, bool) {
	identity, ok := ctx.Value(IdentityContextKey).(*Identity)
	return identity, ok
}
