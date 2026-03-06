package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// JWTClaims represents the decoded JWT token claims
type JWTClaims struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	// Add other JWT fields as needed
}

// DecodeJWT decodes a JWT token without verifying the signature
// @Summary Decode JWT token without verification
// @description
// - Splits the JWT token into its three parts
// - Decodes the payload (second part) from base64
// - Returns the claims as a JWTClaims struct
// @param token - The JWT token to decode
// @returns JWTClaims object containing user information or error if decoding fails
func DecodeJWT(token string) (*JWTClaims, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}

	// JWT tokens have three parts separated by dots
	// We only need to decode the second part (payload)
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode the payload part
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	// Parse the JSON payload
	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	return &claims, nil
}
