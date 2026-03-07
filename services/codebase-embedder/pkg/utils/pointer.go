package utils

import "time"

// BoolPtr 返回布尔值的指针
func BoolPtr(v bool) *bool {
	return &v
}

// StringPtr 返回字符串的指针
func StringPtr(v string) *string {
	return &v
}

// IntPtr 返回 int 的指针
func IntPtr(v int) *int {
	return &v
}

// Int8Ptr 返回 int8 的指针
func Int8Ptr(v int8) *int8 {
	return &v
}

// Int16Ptr 返回 int16 的指针
func Int16Ptr(v int16) *int16 {
	return &v
}

// Int32Ptr 返回 int32 的指针
func Int32Ptr(v int32) *int32 {
	return &v
}

// Int64Ptr 返回 int64 的指针
func Int64Ptr(v int64) *int64 {
	return &v
}

// UintPtr 返回 uint 的指针
func UintPtr(v uint) *uint {
	return &v
}

// Uint8Ptr 返回 uint8 的指针
func Uint8Ptr(v uint8) *uint8 {
	return &v
}

// Uint16Ptr 返回 uint16 的指针
func Uint16Ptr(v uint16) *uint16 {
	return &v
}

// Uint32Ptr 返回 uint32 的指针
func Uint32Ptr(v uint32) *uint32 {
	return &v
}

// Uint64Ptr 返回 uint64 的指针
func Uint64Ptr(v uint64) *uint64 {
	return &v
}

// Float32Ptr 返回 float32 的指针
func Float32Ptr(v float32) *float32 {
	return &v
}

// Float64Ptr 返回 float64 的指针
func Float64Ptr(v float64) *float64 {
	return &v
}

// BytePtr 返回 byte 的指针
func BytePtr(v byte) *byte {
	return &v
}

// RunePtr 返回 rune 的指针
func RunePtr(v rune) *rune {
	return &v
}

// TimePtr 返回 time.Time 的指针
func TimePtr(v time.Time) *time.Time {
	return &v
}

// DurationPtr 返回 time.Duration 的指针
func DurationPtr(v time.Duration) *time.Duration {
	return &v
}
