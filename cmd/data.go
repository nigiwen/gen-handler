package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nigiwen/gen-handler/internal/generator"
	"github.com/nigiwen/gen-handler/internal/tui"
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/updater"
	"github.com/nigiwen/gen-handler/internal/util"
	"github.com/nigiwen/gen-handler/internal/workflow"
)

var (
	updateProviderSet = updater.UpdateProviderSet
	runWireCommand    = generator.RunWireCommand
)

// RunDataCommand 执行 Data 层同步逻辑
func RunDataCommand(config types.Config) {
	// 1. 确定目录路径
	projectRoot := "." // 假设在项目根目录运行
	entityDir := filepath.Join(projectRoot, "internal/model/entity")
	dataDir := filepath.Join(projectRoot, "data")

	// 确保 data 目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("❌ 创建 data 目录失败: %v\n", err)
		return
	}

	// 2. 扫描 entity 目录
	entities, err := scanEntities(entityDir)
	if err != nil {
		fmt.Printf("❌ 扫描 entity 目录失败: %v\n", err)
		return
	}

	if len(entities) == 0 {
		fmt.Printf("⚠️  未在 %s 找到实体文件\n", entityDir)
		return
	}

	// 3. 筛选缺失的实体
	missingEntities := filterPendingEntities(entities, dataDir)

	if len(missingEntities) == 0 {
		fmt.Println("✅ 所有实体已同步，无需生成")
		return
	}

	fmt.Printf("🔍 找到 %d 个待同步的实体\n", len(missingEntities))

	wf := workflow.DataWorkflow{
		Config:            config,
		EntityDir:         entityDir,
		DataDir:           dataDir,
		GenerateEntity:    generator.GenerateEntityFile,
		GenerateRepo:      generator.GenerateRepoFile,
		UpdateProviderSet: updateProviderSet,
		RunWire:           runWireCommand,
		FileExists:        util.FileExists,
	}

	if err := runSession(tui.SessionConfig{
		Title: "Data Sync",
		Items: wf.BuildItems(missingEntities),
		Run:   wf.RunItem,
	}); err != nil {
		fmt.Printf("❌ Data Sync 失败: %v\n", err)
	}
}

func filterPendingEntities(entities []types.EntityInfo, dataDir string) []types.EntityInfo {
	var pending []types.EntityInfo
	for _, entity := range entities {
		repoPath := filepath.Join(dataDir, entity.FileName+".go")
		if !util.FileExists(repoPath) {
			pending = append(pending, entity)
		}
	}

	return pending
}

func syncSelectedEntity(projectRoot string, config types.Config, entity types.EntityInfo) error {
	entityDir := filepath.Join(projectRoot, "internal/model/entity")
	dataDir := filepath.Join(projectRoot, "data")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建 data 目录失败: %w", err)
	}

	entityPath := filepath.Join(entityDir, entity.FileName+".go")
	if !util.FileExists(entityPath) {
		if err := generator.GenerateEntityFile(entityPath); err != nil {
			return fmt.Errorf("生成 entity %s 失败: %w", entity.FileName, err)
		}
	}

	repoPath := filepath.Join(dataDir, entity.FileName+".go")
	if util.FileExists(repoPath) {
		return nil
	}

	if err := generator.GenerateRepoFile(repoPath, entity); err != nil {
		return fmt.Errorf("生成 repo %s 失败: %w", entity.FileName, err)
	}

	dataGoPath := filepath.Join(dataDir, "data.go")
	if err := updateProviderSet(dataGoPath, "New"+entity.EntityName+"Repo", "Repo"); err != nil {
		return fmt.Errorf("注册 %s 到 ProviderSet 失败: %w", entity.EntityName, err)
	}

	if err := runWireCommand(config.WireDir); err != nil {
		return fmt.Errorf("运行 wire 命令失败: %w", err)
	}

	return nil
}
