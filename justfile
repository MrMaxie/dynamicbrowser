set windows-shell := ["powershell.exe", "-NoProfile", "-Command"]

# List available commands.
default:
    @just --list

# Run the Hello World program.
run:
    go run .

# Print the application version.
version:
    go run . --version

# Build the native executable in bin/.
build:
    go build -o bin/ .

# Run the CLI regression tests.
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
