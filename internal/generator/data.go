package generator

import (
	"fmt"
	
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

// GenerateDbsetFile 生成 dbset 文件
func GenerateDbsetFile(path string, entity types.EntityInfo, modulePath string) error {
	// 准备模板数据
	data := struct {
		ModulePath string
		EntityName string
	}{
		ModulePath: modulePath,
		EntityName: entity.EntityName,
	}
	
	content, err := ExecuteTemplate(DbsetTemplate, data)
	if err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}
	
	return util.WriteFile(path, content)
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
