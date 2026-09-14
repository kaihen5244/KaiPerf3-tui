package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

// formField indexes the focusable rows of the client config form.
type formField int

const (
	fieldServer formField = iota
	fieldPort
	fieldDuration
	fieldParallel
	fieldProtocol
	fieldReverse
	fieldBandwidth
	fieldInterval
	fieldCount
)

const (
	actionStart formField = iota + 100
	actionSavePreset
	actionLoadPreset
	actionBack
)

type formModel struct {
	inputs   map[formField]textinput.Model
	focus    formField
	udp      bool
	reverse  bool
	saveName textinput.Model
	saving   bool
	err      string
}

func newFormModel() formModel {
	mk := func(placeholder, value string, charLimit int) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.SetValue(value)
		ti.CharLimit = charLimit
		ti.Prompt = ""
		return ti
	}

	inputs := map[formField]textinput.Model{
		fieldServer:    mk("e.g. 10.0.0.5 or iperf.example.com", "", 64),
		fieldPort:      mk("5201", "5201", 5),
		fieldDuration:  mk("10", "10", 4),
		fieldParallel:  mk("1", "1", 3),
		fieldBandwidth: mk("0 (unlimited)", "0", 12),
		fieldInterval:  mk("1", "1", 4),
	}

	saveName := textinput.New()
	saveName.Placeholder = "preset name"
	saveName.CharLimit = 32
	saveName.Prompt = ""

	f := formModel{inputs: inputs, focus: fieldServer, saveName: saveName}
	f.inputs[fieldServer] = focusInput(f.inputs[fieldServer])
	return f
}

func focusInput(ti textinput.Model) textinput.Model {
	ti.Focus()
	return ti
}

func (f *formModel) config() (iperf.ClientConfig, error) {
	server := strings.TrimSpace(f.inputs[fieldServer].Value())
	if server == "" {
		return iperf.ClientConfig{}, fmt.Errorf("server address is required")
	}
	port, err := strconv.Atoi(strings.TrimSpace(f.inputs[fieldPort].Value()))
	if err != nil || port <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("port must be a positive number")
	}
	duration, err := strconv.Atoi(strings.TrimSpace(f.inputs[fieldDuration].Value()))
	if err != nil || duration <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("duration must be a positive number of seconds")
	}
	parallel, err := strconv.Atoi(strings.TrimSpace(f.inputs[fieldParallel].Value()))
	if err != nil || parallel <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("parallel streams must be a positive number")
	}
	interval, err := strconv.ParseFloat(strings.TrimSpace(f.inputs[fieldInterval].Value()), 64)
	if err != nil || interval <= 0 {
		return iperf.ClientConfig{}, fmt.Errorf("interval must be a positive number of seconds")
	}

	return iperf.ClientConfig{
		Server:    server,
		Port:      port,
		Duration:  duration,
		Parallel:  parallel,
		Reverse:   f.reverse,
		UDP:       f.udp,
		Bandwidth: strings.TrimSpace(f.inputs[fieldBandwidth].Value()),
		Interval:  interval,
	}, nil
}

func (f *formModel) loadConfig(c iperf.ClientConfig) {
	set := func(field formField, value string) {
		ti := f.inputs[field]
		ti.SetValue(value)
		f.inputs[field] = ti
	}
	set(fieldServer, c.Server)
	set(fieldPort, strconv.Itoa(c.Port))
	set(fieldDuration, strconv.Itoa(c.Duration))
	set(fieldParallel, strconv.Itoa(c.Parallel))
	set(fieldBandwidth, c.Bandwidth)
	set(fieldInterval, strconv.FormatFloat(c.Interval, 'f', -1, 64))
	f.udp = c.UDP
	f.reverse = c.Reverse
}

func (f *formModel) fieldOrder() []formField {
	order := []formField{fieldServer, fieldPort, fieldDuration, fieldParallel, fieldProtocol, fieldReverse}
	if f.udp {
		order = append(order, fieldBandwidth)
	}
	order = append(order, fieldInterval, actionStart, actionSavePreset, actionLoadPreset, actionBack)
	return order
}

func (f *formModel) setFocus(target formField) {
	if ti, ok := f.inputs[f.focus]; ok {
		ti.Blur()
		f.inputs[f.focus] = ti
	}
	f.focus = target
	if ti, ok := f.inputs[target]; ok {
		ti.Focus()
		f.inputs[target] = ti
	}
}

func (f *formModel) moveFocus(delta int) {
	order := f.fieldOrder()
	idx := 0
	for i, ff := range order {
		if ff == f.focus {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(order)) % len(order)
	f.setFocus(order[idx])
}

// updateForm handles key events while the client config form is active.
// Returns (cmd, action) where action is non-empty when the user triggered
// start/save/load/back.
func (m *Model) updateForm(msg tea.Msg) tea.Cmd {
	f := &m.form

	if f.saving {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				f.saving = false
				f.saveName.Blur()
				return nil
			case "enter":
				name := strings.TrimSpace(f.saveName.Value())
				if name == "" {
					f.err = "preset name cannot be empty"
					return nil
				}
				cfg, err := f.config()
				if err != nil {
					f.err = err.Error()
					f.saving = false
					return nil
				}
				if err := iperf.UpsertPreset(iperf.Preset{Name: name, Config: cfg}); err != nil {
					f.err = err.Error()
				} else {
					f.err = ""
				}
				f.saving = false
				f.saveName.Blur()
				f.saveName.SetValue("")
				return nil
			}
		}
		var cmd tea.Cmd
		f.saveName, cmd = f.saveName.Update(msg)
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.moveFocus(1)
			return nil
		case "shift+tab", "up":
			f.moveFocus(-1)
			return nil
		case "left", "right", " ":
			switch f.focus {
			case fieldProtocol:
				f.udp = !f.udp
				return nil
			case fieldReverse:
				f.reverse = !f.reverse
				return nil
			}
		case "enter":
			switch f.focus {
			case actionStart:
				cfg, err := f.config()
				if err != nil {
					f.err = err.Error()
					return nil
				}
				f.err = ""
				return m.startClientTest(cfg)
			case actionSavePreset:
				f.saving = true
				f.saveName.Focus()
				return nil
			case actionLoadPreset:
				return m.gotoPresets(true)
			case actionBack:
				m.screen = screenMenu
				return nil
			}
		case "esc":
			m.screen = screenMenu
			return nil
		}
	}

	if ti, ok := f.inputs[f.focus]; ok {
		var cmd tea.Cmd
		ti, cmd = ti.Update(msg)
		f.inputs[f.focus] = ti
		return cmd
	}
	return nil
}

func (m *Model) viewForm() string {
	f := &m.form
	var b strings.Builder

	b.WriteString(titleStyle.Render("KaiPerf3-tui · client test") + "\n")
	b.WriteString(subtitleStyle.Render("Configure a test and run it against a remote iperf3 server") + "\n\n")

	row := func(field formField, label, value string) {
		style := labelStyle
		vstyle := valueStyle
		if f.focus == field {
			style = style.Foreground(colorAccent)
			vstyle = focusedValueStyle
		}
		b.WriteString(style.Render(label) + " " + vstyle.Render(value) + "\n")
	}

	row(fieldServer, "Server", f.inputs[fieldServer].View())
	row(fieldPort, "Port", f.inputs[fieldPort].View())
	row(fieldDuration, "Duration (s)", f.inputs[fieldDuration].View())
	row(fieldParallel, "Streams", f.inputs[fieldParallel].View())

	proto := "TCP"
	if f.udp {
		proto = "UDP"
	}
	row(fieldProtocol, "Protocol", proto+"  (←/→ to toggle)")

	rev := "off"
	if f.reverse {
		rev = "on  (server → client)"
	}
	row(fieldReverse, "Reverse (-R)", rev)

	if f.udp {
		row(fieldBandwidth, "Target rate", f.inputs[fieldBandwidth].View()+" Mbps (0=unlimited)")
	}
	row(fieldInterval, "Report every", f.inputs[fieldInterval].View()+"s")

	b.WriteString("\n")

	actionLine := func(field formField, label string) string {
		if f.focus == field {
			return selectedItemStyle.Render("▶ " + label)
		}
		return itemStyle.Render("  " + label)
	}
	b.WriteString(actionLine(actionStart, "Start test") + "\n")
	b.WriteString(actionLine(actionSavePreset, "Save as preset") + "\n")
	b.WriteString(actionLine(actionLoadPreset, "Load preset") + "\n")
	b.WriteString(actionLine(actionBack, "Back to menu") + "\n")

	if f.saving {
		b.WriteString("\n" + labelStyle.Render("Preset name") + " " + focusedValueStyle.Render(f.saveName.View()) + "\n")
		b.WriteString(subtitleStyle.Render("enter to save · esc to cancel") + "\n")
	}

	if f.err != "" {
		b.WriteString("\n" + errorStyle.Render("✗ "+f.err) + "\n")
	}

	content := borderStyle.Render(b.String())
	help := helpStyle.Render("tab/↑↓ move · ←→ toggle · enter select · esc back · ctrl+c quit")
	return content + "\n" + help
}
