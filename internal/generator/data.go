package generator

import (
	"fmt"

	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

// GenerateEntityFile 生成 entity 占位文件
func GenerateEntityFile(path string) error {
	return util.WriteFile(path, EntityTemplate)
}

// GenerateRepoFile 生成 repo 文件
func GenerateRepoFile(path string, entity types.EntityInfo) error {
	// 准备模板数据
	data := struct {
		EntityName string
	}{
		EntityName: entity.EntityName,
	}

	content, err := ExecuteTemplate(RepoTemplate, data)
	if err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	return util.WriteFile(path, content)
}
