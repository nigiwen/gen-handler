package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// toUpperCamel 将下划线命名转换为大驼峰命名
func toUpperCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// readModuleFromGoMod 从当前目录向上查找 go.mod 文件并读取 module 路径
// 返回 module 路径和是否找到文件
func readModuleFromGoMod(startDir string) (string, bool) {
	// 从当前目录开始，向上查找 go.mod 文件
	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		
		// 检查文件是否存在
		if _, err := os.Stat(goModPath); err == nil {
			// 文件存在，尝试读取
			module, err := parseGoModFile(goModPath)
			if err == nil {
				return module, true
			}
		}
		
		// 获取父目录
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// 已经到达根目录，停止查找
			break
		}
		dir = parentDir
	}
	
	return "", false
}

// parseGoModFile 解析 go.mod 文件，提取 module 路径
func parseGoModFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 查找 module 声明行
		if strings.HasPrefix(line, "module ") {
			// 提取 module 路径（去掉 "module " 前缀）
			module := strings.TrimSpace(line[7:])
			// 去掉可能的注释
			if idx := strings.Index(module, "//"); idx >= 0 {
				module = strings.TrimSpace(module[:idx])
			}
			return module, nil
		}
	}
	
	if err := scanner.Err(); err != nil {
		return "", err
	}
	
	return "", fmt.Errorf("未找到 module 声明")
}

// generateProtoDirFromModule 从 module 路径生成 proto-dir
// 规则：./internal/proto + (module 去掉第一个 / 前面的部分)
// 例如：bsi/axis/devopsx -> ./internal/proto/axis/devopsx
func generateProtoDirFromModule(modulePath string) string {
	// 查找第一个 / 的位置
	idx := strings.Index(modulePath, "/")
	if idx < 0 {
		// 如果没有 /，直接使用整个 module 路径
		return filepath.Join("./internal/proto", modulePath)
	}
	
	// 取第一个 / 后面的所有部分
	suffix := modulePath[idx+1:]
	
	// 组合成 ./internal/proto/{suffix}
	return filepath.Join("./internal/proto", suffix)
}

// generateWireDirFromModule 从 module 路径生成 wire-dir
// 规则：./cmd + (module 最后一个 / 后面的部分)
// 例如：bsi/axis/devopsx -> ./cmd/devopsx
func generateWireDirFromModule(modulePath string) string {
	// 查找最后一个 / 的位置
	idx := strings.LastIndex(modulePath, "/")
	if idx < 0 {
		// 如果没有 /，直接使用整个 module 路径
		return filepath.Join("./cmd", modulePath)
	}
	
	// 取最后一个 / 后面的部分
	suffix := modulePath[idx+1:]
	
	// 组合成 ./cmd/{suffix}
	return filepath.Join("./cmd", suffix)
}
