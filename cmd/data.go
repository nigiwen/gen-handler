package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/nigiwen/gen-handler/internal/generator"
	"github.com/nigiwen/gen-handler/internal/scanner"
	"github.com/nigiwen/gen-handler/internal/selector"
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/updater"
	"github.com/nigiwen/gen-handler/internal/util"
)

// RunDataCommand 执行 Data 层同步逻辑
func RunDataCommand(config types.Config) {
	// 1. 确定目录路径
	projectRoot := "." // 假设在项目根目录运行
	entityDir := filepath.Join(projectRoot, "internal/model/entity")
	dataDir := filepath.Join(projectRoot, "data")
	dbsetDir := filepath.Join(dataDir, "dbset")

	// 确保目录存在
	if err := os.MkdirAll(dbsetDir, 0755); err != nil {
		fmt.Printf("❌ 创建 dbset 目录失败: %v\n", err)
		return
	}

	// 2. 扫描 entity 目录
	entities, err := scanner.ScanEntities(entityDir)
	if err != nil {
		fmt.Printf("❌ 扫描 entity 目录失败: %v\n", err)
		return
	}

	if len(entities) == 0 {
		fmt.Printf("⚠️  未在 %s 找到实体文件\n", entityDir)
		return
	}

	// 3. 筛选缺失的实体
	var missingEntities []types.EntityInfo
	for _, entity := range entities {
		dbsetPath := filepath.Join(dbsetDir, entity.FileName+".go")
		repoPath := filepath.Join(dataDir, entity.FileName+".go")

		// 只要有一个不存在，就认为需要同步
		if !util.FileExists(dbsetPath) || !util.FileExists(repoPath) {
			missingEntities = append(missingEntities, entity)
		}
	}

	if len(missingEntities) == 0 {
		fmt.Println("✅ 所有实体已同步，无需生成")
		return
	}

	fmt.Printf("🔍 找到 %d 个待同步的实体\n", len(missingEntities))

	// 4. 交互式选择
	selected := selector.SelectItems(missingEntities)
	if len(selected) == 0 {
		fmt.Println("\n⏭️  已取消")
		return
	}

	// 5. 逐个处理选中的实体
	for _, entity := range selected {
		// 步骤 A: 同步 dbset 层 (类型别名 + BeforeCreate)
		dbsetPath := filepath.Join(dbsetDir, entity.FileName+".go")
		if !util.FileExists(dbsetPath) {
			if err := generator.GenerateDbsetFile(dbsetPath, entity, config.ModulePath); err != nil {
				fmt.Printf("⚠️  生成 dbset %s 失败: %v\n", entity.FileName, err)
			} else {
				fmt.Printf("✅ 已生成 dbset: %s\n", entity.FileName)
			}
		}

		// 步骤 B: 同步 data 层 (Repo 结构体 + 构造函数)
		repoPath := filepath.Join(dataDir, entity.FileName+".go")
		if !util.FileExists(repoPath) {
			if err := generator.GenerateRepoFile(repoPath, entity); err != nil {
				fmt.Printf("⚠️  生成 repo %s 失败: %v\n", entity.FileName, err)
			} else {
				fmt.Printf("✅ 已生成 repo: %s\n", entity.FileName)

				// 步骤 C: 自动注册到 ProviderSet (仅在生成 repo 时尝试注册)
				dataGoPath := filepath.Join(dataDir, "data.go")
				if err := updater.UpdateProviderSet(dataGoPath, "New"+entity.EntityName+"Repo", "Repo"); err != nil {
					fmt.Printf("⚠️  注册 %s 到 ProviderSet 失败: %v\n", entity.EntityName, err)
				} else {
					fmt.Printf("🔗 已注册到 ProviderSet: New%sRepo\n", entity.EntityName)

					// 步骤 D: 运行 wire 命令
					runWire(config.WireDir)
				}
			}
		}
	}
}

// runWire 在指定目录下运行 wire 命令
func runWire(wireDir string) {
	fmt.Printf("🔧 正在运行 wire 命令: %s\n", wireDir)
	if err := generator.RunWireCommand(wireDir); err != nil {
		fmt.Printf("⚠️  运行 wire 命令失败: %v\n", err)
		fmt.Printf("💡 请手动在 %s 目录下运行 wire 命令\n", wireDir)
	} else {
		fmt.Printf("✅ 已重新生成 wire 依赖注入代码\n")
	}
}
