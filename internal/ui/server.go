package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaihen/iperf3-tui/internal/iperf"
)

type serverFormModel struct {
	port textinput.Model
	err  string
}

func newServerFormModel() serverFormModel {
	ti := textinput.New()
	ti.SetValue("5201")
	ti.CharLimit = 5
	ti.Prompt = ""
	ti.Focus()
	return serverFormModel{port: ti}
}

func (m *Model) updateServerConfig(msg tea.Msg) tea.Cmd {
	s := &m.serverForm
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMenu
			return nil
		case "enter":
			port, err := strconv.Atoi(strings.TrimSpace(s.port.Value()))
			if err != nil || port <= 0 {
				s.err = "port must be a positive number"
				return nil
			}
			s.err = ""
			return m.startServer(iperf.ServerConfig{Port: port})
		}
	}
	var cmd tea.Cmd
	s.port, cmd = s.port.Update(msg)
	return cmd
}

func (m *Model) viewServerConfig() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("iperf3 · server") + "\n")
	b.WriteString(subtitleStyle.Render("start a local iperf3 server for others to test against") + "\n\n")
	b.WriteString(labelStyle.Render("Port") + " " + focusedValueStyle.Render(m.serverForm.port.View()) + "\n\n")

	b.WriteString(subtitleStyle.Render("give one of these addresses to whoever is running the client:") + "\n")
	addrs, err := iperf.LocalIPs()
	if err != nil || len(addrs) == 0 {
		b.WriteString(subtitleStyle.Render("  no network addresses found") + "\n")
	} else {
		for _, a := range addrs {
			b.WriteString("  " + goodStyle.Render(a.IP) + subtitleStyle.Render("  ("+a.Iface+")") + "\n")
		}
	}

	if m.serverForm.err != "" {
		b.WriteString("\n" + errorStyle.Render("✗ "+m.serverForm.err) + "\n")
	}
	content := borderStyle.Render(b.String())
	help := helpStyle.Render("enter to start · esc back · ctrl+c quit")
	return content + "\n" + help
}
