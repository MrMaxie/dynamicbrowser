set windows-shell := ["powershell.exe", "-NoProfile", "-Command"]

# List available commands.
default:
    @just --list

# Build and start the Windows tray application without a console.
run: build
    Start-Process bin/dynamicbrowser.exe

# Print the application version.
version:
    $version = (Get-Content -Raw VERSION).Trim(); go run "-ldflags=-X main.version=$version" ./cmd/dynamicbrowser --version

# Build the native executable in bin/.
build:
    $version = (Get-Content -Raw VERSION).Trim(); go build "-ldflags=-H=windowsgui -X main.version=$version" -o bin/dynamicbrowser.exe ./cmd/dynamicbrowser

# Run regression tests.
test:
    go test ./...

# Format Go source files with gofmt.
fmt:
    golangci-lint fmt ./...

# Check lint rules and formatting without modifying files.
lint:
    golangci-lint config verify
    golangci-lint run ./...

# Validate OpenSpec requirements and changes.
spec-check:
    openspec validate --all --strict --no-interactive

# Run all checks and build the executable.
check: lint test spec-check build
