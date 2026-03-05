package model

import (
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestNewUserInfo(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		want        *UserInfo
		expectError bool
	}{
		{
			name: "Valid token with basic info",
			token: createTestToken(jwt.MapClaims{
				"name":         "John Doe",
				"phone":        "13800138000",
				"universal_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				"email":        "john@example.com",
			}),
			want: &UserInfo{
				UUID:  "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				Phone: "13800138000",
				Email: "john@example.com",
				Name:  "John Doe",
			},
			expectError: false,
		},
		{
			name: "Valid token with GitHub info",
			token: createTestToken(jwt.MapClaims{
				"name":         "John Doe",
				"phone":        "+8613800138000",
				"universal_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				"properties": map[string]interface{}{
					"oauth_GitHub_id":       "12345",
					"oauth_GitHub_username": "johndoe",
				},
			}),
			want: &UserInfo{
				UUID:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				Phone:      "13800138000",
				Name:       "johndoe",
				GithubID:   "12345",
				GithubName: "johndoe",
			},
			expectError: false,
		},
		{
			name: "Valid token with custom login",
			token: createTestToken(jwt.MapClaims{
				"universal_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				"properties": map[string]interface{}{
					"oauth_Custom_username": "john",
					"oauth_Custom_id":       "1001",
					"oauth_Custom_email":    "+8613800138001",
				},
			}),
			want: &UserInfo{
				UUID:           "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				Phone:          "13800138001",
				Name:           "john1001",
				EmployeeNumber: "1001",
			},
			expectError: false,
		},
		{
			name:        "Invalid token format",
			token:       "invalid.token.here",
			want:        nil,
			expectError: true,
		},
		{
			name: "Missing universal_id",
			token: createTestToken(jwt.MapClaims{
				"name": "John Doe",
			}),
			want:        nil,
			expectError: true,
		},
		{
			name: "Invalid universal_id format",
			token: createTestToken(jwt.MapClaims{
				"universal_id": "invalid-uuid",
			}),
			want:        nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserInfo(tt.token)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractLoginFromToken(t *testing.T) {
	tests := []struct {
		name     string
		user     *UserInfo
		expected string
	}{
		{
			name: "GitHub login",
			user: &UserInfo{
				GithubName: "johndoe",
			},
			expected: "github",
		},
		{
			name: "Sangfor login",
			user: &UserInfo{
				EmployeeNumber: "1001",
			},
			expected: "sangfor",
		},
		{
			name: "Phone login",
			user: &UserInfo{
				Phone: "13800138000",
			},
			expected: "phone",
		},
		{
			name:     "Unknown login",
			user:     &UserInfo{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.ExtractLoginFromToken()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Phone with +86",
			input:    "+8613800138000",
			expected: "13800138000",
		},
		{
			name:     "Phone without +86",
			input:    "13800138000",
			expected: "13800138000",
		},
		{
			name:     "Empty phone",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePhone(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create test JWT tokens
func createTestToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}
