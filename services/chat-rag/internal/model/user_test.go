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
				Name:       "John Doe",
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
				Name:           "",
				EmployeeNumber: "1001",
			},
			expectError: false,
		},
		{
			name:        "Invalid token format",
			token:       "invalid.token.here",
			want:        &UserInfo{},
			expectError: true,
		},
		{
			name: "Missing universal_id",
			token: createTestToken(jwt.MapClaims{
				"name": "John Doe",
			}),
			want:        &UserInfo{UUID: "00000000-0000-0000-0000-000000000000", Name: "John Doe"},
			expectError: true,
		},
		{
			name: "Invalid universal_id format",
			token: createTestToken(jwt.MapClaims{
				"universal_id": "invalid-uuid",
			}),
			want:        &UserInfo{UUID: "00000000-0000-0000-0000-000000000000"},
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

func TestUserNameClaimPriority(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
		want   string
	}{
		{"name first", jwt.MapClaims{"name": "Alice", "display_name": "Display", "phone": "+8613800138000"}, "Alice"},
		{"display name fallback", jwt.MapClaims{"name": "", "display_name": "Display", "phone": "+8613800138000"}, "Display"},
		{"phone fallback", jwt.MapClaims{"name": "", "display_name": "", "phone": "+8613800138000"}, "13800138000"},
		{"missing claims", jwt.MapClaims{}, ""},
		{"non-string names", jwt.MapClaims{"name": 123, "display_name": false, "phone": "13800138000"}, "13800138000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.claims["universal_id"] = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
			tt.claims["properties"] = map[string]interface{}{
				"oauth_GitHub_username": "old-github",
				"oauth_Custom_username": "old-custom",
				"oauth_Custom_id":       "1001",
				"oauth_Custom_email":    "+8613900139000",
			}
			got := NewUserInfo(createTestToken(tt.claims))
			assert.Equal(t, tt.want, got.Name)
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
