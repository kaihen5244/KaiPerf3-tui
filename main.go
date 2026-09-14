// Command iperf3-tui is a terminal UI frontend for the iperf3 network
// throughput testing tool.
package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaihen/iperf3-tui/internal/ui"
)

func main() {
	if _, err := exec.LookPath("iperf3"); err != nil {
		fmt.Fprintln(os.Stderr, "iperf3-tui: iperf3 is not installed or not in PATH.")
		fmt.Fprintln(os.Stderr, "Install it first, e.g.:")
		fmt.Fprintln(os.Stderr, "  macOS:   brew install iperf3")
		fmt.Fprintln(os.Stderr, "  Debian:  sudo apt install iperf3")
		fmt.Fprintln(os.Stderr, "  RHEL:    sudo dnf install iperf3")
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "iperf3-tui:", err)
		os.Exit(1)
	}
}
