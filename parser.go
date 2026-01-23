package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// parseGrpcFile 解析 grpc 文件，提取服务接口信息
func parseGrpcFile(filePath string) ([]ServiceInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var services []ServiceInfo

	// 遍历 AST，查找 Server 接口
	ast.Inspect(node, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// 检查是否是 Server 接口（以 Server 结尾）
		typeName := ts.Name.Name
		if !strings.HasSuffix(typeName, "Server") || strings.Contains(typeName, "Unimplemented") {
			return true
		}

		// 检查是否是接口类型
		it, ok := ts.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		// 提取服务信息
		service := extractServiceInfo(typeName, it, fset, node)
		if service != nil {
			services = append(services, *service)
		}

		return true
	})

	return services, nil
}

// extractServiceInfo 从接口定义中提取服务信息
func extractServiceInfo(serverName string, it *ast.InterfaceType, fset *token.FileSet, node *ast.File) *ServiceInfo {
	// 去掉 Server 后缀，得到基础名称
	baseName := strings.TrimSuffix(serverName, "Server")

	// 转换为文件名（驼峰转下划线，全小写）
	fileName := camelToSnake(baseName) + ".go"

	// Handler 名称
	handlerName := baseName + "Handler"

	// 字段名称（首字母小写的驼峰）
	fieldName := toLowerCamel(baseName) + "Srv"

	// Service 名称
	serviceName := baseName + "Service"

	var methods []Method

	// 提取方法
	for _, method := range it.Methods.List {
		if len(method.Names) == 0 {
			continue
		}

		methodName := method.Names[0].Name

		// 跳过 mustEmbedUnimplemented 方法
		if strings.HasPrefix(methodName, "mustEmbed") {
			continue
		}

		// 获取方法注释
		comment := ""
		if method.Doc != nil && len(method.Doc.List) > 0 {
			// 提取注释文本，去掉 // 前缀
			for _, c := range method.Doc.List {
				text := strings.TrimSpace(c.Text)
				if strings.HasPrefix(text, "//") {
					text = strings.TrimSpace(text[2:])
					if text != "" {
						comment = text
						break
					}
				}
			}
		}

		// 解析方法签名
		ft, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		// 提取参数和返回值类型
		var requestType, requestPkg, responseType, responsePkg string

		// 第一个参数是 context.Context，第二个是请求类型
		if ft.Params != nil && len(ft.Params.List) >= 2 {
			if sel, ok := ft.Params.List[1].Type.(*ast.StarExpr); ok {
				requestType, requestPkg = extractTypeInfo(sel.X, node)
			}
		}

		// 返回值：第一个是响应类型，第二个是 error
		if ft.Results != nil && len(ft.Results.List) >= 1 {
			if sel, ok := ft.Results.List[0].Type.(*ast.StarExpr); ok {
				responseType, responsePkg = extractTypeInfo(sel.X, node)
			}
		}

		// 默认包名为 devopsx
		if requestPkg == "" {
			requestPkg = "devopsx"
		}
		if responsePkg == "" {
			responsePkg = "devopsx"
		}

		methods = append(methods, Method{
			Name:         methodName,
			Comment:      comment,
			RequestType:  requestType,
			RequestPkg:   requestPkg,
			ResponseType: responseType,
			ResponsePkg:  responsePkg,
		})
	}

	if len(methods) == 0 {
		return nil
	}

	return &ServiceInfo{
		ServerName:  serverName,
		HandlerName: handlerName,
		FileName:    fileName,
		FieldName:   fieldName,
		ServiceName: serviceName,
		Methods:     methods,
	}
}

// extractTypeInfo 从 AST 节点提取类型信息（类型名和包名）
func extractTypeInfo(expr ast.Expr, node *ast.File) (typeName, pkgName string) {
	switch x := expr.(type) {
	case *ast.Ident:
		// 简单标识符，检查是否是导入的类型
		typeName = x.Name
		// 检查是否是 basic 或 zebra 包的类型
		if typeName == "String" || typeName == "Empty" || typeName == "Int64" || typeName == "PageRequest" {
			pkgName = "basic"
		} else if typeName == "OpenApiInfo" {
			pkgName = "zebra"
		} else {
			pkgName = "devopsx"
		}
	case *ast.SelectorExpr:
		// 选择器表达式，如 basic.String
		if ident, ok := x.X.(*ast.Ident); ok {
			pkgName = ident.Name
			typeName = x.Sel.Name
		}
	}
	return
}
