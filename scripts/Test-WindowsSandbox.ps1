[CmdletBinding()]
param(
    [switch]$KeepOpen,
    [ValidateRange(60, 1800)][int]$TimeoutSeconds = 600
)

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSHOME 'Modules/Microsoft.PowerShell.Utility') -Force
$repo = Split-Path $PSScriptRoot -Parent
$launcher = (Get-Command WindowsSandbox.exe -ErrorAction Stop).Source
if (Get-Process WindowsSandbox, WindowsSandboxClient -ErrorAction SilentlyContinue) {
    throw 'Close existing Windows Sandbox windows before starting this harness.'
}
if ((& go env GOOS) -ne 'windows' -or (& go env GOARCH) -ne 'amd64') {
    throw 'Run this harness with a Windows amd64 Go toolchain.'
}
$goRoot = (& go env GOROOT).Trim()
$run = Join-Path ([IO.Path]::GetTempPath()) ('dynamicbrowser-sandbox-' + [Guid]::NewGuid().ToString('N'))
$inputDir = Join-Path $run 'input'
$source = Join-Path $inputDir 'source'
$results = Join-Path $run 'results'
New-Item -ItemType Directory -Path $source, $results, "$inputDir/tests" -Force | Out-Null
Write-Host "Sandbox files and results: $run"

$files = & git -C $repo ls-files --cached --others --exclude-standard -- cmd internal go.mod go.sum VERSION config.example.yaml
if ($LASTEXITCODE -ne 0) { throw 'Cannot enumerate project source files.' }
foreach ($file in ($files | Sort-Object -Unique)) {
    $from = $repo
    foreach ($component in ($file -split '[/\\]')) {
        if ($component -in '.', '..') { throw "Invalid source path: $file" }
        $from = Join-Path $from $component
        if ((Get-Item -LiteralPath $from -Force).Attributes -band [IO.FileAttributes]::ReparsePoint) {
            throw "Source links are not supported: $file"
        }
    }
    $to = Join-Path $source $file
    New-Item -ItemType Directory (Split-Path $to -Parent) -Force | Out-Null
    Copy-Item -LiteralPath $from -Destination $to
}
& go -C $source mod vendor
if ($LASTEXITCODE -ne 0) { throw 'Preparing offline dependencies failed.' }
$packages = @('config', 'routing', 'autostart', 'app', 'registration')
foreach ($package in $packages) {
    & go -C $source test -mod=vendor -race -c -o "$inputDir/tests/$package.test.exe" "./internal/$package"
    if ($LASTEXITCODE -ne 0) { throw "Building $package tests failed." }
}
$version = [IO.File]::ReadAllText((Join-Path $repo 'VERSION')).Trim()
& go -C $source build -mod=vendor -ldflags "-H=windowsgui -X main.version=$version" -o "$inputDir/dynamicbrowser.exe" ./cmd/dynamicbrowser
if ($LASTEXITCODE -ne 0) { throw 'Building the application failed.' }
& go build -ldflags '-H=windowsgui' -o "$inputDir/browser-probe.exe" (Join-Path $PSScriptRoot 'sandbox/browser-probe/main.go')
if ($LASTEXITCODE -ne 0) { throw 'Building the browser probe failed.' }
Copy-Item (Join-Path $PSScriptRoot 'sandbox/Run-Tests.ps1') "$inputDir/Run-Tests.ps1"
$manifest = @{
    version = $version
    commit = (& git -C $repo rev-parse HEAD).Trim()
    packages = $packages
    files = @(Get-ChildItem $inputDir -Recurse -File | ForEach-Object {
        @{ path = $_.FullName.Substring($inputDir.Length + 1); sha256 = (Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash }
    })
}
$manifest | ConvertTo-Json -Depth 5 | Set-Content (Join-Path $run 'inputs.json') -Encoding UTF8
$guestOption = if ($KeepOpen) { ' -KeepOpen' } else { '' }
$inputXml = [Security.SecurityElement]::Escape($inputDir)
$resultsXml = [Security.SecurityElement]::Escape($results)
$goXml = [Security.SecurityElement]::Escape($goRoot)
@"
<Configuration>
  <VGpu>Disable</VGpu><Networking>Disable</Networking><ClipboardRedirection>Disable</ClipboardRedirection><MemoryInMB>4096</MemoryInMB>
  <MappedFolders>
    <MappedFolder><HostFolder>$inputXml</HostFolder><SandboxFolder>C:\Inputs</SandboxFolder><ReadOnly>true</ReadOnly></MappedFolder>
    <MappedFolder><HostFolder>$goXml</HostFolder><SandboxFolder>C:\Go</SandboxFolder><ReadOnly>true</ReadOnly></MappedFolder>
    <MappedFolder><HostFolder>$resultsXml</HostFolder><SandboxFolder>C:\Results</SandboxFolder><ReadOnly>false</ReadOnly></MappedFolder>
  </MappedFolders>
  <LogonCommand><Command>powershell.exe -NoProfile -ExecutionPolicy Bypass -File C:\Inputs\Run-Tests.ps1$guestOption</Command></LogonCommand>
</Configuration>
"@ | Set-Content (Join-Path $run 'test.wsb') -Encoding UTF8
$watcher = [IO.FileSystemWatcher]::new($results, 'result.json')
$watcher.EnableRaisingEvents = $true
try {
    if (Get-Process WindowsSandbox, WindowsSandboxClient -ErrorAction SilentlyContinue) {
        throw 'A Windows Sandbox opened during staging; close its window before retrying this harness.'
    }
    Start-Process $launcher -ArgumentList ('"' + (Join-Path $run 'test.wsb') + '"')
    $change = $watcher.WaitForChanged(([IO.WatcherChangeTypes]::Created -bor [IO.WatcherChangeTypes]::Renamed), $TimeoutSeconds * 1000)
    if ($change.TimedOut) { throw "Sandbox timed out. Inspect $results; close its window if it is still running." }
    $result = [IO.File]::ReadAllText((Join-Path $results 'result.json')) | ConvertFrom-Json
    if (-not $result.success) { throw "Sandbox failed: $($result.error). Logs: $results" }
    Write-Host "Automated Sandbox tests passed. Results: $results"
    if ($KeepOpen) { Write-Host 'Sandbox remains open for the manual Default Apps checklist in scripts/sandbox/README.md.' }
} finally {
    $watcher.Dispose()
}
