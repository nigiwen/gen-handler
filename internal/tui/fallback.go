package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func RunSession(cfg SessionConfig) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return runFallback(cfg)
	}

	model := NewModel(cfg)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if model.runtime != nil {
		model.runtime.send = program.Send
	}

	_, err := program.Run()
	return err
}

func runFallback(cfg SessionConfig) error {
	if len(cfg.Items) == 0 {
		return nil
	}

	fmt.Println("\n请选择要执行的项（输入编号，多个用逗号分隔，如: 1,3,5；输入 all 执行全部；直接回车默认第一项）：")
	for i, item := range cfg.Items {
		fmt.Printf("%d. %s (%s)\n", i+1, item.Title, item.Description)
	}
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && input == "" {
		input = ""
	}

	selected := fallbackSelect(cfg.Items, input)
	if len(selected) == 0 {
		fmt.Println("\n⏭️  已取消")
		return nil
	}
	if cfg.Run == nil {
		return nil
	}

	for _, item := range selected {
		fmt.Printf("\n▶ %s\n", item.Title)
		result := cfg.Run(item, func(event ProgressEvent) {
			if strings.TrimSpace(event.Step) == "" {
				return
			}
			fmt.Printf("  - %s\n", event.Step)
		})
		switch {
		case result.Err != nil:
			fmt.Printf("  ✗ %v\n", result.Err)
		case result.Skipped:
			fmt.Println("  - skipped")
		default:
			fmt.Println("  ✓ done")
		}
	}

	return nil
}

func fallbackSelect(items []Item, input string) []Item {
	input = strings.TrimSpace(input)
	if input == "" && len(items) > 0 {
		return []Item{items[0]}
	}
	if strings.EqualFold(input, "all") {
		return append([]Item(nil), items...)
	}

	selected := make([]Item, 0, len(items))
	seen := make(map[int]bool, len(items))
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		index, err := strconv.Atoi(part)
		if err != nil || index <= 0 || index > len(items) {
			continue
		}
		if seen[index-1] {
			continue
		}
		seen[index-1] = true
		selected = append(selected, items[index-1])
	}

	return selected
}
