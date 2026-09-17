BINARY     := KaiPerf3-tui
GUI_BINARY := KaiPerf3-gui
DIST       := dist

.PHONY: build gui test cross clean

build:
	go build -o $(BINARY) .

# Builds the GUI for the CURRENT platform only. Fyne uses cgo (OpenGL
# bindings), so unlike the TUI this can't be cross-compiled with a plain
# `GOOS=... go build` from one machine — it needs each target's native
# toolchain (or a tool like fyne-cross). Run this natively on each OS you
# want a GUI build for.
gui:
	go build -o $(GUI_BINARY) ./cmd/gui

test:
	go vet ./...
	go test ./...

# Cross-compile the TUI as a static binary for every platform worth handing
# to teammates. (The GUI is intentionally not included here — see `gui`.)
cross: test
	mkdir -p $(DIST)
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-darwin-amd64 .
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-windows-amd64.exe .
	@echo "binaries written to $(DIST)/"

clean:
	rm -rf $(DIST) $(BINARY) $(GUI_BINARY)
