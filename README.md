# dynamicbrowser

A Windows web-link router that opens URLs in named browser profiles using URL and source-application rules. An optional tray lets you temporarily choose one profile instead of automatic routing.

## Quick start

Install [mise](https://mise.jdx.dev/), then build:

```powershell
mise trust
mise install
mise exec -- just build
```

For a new installation, copy `config.example.yaml` to `bin/config.yaml`, then replace the example executable paths and profile arguments with your browser's values. Keep `config.yaml` next to `dynamicbrowser.exe`; the working directory does not affect configuration lookup. In an extracted ZIP, copy the example to `config.yaml` and use `./dynamicbrowser.exe` instead of `./bin/dynamicbrowser.exe` in the commands below.

Run the tray or open a URL directly:

```powershell
Start-Process .\bin\dynamicbrowser.exe
.\bin\dynamicbrowser.exe 'https://work.example.test/projects/item'
```

URL invocations do not start a tray. If one is running, they use its temporary selection. Closing the tray does not disable later URL routing. Starting a second tray exits silently; `--force` replaces the running tray.

The tray menu contains `Auto`, available browser names, `Edit configuration`, and `Close tray`. A selected browser overrides rules only for that tray session; `Auto`, closing the tray, or removing that browser from configuration clears the selection.

## Configuration

See [config.example.yaml](config.example.yaml). Replace its placeholder paths before using it.

- `browsers` maps names to `exe` and optional `args`, an array of strings. Relative executable paths resolve beside dynamicbrowser; bare executable names can also resolve through `PATH`. Arguments and each URL are passed directly, without a shell.
- `rules` are evaluated in order. Conditions can match `source.process`, `source.window`, `target.url`, `target.domain`, and `target.path`. Specified fields combine with AND; an array within one field combines alternatives with OR.
- Patterns are whole-value globs (`*` and `?`, including path separators) unless prefixed with `regex:`. Regex patterns use Go's regular-expression syntax; add anchors when you want a whole-value match. Backslash escapes literal glob wildcards.
- `default` names the fallback browser when no rule selects an available browser. There is no implicit first-browser fallback. A valid rule or tray override works without a default; needing an unavailable default produces an error dialog.
- `autostart: true` opens the tray at user sign-in. It is not required for URL routing.

A running tray reloads valid edits, including atomic file replacement. Invalid edits retain the last valid configuration. Only available browsers appear in the menu, in YAML declaration order.

When missing, the application creates an empty `config.yaml` and a local `config.schema.json`, preserving existing files. The example and newly created YAML include a modeline for YAML language servers. Additional top-level fields are allowed; known fields and nested properties are validated.

## Default browser on Windows

```powershell
.\bin\dynamicbrowser.exe --register
.\bin\dynamicbrowser.exe --restore
```

`--register` makes dynamicbrowser available and opens Windows Default Apps when selection is needed. Choose it for both HTTP and HTTPS, then confirm the application's dialog. Normal startup never changes your default browser.

`--restore` guides you back to the separately saved HTTP/HTTPS choices and removes registration only after verification. A newer independent choice is not a restoration target. Windows Settings can change both protocols from a single picker; if it also changes an independent choice, dynamicbrowser reports an error and retains recovery state. It does not bypass protected Windows associations.

Registration and restoration cannot be combined with URLs, each other, or `--force`. Cancelling setup returns exit status 2; an error returns 1. `--version` prints the build version without opening the tray or performing setup.

## Development

```powershell
mise exec -- just fmt
mise exec -- just check
mise exec -- just version
mise exec -- just sandbox-test
mise exec -- just package
```

`just check` runs lint, formatting checks, tests, OpenSpec validation, and a build. `just sandbox-test` runs offline tests in Windows Sandbox without changing host associations; see [Windows Sandbox checks](https://github.com/MrMaxie/dynamicbrowser/blob/master/scripts/sandbox/README.md) for the separate Settings/consent checklist. `just package` builds a Windows amd64 ZIP in `bin/` using `VERSION`, the schema, the example, README, and license. It does not include an installed `config.yaml` or publish a release.

[Windows CI](https://github.com/MrMaxie/dynamicbrowser/actions/workflows/windows.yml) runs on pushes and pull requests to `master`, or manually. It uses `mise.toml`, runs `just check`, race tests and `go vet`, then keeps the Windows ZIP and SHA256 checksum as a run artifact for 14 days. It does not publish releases or run the Sandbox/Settings checklist.

Tool versions are pinned in `mise.toml`; Node.js is used by the OpenSpec CLI. In GoLand, open this directory as a Go project and use the pinned Go SDK. Application code lives in `cmd/` and `internal/`; Windows is the supported platform for default-browser setup.

## Project sources

- `VERSION`: application version embedded at build time. Use Semantic Versioning; update it for each release.
- `todo.txt`: quick task intake.
- `openspec/specs/`: accepted product and engineering requirements.
- `openspec/changes/`: proposed changes and archived changes.
- `arcantry.toml`: source configuration.

Licensed under the [Apache License 2.0](LICENSE).
