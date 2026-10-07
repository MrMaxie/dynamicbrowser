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

icon:
    foreach ($size in 16,20,22,24,28,32,36,40,44,48,56,64,128,256) { magick -background none -density ($size * 4) assets/icon.svg -strip "PNG32:internal/app/assets/icon-$size.png"; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE } }
    magick internal/app/assets/icon-256.png internal/app/assets/icon-128.png internal/app/assets/icon-64.png internal/app/assets/icon-48.png internal/app/assets/icon-32.png internal/app/assets/icon-16.png assets/icon.ico
    go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -ico assets/icon.ico -manifest assets/app.manifest -o cmd/dynamicbrowser/icon_windows_amd64.syso

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
