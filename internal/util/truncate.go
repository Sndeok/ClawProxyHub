// Package util — 通用工具函数。
package util

import "unicode/utf8"

// TruncStr 字节长度截断，保证不切断 UTF-8 多字节字符（中文等）。
// n 为字节上限；截断点落在多字节字符中间时回退到前一个字符边界。
func TruncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	// 回退到有效 rune 边界（避免切半个中文字符产生无效 UTF-8）
	for n > 0 && n < len(s) && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}