package ui

import (
	"strings"

	"github.com/kaihen/iperf3-tui/internal/iperf"
)

var exportFormats = []string{"Text report (.txt)", "CSV samples (.csv)", "PDF report (.pdf)"}
var exportExts = []string{"txt", "csv", "pdf"}

// exportState is embedded wherever a completed test result can be exported
// (the running screen right after a test finishes, and the history list).
type exportState struct {
	active bool
	cursor int
	result string
}

// handleKey processes a keypress while the export panel is open. It returns
// true when the user confirmed a format (Enter) so the caller should write
// the file for exportExts[e.cursor].
func (e *exportState) handleKey(key string) (confirm bool) {
	switch key {
	case "up", "k":
		if e.cursor > 0 {
			e.cursor--
		}
	case "down", "j":
		if e.cursor < len(exportFormats)-1 {
			e.cursor++
		}
	case "enter":
		return true
	case "esc":
		e.active = false
		e.result = ""
	}
	return false
}

func (e *exportState) render() string {
	var b strings.Builder
	b.WriteString("\n" + subtitleStyle.Render("export as:") + "\n")
	for i, f := range exportFormats {
		if i == e.cursor {
			b.WriteString(selectedItemStyle.Render("▶ "+f) + "\n")
		} else {
			b.WriteString(itemStyle.Render("  "+f) + "\n")
		}
	}
	b.WriteString(subtitleStyle.Render("enter to save · esc to cancel") + "\n")
	if e.result != "" {
		b.WriteString("\n" + e.result + "\n")
	}
	return b.String()
}

// exportEntry writes entry to disk in the format selected by cursor and
// returns a styled status line describing the result.
func exportEntry(entry iperf.HistoryEntry, cursor int) string {
	ext := exportExts[cursor]
	path, err := iperf.DefaultExportPath(entry, ext)
	if err != nil {
		return errorStyle.Render("✗ " + err.Error())
	}

	var werr error
	switch cursor {
	case 0:
		werr = iperf.ExportTXT(entry, path)
	case 1:
		werr = iperf.ExportCSV(entry, path)
	case 2:
		werr = iperf.ExportPDF(entry, path)
	}
	if werr != nil {
		return errorStyle.Render("✗ " + werr.Error())
	}
	return goodStyle.Render("✓ saved to " + path)
}
