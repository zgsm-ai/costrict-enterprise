package utils

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/zgsm-ai/oidc-auth/internal/constants"
	"github.com/zgsm-ai/oidc-auth/internal/repository"
)

// GenerateInviteCode generates an invite code
func GenerateInviteCode() (string, error) {
	chars := constants.InviteCodeChars
	length := constants.InviteCodeLength

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to get invite code: %w", err)
	}

	result := make([]byte, length)
	for i := range b {
		result[i] = chars[int(b[i])%len(chars)]
	}

	return string(result), nil
}

// ValidateInviteCode validates invite code and returns inviter information
func ValidateInviteCode(ctx context.Context, inviteCode string) (*repository.AuthUser, error) {
	if inviteCode == "" {
		return nil, fmt.Errorf("invite code cannot be empty")
	}

	// Find user by invite code
	user, err := repository.GetDB().GetUserByField(ctx, "invite_code", inviteCode)
	if err != nil {
		return nil, fmt.Errorf("failed to query invite code: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("invalid invite code")
	}

	return user, nil
}

// GenerateUniqueInviteCode generates a unique invite code (ensures no duplicates)
func GenerateUniqueInviteCode(ctx context.Context) (string, error) {
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		code, err := GenerateInviteCode()
		if err != nil {
			return "", err
		}

		// Check if invite code already exists
		existingUser, err := repository.GetDB().GetUserByField(ctx, "invite_code", code)
		if err != nil {
			return "", fmt.Errorf("failed to check invite code uniqueness: %w", err)
		}

		if existingUser == nil {
			// Invite code doesn't exist, can be used
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique invite code after maximum retries")
}
