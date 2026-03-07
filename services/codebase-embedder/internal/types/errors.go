package types

import "errors"

var (
	// ErrRateLimitReached 限流达到上限错误
	ErrRateLimitReached = errors.New("The system is busy. Please try again later (maximum number of concurrent tasks reached).")
)
