package utils

import "time"

func CurrentTime() *time.Time {
	now := time.Now()
	return &now
}
