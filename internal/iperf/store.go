package iperf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".iperf3-tui")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// SamplePoint is one interval report from a test, kept for charts/exports.
type SamplePoint struct {
	StartSec float64 `json:"start_sec"`
	EndSec   float64 `json:"end_sec"`
	Mbps     float64 `json:"mbps"`
}

// HistoryEntry is one completed test result, persisted for later review.
type HistoryEntry struct {
	Time        time.Time     `json:"time"`
	Config      string        `json:"config"` // human-readable summary
	AvgMbpsUp   float64       `json:"avg_mbps_up"`
	AvgMbpsDown float64       `json:"avg_mbps_down"`
	Retransmits int           `json:"retransmits"`
	LossPct     float64       `json:"loss_pct"`
	Aborted     bool          `json:"aborted"`
	Samples     []SamplePoint `json:"samples,omitempty"`
	RawLog      []string      `json:"raw_log,omitempty"`
}

// ClientRunResult summarizes a finished (or aborted) client test, as
// collected by whichever UI ran it. BuildHistoryEntry turns this into a
// HistoryEntry, so every frontend records/exports results the same way.
type ClientRunResult struct {
	Config     ClientConfig
	Sender     *Sample // nil if no sender summary line was seen
	Receiver   *Sample // nil if no receiver summary line was seen
	Points     []SamplePoint
	RawLog     []string
	Aborted    bool
	FinishedAt time.Time
}

// BuildHistoryEntry converts a ClientRunResult into a HistoryEntry.
func BuildHistoryEntry(r ClientRunResult) HistoryEntry {
	t := r.FinishedAt
	if t.IsZero() {
		t = time.Now()
	}
	entry := HistoryEntry{
		Time:    t,
		Config:  r.Config.Summary(),
		Aborted: r.Aborted,
		Samples: r.Points,
		RawLog:  r.RawLog,
	}
	if r.Sender != nil {
		entry.AvgMbpsUp = r.Sender.Mbps
		entry.Retransmits = r.Sender.Retransmits
		if r.Sender.LossPct >= 0 {
			entry.LossPct = r.Sender.LossPct
		}
	}
	if r.Receiver != nil {
		entry.AvgMbpsDown = r.Receiver.Mbps
	}
	return entry
}

func historyPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

func LoadHistory() ([]HistoryEntry, error) {
	path, err := historyPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func AppendHistory(e HistoryEntry) error {
	entries, err := LoadHistory()
	if err != nil {
		return err
	}
	entries = append([]HistoryEntry{e}, entries...) // newest first
	if len(entries) > 200 {
		entries = entries[:200]
	}
	path, err := historyPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Preset is a saved client configuration for quick reuse.
type Preset struct {
	Name   string       `json:"name"`
	Config ClientConfig `json:"config"`
}

func presetsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "presets.json"), nil
}

func LoadPresets() ([]Preset, error) {
	path, err := presetsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var presets []Preset
	if err := json.Unmarshal(data, &presets); err != nil {
		return nil, err
	}
	return presets, nil
}

func SavePresets(presets []Preset) error {
	path, err := presetsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func UpsertPreset(p Preset) error {
	presets, err := LoadPresets()
	if err != nil {
		return err
	}
	for i, existing := range presets {
		if existing.Name == p.Name {
			presets[i] = p
			return SavePresets(presets)
		}
	}
	presets = append(presets, p)
	return SavePresets(presets)
}

func DeletePreset(name string) error {
	presets, err := LoadPresets()
	if err != nil {
		return err
	}
	out := presets[:0]
	for _, p := range presets {
		if p.Name != name {
			out = append(out, p)
		}
	}
	return SavePresets(out)
}
