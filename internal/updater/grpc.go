package updater

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

// UpdateGrpcFile 更新 grpc.go 文件，添加新的 Handler 到 ProviderSet 和 NewGRPCServer
func UpdateGrpcFile(service types.ServiceInfo, grpcFilePath string) error {
	// 读取文件内容
	content, err := util.ReadFile(grpcFilePath)
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

	// 转换为字符串进行修改
	fileContent := content

	// 1. 在 ProviderSet 中添加 NewXXXHandler
	fileContent, err = addToProviderSet(fileContent, service)
	if err != nil {
		return err
	}

	// 2. 在 NewGRPCServer 参数中添加 handler
	fileContent = addToNewGRPCServerParams(fileContent, service)

	// 3. 在 NewGRPCServer 函数体中添加 RegisterXXXServer
	fileContent = addToNewGRPCServerBody(fileContent, service)

	// 格式化代码
	formatted, err := util.FormatGoFile(fileContent)
	if err != nil {
		// 如果格式化失败，使用原始内容
		formatted = fileContent
	}

	// 写回文件
	if err := util.WriteFile(grpcFilePath, formatted); err != nil {
		return fmt.Errorf("写入 grpc.go 文件失败: %v", err)
	}

	return nil
}

// handlerExistsInGrpc 检查 Handler 是否已经在 grpc.go 中存在
// 改为始终返回 false，让后续的添加函数自己判断是否已存在
// 这样可以确保即使部分添加失败，下次运行时也能补充缺失的部分
func handlerExistsInGrpc(node *ast.File, service types.ServiceInfo) bool {
	// 始终返回 false，让添加流程执行
	// addToProviderSet、addToNewGRPCServerParams 和 addToNewGRPCServerBody
	// 都有自己的检查逻辑，不会重复添加
	return false
}

// addToProviderSet 在 ProviderSet 中添加 NewXXXHandler
func addToProviderSet(content string, service types.ServiceInfo) (string, error) {
	newHandlerName := "New" + service.HandlerName
	return util.InsertIntoProviderSet(content, newHandlerName, "Handler")
}

// addToNewGRPCServerParams 在 NewGRPCServer 参数中添加 handler
func addToNewGRPCServerParams(content string, service types.ServiceInfo) string {
	handlerVarName := util.ToLowerCamel(service.HandlerName)

	// 查找 NewGRPCServer 函数定义
	funcStart := strings.Index(content, "func NewGRPCServer(")
	if funcStart == -1 {
		return content
	}

	// 查找参数列表的结束位置
	paramStart := funcStart + len("func NewGRPCServer(")
	paramEnd := util.FindMatchingBrace(content, paramStart-1)
	if paramEnd == -1 {
		return content
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
func addToNewGRPCServerBody(content string, service types.ServiceInfo) string {
	handlerVarName := util.ToLowerCamel(service.HandlerName)
	protoPackage := grpcProtoPackage(service)
	// ServerName 已经包含了 "Server" 后缀，所以直接用 Register%s
	registerCall := fmt.Sprintf("\t%s.Register%s(srv, %s)", protoPackage, service.ServerName, handlerVarName)

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
	bodyEnd := util.FindMatchingCurlyBrace(content, bodyStart-1)
	if bodyEnd == -1 {
		return content
	}

	// 检查是否已经存在 Register 调用（只检查 RegisterXXXServer，不检查参数）
	bodyContent := content[bodyStart:bodyEnd]
	registerPattern := protoPackage + ".Register" + service.ServerName + "("
	if strings.Contains(bodyContent, registerPattern) {
		return content // 已经存在
	}

	// 查找 return srv 的位置（应该在最后）
	returnPos := strings.LastIndex(bodyContent, "return srv")
	if returnPos == -1 {
		return content
	}

	// 找到 return srv 行的开始位置（向前找到换行符或开头）
	returnLineStart := returnPos
	for returnLineStart > 0 && bodyContent[returnLineStart-1] != '\n' {
		returnLineStart--
	}

	// 在 return srv 之前插入新的 Register 调用
	insertPos := bodyStart + returnLineStart
	content = content[:insertPos] + registerCall + "\n" + content[insertPos:]

	return content
}

func grpcProtoPackage(service types.ServiceInfo) string {
	if service.ProtoPackage != "" {
		return service.ProtoPackage
	}
	for _, method := range service.Methods {
		for _, pkg := range []string{method.RequestPkg, method.ResponsePkg} {
			if pkg != "" && pkg != "basic" && pkg != "zebra" {
				return pkg
			}
		}
	}
	return "devopsx"
}
