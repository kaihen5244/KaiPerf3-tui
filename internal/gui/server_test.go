package gui

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// TestServerTabConnectionStatus drives the server tab exactly as a click on
// "Start server" would, then connects a real iperf3 client and checks the
// bold status indicator flips LISTENING -> CONNECTED -> LISTENING, which is
// what a person watching the window is meant to see.
func TestServerTabConnectionStatus(t *testing.T) {
	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed")
	}

	testApp := test.NewApp()
	defer test.NewApp()

	win := testApp.NewWindow("test")
	app := &App{fyneApp: testApp, win: win}
	server := newServerTab(app)

	port := 15801
	server.portEntry.SetText(strconv.Itoa(port))
	server.start()
	if server.run == nil {
		t.Fatal("start() did not set up a run")
	}
	defer server.stop()

	if server.statusText.Text == "" {
		t.Fatal("expected a status message immediately after starting")
	}
	waitForStatusContains(t, server, "LISTENING", 3*time.Second)

	client := exec.Command("iperf3", "-c", "127.0.0.1", "-p", strconv.Itoa(port), "-t", "2")
	if err := client.Start(); err != nil {
		t.Fatalf("failed to start test iperf3 client: %v", err)
	}
	defer client.Wait()

	waitForStatusContains(t, server, "CONNECTED", 5*time.Second)
	if !server.run.connected {
		t.Error("expected run.connected to be true while a client is transferring")
	}

	waitForStatusContains(t, server, "LISTENING", 5*time.Second)
	if server.run.connected {
		t.Error("expected run.connected to be false again after the transfer finished")
	}
	if server.run.lastClientAddr != "127.0.0.1" {
		t.Errorf("expected lastClientAddr to be recorded, got %q", server.run.lastClientAddr)
	}
}

func waitForStatusContains(t *testing.T, s *serverTab, substr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(s.statusText.Text, substr) {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatalf("status text did not contain %q within %s (last seen: %q)", substr, timeout, s.statusText.Text)
}
