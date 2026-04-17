package workflow

import appTypes "github.com/nigiwen/gen-handler/internal/types"

type HandlerWorkflow struct {
	Config                appTypes.Config
	EnsureHandlerFile     func(service appTypes.ServiceInfo, config appTypes.Config) (bool, error)
	UpdateGrpcProvider    func(service appTypes.ServiceInfo, outputDir string) error
	EnsureCoreServiceFile func(service appTypes.ServiceInfo, config appTypes.Config) (bool, error)
	UpdateCoreProvider    func(service appTypes.ServiceInfo, coreDir string) error
	RunWire               func(wireDir string) error
}

func (wf HandlerWorkflow) BuildItems(services []appTypes.ServiceInfo) []Item {
	items := make([]Item, 0, len(services))
	for _, service := range services {
		items = append(items, Item{
			ID:          service.FileName,
			Title:       service.FileName,
			Description: "Handler: " + service.HandlerName,
			Keywords:    []string{service.FileName, service.HandlerName, service.ServiceName},
			Payload:     service,
		})
	}
	return items
}

func (wf HandlerWorkflow) RunItem(item Item, emit func(ProgressEvent)) RunResult {
	service := item.Payload.(appTypes.ServiceInfo)

	emit(ProgressEvent{ItemID: item.ID, Step: "处理 handler 文件"})
	grpcCreated, err := wf.EnsureHandlerFile(service, wf.Config)
	if err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	if grpcCreated {
		emit(ProgressEvent{ItemID: item.ID, Step: "更新 grpc.go"})
		if err := wf.UpdateGrpcProvider(service, wf.Config.OutputDir); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "处理 core service"})
	coreCreated, err := wf.EnsureCoreServiceFile(service, wf.Config)
	if err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	if coreCreated {
		emit(ProgressEvent{ItemID: item.ID, Step: "更新 core ProviderSet"})
		if err := wf.UpdateCoreProvider(service, wf.Config.CoreDir); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	if grpcCreated || coreCreated {
		emit(ProgressEvent{ItemID: item.ID, Step: "运行 wire"})
		if err := wf.RunWire(wf.Config.WireDir); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	return RunResult{ItemID: item.ID, Title: item.Title, Success: true}
}
