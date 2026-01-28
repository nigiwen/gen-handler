package util

import "strings"

// CamelToSnake 将驼峰命名转换为下划线命名（全小写）
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ToLowerCamel 将首字母转为小写
func ToLowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// ToUpperCamel 将下划线命名转换为大驼峰命名
func ToUpperCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// TrimSrv 去掉 Srv 后缀（用于模板函数）
func TrimSrv(s string) string {
	return strings.TrimSuffix(s, "Srv")
}

// FirstChar 返回字符串的首字母小写（用于模板函数）
func FirstChar(s string) string {
	if len(s) == 0 {
		return "x"
	}
	return strings.ToLower(s[:1])
}
