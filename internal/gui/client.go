package gui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

const maxLogLines = 200

// clientRun tracks the state of one in-progress or finished client test.
type clientRun struct {
	cfg        iperf.ClientConfig
	process    *iperf.Process
	lines      []string
	samples    []float64
	points     []iperf.SamplePoint
	sender     *iperf.Sample
	receiver   *iperf.Sample
	latest     *iperf.Sample
	finished   bool
	aborted    bool
	recorded   bool
	finishedAt time.Time
	err        error
}

type clientTab struct {
	app     *App
	content fyne.CanvasObject

	serverEntry    *widget.Entry
	portEntry      *widget.Entry
	durationEntry  *widget.Entry
	parallelEntry  *widget.Entry
	protocolSelect *widget.Select
	reverseCheck   *widget.Check
	bandwidthEntry *widget.Entry
	intervalEntry  *widget.Entry

	startBtn *widget.Button
	stopBtn  *widget.Button

	chart        *LineChart
	statusLabel  *widget.Label
	summaryLabel *widget.Label
	logBox       *widget.Entry
	exportTxtBtn *widget.Button
	exportCsvBtn *widget.Button
	exportPdfBtn *widget.Button
	exportStatus *widget.Label

	run *clientRun
}

func newClientTab(app *App) *clientTab {
	c := &clientTab{app: app}

	c.serverEntry = widget.NewEntry()
	c.serverEntry.SetPlaceHolder("e.g. 10.0.0.5 or iperf.example.com")
	c.portEntry = widget.NewEntry()
	c.portEntry.SetText("5201")
	c.durationEntry = widget.NewEntry()
	c.durationEntry.SetText("10")
	c.parallelEntry = widget.NewEntry()
	c.parallelEntry.SetText("1")
	c.protocolSelect = widget.NewSelect([]string{"TCP", "UDP"}, nil)
	c.protocolSelect.SetSelected("TCP")
	c.reverseCheck = widget.NewCheck("reverse: server sends, client receives (-R)", nil)
	c.bandwidthEntry = widget.NewEntry()
	c.bandwidthEntry.SetText("0")
	c.bandwidthEntry.SetPlaceHolder("UDP only; 0 = unlimited")
	c.intervalEntry = widget.NewEntry()
	c.intervalEntry.SetText("1")

	form := widget.NewForm(
		widget.NewFormItem("Server", c.serverEntry),
		widget.NewFormItem("Port", c.portEntry),
		widget.NewFormItem("Duration (s)", c.durationEntry),
		widget.NewFormItem("Streams", c.parallelEntry),
		widget.NewFormItem("Protocol", c.protocolSelect),
		widget.NewFormItem("", c.reverseCheck),
		widget.NewFormItem("Target rate (Mbps)", c.bandwidthEntry),
		widget.NewFormItem("Report every (s)", c.intervalEntry),
	)

	c.startBtn = widget.NewButtonWithIcon("Start test", nil, c.start)
	c.startBtn.Importance = widget.HighImportance
	c.stopBtn = widget.NewButton("Stop", c.stop)
	c.stopBtn.Disable()
	saveBtn := widget.NewButton("Save as preset", c.showSavePreset)
	loadBtn := widget.NewButton("Load preset", func() { app.tabs.SelectIndex(tabIndexPresets) })
	buttons := container.NewHBox(c.startBtn, c.stopBtn, saveBtn, loadBtn)

	c.chart = NewLineChart()
	c.statusLabel = widget.NewLabel("")
	c.summaryLabel = widget.NewLabel("")
	c.summaryLabel.Wrapping = fyne.TextWrapWord

	c.logBox = widget.NewMultiLineEntry()
	c.logBox.Disable()
	logScroll := container.NewVScroll(c.logBox)
	logScroll.SetMinSize(fyne.NewSize(0, 160))

	c.exportTxtBtn = widget.NewButton("Export .txt", func() { c.export(0) })
	c.exportCsvBtn = widget.NewButton("Export .csv", func() { c.export(1) })
	c.exportPdfBtn = widget.NewButton("Export .pdf", func() { c.export(2) })
	c.setExportEnabled(false)
	c.exportStatus = widget.NewLabel("")
	exportRow := container.NewHBox(c.exportTxtBtn, c.exportCsvBtn, c.exportPdfBtn, c.exportStatus)

	results := container.NewVBox(
		c.chart,
		c.statusLabel,
		c.summaryLabel,
		exportRow,
		widget.NewLabel("Log:"),
		logScroll,
	)

	c.content = container.NewVBox(
		form,
		buttons,
		widget.NewSeparator(),
		results,
	)
	return c
}

func (c *clientTab) setExportEnabled(enabled bool) {
	for _, b := range []*widget.Button{c.exportTxtBtn, c.exportCsvBtn, c.exportPdfBtn} {
		if enabled {
			b.Enable()
		} else {
			b.Disable()
		}
	}
}

func (c *clientTab) config() (iperf.ClientConfig, error) {
	server := strings.TrimSpace(c.serverEntry.Text)
	if server == "" {
		return iperf.ClientConfig{}, fmt.Errorf("server address is required")
	}
	port, err := strconv.Atoi(strings.TrimSpace(c.portEntry.Text))
	if err != nil || port <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("port must be a positive number")
	}
	duration, err := strconv.Atoi(strings.TrimSpace(c.durationEntry.Text))
	if err != nil || duration <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("duration must be a positive number of seconds")
	}
	parallel, err := strconv.Atoi(strings.TrimSpace(c.parallelEntry.Text))
	if err != nil || parallel <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("streams must be a positive number")
	}
	interval, err := strconv.ParseFloat(strings.TrimSpace(c.intervalEntry.Text), 64)
	if err != nil || interval <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("report interval must be a positive number of seconds")
	}
	return iperf.ClientConfig{
		Server:    server,
		Port:      port,
		Duration:  duration,
		Parallel:  parallel,
		Reverse:   c.reverseCheck.Checked,
		UDP:       c.protocolSelect.Selected == "UDP",
		Bandwidth: strings.TrimSpace(c.bandwidthEntry.Text),
		Interval:  interval,
	}, nil
}

// loadConfig populates the form from a saved preset.
func (c *clientTab) loadConfig(cfg iperf.ClientConfig) {
	c.serverEntry.SetText(cfg.Server)
	c.portEntry.SetText(strconv.Itoa(cfg.Port))
	c.durationEntry.SetText(strconv.Itoa(cfg.Duration))
	c.parallelEntry.SetText(strconv.Itoa(cfg.Parallel))
	if cfg.UDP {
		c.protocolSelect.SetSelected("UDP")
	} else {
		c.protocolSelect.SetSelected("TCP")
	}
	c.reverseCheck.SetChecked(cfg.Reverse)
	c.bandwidthEntry.SetText(cfg.Bandwidth)
	c.intervalEntry.SetText(strconv.FormatFloat(cfg.Interval, 'f', -1, 64))
}

func (c *clientTab) showSavePreset() {
	name := widget.NewEntry()
	name.SetPlaceHolder("preset name")
	d := dialog.NewCustomConfirm("Save preset", "Save", "Cancel", name, func(ok bool) {
		if !ok {
			return
		}
		trimmed := strings.TrimSpace(name.Text)
		if trimmed == "" {
			return
		}
		cfg, err := c.config()
		if err != nil {
			dialog.ShowError(err, c.app.win)
			return
		}
		if err := iperf.UpsertPreset(iperf.Preset{Name: trimmed, Config: cfg}); err != nil {
			dialog.ShowError(err, c.app.win)
			return
		}
	}, c.app.win)
	d.Show()
}

func (c *clientTab) start() {
	cfg, err := c.config()
	if err != nil {
		dialog.ShowError(err, c.app.win)
		return
	}
	proc, err := iperf.Start(cfg.Args()...)
	if err != nil {
		dialog.ShowError(err, c.app.win)
		return
	}

	c.run = &clientRun{cfg: cfg, process: proc}
	c.chart.SetSamples(nil)
	c.logBox.SetText("")
	c.statusLabel.SetText("running: " + cfg.Summary())
	c.summaryLabel.SetText("")
	c.exportStatus.SetText("")
	c.setExportEnabled(false)
	c.startBtn.Disable()
	c.stopBtn.Enable()

	go c.consume(proc)
}

func (c *clientTab) stop() {
	if c.run == nil || c.run.finished {
		return
	}
	c.run.aborted = true
	c.run.process.Stop()
	c.stopBtn.Disable()
}

// consume drains proc.Events on a background goroutine; each event is
// applied to UI state via fyne.Do so widget mutation stays on the UI thread.
func (c *clientTab) consume(proc *iperf.Process) {
	for e := range proc.Events {
		ev := e
		fyne.Do(func() { c.handleEvent(ev) })
	}
}

func (c *clientTab) handleEvent(e iperf.Event) {
	r := c.run
	if r == nil {
		return
	}

	if e.Line != "" {
		r.lines = append(r.lines, e.Line)
		if len(r.lines) > maxLogLines {
			r.lines = r.lines[len(r.lines)-maxLogLines:]
		}
		c.logBox.SetText(strings.Join(r.lines, "\n"))

		if sample, ok := iperf.ParseLine(e.Line); ok {
			r.latest = &sample
			if sample.IsSummary {
				s := sample
				if s.IsSender {
					r.sender = &s
				} else if s.IsReceiver {
					r.receiver = &s
				}
			} else if sample.IsChartable(maxInt(r.cfg.Parallel, 1)) {
				r.samples = append(r.samples, sample.Mbps)
				r.points = append(r.points, iperf.SamplePoint{
					StartSec: sample.StartSec,
					EndSec:   sample.EndSec,
					Mbps:     sample.Mbps,
				})
				c.chart.SetSamples(r.samples)
			}
		}
		c.updateSummary()
	}

	if e.Done != nil {
		r.finished = true
		r.finishedAt = time.Now()
		if e.Done.Err != nil && !r.aborted {
			r.err = e.Done.Err
		}
		c.startBtn.Enable()
		c.stopBtn.Disable()
		c.recordHistory()
		c.setExportEnabled(true)
		c.updateSummary()

		switch {
		case r.aborted:
			c.statusLabel.SetText("aborted by user: " + r.cfg.Summary())
		case r.err != nil:
			c.statusLabel.SetText("failed: " + r.err.Error())
		default:
			c.statusLabel.SetText("complete: " + r.cfg.Summary())
		}
		c.app.history.reload()
	}
}

func (c *clientTab) updateSummary() {
	r := c.run
	if r == nil {
		return
	}
	var b strings.Builder
	if r.sender != nil {
		fmt.Fprintf(&b, "Upload: %.1f Mbps", r.sender.Mbps)
		if r.sender.Retransmits > 0 {
			fmt.Fprintf(&b, "  (retransmits: %d)", r.sender.Retransmits)
		}
		b.WriteString("\n")
	}
	if r.receiver != nil {
		fmt.Fprintf(&b, "Download: %.1f Mbps\n", r.receiver.Mbps)
	}
	if r.latest != nil && r.latest.LossPct >= 0 {
		fmt.Fprintf(&b, "Loss: %.2f%%   Jitter: %.3f ms\n", r.latest.LossPct, r.latest.JitterMs)
	}
	c.summaryLabel.SetText(strings.TrimRight(b.String(), "\n"))
}

func (c *clientTab) recordHistory() {
	r := c.run
	if r == nil || r.recorded {
		return
	}
	r.recorded = true
	entry := c.historyEntry()
	_ = iperf.AppendHistory(entry)
}

func (c *clientTab) historyEntry() iperf.HistoryEntry {
	r := c.run
	return iperf.BuildHistoryEntry(iperf.ClientRunResult{
		Config:     r.cfg,
		Sender:     r.sender,
		Receiver:   r.receiver,
		Points:     r.points,
		RawLog:     r.lines,
		Aborted:    r.aborted,
		FinishedAt: r.finishedAt,
	})
}

func (c *clientTab) export(format int) {
	if c.run == nil || !c.run.finished {
		return
	}
	c.exportStatus.SetText(exportEntry(c.historyEntry(), format))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
