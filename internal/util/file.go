package util

import (
	"go/format"
	"os"
)

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile 读取文件内容
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile 写入文件内容
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// FormatGoFile 格式化 Go 代码
func FormatGoFile(content string) (string, error) {
	formatted, err := format.Source([]byte(content))
	if err != nil {
		// 格式化失败，返回原始内容
		return content, err
	}
	return string(formatted), nil
}
