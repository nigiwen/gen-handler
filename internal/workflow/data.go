package workflow

import (
	"path/filepath"

	appTypes "github.com/nigiwen/gen-handler/internal/types"
)

type DataWorkflow struct {
	Config            appTypes.Config
	EntityDir         string
	DataDir           string
	GenerateEntity    func(path string) error
	GenerateRepo      func(path string, entity appTypes.EntityInfo) error
	UpdateProviderSet func(filePath, newItem, itemType string) error
	RunWire           func(wireDir string) error
	FileExists        func(path string) bool
}

func (wf DataWorkflow) BuildItems(entities []appTypes.EntityInfo) []Item {
	items := make([]Item, 0, len(entities))
	for _, entity := range entities {
		items = append(items, Item{
			ID:          entity.FileName,
			Title:       entity.FileName,
			Description: "Entity: " + entity.EntityName,
			Keywords:    []string{entity.FileName, entity.EntityName},
			Payload:     entity,
		})
	}
	return items
}

func (wf DataWorkflow) RunItem(item Item, emit func(ProgressEvent)) RunResult {
	entity := item.Payload.(appTypes.EntityInfo)
	entityPath := filepath.Join(wf.EntityDir, entity.FileName+".go")
	repoPath := filepath.Join(wf.DataDir, entity.FileName+".go")

	emit(ProgressEvent{ItemID: item.ID, Step: "检查 entity 占位文件"})
	if !wf.FileExists(entityPath) {
		emit(ProgressEvent{ItemID: item.ID, Step: "生成 entity 占位文件"})
		if err := wf.GenerateEntity(entityPath); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "生成 repo"})
	if wf.FileExists(repoPath) {
		return RunResult{ItemID: item.ID, Title: item.Title, Success: true, Skipped: true}
	}
	if err := wf.GenerateRepo(repoPath, entity); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "更新 ProviderSet"})
	if err := wf.UpdateProviderSet(filepath.Join(wf.DataDir, "data.go"), "New"+entity.EntityName+"Repo", "Repo"); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "运行 wire"})
	if err := wf.RunWire(wf.Config.WireDir); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	return RunResult{ItemID: item.ID, Title: item.Title, Success: true}
}
