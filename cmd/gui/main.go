// Command KaiPerf3-gui is the graphical (Fyne) frontend for iperf3.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kaihen/kaiperf3-tui/internal/gui"
)

func main() {
	if _, err := exec.LookPath("iperf3"); err != nil {
		fmt.Fprintln(os.Stderr, "KaiPerf3-gui: iperf3 is not installed or not in PATH.")
		fmt.Fprintln(os.Stderr, "Install it first, e.g.:")
		fmt.Fprintln(os.Stderr, "  macOS:   brew install iperf3")
		fmt.Fprintln(os.Stderr, "  Debian:  sudo apt install iperf3")
		fmt.Fprintln(os.Stderr, "  RHEL:    sudo dnf install iperf3")
		os.Exit(1)
	}

	gui.NewApp().Run()
}
