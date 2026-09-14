package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenMenu screen = iota
	screenForm
	screenServerConfig
	screenRunning
	screenHistory
	screenPresets
)

// Model is the root bubbletea model for iperf3-tui.
type Model struct {
	screen screen
	width  int
	height int

	menuCursor int

	form       formModel
	serverForm serverFormModel
	running    *runningModel
	history    historyModel
	presets    presetsModel

	historyDirty bool
	quitting     bool
	err          error
}

func NewModel() Model {
	return Model{
		screen:     screenMenu,
		form:       newFormModel(),
		serverForm: newServerFormModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.running != nil && !m.running.finished {
				m.running.process.Stop()
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenMenu:
		cmd = m.updateMenu(msg)
	case screenForm:
		cmd = m.updateForm(msg)
	case screenServerConfig:
		cmd = m.updateServerConfig(msg)
	case screenRunning:
		cmd = m.updateRunning(msg)
	case screenHistory:
		cmd = m.updateHistory(msg)
	case screenPresets:
		cmd = m.updatePresets(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case screenMenu:
		return m.viewMenu()
	case screenForm:
		return m.viewForm()
	case screenServerConfig:
		return m.viewServerConfig()
	case screenRunning:
		return m.viewRunning()
	case screenHistory:
		return m.viewHistory()
	case screenPresets:
		return m.viewPresets()
	}
	return ""
}
