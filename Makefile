BINARY  := godot-cli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/Sqoff/godot-cli/cmd.Version=$(VERSION)

.PHONY: build clean snapshot release

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY).exe .

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-linux-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-amd64 .

clean:
	go clean
	rm -f $(BINARY) $(BINARY).exe $(BINARY)-linux-amd64 $(BINARY)-darwin-arm64 $(BINARY)-darwin-amd64
	rm -rf dist/

snapshot:
	goreleaser build --clean --snapshot

release:
	goreleaser release --clean
