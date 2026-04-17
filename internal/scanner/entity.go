package scanner

import (
	"os"
	"strings"

	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

// ScanEntities 扫描目录下的实体文件
func ScanEntities(dir string) ([]types.EntityInfo, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entities []types.EntityInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".gen.go") {
			continue
		}

		fileName := strings.TrimSuffix(file.Name(), ".gen.go")
		if fileName == "" {
			continue
		}

		entities = append(entities, types.EntityInfo{
			EntityName: util.ToUpperCamel(fileName),
			FileName:   fileName,
		})
	}
	return entities, nil
}
