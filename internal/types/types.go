package types

import "fmt"

// Config 工具配置
type Config struct {
	// 路径配置
	ProtoDir  string // proto 生成的 grpc 文件目录，如 ./internal/proto/axis/devopsx
	OutputDir string // handler 输出目录，如 ./api/grpc
	CoreDir   string // core service 输出目录，如 ./core
	WireDir   string // wire 命令执行目录，如 ./cmd/devopsx

	// 包名配置（用于代码生成）
	ModulePath string // Go 模块路径，如 bsi/axis/devopsx
}

// ServiceInfo 存储服务接口信息
type ServiceInfo struct {
	ServerName   string   // 如 TestCaseServer
	ProtoPackage string   // 主 proto 包名，如 devopsx
	HandlerName  string   // 如 TestCaseHandler
	FileName     string   // 如 test_case.go
	FieldName    string   // 如 testCaseSrv
	ServiceName  string   // 如 TestCaseService
	Methods      []Method // 方法列表
}

// Method 存储方法信息
type Method struct {
	Name         string // 方法名
	Comment      string // 注释（从 proto 中提取）
	RequestType  string // 请求类型（完整类型名，如 CreateTestCaseRequest）
	RequestPkg   string // 请求类型包名（devopsx 或 basic）
	ResponseType string // 响应类型（完整类型名）
	ResponsePkg  string // 响应类型包名（devopsx, basic 或 zebra）
}

// EntityInfo 实体信息
type EntityInfo struct {
	EntityName string // 如 "Project"
	FileName   string // 如 "project"
}

// GetDisplayName 实现 Selectable 接口
func (s ServiceInfo) GetDisplayName() string {
	return s.FileName
}

// GetDescription 实现 Selectable 接口
func (s ServiceInfo) GetDescription() string {
	return fmt.Sprintf("Handler: %s, %d 个方法", s.HandlerName, len(s.Methods))
}

// GetDisplayName 实现 Selectable 接口
func (e EntityInfo) GetDisplayName() string {
	return e.FileName + ".go"
}

// GetDescription 实现 Selectable 接口
func (e EntityInfo) GetDescription() string {
	return "Entity: " + e.EntityName
}
