// Package gui is a Fyne-based graphical frontend for iperf3, sharing all of
// its test-running, parsing, history, preset, and export logic with the
// terminal frontend via the internal/iperf package.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

const (
	tabIndexClient = iota
	tabIndexServer
	tabIndexHistory
	tabIndexPresets
)

// App is the root of the GUI: one window with a tab per feature area.
type App struct {
	fyneApp fyne.App
	win     fyne.Window
	tabs    *container.AppTabs

	client  *clientTab
	server  *serverTab
	history *historyTab
	presets *presetsTab
}

// NewApp builds the window and all tabs, wiring cross-tab actions (loading a
// preset into the client form, refreshing history after a test finishes).
func NewApp() *App {
	a := &App{fyneApp: app.NewWithID("com.kaihen.kaiperf3-tui")}
	a.win = a.fyneApp.NewWindow("KaiPerf3-tui")

	a.client = newClientTab(a)
	a.server = newServerTab(a)
	a.history = newHistoryTab(a)
	a.presets = newPresetsTab(a)

	a.tabs = container.NewAppTabs(
		container.NewTabItem("Client Test", a.client.content),
		container.NewTabItem("Server", a.server.content),
		container.NewTabItem("History", a.history.content),
		container.NewTabItem("Presets", a.presets.content),
	)
	a.tabs.OnSelected = func(item *container.TabItem) {
		if a.tabs.SelectedIndex() == tabIndexHistory {
			a.history.reload()
		}
		if a.tabs.SelectedIndex() == tabIndexPresets {
			a.presets.reload()
		}
	}

	a.win.SetContent(a.tabs)
	a.win.Resize(fyne.NewSize(720, 640))
	return a
}

// Run shows the window and blocks until it's closed.
func (a *App) Run() {
	a.win.ShowAndRun()
}
