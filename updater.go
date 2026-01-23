package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// updateGrpcFile 更新 grpc.go 文件，添加新的 Handler 到 ProviderSet 和 NewGRPCServer
func updateGrpcFile(service ServiceInfo, grpcFilePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(grpcFilePath)
	if err != nil {
		return fmt.Errorf("读取 grpc.go 文件失败: %v", err)
	}

	// 解析文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, grpcFilePath, content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析 grpc.go 文件失败: %v", err)
	}

	// 检查是否已经存在
	if handlerExistsInGrpc(node, service) {
		return nil // 已经存在，不需要更新
	}

	// 转换为字符串进行修改（更简单可靠）
	fileContent := string(content)

	// 1. 在 ProviderSet 中添加 NewXXXHandler
	fileContent = addToProviderSet(fileContent, service)

	// 2. 在 NewGRPCServer 参数中添加 handler
	fileContent = addToNewGRPCServerParams(fileContent, service)

	// 3. 在 NewGRPCServer 函数体中添加 RegisterXXXServer
	fileContent = addToNewGRPCServerBody(fileContent, service)

	// 格式化代码
	formatted, err := format.Source([]byte(fileContent))
	if err != nil {
		// 如果格式化失败，使用原始内容
		formatted = []byte(fileContent)
	}

	// 写回文件
	if err := os.WriteFile(grpcFilePath, formatted, 0644); err != nil {
		return fmt.Errorf("写入 grpc.go 文件失败: %v", err)
	}

	return nil
}

// handlerExistsInGrpc 检查 Handler 是否已经在 grpc.go 中存在
func handlerExistsInGrpc(node *ast.File, service ServiceInfo) bool {
	newHandlerName := "New" + service.HandlerName

	// 检查 ProviderSet 中是否存在
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// 检查是否是 wire.NewSet 调用
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "NewSet" {
				// 检查参数列表中是否包含 NewXXXHandler
				for _, arg := range call.Args {
					if ident, ok := arg.(*ast.Ident); ok {
						if ident.Name == newHandlerName {
							return false // 找到了，停止遍历
						}
					}
				}
			}
		}
		return true
	})

	return false
}

// addToProviderSet 在 ProviderSet 中添加 NewXXXHandler
func addToProviderSet(content string, service ServiceInfo) string {
	newHandlerName := "New" + service.HandlerName

	// 查找 ProviderSet 的位置
	providerSetStart := strings.Index(content, "var ProviderSet = wire.NewSet(")
	if providerSetStart == -1 {
		return content
	}

	// 查找 ProviderSet 的结束位置（找到对应的右括号）
	start := providerSetStart + len("var ProviderSet = wire.NewSet(")
	parenCount := 1
	end := start
	for i := start; i < len(content) && parenCount > 0; i++ {
		if content[i] == '(' {
			parenCount++
		} else if content[i] == ')' {
			parenCount--
			if parenCount == 0 {
				end = i
				break
			}
		}
	}

	// 检查是否已经存在
	providerSetContent := content[start:end]
	if strings.Contains(providerSetContent, newHandlerName) {
		return content // 已经存在
	}

	// 找到最后一个 Handler 行的位置（最后一个包含 "New" 和 "Handler" 的行）
	lines := strings.Split(providerSetContent, "\n")
	lastHandlerLineIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "New") && strings.Contains(line, "Handler") {
			lastHandlerLineIdx = i
			break
		}
	}

	if lastHandlerLineIdx == -1 {
		// 没找到，在最后添加
		insertContent := "\n\t" + newHandlerName + ","
		content = content[:end] + insertContent + content[end:]
		return content
	}

	// 找到最后一个 Handler 行的结束位置
	lineStartPos := start
	for i := 0; i < lastHandlerLineIdx; i++ {
		lineStartPos += len(lines[i]) + 1 // +1 for \n
	}
	lastLine := lines[lastHandlerLineIdx]
	lastLineEnd := lineStartPos + len(lastLine)

	// 检查最后一行是否已经有逗号
	trimmedLine := strings.TrimSpace(lastLine)
	insertContent := "\n\t" + newHandlerName + ","
	if !strings.HasSuffix(trimmedLine, ",") {
		// 需要在最后一行添加逗号
		content = content[:lastLineEnd] + "," + content[lastLineEnd:]
		lastLineEnd++
	}

	// 在最后一行后插入新行
	content = content[:lastLineEnd] + insertContent + content[lastLineEnd:]

	return content
}

// addToNewGRPCServerParams 在 NewGRPCServer 参数中添加 handler
func addToNewGRPCServerParams(content string, service ServiceInfo) string {
	handlerVarName := toLowerCamel(service.HandlerName)

	// 查找 NewGRPCServer 函数定义
	funcStart := strings.Index(content, "func NewGRPCServer(")
	if funcStart == -1 {
		return content
	}

	// 查找参数列表的结束位置
	paramStart := funcStart + len("func NewGRPCServer(")
	parenCount := 1
	paramEnd := paramStart
	for i := paramStart; i < len(content) && parenCount > 0; i++ {
		if content[i] == '(' {
			parenCount++
		} else if content[i] == ')' {
			parenCount--
			if parenCount == 0 {
				paramEnd = i
				break
			}
		}
	}

	// 检查参数是否已经存在
	paramsContent := content[paramStart:paramEnd]
	if strings.Contains(paramsContent, handlerVarName+" *"+service.HandlerName) {
		return content // 已经存在
	}

	// 找到最后一个参数行的位置（最后一个包含 "*Handler" 的行）
	lines := strings.Split(paramsContent, "\n")
	lastParamLineIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "*") && strings.Contains(line, "Handler") {
			lastParamLineIdx = i
			break
		}
	}

	if lastParamLineIdx == -1 {
		// 没找到，在参数列表末尾添加
		insertContent := "\n\t" + handlerVarName + " *" + service.HandlerName + ","
		content = content[:paramEnd] + insertContent + content[paramEnd:]
		return content
	}

	// 找到最后一个参数行的结束位置
	lineStartPos := paramStart
	for i := 0; i < lastParamLineIdx; i++ {
		lineStartPos += len(lines[i]) + 1 // +1 for \n
	}
	lastLine := lines[lastParamLineIdx]
	lastLineEnd := lineStartPos + len(lastLine)

	// 检查最后一行是否已经有逗号
	trimmedLine := strings.TrimSpace(lastLine)
	insertContent := "\n\t" + handlerVarName + " *" + service.HandlerName + ","
	if !strings.HasSuffix(trimmedLine, ",") {
		// 需要在最后一行添加逗号
		content = content[:lastLineEnd] + "," + content[lastLineEnd:]
		lastLineEnd++
	}

	// 在最后一行后插入新参数
	content = content[:lastLineEnd] + insertContent + content[lastLineEnd:]

	return content
}

// addToNewGRPCServerBody 在 NewGRPCServer 函数体中添加 RegisterXXXServer
func addToNewGRPCServerBody(content string, service ServiceInfo) string {
	handlerVarName := toLowerCamel(service.HandlerName)
	// ServerName 已经包含了 "Server" 后缀，所以直接用 Register%s
	registerCall := fmt.Sprintf("\tdevopsx.Register%s(srv, %s)", service.ServerName, handlerVarName)

	// 查找 NewGRPCServer 函数体
	funcStart := strings.Index(content, "func NewGRPCServer(")
	if funcStart == -1 {
		return content
	}

	// 找到函数体的开始位置（第一个 {）
	bodyStart := strings.Index(content[funcStart:], "{")
	if bodyStart == -1 {
		return content
	}
	bodyStart += funcStart + 1

	// 找到函数体的结束位置（匹配的 }）
	braceCount := 1
	bodyEnd := bodyStart
	for i := bodyStart; i < len(content) && braceCount > 0; i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				bodyEnd = i
				break
			}
		}
	}

	// 检查是否已经存在 Register 调用
	bodyContent := content[bodyStart:bodyEnd]
	if strings.Contains(bodyContent, "Register"+service.ServerName+"Server") {
		return content // 已经存在
	}

	// 找到最后一个 Register 调用的位置
	lastRegisterPos := strings.LastIndex(bodyContent, "devopsx.Register")
	if lastRegisterPos == -1 {
		// 如果没有找到 Register，查找 return srv 之前
		returnPos := strings.LastIndex(bodyContent, "\treturn srv")
		if returnPos != -1 {
			insertPos := bodyStart + returnPos
			content = content[:insertPos] + registerCall + "\n" + content[insertPos:]
			return content
		}
		return content
	}

	// 找到最后一个 Register 行的结束位置（包括换行符）
	lastRegisterLineEnd := lastRegisterPos
	for lastRegisterLineEnd < len(bodyContent) && bodyContent[lastRegisterLineEnd] != '\n' {
		lastRegisterLineEnd++
	}
	if lastRegisterLineEnd < len(bodyContent) {
		lastRegisterLineEnd++ // 包含换行符
	}

	// 在最后一个 Register 调用后添加新调用
	insertPos := bodyStart + lastRegisterLineEnd
	content = content[:insertPos] + registerCall + "\n" + content[insertPos:]

	return content
}
