# dynamicbrowser

A Windows web-link router that opens URLs in named browser profiles using URL and source-application rules. An optional tray lets you temporarily choose one profile instead of automatic routing.

## Set up on Windows

You do not need developer tools to use a packaged build. Any browser that accepts a URL on its command line can be configured; Brave, Chrome, and Firefox are examples below.

### 1. Download and unpack

Open [Windows CI](https://github.com/MrMaxie/dynamicbrowser/actions/workflows/windows.yml), choose a successful run for `master`, and download the `dynamicbrowser-windows-amd64` artifact under **Artifacts** (GitHub sign-in required). Artifacts expire after 14 days. Unpack the download, then unpack the application ZIP inside it into a permanent folder. Builds are currently distributed as CI artifacts, not published releases.

Keep the folder in place after registering dynamicbrowser as your default browser. For the commands below, open that folder in PowerShell (right-click inside it and choose **Open in Terminal**).

### 2. Configure your browsers

Copy `config.example.yaml` to `config.yaml` beside `dynamicbrowser.exe` and open it in a text editor. You can replace its contents with this example:

```yaml
default: brave
browsers:
  brave:
    exe: 'C:/Program Files/BraveSoftware/Brave-Browser/Application/brave.exe'
    args: ['--profile-directory=Default']
  chrome:
    exe: 'C:/Program Files/Google/Chrome/Application/chrome.exe'
    args: ['--profile-directory=Default']
  firefox:
    exe: 'C:/Program Files/Mozilla Firefox/firefox.exe'
    args: ['-P', 'default-release']
rules:
  - browser: chrome
    target:
      domain: 'work.example.test'
  - browser: firefox
    target:
      domain: 'docs.example.test'
```

These are sample installation paths and profiles, not detected settings. Keep only browsers you have installed, correct each `exe` path, and set `default` to one of the names under `browsers`. The names are your own labels, not browser identifiers. Replace the example domains with sites you want to route, or remove `rules` to send every link to your default entry.

To find the right executable and profile:

- **Brave:** open `brave://version` in the profile you want. Copy **Executable Path** into `exe`; use the last folder from **Profile Path**, such as `Default` or `Profile 1`, in `--profile-directory=...`.
- **Chrome:** do the same at `chrome://version`. Use the profile folder name, not its visible account name.
- **Firefox:** find `firefox.exe` through the shortcut's **Properties → Target** (copy only the executable path). Open `about:profiles` and use the profile's name in `['-P', 'your-profile-name']`, not its folder name. See [Firefox profile management](https://support.mozilla.org/en-US/kb/profile-manager-create-remove-switch-firefox-profiles). Check routing both with Firefox closed and already running before relying on a multi-profile setup.

Use forward slashes in YAML paths as shown above. To use a browser's normal startup profile without selecting one explicitly, omit `args`. Multiple entries can point to the same executable with different profile arguments.

### 3. Test before changing your default browser

In PowerShell, run:

```powershell
.\dynamicbrowser.exe 'https://example.com'
```

The browser selected by `default` should open. Test your rule domains too, and confirm that the expected profile opens. If you see an error dialog, check the executable path, profile arguments, and browser names in `config.yaml` before continuing.

### 4. Make it your default link handler

```powershell
.\dynamicbrowser.exe --register
```

In Windows Default Apps, select dynamicbrowser for **HTTP** and **HTTPS**, then confirm the application's dialog. Links opened from other applications will now follow your configuration; the tray does not need to be running.

To temporarily choose a browser instead of following rules, double-click `dynamicbrowser.exe` and select an entry from its tray menu. Choose `Auto` to return to rules. Add `autostart: true` to `config.yaml` if you want this tray at sign-in.

To undo default-browser setup:

```powershell
.\dynamicbrowser.exe --restore
```

Follow the Windows prompts to restore your previous HTTP/HTTPS choices before deleting or moving the application folder.

## Tray behavior

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
.\dynamicbrowser.exe --register
.\dynamicbrowser.exe --restore
```

`--register` makes dynamicbrowser available and opens Windows Default Apps when selection is needed. Choose it for both HTTP and HTTPS, then confirm the application's dialog. Normal startup never changes your default browser.

`--restore` guides you back to the separately saved HTTP/HTTPS choices and removes registration only after verification. A newer independent choice is not a restoration target. Windows Settings can change both protocols from a single picker; if it also changes an independent choice, dynamicbrowser reports an error and retains recovery state. It does not bypass protected Windows associations.

Registration and restoration cannot be combined with URLs, each other, or `--force`. Cancelling setup returns exit status 2; an error returns 1. `--version` prints the build version without opening the tray or performing setup.

## Development

To build from source on Windows, install [mise](https://mise.jdx.dev/), clone this repository, and run these commands from its root:

```powershell
mise trust
mise install
mise exec -- just build
```

The executable is created in `bin/`. Follow the setup instructions above using that folder: copy `config.example.yaml` from the repository root to `bin/config.yaml`, customize it, then run commands from `bin/`. Configuration is always loaded beside the executable, regardless of the working directory.

Developer checks and packaging:

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
