package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaihen/iperf3-tui/internal/iperf"
)

type presetsModel struct {
	items      []iperf.Preset
	cursor     int
	loadIntoForm bool // true when entered from the form's "Load preset" action
	err        string
}

func (m *Model) gotoPresets(loadIntoForm bool) tea.Cmd {
	presets, err := iperf.LoadPresets()
	m.presets = presetsModel{items: presets, loadIntoForm: loadIntoForm}
	if err != nil {
		m.presets.err = err.Error()
	}
	m.screen = screenPresets
	return nil
}

func (m *Model) updatePresets(msg tea.Msg) tea.Cmd {
	p := &m.presets
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.cursor++
			}
		case "d":
			if len(p.items) == 0 {
				return nil
			}
			name := p.items[p.cursor].Name
			if err := iperf.DeletePreset(name); err != nil {
				p.err = err.Error()
				return nil
			}
			p.items, _ = iperf.LoadPresets()
			if p.cursor >= len(p.items) && p.cursor > 0 {
				p.cursor--
			}
		case "enter":
			if len(p.items) == 0 {
				return nil
			}
			cfg := p.items[p.cursor].Config
			if p.loadIntoForm {
				m.form.loadConfig(cfg)
				m.screen = screenForm
				return nil
			}
			m.form.loadConfig(cfg)
			m.screen = screenForm
		case "esc":
			if p.loadIntoForm {
				m.screen = screenForm
			} else {
				m.screen = screenMenu
			}
			return nil
		}
	}
	return nil
}

func (m *Model) viewPresets() string {
	p := &m.presets
	var b strings.Builder
	b.WriteString(titleStyle.Render("iperf3-tui · presets") + "\n")
	b.WriteString(subtitleStyle.Render("saved client configurations") + "\n\n")

	if len(p.items) == 0 {
		b.WriteString(subtitleStyle.Render("no presets saved yet — save one from the client test form") + "\n")
	}
	for i, preset := range p.items {
		line := preset.Name + "  " + subtitleStyle.Render(preset.Config.Summary())
		if i == p.cursor {
			b.WriteString(selectedItemStyle.Render("▶ ") + line + "\n")
		} else {
			b.WriteString(itemStyle.Render("  ") + line + "\n")
		}
	}

	if p.err != "" {
		b.WriteString("\n" + errorStyle.Render("✗ "+p.err) + "\n")
	}

	content := borderStyle.Render(b.String())
	help := helpStyle.Render("↑↓/jk move · enter load · d delete · esc back")
	return content + "\n" + help
}
