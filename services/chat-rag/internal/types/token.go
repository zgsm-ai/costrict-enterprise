package types

import "math"

// TokenStats represents detailed token statistics
type TokenStats struct {
	SystemTokens int `json:"system_tokens"`
	UserTokens   int `json:"user_tokens"`
	All          int `json:"all"`
}

// TokenRatio represents the ratio between processed and original tokens
type TokenRatio struct {
	SystemRatio float64 `json:"system_ratio"`
	UserRatio   float64 `json:"user_ratio"`
	AllRatio    float64 `json:"all_ratio"`
}

// TokenMetrics represents complete token statistics including original, processed and ratios
type TokenMetrics struct {
	Original  TokenStats `json:"original"`
	Processed TokenStats `json:"processed"`
	Ratios    TokenRatio `json:"ratios"`
}

// CalculateRatios calculates the token ratios between processed and original tokens
func (tm *TokenMetrics) CalculateRatios() {
	if tm.Original.All > 0 {
		ratio := float64(tm.Processed.All) / float64(tm.Original.All)
		tm.Ratios.AllRatio = math.Round(ratio*100) / 100
	}
	if tm.Original.SystemTokens > 0 {
		ratio := float64(tm.Processed.SystemTokens) / float64(tm.Original.SystemTokens)
		tm.Ratios.SystemRatio = math.Round(ratio*100) / 100
	}
	if tm.Original.UserTokens > 0 {
		ratio := float64(tm.Processed.UserTokens) / float64(tm.Original.UserTokens)
		tm.Ratios.UserRatio = math.Round(ratio*100) / 100
	}
}
