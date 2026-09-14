# iperf3-tui

A terminal UI frontend for [iperf3](https://iperf.fr/), the network throughput
testing tool. It wraps the `iperf3` CLI so you don't have to remember flags:
pick client or server mode, fill in a form, and watch a live throughput chart
while the test runs.

![requires iperf3](https://img.shields.io/badge/requires-iperf3-blue)

## Features

- **Client tests** — TCP or UDP, configurable duration, parallel streams,
  reverse mode (`-R`), target bandwidth, and report interval
- **Live throughput chart** rendered in the terminal as the test runs
- **Server mode** — start `iperf3 -s` and watch incoming connections
- **Shows this host's IP addresses** on the main menu and server screens, so
  whoever's on the other end doesn't have to run `ifconfig`/`ipconfig` to
  find out what to connect to
- **History** — every completed test is saved to `~/.iperf3-tui/history.json`
  so you can review past results
- **Export results** — save a completed test as a `.txt` report, a `.csv` of
  the interval samples, or a formatted `.pdf` (with a throughput chart), to
  `~/iperf3-tui-exports/`, for sharing outside the terminal
- **Presets** — save frequently-used client configs and reload them by name

## Requirements

`iperf3` itself must be installed and on your `PATH`. This tool is just a
frontend for it. Version 3.1+ is required for live-updating results (it
relies on iperf3's `--forceflush` flag).

| OS / Distro | Command |
|---|---|
| Debian / Ubuntu / Mint / Pop!_OS | `sudo apt install iperf3` |
| Fedora | `sudo dnf install iperf3` |
| RHEL / CentOS / Rocky / AlmaLinux | `sudo dnf install epel-release && sudo dnf install iperf3` (EPEL needed on RHEL-family; Fedora doesn't need it) |
| Arch / Manjaro | `sudo pacman -S iperf3` |
| openSUSE | `sudo zypper install iperf3` |
| Alpine | `sudo apk add iperf3` |
| macOS | `brew install iperf3` |
| Windows | `winget install iperf3` if available, otherwise download the official binary from [iperf.fr](https://iperf.fr/iperf-download.php) and put it on `PATH` |

## Installing

### Option A: use a prebuilt binary (easiest for sharing with a team)

Grab the binary for your platform from `dist/` (see "Building" below to
produce them), make it executable, and run it:

**macOS / Linux:**
```
chmod +x iperf3-tui-<platform>
./iperf3-tui-<platform>

# optional: put it on PATH so you can just run `iperf3-tui`
sudo mv iperf3-tui-<platform> /usr/local/bin/iperf3-tui
```

**Windows:** no install step needed — double-click
`iperf3-tui-windows-amd64.exe`, or run it from PowerShell/cmd. No admin
rights required.

No other runtime dependencies — it's a single static binary.

### Option B: build from source

Requires Go 1.21+.

```
go build -o iperf3-tui .
./iperf3-tui
```

## Building binaries to share

```
make cross
```

This produces static binaries for macOS (arm64/amd64), Linux (amd64/arm64),
and Windows (amd64) in `dist/`. Hand teammates the one matching their machine
— nothing else needs to be installed except `iperf3`.

## Usage

Run `iperf3-tui` and use the arrow keys / `j`/`k` to navigate, `Enter` to
select, `Tab`/`Shift+Tab` to move between form fields, `←`/`→` or `Space` to
toggle options, and `Esc` to go back.

- **Run client test** — enter the server address/port of a machine already
  running `iperf3 -s` (or this tool's own server mode), configure the test,
  and start it. Press `q` at any time to stop and return.
- **Run server** — listen for incoming test connections on a port.
- **History** — browse past client test results. Press `e` on a highlighted
  entry to export it.
- **Presets** — save a client config from the test form (`Save as preset`)
  and reload it later, from either the form (`Load preset`) or the main menu.
- **Export** — once a client test finishes, press `e` to save it as a text
  report, CSV, or PDF. Files go to `~/iperf3-tui-exports/`, named
  `iperf3-report-<timestamp>.<ext>`; the exact path is shown after saving.

## Data storage

Config is kept in `~/.iperf3-tui/`, and exported reports in
`~/iperf3-tui-exports/`:

- `history.json` — the last 200 completed client test results
- `presets.json` — saved client configurations

Nothing is sent anywhere else; these are local, plain-text JSON files you can
inspect, back up, or delete freely.
