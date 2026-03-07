package utils

func Values[T any](m map[string]T) []T {
	result := make([]T, 0, len(m)) // 预分配切片，避免多次扩容
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
