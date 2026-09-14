package iperf

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

// ClientConfig holds the parameters for an `iperf3 -c` run.
type ClientConfig struct {
	Server    string
	Port      int
	Duration  int // seconds
	Parallel  int
	Reverse   bool
	UDP       bool
	Bandwidth string // e.g. "0" (unlimited) or "10M"; UDP only
	Interval  float64
}

func (c ClientConfig) Args() []string {
	args := []string{"-c", c.Server, "-p", strconv.Itoa(c.Port), "-t", strconv.Itoa(c.Duration)}
	if c.Parallel > 1 {
		args = append(args, "-P", strconv.Itoa(c.Parallel))
	}
	if c.Reverse {
		args = append(args, "-R")
	}
	if c.UDP {
		args = append(args, "-u")
		if c.Bandwidth != "" {
			args = append(args, "-b", c.Bandwidth)
		}
	}
	interval := c.Interval
	if interval <= 0 {
		interval = 1
	}
	args = append(args, "-i", strconv.FormatFloat(interval, 'f', -1, 64))
	// iperf3 fully buffers stdout when it isn't a terminal (i.e. always, when
	// we run it as a subprocess), so without this every line arrives in one
	// burst at exit instead of as the test progresses.
	args = append(args, "--forceflush")
	return args
}

func (c ClientConfig) Summary() string {
	proto := "TCP"
	if c.UDP {
		proto = "UDP"
	}
	dir := "→"
	if c.Reverse {
		dir = "←"
	}
	return fmt.Sprintf("%s %s:%d %s %ds x%d", proto, c.Server, c.Port, dir, c.Duration, c.Parallel)
}

// ServerConfig holds the parameters for an `iperf3 -s` run.
type ServerConfig struct {
	Port int
}

func (s ServerConfig) Args() []string {
	return []string{"-s", "-p", strconv.Itoa(s.Port), "--forceflush"}
}

// DoneMsg signals the process has exited.
type DoneMsg struct {
	Err      error
	ExitCode int
}

// Event is one item from a running Process: either a line of output or,
// as the final event, the process's exit status. Using a single channel
// (rather than separate line/done channels) guarantees the consumer sees
// every line before it sees Done — two channels racing against each other
// can deliver Done before a concurrently-running reader has drained the
// last buffered lines.
type Event struct {
	Line string
	Done *DoneMsg // non-nil only on the final event
}

// Process wraps a running iperf3 command, streaming output line-by-line.
type Process struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	Events chan Event
}

// Start launches `iperf3` with the given args and begins streaming output.
func Start(args ...string) (*Process, error) {
	binPath, err := exec.LookPath("iperf3")
	if err != nil {
		return nil, fmt.Errorf("iperf3 not found in PATH: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	cmd.Stderr = cmd.Stdout // merge stderr into the same stream

	p := &Process{
		cmd:    cmd,
		cancel: cancel,
		Events: make(chan Event, 256),
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			p.Events <- Event{Line: scanner.Text()}
		}
		err := cmd.Wait()
		exitCode := 0
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
		p.Events <- Event{Done: &DoneMsg{Err: err, ExitCode: exitCode}}
		close(p.Events)
	}()

	return p, nil
}

// Stop terminates the process if it's still running.
func (p *Process) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}
