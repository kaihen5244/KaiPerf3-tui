package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

type menuItem struct {
	label string
	desc  string
}

var menuItems = []menuItem{
	{"Run client test", "connect to a remote iperf3 server and measure throughput"},
	{"Run server", "listen for incoming iperf3 client connections"},
	{"History", "review past test results"},
	{"Presets", "manage saved client configurations"},
	{"Quit", "exit KaiPerf3-tui"},
}

func (m *Model) updateMenu(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < len(menuItems)-1 {
				m.menuCursor++
			}
		case "enter":
			switch m.menuCursor {
			case 0:
				m.screen = screenForm
			case 1:
				m.screen = screenServerConfig
			case 2:
				return m.gotoHistory()
			case 3:
				return m.gotoPresets(false)
			case 4:
				m.quitting = true
				return tea.Quit
			}
		case "q":
			m.quitting = true
			return tea.Quit
		}
	}
	return nil
}

func (m *Model) viewMenu() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("KaiPerf3-tui") + "\n")
	b.WriteString(subtitleStyle.Render("a friendly terminal frontend for iperf3") + "\n")
	b.WriteString(subtitleStyle.Render("this machine: "+iperf.LocalIPsSummary()) + "\n\n")

	for i, item := range menuItems {
		line := item.label
		if i == m.menuCursor {
			b.WriteString(selectedItemStyle.Render("▶ "+line) + "  " + subtitleStyle.Render(item.desc) + "\n")
		} else {
			b.WriteString(itemStyle.Render("  "+line) + "  " + subtitleStyle.Render(item.desc) + "\n")
		}
	}

	content := borderStyle.Render(b.String())
	help := helpStyle.Render("↑↓/jk move · enter select · q quit")
	return content + "\n" + help
}
