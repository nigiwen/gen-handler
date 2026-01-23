package main

import (
	"os"
	"path/filepath"
	"strings"
)

// findMissingHandlers 找出还没有生成 handler 文件的服务
func findMissingHandlers(services []ServiceInfo, outputDir string) []ServiceInfo {
	var missing []ServiceInfo

	for _, service := range services {
		if !handlerExists(service, outputDir) {
			missing = append(missing, service)
		}
	}

	return missing
}

// handlerExists 检查 handler 文件是否已存在
func handlerExists(service ServiceInfo, outputDir string) bool {
	// 检查文件是否已存在
	filePath := filepath.Join(outputDir, service.FileName)
	if _, err := os.Stat(filePath); err == nil {
		return true
	}

	// 检查所有可能的文件变体
	baseName := strings.TrimSuffix(service.FileName, ".go")
	variants := []string{
		strings.Title(baseName) + ".go",                      // TestCase.go
		strings.ToUpper(baseName[:1]) + baseName[1:] + ".go", // Test_case.go (首字母大写)
	}

	for _, variant := range variants {
		variantPath := filepath.Join(outputDir, variant)
		if _, err := os.Stat(variantPath); err == nil {
			return true
		}
	}

	// 检查目录中是否有相同 Handler 名称的文件
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// 读取文件内容，检查是否有相同的 Handler 类型定义
		content, err := os.ReadFile(filepath.Join(outputDir, file.Name()))
		if err == nil {
			if strings.Contains(string(content), "type "+service.HandlerName+" struct") {
				return true
			}
		}
	}

	return false
}
