package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

type historyModel struct {
	items  []iperf.HistoryEntry
	cursor int
	err    string
	export exportState
}

func (m *Model) gotoHistory() tea.Cmd {
	entries, err := iperf.LoadHistory()
	m.history = historyModel{items: entries}
	if err != nil {
		m.history.err = err.Error()
	}
	m.screen = screenHistory
	return nil
}

func (m *Model) updateHistory(msg tea.Msg) tea.Cmd {
	h := &m.history
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if h.export.active {
			if h.export.handleKey(msg.String()) && h.cursor < len(h.items) {
				h.export.result = exportEntry(h.items[h.cursor], h.export.cursor)
			}
			return nil
		}
		switch msg.String() {
		case "up", "k":
			if h.cursor > 0 {
				h.cursor--
			}
		case "down", "j":
			if h.cursor < len(h.items)-1 {
				h.cursor++
			}
		case "e":
			if len(h.items) > 0 {
				h.export = exportState{active: true}
			}
		case "esc", "q":
			m.screen = screenMenu
			return nil
		}
	}
	return nil
}

func (m *Model) viewHistory() string {
	h := &m.history
	var b strings.Builder
	b.WriteString(titleStyle.Render("KaiPerf3-tui · history") + "\n")
	b.WriteString(subtitleStyle.Render("past test results (newest first)") + "\n\n")

	if len(h.items) == 0 {
		b.WriteString(subtitleStyle.Render("no tests recorded yet") + "\n")
	}

	for i, e := range h.items {
		status := goodStyle.Render("✓")
		if e.Aborted {
			status = warnStyle.Render("⚠")
		}
		line := fmt.Sprintf("%s %s   %s   up %.1f Mbps / down %.1f Mbps",
			status, e.Time.Format("Jan 2 15:04"), e.Config, e.AvgMbpsUp, e.AvgMbpsDown)
		if e.Retransmits > 0 {
			line += fmt.Sprintf("   retx %d", e.Retransmits)
		}
		if e.LossPct > 0 {
			line += fmt.Sprintf("   loss %.2f%%", e.LossPct)
		}
		if i == h.cursor {
			b.WriteString(selectedItemStyle.Render("▶ "+line) + "\n")
		} else {
			b.WriteString(itemStyle.Render("  "+line) + "\n")
		}
	}

	if h.err != "" {
		b.WriteString("\n" + errorStyle.Render("✗ "+h.err) + "\n")
	}

	if h.export.active {
		b.WriteString(h.export.render())
	}

	content := borderStyle.Render(b.String())
	help := "↑↓/jk move · esc/q back"
	if len(h.items) > 0 {
		help = "↑↓/jk move · e export · esc/q back"
	}
	return content + "\n" + helpStyle.Render(help)
}
