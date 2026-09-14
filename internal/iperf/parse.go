// Package iperf wraps the iperf3 CLI: building commands, running them, and
// parsing the human-readable interval/summary lines it prints as it runs.
package iperf

import (
	"regexp"
	"strconv"
	"strings"
)

// Sample is one parsed line of iperf3 output.
type Sample struct {
	Raw         string
	Stream      string // "SUM" or a stream id like "5"
	IsSummary   bool
	IsSender    bool
	IsReceiver  bool
	StartSec    float64
	EndSec      float64
	Mbps        float64
	Retransmits int     // TCP only, -1 if not present
	LossPct     float64 // UDP only, -1 if not present
	JitterMs    float64 // UDP only, -1 if not present
}

// IsChartable reports whether this sample should be plotted on the live
// throughput chart given how many parallel streams the test is using: with
// multiple streams only the [SUM] line reflects total throughput, otherwise
// the single stream's own line is the total.
func (s Sample) IsChartable(parallel int) bool {
	if parallel > 1 {
		return s.Stream == "SUM"
	}
	return s.Stream != "SUM"
}

// intervalRe matches the per-stream interval/summary lines iperf3 prints, e.g.:
//
//	[  5]   0.00-1.00   sec   112 MBytes   941 Mbits/sec
//	[  5]   0.00-1.00   sec   112 MBytes   941 Mbits/sec    0    850 KBytes
//	[  5]   0.00-1.00   sec  1.17 MBytes  9.80 Mbits/sec  0.032 ms  0/850 (0%)
//	[SUM]   0.00-10.00  sec  1.10 GBytes   941 Mbits/sec                  sender
var intervalRe = regexp.MustCompile(
	`^\[\s*(SUM|\d+)\]\s+([\d.]+)-([\d.]+)\s+sec\s+[\d.]+\s+\w?Bytes\s+([\d.]+)\s+(\w?)bits/sec` +
		`(?:\s+(\d+))?` + // TCP retransmits (client only)
		`(?:\s+[\d.]+\s+\w?Bytes)?` + // TCP congestion window, e.g. "850 KBytes" (interval lines only)
		`(?:\s+[\d.]+\s+ms)?` + // UDP jitter, captured separately below
		`(?:\s+\d+/\d+\s+\(([\d.]+)%\))?` + // UDP loss pct
		`\s*(sender|receiver)?\s*$`,
)

var jitterRe = regexp.MustCompile(`([\d.]+)\s+ms\s+\d+/\d+\s+\(`)

// ParseLine attempts to parse a single line of iperf3 stdout into a Sample.
// The second return value is false for lines that aren't interval/summary
// reports (banners, connection messages, errors, etc).
func ParseLine(line string) (Sample, bool) {
	line = strings.TrimRight(line, "\r\n")
	m := intervalRe.FindStringSubmatch(line)
	if m == nil {
		return Sample{}, false
	}

	stream := m[1]
	start, _ := strconv.ParseFloat(m[2], 64)
	end, _ := strconv.ParseFloat(m[3], 64)
	rate, _ := strconv.ParseFloat(m[4], 64)
	unit := m[5]

	mbps := rate
	switch unit {
	case "G":
		mbps = rate * 1000
	case "K":
		mbps = rate / 1000
	case "":
		mbps = rate / 1_000_000 // bare "bits/sec", no unit prefix (e.g. a zero-throughput interval)
	}

	s := Sample{
		Raw:         line,
		Stream:      stream,
		StartSec:    start,
		EndSec:      end,
		Mbps:        mbps,
		Retransmits: -1,
		LossPct:     -1,
		JitterMs:    -1,
	}

	if m[6] != "" {
		s.Retransmits, _ = strconv.Atoi(m[6])
	}
	if m[7] != "" {
		s.LossPct, _ = strconv.ParseFloat(m[7], 64)
	}
	if jm := jitterRe.FindStringSubmatch(line); jm != nil {
		s.JitterMs, _ = strconv.ParseFloat(jm[1], 64)
	}

	switch m[8] {
	case "sender":
		s.IsSummary = true
		s.IsSender = true
	case "receiver":
		s.IsSummary = true
		s.IsReceiver = true
	}

	return s, true
}
