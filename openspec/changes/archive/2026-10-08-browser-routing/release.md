---
impact: minor
visibility: public
components:
  - browser-routing
  - tray-routing
  - browser-registration
  - application-cli
---

## Added

### Route links to named browser profiles

Open HTTP/HTTPS URLs using ordered source-application and URL rules, with glob or `regex:` patterns and an explicit fallback. Routing works without a resident tray.

### Temporarily select a profile from the tray

Use `Auto` or a named browser configuration for the current tray session. Edit configuration from the menu, optionally start the tray at sign-in, and replace a running tray with `--force`.

### Set up and undo the default browser on Windows

Use explicit registration and restoration commands with Windows Default Apps. Previous HTTP/HTTPS choices are saved separately, and setup verifies the resulting associations. Incomplete restoration or an unexpected change to an independent protocol reports an error and preserves recovery state.

## Changed

### Configure routing with executable-adjacent YAML

Use `config.yaml` and a local editor schema beside the executable. Valid edits reload while the tray is running; invalid edits keep the last valid configuration. Browser arguments are string arrays passed directly to the executable.

## Removed

### Startup greeting

Starting without URLs now opens the singleton tray rather than printing a greeting. `--version` remains available without resident startup or default-browser changes.
