[CmdletBinding()]
param([switch]$KeepOpen)

$ErrorActionPreference = 'Stop'
if ($env:USERNAME -ne 'WDAGUtilityAccount' -or -not (Test-Path C:\Inputs\tests)) {
    throw 'This script must run inside the harness Windows Sandbox.'
}
Start-Transcript -Path C:\Results\guest.log -Force
$result = @{ success = $false; packages = @(); manualChecks = 'not run' }
try {
    New-Item -ItemType Directory C:\Work -Force | Out-Null
    Copy-Item C:\Inputs\source C:\Work\source -Recurse
    New-Item -ItemType Directory C:\Work\bin -Force | Out-Null
    Copy-Item C:\Inputs\dynamicbrowser.exe, C:\Inputs\browser-probe.exe C:\Work\bin
    $env:GOROOT = 'C:\Go'
    $env:PATH = 'C:\Go\bin;' + $env:PATH
    $env:GOTOOLCHAIN = 'local'
    $env:GOPROXY = 'off'
    $env:GOSUMDB = 'off'
    $env:GOWORK = 'off'
    $env:GOFLAGS = '-mod=vendor'
    $env:CGO_ENABLED = '0'
    $env:DYNAMICBROWSER_SANDBOX_REGISTRATION = '1'
    foreach ($package in 'config', 'routing', 'autostart', 'app', 'registration') {
        Set-Location "C:\Work\source\internal\$package"
        & "C:\Inputs\tests\$package.test.exe" '-test.v' '-test.timeout=180s' *> "C:\Results\$package.log"
        $exitCode = $LASTEXITCODE
        $log = "C:\Results\$package.log"
        [IO.File]::WriteAllText($log, [IO.File]::ReadAllText($log), [Text.UTF8Encoding]::new($false))
        if ($exitCode -ne 0) { throw "$package tests exited with status $exitCode" }
        $result.packages += $package
    }
    @'
# yaml-language-server: $schema=./config.schema.json
default: first
browsers:
  first:
    exe: C:/Work/bin/browser-probe.exe
    args: ['first profile']
  second:
    exe: C:/Work/bin/browser-probe.exe
    args: ['second profile']
rules:
  - browser: second
    target:
      path: '/work/*'
'@ | Set-Content C:\Work\bin\config.yaml -Encoding UTF8
    Copy-Item C:\Work\source\internal\config\config.schema.json C:\Work\bin
    $url = 'https://example.test/work/sandbox-check'
    $process = Start-Process C:\Work\bin\dynamicbrowser.exe -ArgumentList $url -PassThru
    if (-not $process.WaitForExit(10000) -or $process.ExitCode -ne 0) { throw 'Fixture URL invocation failed.' }
    $log = 'C:\Work\bin\browser-launches.jsonl'
    $deadline = [DateTime]::UtcNow.AddSeconds(5)
    $launch = $null
    do {
        if (Test-Path $log) {
            $text = [IO.File]::ReadAllText($log)
            if ($text.EndsWith("`n")) { $launch = $text | ConvertFrom-Json; break }
        }
        Start-Sleep -Milliseconds 50
    } while ([DateTime]::UtcNow -lt $deadline)
    if ($null -eq $launch) { throw 'The browser probe did not record a complete launch.' }
    if ($launch.Count -ne 2 -or $launch[0] -ne 'second profile' -or $launch[1] -ne $url) { throw 'Unexpected fixture routing arguments.' }
    Copy-Item $log C:\Results\fixture-routing.jsonl
    $result.fixtureRouting = 'passed'
    $result.success = $true
} catch {
    $result.error = $_.Exception.Message
    $_ | Out-String | Set-Content C:\Results\failure.txt
} finally {
    Stop-Transcript
    $result | ConvertTo-Json -Depth 4 | Set-Content C:\Results\result.tmp -Encoding UTF8
    Move-Item C:\Results\result.tmp C:\Results\result.json -Force
    if ($KeepOpen) {
        Write-Host 'Automated results: C:\Results\result.json'
        Write-Host 'Manual-test application and fixture: C:\Work\bin'
        Write-Host 'Close this Sandbox window when finished.'
    } else {
        shutdown.exe /s /t 5
    }
}
