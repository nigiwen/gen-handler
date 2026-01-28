package util

import (
	"fmt"
	"strings"
)

// FindMatchingBrace 查找匹配的右圆括号
// content: 源代码字符串
// start: 开始位置（第一个左圆括号的位置）
// 返回匹配的右圆括号位置，如果未找到返回 -1
func FindMatchingBrace(content string, start int) int {
	parenCount := 1
	for i := start + 1; i < len(content); i++ {
		if content[i] == '(' {
			parenCount++
		} else if content[i] == ')' {
			parenCount--
			if parenCount == 0 {
				return i
			}
		}
	}
	return -1
}

// FindMatchingCurlyBrace 查找匹配的右花括号
// content: 源代码字符串
// start: 开始位置（第一个左花括号 { 的位置）
// 返回匹配的右花括号位置，如果未找到返回 -1
func FindMatchingCurlyBrace(content string, start int) int {
	braceCount := 1
	for i := start + 1; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return i
			}
		}
	}
	return -1
}

// FindProviderSet 查找 ProviderSet 的起始和结束位置
// content: 文件内容
// 返回 ProviderSet 内容的起始位置、结束位置（右括号位置）
func FindProviderSet(content string) (start, end int, err error) {
	// 查找 ProviderSet 的位置
	providerSetIdx := strings.Index(content, "var ProviderSet = wire.NewSet(")
	if providerSetIdx == -1 {
		// 尝试没有空格的版本
		providerSetIdx = strings.Index(content, "var ProviderSet=wire.NewSet(")
		if providerSetIdx == -1 {
			return 0, 0, fmt.Errorf("未找到 ProviderSet")
		}
	}
	
	// 找到 wire.NewSet( 后的位置
	openParenIdx := strings.Index(content[providerSetIdx:], "wire.NewSet(")
	if openParenIdx == -1 {
		return 0, 0, fmt.Errorf("未找到 wire.NewSet")
	}
	
	// 计算 ( 的绝对位置
	openParenIdx = providerSetIdx + openParenIdx + len("wire.NewSet(")
	
	// 查找匹配的右括号
	closeParenIdx := FindMatchingBrace(content, openParenIdx-1)
	if closeParenIdx == -1 {
		return 0, 0, fmt.Errorf("未找到 ProviderSet 的闭合括号")
	}
	
	return openParenIdx, closeParenIdx, nil
}

// InsertIntoProviderSet 在 ProviderSet 中插入新项
// content: 文件内容
// newItem: 要插入的新项（如 "NewTestCaseHandler"）
// itemType: 项的类型（用于查找插入位置，如 "Handler", "Service", "Repo"）
// 返回更新后的文件内容
func InsertIntoProviderSet(content, newItem, itemType string) (string, error) {
	// 检查是否已存在
	if strings.Contains(content, newItem) {
		return content, nil
	}
	
	// 查找 ProviderSet
	start, end, err := FindProviderSet(content)
	if err != nil {
		return "", err
	}
	
	providerSetContent := content[start:end]
	
	// 找到同类型的最后一行
	lines := strings.Split(providerSetContent, "\n")
	lastItemLineIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "New") && strings.Contains(line, itemType) {
			lastItemLineIdx = i
			break
		}
	}
	
	if lastItemLineIdx == -1 {
		// 没找到同类型的，在最后添加
		insertContent := "\n\t" + newItem + ","
		return content[:end] + insertContent + content[end:], nil
	}
	
	// 找到最后一行的结束位置
	lineStartPos := start
	for i := 0; i < lastItemLineIdx; i++ {
		lineStartPos += len(lines[i]) + 1 // +1 for \n
	}
	lastLine := lines[lastItemLineIdx]
	lastLineEnd := lineStartPos + len(lastLine)
	
	// 检查最后一行是否有逗号
	trimmedLine := strings.TrimSpace(lastLine)
	insertContent := "\n\t" + newItem + ","
	if !strings.HasSuffix(trimmedLine, ",") {
		// 需要在最后一行添加逗号
		content = content[:lastLineEnd] + "," + content[lastLineEnd:]
		lastLineEnd++
	}
	
	// 在最后一行后插入新行
	return content[:lastLineEnd] + insertContent + content[lastLineEnd:], nil
}
