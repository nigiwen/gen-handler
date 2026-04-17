package tui

type Item struct {
	ID          string
	Title       string
	Description string
	Keywords    []string
	Payload     any
}

type Status string

const (
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

type Runner func(item Item, emit func(ProgressEvent)) RunResult

type ProgressEvent struct {
	ItemID  string
	Step    string
	Status  Status
	Message string
	Err     error
}

type RunResult struct {
	ItemID  string
	Title   string
	Success bool
	Skipped bool
	Err     error
}

type runItemFinishedMsg struct {
	Result RunResult
}

type runFinishedMsg struct{}

type SessionConfig struct {
	Title string
	Items []Item
	Run   Runner
}
