package service

import "strings"

// normalizeSpaces 合并连续空白并去掉首尾空白（用于标题等展示字段）。
func normalizeSpaces(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
