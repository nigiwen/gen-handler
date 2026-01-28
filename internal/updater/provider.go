package updater

import (
	"fmt"
	
	"github.com/nigiwen/gen-handler/internal/util"
)

// UpdateProviderSet 通用的 ProviderSet 更新函数
// filePath: 要更新的文件路径（如 grpc.go、core.go、data.go）
// newItem: 要添加的新项（如 "NewTestCaseHandler"）
// itemType: 项的类型（用于智能插入位置，如 "Handler", "Service", "Repo"）
func UpdateProviderSet(filePath, newItem, itemType string) error {
	// 读取文件
	content, err := util.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	// 插入新项
	updatedContent, err := util.InsertIntoProviderSet(content, newItem, itemType)
	if err != nil {
		return fmt.Errorf("更新 ProviderSet 失败: %v", err)
	}
	
	// 格式化代码
	formatted, err := util.FormatGoFile(updatedContent)
	if err != nil {
		// 格式化失败不影响主流程，使用未格式化的内容
		formatted = updatedContent
	}
	
	// 写回文件
	if err := util.WriteFile(filePath, formatted); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	
	return nil
}
