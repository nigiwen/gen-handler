package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EntityInfo 实体信息
type EntityInfo struct {
	EntityName string // 如 "Project"
	FileName   string // 如 "project"
}

// runDataSync 执行 Data 层同步逻辑
func runDataSync(config Config) {
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
	var missingEntities []EntityInfo
	for _, entity := range entities {
		dbsetPath := filepath.Join(dbsetDir, entity.FileName+".go")
		repoPath := filepath.Join(dataDir, entity.FileName+".go")

		// 只要有一个不存在，就认为需要同步
		if !fileExists(dbsetPath) || !fileExists(repoPath) {
			missingEntities = append(missingEntities, entity)
		}
	}

	if len(missingEntities) == 0 {
		fmt.Println("✅ 所有实体已同步，无需生成")
		return
	}

	fmt.Printf("🔍 找到 %d 个待同步的实体\n", len(missingEntities))

	// 4. 交互式选择
	selected := selectItems(missingEntities)
	if len(selected) == 0 {
		fmt.Println("\n⏭️  已取消")
		return
	}

	// 5. 逐个处理选中的实体
	for _, entity := range selected {
		// 步骤 A: 同步 dbset 层 (类型别名 + BeforeCreate)
		dbsetPath := filepath.Join(dbsetDir, entity.FileName+".go")
		if !fileExists(dbsetPath) {
			if err := generateDbsetFile(dbsetPath, entity, config.ModulePath); err != nil {
				fmt.Printf("⚠️  生成 dbset %s 失败: %v\n", entity.FileName, err)
			} else {
				fmt.Printf("✅ 已生成 dbset: %s\n", entity.FileName)
			}
		}

		// 步骤 B: 同步 data 层 (Repo 结构体 + 构造函数)
		repoPath := filepath.Join(dataDir, entity.FileName+".go")
		if !fileExists(repoPath) {
			if err := generateRepoFile(repoPath, entity); err != nil {
				fmt.Printf("⚠️  生成 repo %s 失败: %v\n", entity.FileName, err)
			} else {
				fmt.Printf("✅ 已生成 repo: %s\n", entity.FileName)

				// 步骤 C: 自动注册到 ProviderSet (仅在生成 repo 时尝试注册)
				dataGoPath := filepath.Join(dataDir, "data.go")
				if err := registerToProviderSet(dataGoPath, "New"+entity.EntityName+"Repo"); err != nil {
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

// scanEntities 扫描目录下的实体文件
func scanEntities(dir string) ([]EntityInfo, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entities []EntityInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		// 排除一些通用的文件
		fileName := strings.TrimSuffix(file.Name(), ".go")
		if fileName == "entity" || strings.HasSuffix(fileName, "_test") {
			continue
		}

		entities = append(entities, EntityInfo{
			EntityName: toUpperCamel(fileName),
			FileName:   fileName,
		})
	}
	return entities, nil
}

// generateDbsetFile 生成 dbset 文件
func generateDbsetFile(path string, entity EntityInfo, modulePath string) error {
	content := fmt.Sprintf(`package dbset

import (
	jujuerrors "github.com/juju/errors"
	"gorm.io/gorm"

	"%s/internal/model/entity"
	"bsi/kratos/micro/algo/snow"
)

type %s struct {
	entity.%s
}

func (t *%s) BeforeCreate(tx *gorm.DB) error {
	if t.ID > 0 {
		return nil
	}

	id, err := snow.ID()
	if err != nil {
		return jujuerrors.Trace(err)
	}
	t.ID = id

	return nil
}
`, modulePath, entity.EntityName, entity.EntityName, entity.EntityName)

	return os.WriteFile(path, []byte(content), 0644)
}

// generateRepoFile 生成 repo 文件
func generateRepoFile(path string, entity EntityInfo) error {
	content := fmt.Sprintf(`package data

import (
	"github.com/go-kratos/kratos/v2/log"

	mgorm "bsi/kratos/micro/gorm"
)

type %sRepo struct {
	log *log.Helper
	*mgorm.GormDB
}

func New%sRepo(logger log.Logger, db *mgorm.GormDB) *%sRepo {
	return &%sRepo{
		log:    log.NewHelper(logger),
		GormDB: db,
	}
}
`, entity.EntityName, entity.EntityName, entity.EntityName, entity.EntityName)

	return os.WriteFile(path, []byte(content), 0644)
}

// registerToProviderSet 自动注册到 ProviderSet
func registerToProviderSet(filePath, repoName string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	sContent := string(content)

	// 1. 检查是否已存在
	if strings.Contains(sContent, repoName) {
		return nil
	}

	// 2. 寻找 ProviderSet = wire.NewSet( 的位置
	startIdx := strings.Index(sContent, "ProviderSet = wire.NewSet(")
	if startIdx == -1 {
		// 尝试匹配没有空格的情况
		startIdx = strings.Index(sContent, "ProviderSet=wire.NewSet(")
		if startIdx == -1 {
			return fmt.Errorf("未找到 ProviderSet = wire.NewSet( 定义")
		}
	}

	// 3. 寻找 wire.NewSet(...) 的闭合括号
	// 我们需要找到与 wire.NewSet( 匹配的那个右括号
	// 逻辑：从 startIdx 开始，记录左括号数量，直到找到匹配的右括号
	bracketCount := 0
	endIdx := -1
	for i := startIdx; i < len(sContent); i++ {
		if sContent[i] == '(' {
			bracketCount++
		} else if sContent[i] == ')' {
			bracketCount--
			if bracketCount == 0 {
				endIdx = i
				break
			}
		}
	}

	if endIdx == -1 {
		return fmt.Errorf("未找到 ProviderSet 的闭合括号")
	}

	// 4. 在闭合括号前插入新项
	prefix := sContent[:endIdx]
	suffix := sContent[endIdx:]

	// 检查前缀最后一个非空字符是否是逗号
	trimmedPrefix := strings.TrimRight(prefix, " \t\n\r")
	if !strings.HasSuffix(trimmedPrefix, ",") && !strings.HasSuffix(trimmedPrefix, "(") {
		prefix = trimmedPrefix + ","
	} else {
		prefix = trimmedPrefix
	}

	newContent := prefix + "\n\t" + repoName + ",\n" + suffix
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// runWire 在指定目录下运行 wire 命令
func runWire(wireDir string) {
	fmt.Printf("🔧 正在运行 wire 命令: %s\n", wireDir)
	if err := runWireCommand(wireDir); err != nil {
		fmt.Printf("⚠️  运行 wire 命令失败: %v\n", err)
		fmt.Printf("💡 请手动在 %s 目录下运行 wire 命令\n", wireDir)
	} else {
		fmt.Printf("✅ 已重新生成 wire 依赖注入代码\n")
	}
}
