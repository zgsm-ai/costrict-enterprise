package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// VoucherData represents the data structure in voucher code
type VoucherData struct {
	GiverID         string             `json:"giver_id"`
	GiverName       string             `json:"giver_name"`
	GiverPhone      string             `json:"giver_phone"`
	GiverGithub     string             `json:"giver_github"`
	GiverGithubStar string             `json:"giver_github_star"` // Comma-separated list of starred projects
	ReceiverID      string             `json:"receiver_id"`
	QuotaList       []VoucherQuotaItem `json:"quota_list"`
	Timestamp       int64              `json:"timestamp"`
}

// VoucherQuotaItem represents quota item in voucher
type VoucherQuotaItem struct {
	Amount     float64   `json:"amount"`
	ExpiryDate time.Time `json:"expiry_date"`
}

// VoucherService handles voucher code generation and validation
type VoucherService struct {
	signingKey []byte
}

// NewVoucherService creates a new voucher service
func NewVoucherService(signingKey string) *VoucherService {
	return &VoucherService{
		signingKey: []byte(signingKey),
	}
}

// GenerateVoucher generates a voucher code
func (s *VoucherService) GenerateVoucher(data *VoucherData) (string, error) {
	// Set timestamp
	data.Timestamp = time.Now().Unix()

	// Serialize to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal voucher data: %w", err)
	}

	// Generate HMAC signature
	signature := s.generateSignature(jsonData)

	// Combine JSON and signature with "|" separator
	combined := string(jsonData) + "|" + hex.EncodeToString(signature)

	// Base64URL encode
	voucherCode := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(combined))

	return voucherCode, nil
}

// generateSignature generates HMAC-SHA256 signature
func (s *VoucherService) generateSignature(data []byte) []byte {
	h := hmac.New(sha256.New, s.signingKey)
	h.Write(data)
	return h.Sum(nil)
}
