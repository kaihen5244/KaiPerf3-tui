package iperf

import "testing"

func TestParseLine(t *testing.T) {
	cases := []struct {
		name       string
		line       string
		wantOK     bool
		wantMbps   float64
		wantStream string
		wantSum    bool
		wantSender bool
	}{
		{
			name:       "tcp interval with cwnd",
			line:       "[  5]   0.00-1.00   sec   112 MBytes   941 Mbits/sec    0    850 KBytes",
			wantOK:     true,
			wantMbps:   941,
			wantStream: "5",
		},
		{
			name:       "tcp interval no cwnd (receiver side)",
			line:       "[  5]   0.00-1.00   sec   112 MBytes   941 Mbits/sec",
			wantOK:     true,
			wantMbps:   941,
			wantStream: "5",
		},
		{
			name:       "udp interval with jitter and loss",
			line:       "[  5]   0.00-1.00   sec  1.17 MBytes  9.80 Mbits/sec  0.032 ms  0/850 (0%)",
			wantOK:     true,
			wantMbps:   9.80,
			wantStream: "5",
		},
		{
			name:       "gbit interval",
			line:       "[  5]   2.00-3.00   sec  17.9 GBytes   154 Gbits/sec    0   4.00 MBytes",
			wantOK:     true,
			wantMbps:   154000,
			wantStream: "5",
		},
		{
			name:       "sum summary sender",
			line:       "[SUM]   0.00-10.00  sec  1.10 GBytes   941 Mbits/sec                  sender",
			wantOK:     true,
			wantMbps:   941,
			wantStream: "SUM",
			wantSum:    true,
			wantSender: true,
		},
		{
			name:       "single stream summary receiver",
			line:       "[  5]   0.00-4.00   sec  70.7 GBytes   152 Gbits/sec                  receiver",
			wantOK:     true,
			wantMbps:   152000,
			wantStream: "5",
			wantSum:    true,
		},
		{
			name:       "zero throughput udp interval (no unit prefix)",
			line:       "[  5]   1.01-2.00   sec  0.00 Bytes  0.00 bits/sec  0",
			wantOK:     true,
			wantMbps:   0,
			wantStream: "5",
		},
		{
			name:   "not an interval line",
			line:   "Connecting to host 127.0.0.1, port 5301",
			wantOK: false,
		},
		{
			name:   "header line",
			line:   "[ ID] Interval           Transfer     Bitrate         Retr",
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, ok := ParseLine(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ParseLine(%q) ok = %v, want %v", tc.line, ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			if s.Mbps != tc.wantMbps {
				t.Errorf("Mbps = %v, want %v", s.Mbps, tc.wantMbps)
			}
			if s.Stream != tc.wantStream {
				t.Errorf("Stream = %q, want %q", s.Stream, tc.wantStream)
			}
			if s.IsSummary != tc.wantSum {
				t.Errorf("IsSummary = %v, want %v", s.IsSummary, tc.wantSum)
			}
			if tc.wantSender && !s.IsSender {
				t.Errorf("expected IsSender=true")
			}
		})
	}
}

func TestIsChartable(t *testing.T) {
	single, _ := ParseLine("[  5]   0.00-1.00   sec   112 MBytes   941 Mbits/sec")
	sum, _ := ParseLine("[SUM]   0.00-1.00   sec   112 MBytes   941 Mbits/sec")

	if !single.IsChartable(1) {
		t.Error("single stream sample should be chartable when parallel=1")
	}
	if single.IsChartable(4) {
		t.Error("single stream sample should NOT be chartable when parallel=4 (should use SUM)")
	}
	if !sum.IsChartable(4) {
		t.Error("SUM sample should be chartable when parallel=4")
	}
}
