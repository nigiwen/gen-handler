package main

import (
	"fmt"
	"strings"
)

// camelToSnake 将驼峰命名转换为下划线命名（全小写）
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// toLowerCamel 将首字母转为小写
func toLowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// parseInt 解析整数（支持简单格式）
func parseInt(s string) int {
	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}
