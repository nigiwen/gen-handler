package main

import (
	"fmt"
	"go/format"
	"os"
	"strings"
)

// updateCoreProviderSet 更新 core/core.go 的 ProviderSet
func updateCoreProviderSet(service ServiceInfo, coreGoPath string) error {
	// 读取文件内容
	content, err := os.ReadFile(coreGoPath)
	if err != nil {
		return fmt.Errorf("读取 core.go 文件失败: %v", err)
	}

	// 检查是否已经存在
	fileContent := string(content)
	if strings.Contains(fileContent, "New"+service.ServiceName) {
		return nil // 已经存在，不需要更新
	}

	// 查找 ProviderSet 的位置
	providerSetStart := strings.Index(fileContent, "var ProviderSet = wire.NewSet(")
	if providerSetStart == -1 {
		return fmt.Errorf("未找到 ProviderSet")
	}

	// 查找 ProviderSet 的结束位置（找到对应的右括号）
	start := providerSetStart + len("var ProviderSet = wire.NewSet(")
	parenCount := 1
	end := start
	for i := start; i < len(fileContent) && parenCount > 0; i++ {
		if fileContent[i] == '(' {
			parenCount++
		} else if fileContent[i] == ')' {
			parenCount--
			if parenCount == 0 {
				end = i
				break
			}
		}
	}

	// 找到最后一个 Service 行的位置（最后一个包含 "New" 和 "Service" 的行）
	providerSetContent := fileContent[start:end]
	lines := strings.Split(providerSetContent, "\n")
	lastServiceLineIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "New") && strings.Contains(line, "Service") {
			lastServiceLineIdx = i
			break
		}
	}

	if lastServiceLineIdx == -1 {
		// 没找到，在最后添加
		insertContent := "\n\tNew" + service.ServiceName + ","
		fileContent = fileContent[:end] + insertContent + fileContent[end:]
	} else {
		// 找到最后一个 Service 行的结束位置
		lineStartPos := start
		for i := 0; i < lastServiceLineIdx; i++ {
			lineStartPos += len(lines[i]) + 1 // +1 for \n
		}
		lastLine := lines[lastServiceLineIdx]
		lastLineEnd := lineStartPos + len(lastLine)

		// 检查最后一行是否已经有逗号
		trimmedLine := strings.TrimSpace(lastLine)
		insertContent := "\n\tNew" + service.ServiceName + ","
		if !strings.HasSuffix(trimmedLine, ",") {
			// 需要在最后一行添加逗号
			fileContent = fileContent[:lastLineEnd] + "," + fileContent[lastLineEnd:]
			lastLineEnd++
		}

		// 在最后一行后插入新行
		fileContent = fileContent[:lastLineEnd] + insertContent + fileContent[lastLineEnd:]
	}

	// 格式化代码
	formatted, err := format.Source([]byte(fileContent))
	if err != nil {
		// 如果格式化失败，使用原始内容
		formatted = []byte(fileContent)
	}

	// 写回文件
	if err := os.WriteFile(coreGoPath, formatted, 0644); err != nil {
		return fmt.Errorf("写入 core.go 文件失败: %v", err)
	}

	return nil
}
