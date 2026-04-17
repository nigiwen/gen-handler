package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateSelecting state = iota
	stateSearching
	stateRunning
	stateSummary
)

type Model struct {
	state            state
	title            string
	allItems         []Item
	visible          []Item
	selected         map[string]bool
	cursor           int
	listOffset       int
	windowHeight     int
	runQueue         []Item
	searchInput      textinput.Model
	filterQuery      string
	spinner          spinner.Model
	completed        int
	succeeded        int
	failed           int
	skipped          int
	currentItem      string
	currentStep      string
	recent           []recentEvent
	results          []RunResult
	run              Runner
	runtime          *sessionRuntime
	stopAfterCurrent bool
}

type recentEvent struct {
	ItemID  string
	Step    string
	Status  Status
	Message string
}

func NewModel(cfg SessionConfig) Model {
	searchInput := textinput.New()
	searchInput.Prompt = "/ "
	spin := spinner.New()
	spin.Spinner = spinner.Dot

	return Model{
		state:       stateSelecting,
		title:       cfg.Title,
		allItems:    append([]Item(nil), cfg.Items...),
		visible:     append([]Item(nil), cfg.Items...),
		selected:    make(map[string]bool),
		searchInput: searchInput,
		spinner:     spin,
		run:         cfg.Run,
		runtime:     newSessionRuntime(cfg.Run),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchChangedMsg:
		m.filterQuery = string(msg)
		m.applyFilter()
		return m, nil
	case ProgressEvent:
		m.currentItem = msg.ItemID
		m.currentStep = msg.Step
		m.recent = append([]recentEvent{{
			ItemID:  msg.ItemID,
			Step:    msg.Step,
			Status:  msg.Status,
			Message: msg.Message,
		}}, m.recent...)
		if len(m.recent) > 3 {
			m.recent = m.recent[:3]
		}
		return m, nil
	case runItemFinishedMsg:
		m.completed++
		m.results = append(m.results, msg.Result)
		switch {
		case msg.Result.Skipped:
			m.skipped++
		case msg.Result.Success:
			m.succeeded++
		default:
			m.failed++
		}
		return m, nil
	case runFinishedMsg:
		m.state = stateSummary
		return m, nil
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.ensureCursorVisible()
		return m, nil
	case tea.KeyMsg:
		switch m.state {
		case stateSelecting:
			return m.updateSelecting(msg)
		case stateSearching:
			return m.updateSearching(msg)
		case stateRunning:
			return m.updateRunning(msg)
		case stateSummary:
			return m.updateSummary(msg)
		default:
			return m, nil
		}
	case spinner.TickMsg:
		if m.state != stateRunning {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

type searchChangedMsg string

func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.filterQuery))
	if query == "" {
		m.visible = append([]Item(nil), m.allItems...)
		m.cursor = 0
		m.listOffset = 0
		m.ensureCursorVisible()
		return
	}

	var filtered []Item
	for _, item := range m.allItems {
		blob := strings.ToLower(item.Title + " " + item.Description + " " + strings.Join(item.Keywords, " "))
		if strings.Contains(blob, query) {
			filtered = append(filtered, item)
		}
	}

	m.visible = filtered
	if len(m.visible) == 0 {
		m.cursor = 0
		m.listOffset = 0
		return
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}
	m.ensureCursorVisible()
}

func (m Model) updateSelecting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEsc:
		return m, tea.Quit
	case msg.Type == tea.KeyCtrlC:
		return m, tea.Quit
	case isSelectToggleKey(msg):
		return m.toggleCurrentSelection(), nil
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q':
		return m, tea.Quit
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '/':
		m.state = stateSearching
		m.searchInput.Focus()
		return m, textinput.Blink
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'a':
		allSelected := true
		for _, item := range m.visible {
			if !m.selected[item.ID] {
				allSelected = false
				break
			}
		}
		for _, item := range m.visible {
			m.selected[item.ID] = !allSelected
		}
		return m, nil
	}

	return m.updateSelectingKeys(msg)
}

func (m Model) updateSelectingKeys(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSelecting:
		switch keyMsg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
			}
		case tea.KeyDown:
			if m.cursor < len(m.visible)-1 {
				m.cursor++
				m.ensureCursorVisible()
			}
		case keySelect:
			if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
				return m, nil
			}

			item := m.visible[m.cursor]
			m.selected[item.ID] = !m.selected[item.ID]
		case keySubmit:
			m.runQueue = m.selectedItemsOrCurrent()
			if len(m.runQueue) == 0 {
				return m, nil
			}
			m.state = stateRunning
			if m.runtime != nil {
				m.runtime.reset()
				return m, tea.Batch(m.spinner.Tick, m.startRun())
			}
			return m, m.spinner.Tick
		}
	}

	return m, nil
}

func (m Model) updateSearching(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchInput.SetValue("")
		m.filterQuery = ""
		m.applyFilter()
		m.searchInput.Blur()
		m.state = stateSelecting
		return m, nil
	case tea.KeyEnter:
		m.searchInput.Blur()
		m.state = stateSelecting
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.filterQuery = m.searchInput.Value()
	m.applyFilter()
	return m, cmd
}

func (m Model) updateRunning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyCtrlC:
		return m, tea.Quit
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q':
		m.stopAfterCurrent = true
		if m.runtime != nil {
			m.runtime.requestStop()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEnter:
		return m, tea.Quit
	case msg.Type == tea.KeyCtrlC:
		return m, tea.Quit
	case msg.Type == tea.KeyEsc:
		return m, tea.Quit
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q':
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) selectedItemsOrCurrent() []Item {
	items := make([]Item, 0, len(m.visible))
	for _, item := range m.visible {
		if m.selected[item.ID] {
			items = append(items, item)
		}
	}
	if len(items) > 0 {
		return items
	}
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil
	}
	return []Item{m.visible[m.cursor]}
}

func (m Model) selectedCount() int {
	count := 0
	for _, picked := range m.selected {
		if picked {
			count++
		}
	}
	return count
}

func isSelectToggleKey(msg tea.KeyMsg) bool {
	if msg.Type == keySelect {
		return true
	}
	return msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == ' '
}

func (m Model) toggleCurrentSelection() Model {
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return m
	}

	item := m.visible[m.cursor]
	m.selected[item.ID] = !m.selected[item.ID]
	return m
}

const (
	selectViewReservedRows = 7
	selectViewItemRows     = 2
)

func (m *Model) ensureCursorVisible() {
	if len(m.visible) == 0 {
		m.cursor = 0
		m.listOffset = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}

	capacity := m.visibleItemCapacity()
	if capacity <= 0 || capacity >= len(m.visible) {
		m.listOffset = 0
		return
	}

	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+capacity {
		m.listOffset = m.cursor - capacity + 1
	}

	maxOffset := len(m.visible) - capacity
	if m.listOffset < 0 {
		m.listOffset = 0
	}
	if m.listOffset > maxOffset {
		m.listOffset = maxOffset
	}
}

func (m Model) visibleItemCapacity() int {
	if m.windowHeight <= 0 {
		return 0
	}

	availableRows := m.windowHeight - selectViewReservedRows
	if availableRows <= 0 {
		return 1
	}

	capacity := availableRows / selectViewItemRows
	if capacity < 1 {
		return 1
	}
	return capacity
}

func (m Model) visibleWindow() []Item {
	capacity := m.visibleItemCapacity()
	if capacity <= 0 || capacity >= len(m.visible) {
		return m.visible
	}

	start := m.listOffset
	if start < 0 {
		start = 0
	}
	if start > len(m.visible)-capacity {
		start = len(m.visible) - capacity
	}
	if start < 0 {
		start = 0
	}

	end := start + capacity
	if end > len(m.visible) {
		end = len(m.visible)
	}
	return m.visible[start:end]
}

func (m Model) startRun() tea.Cmd {
	if m.runtime == nil || len(m.runQueue) == 0 {
		return nil
	}

	queue := append([]Item(nil), m.runQueue...)
	runtime := m.runtime

	return func() tea.Msg {
		go runtime.runQueue(queue)
		return nil
	}
}
