package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

type presetsTab struct {
	app     *App
	content fyne.CanvasObject

	list     *widget.List
	presets  []iperf.Preset
	selected int

	detailLabel *widget.Label
	loadBtn     *widget.Button
	deleteBtn   *widget.Button
}

func newPresetsTab(app *App) *presetsTab {
	p := &presetsTab{app: app, selected: -1}

	p.list = widget.NewList(
		func() int { return len(p.presets) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, o fyne.CanvasObject) {
			pr := p.presets[id]
			o.(*widget.Label).SetText(pr.Name + "  —  " + pr.Config.Summary())
		},
	)
	p.list.OnSelected = func(id widget.ListItemID) {
		p.selected = id
		p.updateButtons()
	}

	p.detailLabel = widget.NewLabel("Select a preset to load or delete it.")
	p.detailLabel.Wrapping = fyne.TextWrapWord

	p.loadBtn = widget.NewButtonWithIcon("Load into Client tab", nil, p.load)
	p.loadBtn.Importance = widget.HighImportance
	p.deleteBtn = widget.NewButton("Delete", p.delete)
	p.updateButtons()

	detail := container.NewVBox(p.detailLabel, container.NewHBox(p.loadBtn, p.deleteBtn))
	refreshBtn := widget.NewButton("Refresh", p.reload)

	split := container.NewHSplit(p.list, detail)
	split.Offset = 0.55

	p.content = container.NewBorder(refreshBtn, nil, nil, nil, split)
	p.reload()
	return p
}

func (p *presetsTab) updateButtons() {
	enabled := p.selected >= 0 && p.selected < len(p.presets)
	if enabled {
		p.loadBtn.Enable()
		p.deleteBtn.Enable()
		p.detailLabel.SetText(p.presets[p.selected].Config.Summary())
	} else {
		p.loadBtn.Disable()
		p.deleteBtn.Disable()
		p.detailLabel.SetText("Select a preset to load or delete it.")
	}
}

func (p *presetsTab) reload() {
	presets, _ := iperf.LoadPresets()
	p.presets = presets
	if p.selected >= len(p.presets) {
		p.selected = -1
	}
	p.list.Refresh()
	p.updateButtons()
}

func (p *presetsTab) load() {
	if p.selected < 0 || p.selected >= len(p.presets) {
		return
	}
	p.app.client.loadConfig(p.presets[p.selected].Config)
	p.app.tabs.SelectIndex(tabIndexClient)
}

func (p *presetsTab) delete() {
	if p.selected < 0 || p.selected >= len(p.presets) {
		return
	}
	name := p.presets[p.selected].Name
	if err := iperf.DeletePreset(name); err != nil {
		dialog.ShowError(err, p.app.win)
		return
	}
	p.selected = -1
	p.reload()
}
