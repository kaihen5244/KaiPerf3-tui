package iperf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testEntry() HistoryEntry {
	return HistoryEntry{
		Time:        time.Date(2026, 9, 10, 23, 0, 0, 0, time.UTC),
		Config:      "TCP 127.0.0.1:5201 → 10s x1",
		AvgMbpsUp:   940.5,
		AvgMbpsDown: 935.2,
		Retransmits: 3,
		Samples: []SamplePoint{
			{StartSec: 0, EndSec: 1, Mbps: 900},
			{StartSec: 1, EndSec: 2, Mbps: 950},
			{StartSec: 2, EndSec: 3, Mbps: 940},
		},
		RawLog: []string{"Connecting to host 127.0.0.1, port 5201", "iperf Done."},
	}
}

func TestPDFSafeStripsArrows(t *testing.T) {
	got := pdfSafe("TCP 127.0.0.1:5201 → 10s x1 ← reverse")
	if strings.ContainsAny(got, "→←") {
		t.Errorf("pdfSafe left non-ASCII arrow characters: %q", got)
	}
	if !strings.Contains(got, "->") || !strings.Contains(got, "<-") {
		t.Errorf("pdfSafe did not substitute ASCII arrows: %q", got)
	}
}

func TestExportTXT(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.txt")
	if err := ExportTXT(testEntry(), path); err != nil {
		t.Fatalf("ExportTXT: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "iperf3 Test Report") || !strings.Contains(s, "940.5") {
		t.Errorf("txt export missing expected content: %q", s)
	}
}

func TestExportCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.csv")
	if err := ExportCSV(testEntry(), path); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 { // header + 3 samples
		t.Fatalf("expected 4 CSV lines, got %d: %q", len(lines), data)
	}
	if lines[0] != "start_sec,end_sec,mbps" {
		t.Errorf("unexpected CSV header: %q", lines[0])
	}
}

func TestExportPDF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.pdf")
	if err := ExportPDF(testEntry(), path); err != nil {
		t.Fatalf("ExportPDF: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF-") {
		t.Errorf("exported file does not look like a PDF")
	}
}

func TestDefaultExportPath(t *testing.T) {
	path, err := DefaultExportPath(testEntry(), "pdf")
	if err != nil {
		t.Fatalf("DefaultExportPath: %v", err)
	}
	if !strings.HasSuffix(path, ".pdf") {
		t.Errorf("expected .pdf suffix, got %q", path)
	}
}
