package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSessionRuntimeRunQueueProcessesItemsUntilFinished(t *testing.T) {
	var msgs []tea.Msg

	rt := newSessionRuntime(func(item Item, emit func(ProgressEvent)) RunResult {
		emit(ProgressEvent{ItemID: item.ID, Step: "生成 repo", Status: StatusSuccess})
		return RunResult{ItemID: item.ID, Title: item.Title, Success: true}
	})
	rt.send = func(msg tea.Msg) {
		msgs = append(msgs, msg)
	}

	rt.runQueue([]Item{{ID: "field_module", Title: "field_module"}})

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	if _, ok := msgs[0].(ProgressEvent); !ok {
		t.Fatalf("expected first message to be ProgressEvent, got %T", msgs[0])
	}

	resultMsg, ok := msgs[1].(runItemFinishedMsg)
	if !ok {
		t.Fatalf("expected second message to be runItemFinishedMsg, got %T", msgs[1])
	}
	if !resultMsg.Result.Success {
		t.Fatalf("expected successful result, got %+v", resultMsg.Result)
	}

	if _, ok := msgs[2].(runFinishedMsg); !ok {
		t.Fatalf("expected third message to be runFinishedMsg, got %T", msgs[2])
	}
}

func TestSessionRuntimeRunQueueStopsAfterCurrentItem(t *testing.T) {
	var (
		ran []string
		rt  *sessionRuntime
	)

	rt = newSessionRuntime(func(item Item, emit func(ProgressEvent)) RunResult {
		ran = append(ran, item.ID)
		if item.ID == "first" {
			rt.requestStop()
		}
		emit(ProgressEvent{ItemID: item.ID, Step: "生成 repo", Status: StatusSuccess})
		return RunResult{ItemID: item.ID, Title: item.Title, Success: true}
	})
	rt.send = func(msg tea.Msg) {}

	rt.runQueue([]Item{
		{ID: "first", Title: "first"},
		{ID: "second", Title: "second"},
	})

	if len(ran) != 1 || ran[0] != "first" {
		t.Fatalf("expected only first item to run, got %v", ran)
	}
}
