package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guptarohit/asciigraph"
	"github.com/kaihen/iperf3-tui/internal/iperf"
)

type runMode int

const (
	runClient runMode = iota
	runServer
)

type runningModel struct {
	mode     runMode
	process  *iperf.Process
	clientCfg iperf.ClientConfig
	serverCfg iperf.ServerConfig

	lines   []string
	samples []float64
	points  []iperf.SamplePoint
	sender  *iperf.Sample
	receiver *iperf.Sample
	latest  *iperf.Sample

	startedAt  time.Time
	finishedAt time.Time
	finished   bool
	aborted    bool
	recorded   bool
	err        error

	export exportState
}

const maxLogLines = 200
const maxChartPoints = 120

// processEventMsg wraps one iperf.Event as a bubbletea message. Events are
// consumed strictly one at a time (the next listenForEvent is only issued
// after the current one is fully processed), so lines and the final Done
// event are always seen in the order iperf3 produced them.
type processEventMsg iperf.Event

func listenForEvent(p *iperf.Process) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-p.Events
		if !ok {
			return nil
		}
		return processEventMsg(e)
	}
}

func (m *Model) startClientTest(cfg iperf.ClientConfig) tea.Cmd {
	proc, err := iperf.Start(cfg.Args()...)
	if err != nil {
		m.form.err = err.Error()
		return nil
	}
	m.running = &runningModel{
		mode:      runClient,
		process:   proc,
		clientCfg: cfg,
		startedAt: time.Now(),
	}
	m.screen = screenRunning
	return listenForEvent(proc)
}

func (m *Model) startServer(cfg iperf.ServerConfig) tea.Cmd {
	proc, err := iperf.Start(cfg.Args()...)
	if err != nil {
		m.err = err
		return nil
	}
	m.running = &runningModel{
		mode:      runServer,
		process:   proc,
		serverCfg: cfg,
		startedAt: time.Now(),
	}
	m.screen = screenRunning
	return listenForEvent(proc)
}

func (m *Model) updateRunning(msg tea.Msg) tea.Cmd {
	r := m.running
	if r == nil {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if r.export.active {
			if r.export.handleKey(msg.String()) {
				r.export.result = exportEntry(buildHistoryEntry(r), r.export.cursor)
			}
			return nil
		}
		switch msg.String() {
		case "e":
			if r.finished && r.mode == runClient {
				r.export = exportState{active: true}
			}
			return nil
		case "q", "esc":
			if !r.finished {
				r.process.Stop()
				r.aborted = true
				r.finishedAt = time.Now()
			}
			if r.mode == runClient {
				m.recordHistory()
				m.screen = screenForm
			} else {
				m.screen = screenMenu
			}
			m.running = nil
			return nil
		case "ctrl+c":
			if !r.finished {
				r.process.Stop()
			}
			return tea.Quit
		}

	case processEventMsg:
		if msg.Line != "" {
			r.lines = append(r.lines, msg.Line)
			if len(r.lines) > maxLogLines {
				r.lines = r.lines[len(r.lines)-maxLogLines:]
			}
			if sample, ok := iperf.ParseLine(msg.Line); ok {
				r.latest = &sample
				if sample.IsSummary {
					s := sample
					if s.IsSender {
						r.sender = &s
					} else if s.IsReceiver {
						r.receiver = &s
					}
				} else if sample.IsChartable(maxInt(r.clientCfg.Parallel, 1)) {
					r.samples = append(r.samples, sample.Mbps)
					if len(r.samples) > maxChartPoints {
						r.samples = r.samples[len(r.samples)-maxChartPoints:]
					}
					r.points = append(r.points, iperf.SamplePoint{
						StartSec: sample.StartSec,
						EndSec:   sample.EndSec,
						Mbps:     sample.Mbps,
					})
				}
			}
		}

		if msg.Done != nil {
			r.finished = true
			r.finishedAt = time.Now()
			if msg.Done.Err != nil && !r.aborted {
				r.err = msg.Done.Err
			}
			if r.mode == runClient {
				m.recordHistoryFromRun(r)
			}
			return nil
		}
		return listenForEvent(r.process)
	}

	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Model) recordHistory() {
	if m.running != nil {
		m.recordHistoryFromRun(m.running)
	}
}

// buildHistoryEntry snapshots a runningModel into a HistoryEntry, used both
// to persist a completed client test and to export it on demand.
func buildHistoryEntry(r *runningModel) iperf.HistoryEntry {
	t := r.finishedAt
	if t.IsZero() {
		t = time.Now()
	}
	entry := iperf.HistoryEntry{
		Time:    t,
		Config:  r.clientCfg.Summary(),
		Aborted: r.aborted,
		Samples: r.points,
		RawLog:  r.lines,
	}
	if r.sender != nil {
		entry.AvgMbpsUp = r.sender.Mbps
		entry.Retransmits = r.sender.Retransmits
		if r.sender.LossPct >= 0 {
			entry.LossPct = r.sender.LossPct
		}
	}
	if r.receiver != nil {
		entry.AvgMbpsDown = r.receiver.Mbps
	}
	return entry
}

func (m *Model) recordHistoryFromRun(r *runningModel) {
	if r == nil || r.mode != runClient || r.recorded {
		return
	}
	r.recorded = true
	_ = iperf.AppendHistory(buildHistoryEntry(r))
	m.historyDirty = true
}

func (m *Model) viewRunning() string {
	r := m.running
	if r == nil {
		return ""
	}
	var b strings.Builder

	if r.mode == runClient {
		b.WriteString(titleStyle.Render("iperf3 · running client test") + "\n")
		b.WriteString(subtitleStyle.Render(r.clientCfg.Summary()) + "\n\n")
	} else {
		b.WriteString(titleStyle.Render(fmt.Sprintf("iperf3 · server listening on :%d", r.serverCfg.Port)) + "\n")
		b.WriteString(subtitleStyle.Render("reachable at: "+iperf.LocalIPsSummary()) + "\n\n")
	}

	if len(r.samples) >= 2 {
		graph := asciigraph.Plot(r.samples,
			asciigraph.Height(10),
			asciigraph.Width(60),
			asciigraph.Caption("Mbps"),
		)
		b.WriteString(graph + "\n\n")
	}

	if r.latest != nil && !r.latest.IsSummary {
		b.WriteString(fmt.Sprintf("current: %s\n", goodStyle.Render(fmt.Sprintf("%.1f Mbps", r.latest.Mbps))))
	}

	if r.sender != nil {
		b.WriteString(fmt.Sprintf("sender:   %-14s", fmt.Sprintf("%.1f Mbps", r.sender.Mbps)))
		if r.sender.Retransmits >= 0 {
			b.WriteString(fmt.Sprintf("  retransmits: %d", r.sender.Retransmits))
		}
		b.WriteString("\n")
	}
	if r.receiver != nil {
		b.WriteString(fmt.Sprintf("receiver: %-14s\n", fmt.Sprintf("%.1f Mbps", r.receiver.Mbps)))
	}
	if r.latest != nil && r.latest.LossPct >= 0 {
		lossStyle := goodStyle
		if r.latest.LossPct > 1 {
			lossStyle = warnStyle
		}
		if r.latest.LossPct > 5 {
			lossStyle = badStyle
		}
		b.WriteString(fmt.Sprintf("loss: %s   jitter: %.3f ms\n", lossStyle.Render(fmt.Sprintf("%.2f%%", r.latest.LossPct)), r.latest.JitterMs))
	}

	if r.finished {
		if r.aborted {
			b.WriteString("\n" + warnStyle.Render("aborted by user") + "\n")
		} else if r.err != nil {
			b.WriteString("\n" + errorStyle.Render("✗ "+r.err.Error()) + "\n")
		} else {
			b.WriteString("\n" + goodStyle.Render("✓ test complete") + "\n")
		}
	}

	if r.export.active {
		b.WriteString(r.export.render())
	} else {
		b.WriteString("\n" + subtitleStyle.Render("recent log:") + "\n")
		tail := r.lines
		if len(tail) > 8 {
			tail = tail[len(tail)-8:]
		}
		for _, l := range tail {
			b.WriteString(subtitleStyle.Render(l) + "\n")
		}
	}

	content := borderStyle.Render(b.String())
	action := "q to stop and return"
	if r.finished {
		action = "q to return"
		if r.mode == runClient {
			action = "e export · " + action
		}
	}
	help := helpStyle.Render(action + " · ctrl+c quit")
	return content + "\n" + help
}
