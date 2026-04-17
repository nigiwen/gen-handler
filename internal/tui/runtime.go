package tui

import (
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionRuntime struct {
	run           Runner
	send          func(tea.Msg)
	stopRequested atomic.Bool
}

func newSessionRuntime(run Runner) *sessionRuntime {
	if run == nil {
		return nil
	}
	return &sessionRuntime{run: run}
}

func (rt *sessionRuntime) reset() {
	if rt == nil {
		return
	}
	rt.stopRequested.Store(false)
}

func (rt *sessionRuntime) requestStop() {
	if rt == nil {
		return
	}
	rt.stopRequested.Store(true)
}

func (rt *sessionRuntime) shouldStop() bool {
	if rt == nil {
		return false
	}
	return rt.stopRequested.Load()
}

func (rt *sessionRuntime) sendMsg(msg tea.Msg) {
	if rt == nil || rt.send == nil {
		return
	}
	rt.send(msg)
}

func (rt *sessionRuntime) runQueue(items []Item) {
	if rt == nil || rt.run == nil {
		return
	}

	for _, item := range items {
		if rt.shouldStop() {
			break
		}

		result := rt.run(item, func(event ProgressEvent) {
			rt.sendMsg(event)
		})
		if result.ItemID == "" {
			result.ItemID = item.ID
		}
		if result.Title == "" {
			result.Title = item.Title
		}
		rt.sendMsg(runItemFinishedMsg{Result: result})

		if rt.shouldStop() {
			break
		}
	}

	rt.sendMsg(runFinishedMsg{})
}
