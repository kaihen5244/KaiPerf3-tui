package gui

import "github.com/kaihen/kaiperf3-tui/internal/iperf"

var exportExts = []string{"txt", "csv", "pdf"}

// exportEntry writes entry to disk in the given format (0=txt, 1=csv, 2=pdf)
// and returns a short status string describing the result.
func exportEntry(entry iperf.HistoryEntry, format int) string {
	ext := exportExts[format]
	path, err := iperf.DefaultExportPath(entry, ext)
	if err != nil {
		return "error: " + err.Error()
	}

	var werr error
	switch format {
	case 0:
		werr = iperf.ExportTXT(entry, path)
	case 1:
		werr = iperf.ExportCSV(entry, path)
	case 2:
		werr = iperf.ExportPDF(entry, path)
	}
	if werr != nil {
		return "error: " + werr.Error()
	}
	return "saved to " + path
}
