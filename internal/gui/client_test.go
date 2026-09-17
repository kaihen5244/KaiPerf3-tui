package gui

import (
	"os/exec"
	"strconv"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"github.com/kaihen/kaiperf3-tui/internal/iperf"
)

// TestClientTabRunsRealTest drives the client tab exactly as button clicks
// would (fill fields, call start), against a real local iperf3 server, and
// waits for the background consume/handleEvent goroutine to finish. Fyne's
// test driver runs fyne.Do callbacks synchronously, so this exercises the
// real event-handling path without needing a display or simulated clicks.
func TestClientTabRunsRealTest(t *testing.T) {
	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed")
	}

	// Isolate history/export files from the real user's data.
	t.Setenv("HOME", t.TempDir())

	port := 15701
	srv := exec.Command("iperf3", "-s", "-p", strconv.Itoa(port))
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start test iperf3 server: %v", err)
	}
	defer func() { _ = srv.Process.Kill() }()
	time.Sleep(300 * time.Millisecond) // let the server bind before connecting

	testApp := test.NewApp()
	defer test.NewApp() // reset the current app for other tests

	win := testApp.NewWindow("test")
	app := &App{fyneApp: testApp, win: win}
	app.history = newHistoryTab(app)
	app.client = newClientTab(app)

	app.client.serverEntry.SetText("127.0.0.1")
	app.client.portEntry.SetText(strconv.Itoa(port))
	app.client.durationEntry.SetText("2")
	app.client.parallelEntry.SetText("1")
	app.client.intervalEntry.SetText("1")

	app.client.start()
	if app.client.run == nil {
		t.Fatal("start() did not set up a run (check statusLabel/dialog for the error)")
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && !app.client.run.finished {
		time.Sleep(50 * time.Millisecond)
	}
	if !app.client.run.finished {
		t.Fatal("test did not finish within 10s")
	}

	r := app.client.run
	if r.err != nil {
		t.Fatalf("run finished with error: %v", r.err)
	}
	if r.sender == nil || r.sender.Mbps <= 0 {
		t.Errorf("expected a positive sender throughput, got %+v", r.sender)
	}
	if len(r.samples) == 0 {
		t.Errorf("expected at least one chart sample to have been recorded")
	}
	if app.client.logBox.Text == "" {
		t.Errorf("expected log box to contain iperf3 output")
	}

	entries, err := iperf.LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one history entry to be recorded, got %d", len(entries))
	}
	if entries[0].AvgMbpsUp <= 0 {
		t.Errorf("recorded history entry has no upload throughput: %+v", entries[0])
	}

	app.client.export(0)
	if app.client.exportStatus.Text == "" {
		t.Errorf("expected export status to be set after export")
	}
}
