package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

type serverRun struct {
	process        *iperf.Process
	samples        []float64
	finished       bool
	connected      bool
	lastClientAddr string
}

type serverTab struct {
	app     *App
	content fyne.CanvasObject

	portEntry  *widget.Entry
	startBtn   *widget.Button
	stopBtn    *widget.Button
	addrsLabel *widget.Label
	statusText *canvas.Text
	chart      *LineChart
	logBox     *widget.Entry

	run *serverRun
}

func newServerTab(app *App) *serverTab {
	s := &serverTab{app: app}

	s.portEntry = widget.NewEntry()
	s.portEntry.SetText("5201")
	form := widget.NewForm(widget.NewFormItem("Port", s.portEntry))

	s.startBtn = widget.NewButtonWithIcon("Start server", nil, s.start)
	s.startBtn.Importance = widget.HighImportance
	s.stopBtn = widget.NewButton("Stop", s.stop)
	s.stopBtn.Disable()
	buttons := container.NewHBox(s.startBtn, s.stopBtn)

	s.addrsLabel = widget.NewLabel(addressSummary())
	s.addrsLabel.Wrapping = fyne.TextWrapWord
	refreshBtn := widget.NewButton("Refresh addresses", func() {
		s.addrsLabel.SetText(addressSummary())
	})

	s.statusText = canvas.NewText("○ not running", theme.Color(theme.ColorNameDisabled))
	s.statusText.TextStyle = fyne.TextStyle{Bold: true}
	s.statusText.TextSize = 14

	s.chart = NewLineChart()
	s.logBox = widget.NewMultiLineEntry()
	s.logBox.Disable()
	logScroll := container.NewVScroll(s.logBox)
	logScroll.SetMinSize(fyne.NewSize(0, 220))

	s.content = container.NewVBox(
		form,
		buttons,
		widget.NewSeparator(),
		widget.NewLabel("This machine is reachable at:"),
		s.addrsLabel,
		refreshBtn,
		widget.NewSeparator(),
		s.statusText,
		s.chart,
		widget.NewLabel("Log:"),
		logScroll,
	)
	return s
}

// setStatus updates the bold status line and its color together, since a
// canvas.Text needs an explicit Refresh() after either changes.
func (s *serverTab) setStatus(text string, colorName fyne.ThemeColorName) {
	s.statusText.Text = text
	s.statusText.Color = theme.Color(colorName)
	s.statusText.Refresh()
}

func addressSummary() string {
	addrs, err := iperf.LocalIPs()
	if err != nil || len(addrs) == 0 {
		return "no network addresses found"
	}
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = fmt.Sprintf("%s (%s)", a.IP, a.Iface)
	}
	return strings.Join(parts, "\n")
}

func (s *serverTab) start() {
	port, err := strconv.Atoi(strings.TrimSpace(s.portEntry.Text))
	if err != nil || port <= 0 {
		dialog.ShowError(fmt.Errorf("port must be a positive number"), s.app.win)
		return
	}
	proc, err := iperf.Start(iperf.ServerConfig{Port: port}.Args()...)
	if err != nil {
		dialog.ShowError(err, s.app.win)
		return
	}
	s.run = &serverRun{process: proc}
	s.chart.SetSamples(nil)
	s.logBox.SetText("")
	s.startBtn.Disable()
	s.stopBtn.Enable()
	s.setStatus("● LISTENING — waiting for a connection...", theme.ColorNameWarning)

	go s.consume(proc)
}

func (s *serverTab) stop() {
	if s.run == nil || s.run.finished {
		return
	}
	s.run.process.Stop()
	s.stopBtn.Disable()
}

func (s *serverTab) consume(proc *iperf.Process) {
	var lines []string
	for e := range proc.Events {
		ev := e
		fyne.Do(func() {
			r := s.run
			if r == nil {
				return
			}
			if ev.Line != "" {
				lines = append(lines, ev.Line)
				if len(lines) > maxLogLines {
					lines = lines[len(lines)-maxLogLines:]
				}
				s.logBox.SetText(strings.Join(lines, "\n"))

				if addr, ok := iperf.ParseAcceptedConnection(ev.Line); ok {
					r.connected = true
					r.lastClientAddr = addr
					s.setStatus("● CONNECTED — receiving from "+addr, theme.ColorNameSuccess)
				} else if iperf.IsServerListeningLine(ev.Line) {
					r.connected = false
					status := "● LISTENING — waiting for a connection..."
					if r.lastClientAddr != "" {
						status += " (last: " + r.lastClientAddr + ")"
					}
					s.setStatus(status, theme.ColorNameWarning)
				}

				if sample, ok := iperf.ParseLine(ev.Line); ok && !sample.IsSummary {
					r.samples = append(r.samples, sample.Mbps)
					s.chart.SetSamples(r.samples)
				}
			}
			if ev.Done != nil {
				r.finished = true
				s.startBtn.Enable()
				s.stopBtn.Disable()
				s.setStatus("■ stopped", theme.ColorNameDisabled)
			}
		})
	}
}
