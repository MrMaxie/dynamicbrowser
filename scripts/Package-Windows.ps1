[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSHOME 'Modules/Microsoft.PowerShell.Utility') -Force
$repo = Split-Path $PSScriptRoot -Parent
$version = [IO.File]::ReadAllText((Join-Path $repo 'VERSION')).Trim()
if ($version -notmatch '^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$') {
    throw 'VERSION must contain a semantic version.'
}
$stage = Join-Path ([IO.Path]::GetTempPath()) ('dynamicbrowser-package-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory $stage | Out-Null
$environment = @{}
foreach ($name in 'GOOS', 'GOARCH', 'CGO_ENABLED') {
    $environment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}
try {
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    & go -C $repo build -trimpath -ldflags "-H=windowsgui -X main.version=$version" -o "$stage/dynamicbrowser.exe" ./cmd/dynamicbrowser
    if ($LASTEXITCODE -ne 0) { throw 'Windows build failed.' }
    Copy-Item (Join-Path $repo 'internal/config/config.schema.json') "$stage/config.schema.json"
    foreach ($file in 'config.example.yaml', 'README.md', 'LICENSE') {
        Copy-Item (Join-Path $repo $file) (Join-Path $stage $file)
    }
    $files = @('dynamicbrowser.exe', 'config.schema.json', 'config.example.yaml', 'README.md', 'LICENSE')
    $paths = @($files | ForEach-Object { Join-Path $stage $_ })
    $archive = Join-Path $stage 'package.zip'
    Compress-Archive -LiteralPath $paths -DestinationPath $archive
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip = [IO.Compression.ZipFile]::OpenRead($archive)
    try {
        $actual = @($zip.Entries | ForEach-Object { $_.FullName } | Sort-Object)
        if (Compare-Object ($files | Sort-Object) $actual) { throw 'Unexpected package contents.' }
    } finally {
        $zip.Dispose()
    }
    $output = Join-Path $repo 'bin'
    New-Item -ItemType Directory $output -Force | Out-Null
    $destination = Join-Path $output "dynamicbrowser-$version-windows-amd64.zip"
    Copy-Item $archive $destination -Force
    $hash = (Get-FileHash -LiteralPath $destination -Algorithm SHA256).Hash.ToLowerInvariant()
    [IO.File]::WriteAllText("$destination.sha256", "$hash  $([IO.Path]::GetFileName($destination))`n", [Text.UTF8Encoding]::new($false))
    Write-Host "Package: $destination"
} finally {
    foreach ($name in $environment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $environment[$name], 'Process')
    }
    Remove-Item $stage -Recurse -Force
}
