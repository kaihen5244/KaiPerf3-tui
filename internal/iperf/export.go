package iperf

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

// pdfSafe replaces characters outside fpdf's core-font (Latin-1-ish) encoding
// with plain-ASCII equivalents. Without this, e.g. the "→"/"←" arrows in a
// config summary render as mojibake since the core Helvetica/Courier fonts
// aren't UTF-8.
var pdfSafeReplacer = strings.NewReplacer("→", "->", "←", "<-")

func pdfSafe(s string) string {
	return pdfSafeReplacer.Replace(s)
}

// ExportDir returns (creating if needed) the directory reports are saved to:
// a plain, visible folder in the user's home directory, easy to find in a
// file browser without hunting through dotfiles.
func ExportDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "iperf3-tui-exports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// DefaultExportPath builds a timestamped, collision-resistant path for
// exporting the given entry in the given format ("txt", "csv", or "pdf").
func DefaultExportPath(e HistoryEntry, ext string) (string, error) {
	dir, err := ExportDir()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("iperf3-report-%s.%s", e.Time.Format("20060102-150405"), ext)
	return filepath.Join(dir, name), nil
}

// ExportTXT writes a plain-text summary of the test, including every
// interval sample and the raw iperf3 output, if available.
func ExportTXT(e HistoryEntry, path string) error {
	var b strings.Builder

	fmt.Fprintf(&b, "iperf3 Test Report\n")
	fmt.Fprintf(&b, "===================\n\n")
	fmt.Fprintf(&b, "Time:   %s\n", e.Time.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "Config: %s\n", e.Config)
	if e.Aborted {
		fmt.Fprintf(&b, "Status: aborted by user\n")
	} else {
		fmt.Fprintf(&b, "Status: complete\n")
	}

	fmt.Fprintf(&b, "\nSummary\n-------\n")
	fmt.Fprintf(&b, "Upload (sender):     %.2f Mbps\n", e.AvgMbpsUp)
	fmt.Fprintf(&b, "Download (receiver): %.2f Mbps\n", e.AvgMbpsDown)
	if e.Retransmits > 0 {
		fmt.Fprintf(&b, "Retransmits:         %d\n", e.Retransmits)
	}
	if e.LossPct > 0 {
		fmt.Fprintf(&b, "Packet loss:         %.2f%%\n", e.LossPct)
	}

	if len(e.Samples) > 0 {
		fmt.Fprintf(&b, "\nInterval samples\n----------------\n")
		for _, s := range e.Samples {
			fmt.Fprintf(&b, "%6.2f-%6.2f sec   %9.2f Mbps\n", s.StartSec, s.EndSec, s.Mbps)
		}
	}

	if len(e.RawLog) > 0 {
		fmt.Fprintf(&b, "\nRaw iperf3 output\n-----------------\n")
		for _, l := range e.RawLog {
			b.WriteString(l)
			b.WriteString("\n")
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// ExportCSV writes the per-interval throughput samples as CSV, for
// spreadsheets or plotting elsewhere.
func ExportCSV(e HistoryEntry, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"start_sec", "end_sec", "mbps"}); err != nil {
		return err
	}
	for _, s := range e.Samples {
		row := []string{
			strconv.FormatFloat(s.StartSec, 'f', 2, 64),
			strconv.FormatFloat(s.EndSec, 'f', 2, 64),
			strconv.FormatFloat(s.Mbps, 'f', 3, 64),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// ExportPDF writes a formatted PDF report: a summary section, a throughput
// chart (if samples are available), and the raw iperf3 output.
func ExportPDF(e HistoryEntry, path string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, "iperf3 Test Report", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 6, e.Time.Format("Jan 2, 2006 15:04:05"), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, pdfSafe(e.Config), "", 1, "L", false, 0, "")
	status := "complete"
	if e.Aborted {
		status = "aborted by user"
	}
	pdf.CellFormat(0, 6, "Status: "+status, "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 8, "Summary", "", 1, "L", false, 0, "")

	row := func(label, value string) {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(55, 7, label, "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(0, 7, value, "", 1, "L", false, 0, "")
	}
	row("Upload (sender):", fmt.Sprintf("%.1f Mbps", e.AvgMbpsUp))
	row("Download (receiver):", fmt.Sprintf("%.1f Mbps", e.AvgMbpsDown))
	if e.Retransmits > 0 {
		row("Retransmits:", fmt.Sprintf("%d", e.Retransmits))
	}
	if e.LossPct > 0 {
		row("Packet loss:", fmt.Sprintf("%.2f%%", e.LossPct))
	}
	pdf.Ln(6)

	if len(e.Samples) >= 2 {
		pdf.SetFont("Helvetica", "B", 13)
		pdf.CellFormat(0, 8, "Throughput over time", "", 1, "L", false, 0, "")
		pdf.Ln(3)
		drawThroughputChart(pdf, e.Samples)
		pdf.Ln(10)
	}

	if len(e.RawLog) > 0 {
		pdf.AddPage()
		pdf.SetFont("Helvetica", "B", 13)
		pdf.CellFormat(0, 8, "Raw iperf3 output", "", 1, "L", false, 0, "")
		pdf.Ln(2)
		pdf.SetFont("Courier", "", 8)
		for _, l := range e.RawLog {
			pdf.CellFormat(0, 4, l, "", 1, "L", false, 0, "")
		}
	}

	return pdf.OutputFileAndClose(path)
}

// drawThroughputChart draws a simple line chart of Mbps over time starting
// at the PDF's current cursor position, using fpdf's basic line-drawing
// primitives (no charting library needed).
func drawThroughputChart(pdf *fpdf.Fpdf, samples []SamplePoint) {
	const chartW = 170.0
	const chartH = 55.0
	x0, y0 := pdf.GetX(), pdf.GetY()

	maxMbps := samples[0].Mbps
	for _, s := range samples {
		if s.Mbps > maxMbps {
			maxMbps = s.Mbps
		}
	}
	if maxMbps <= 0 {
		maxMbps = 1
	}
	minT := samples[0].StartSec
	maxT := samples[len(samples)-1].EndSec
	tSpan := maxT - minT
	if tSpan <= 0 {
		tSpan = 1
	}

	mapX := func(t float64) float64 { return x0 + (t-minT)/tSpan*chartW }
	mapY := func(v float64) float64 { return y0 + chartH - v/maxMbps*chartH }

	// horizontal gridlines + Mbps labels
	pdf.SetFont("Helvetica", "", 7)
	for i := 0; i <= 4; i++ {
		frac := float64(i) / 4
		y := y0 + chartH - frac*chartH
		v := frac * maxMbps
		pdf.SetDrawColor(225, 225, 225)
		pdf.SetLineWidth(0.1)
		pdf.Line(x0, y, x0+chartW, y)
		pdf.SetTextColor(130, 130, 130)
		pdf.Text(x0-16, y+1.2, fmt.Sprintf("%.0f", v))
	}

	// axes
	pdf.SetDrawColor(160, 160, 160)
	pdf.SetLineWidth(0.2)
	pdf.Line(x0, y0, x0, y0+chartH)
	pdf.Line(x0, y0+chartH, x0+chartW, y0+chartH)

	// throughput line
	pdf.SetDrawColor(37, 99, 235)
	pdf.SetLineWidth(0.6)
	for i := 1; i < len(samples); i++ {
		t1 := (samples[i-1].StartSec + samples[i-1].EndSec) / 2
		t2 := (samples[i].StartSec + samples[i].EndSec) / 2
		pdf.Line(mapX(t1), mapY(samples[i-1].Mbps), mapX(t2), mapY(samples[i].Mbps))
	}

	// x-axis labels
	pdf.SetTextColor(130, 130, 130)
	pdf.SetFont("Helvetica", "", 7)
	pdf.Text(x0, y0+chartH+5, fmt.Sprintf("%.0fs", minT))
	pdf.Text(x0+chartW-8, y0+chartH+5, fmt.Sprintf("%.0fs", maxT))
	pdf.Text(x0+chartW/2-6, y0+chartH+5, "Mbps / time")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(x0, y0+chartH+12)
}
