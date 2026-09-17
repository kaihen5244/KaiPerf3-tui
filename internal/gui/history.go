package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

type historyTab struct {
	app     *App
	content fyne.CanvasObject

	list     *widget.List
	entries  []iperf.HistoryEntry
	selected int

	detailLabel  *widget.Label
	exportStatus *widget.Label
	exportBtns   []*widget.Button
}

func newHistoryTab(app *App) *historyTab {
	h := &historyTab{app: app, selected: -1}

	h.list = widget.NewList(
		func() int { return len(h.entries) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, o fyne.CanvasObject) {
			e := h.entries[id]
			status := "ok"
			if e.Aborted {
				status = "aborted"
			}
			o.(*widget.Label).SetText(fmt.Sprintf("%s  %s  up %.1f / down %.1f Mbps  (%s)",
				e.Time.Format("Jan 2 15:04"), e.Config, e.AvgMbpsUp, e.AvgMbpsDown, status))
		},
	)
	h.list.OnSelected = func(id widget.ListItemID) {
		h.selected = id
		h.updateDetail()
	}

	h.detailLabel = widget.NewLabel("Select a result to see details.")
	h.detailLabel.Wrapping = fyne.TextWrapWord

	btnTxt := widget.NewButton(".txt", func() { h.export(0) })
	btnCsv := widget.NewButton(".csv", func() { h.export(1) })
	btnPdf := widget.NewButton(".pdf", func() { h.export(2) })
	h.exportBtns = []*widget.Button{btnTxt, btnCsv, btnPdf}
	h.setExportEnabled(false)
	h.exportStatus = widget.NewLabel("")

	detail := container.NewVBox(
		h.detailLabel,
		container.NewHBox(widget.NewLabel("Export:"), btnTxt, btnCsv, btnPdf),
		h.exportStatus,
	)

	refreshBtn := widget.NewButton("Refresh", h.reload)

	split := container.NewHSplit(h.list, detail)
	split.Offset = 0.55

	h.content = container.NewBorder(refreshBtn, nil, nil, nil, split)
	h.reload()
	return h
}

func (h *historyTab) setExportEnabled(enabled bool) {
	for _, b := range h.exportBtns {
		if enabled {
			b.Enable()
		} else {
			b.Disable()
		}
	}
}

func (h *historyTab) reload() {
	entries, _ := iperf.LoadHistory()
	h.entries = entries
	if h.selected >= len(h.entries) {
		h.selected = -1
	}
	h.list.Refresh()
	h.updateDetail()
}

func (h *historyTab) updateDetail() {
	if h.selected < 0 || h.selected >= len(h.entries) {
		h.detailLabel.SetText("Select a result to see details.")
		h.setExportEnabled(false)
		return
	}
	e := h.entries[h.selected]

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n\n", e.Time.Format("2006-01-02 15:04:05"), e.Config)
	fmt.Fprintf(&b, "Upload:   %.1f Mbps\n", e.AvgMbpsUp)
	fmt.Fprintf(&b, "Download: %.1f Mbps\n", e.AvgMbpsDown)
	if e.Retransmits > 0 {
		fmt.Fprintf(&b, "Retransmits: %d\n", e.Retransmits)
	}
	if e.LossPct > 0 {
		fmt.Fprintf(&b, "Loss: %.2f%%\n", e.LossPct)
	}
	if e.Aborted {
		b.WriteString("\n(aborted by user)\n")
	}
	h.detailLabel.SetText(b.String())
	h.setExportEnabled(true)
	h.exportStatus.SetText("")
}

func (h *historyTab) export(format int) {
	if h.selected < 0 || h.selected >= len(h.entries) {
		return
	}
	h.exportStatus.SetText(exportEntry(h.entries[h.selected], format))
}
