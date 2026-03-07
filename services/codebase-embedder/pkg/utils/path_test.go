package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestToUnixPath(t *testing.T) {
	// 测试用例
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal", "a//b/c", "a/b/c"},
		{"with dot", "a/./b/c", "a/b/c"},
		{"with parent", "a/b/../c", "a/c"},
		{"mixed separators", "a\\b\\c", "a/b/c"}, // 自动转换为 /
		{"root absolute", "/a/b/c", "/a/b/c"},    // 转为相对路径
		{"current dir", ".", "."},
		{"parent dir", "..", ".."},
		{"complex", "../../a/./b//c/..", "../../a/b"},
	}

	for _, tc := range testCases {
		result := ToUnixPath(tc.input)
		assert.Equal(t, tc.expected, result, fmt.Sprintf("Test case: %s", tc.name))
	}
}

func TestIsChild(t *testing.T) {
	tests := []struct {
		parent string
		path   string
		want   bool
	}{
		// 基本直接子文件
		{"a", "a/b.txt", true},
		{"a/", "a/b.txt", true},  // 修复：处理末尾斜杠
		{"a", "a/b/c.txt", true}, // 修复：处理多级子目录

		// 边缘情况
		{"a", "a", false},            // 修复：相同路径返回 false
		{"a", "b.txt", false},        // 正确
		{"a/b/", "b/a/b.txt", false}, // 正确
		{"a/b", "a/b/c/d.txt", true}, // 修复：深层子目录

		// 路径规范化
		{"a/b", "a/b/c/../d.txt", true}, // 修复：处理 ".."
		{"/a", "/a/b", true},            // 处理绝对路径
		{"a", "a/./b", true},            // 处理 "."
	}

	for _, tt := range tests {
		t.Run(tt.parent+"_"+tt.path, func(t *testing.T) {
			if got := IsChild(tt.parent, tt.path); got != tt.want {
				t.Errorf("IsChild() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHidden(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{".gitignore", true},         // 隐藏文件
		{"./.test", true},            // 隐藏文件（带前缀）
		{"/home/user/.config", true}, // 隐藏目录
		{"README.md", false},         // 普通文件
		{"./dir/file.txt", false},    // 普通文件路径
		{"/../.cache/file", true},    // 隐藏文件（带上级目录）
		{"//dir//.hidden//", true},   // 隐藏目录（带多余斜杠）
		{"/", false},                 // 根目录
		{"..", false},                // 上级目录
		{"./", false},                // 当前目录
		{"...hidden", true},          // 非隐藏文件（.在中间）
	}

	for _, tc := range testCases {
		tc := tc // 捕获循环变量
		t.Run(tc.path, func(t *testing.T) {
			result := IsHiddenFile(tc.path)
			require.Equal(t, tc.expected, result,
				"expected %v for path %q, got %v",
				tc.expected, tc.path, result)
		})
	}
}
