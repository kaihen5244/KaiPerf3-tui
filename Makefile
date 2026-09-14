BINARY := KaiPerf3-tui
DIST   := dist

.PHONY: build test cross clean

build:
	go build -o $(BINARY) .

test:
	go vet ./...
	go test ./...

# Cross-compile a static binary for every platform worth handing to teammates.
cross: test
	mkdir -p $(DIST)
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-darwin-amd64 .
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST)/$(BINARY)-windows-amd64.exe .
	@echo "binaries written to $(DIST)/"

clean:
	rm -rf $(DIST) $(BINARY)
