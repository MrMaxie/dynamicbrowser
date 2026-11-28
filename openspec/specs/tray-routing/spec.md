# tray-routing Specification

## Purpose

Define the singleton tray, temporary browser selection, configuration editing and sign-in startup.
## Requirements
### Requirement: Optional resident tray

Invoking the application without URL arguments SHALL open a singleton tray on supported desktops, without a terminal window when launched from the Windows executable or the Linux desktop entry. Further ordinary tray invocations SHALL exit silently; `--force` SHALL stop the existing tray before replacing it. Windows SHALL preserve left-click menu activation without a notification. Linux SHALL use the desktop's native tray menu interaction and SHALL report that tray mode is unavailable when the session provides no supported tray host. On supported Linux desktops, the tray SHALL provide the same menu actions, session-only override, live configuration reload, singleton behavior, and opt-in sign-in startup as on Windows. Running the tray is optional for users, but tray support is required for the Linux target. Tray presence SHALL not be required for URL routing.

#### Scenario: Replace a resident tray

- **GIVEN** one tray instance is running
- **WHEN** the application is invoked with `--force`
- **THEN** the old tray closes before a new singleton tray starts

#### Scenario: Use the Linux tray to control routing

- **GIVEN** a supported Linux desktop with a tray host and a configured alternate browser profile
- **WHEN** the user launches dynamicbrowser without a URL and selects that profile from the visible tray menu
- **THEN** the next desktop link opens in that profile
- **AND** selecting `Auto` restores rule-based routing
- **AND** closing the tray clears the override without disabling subsequent URL routing

#### Scenario: Route without a Linux tray host

- **GIVEN** the Linux session has no supported tray host
- **WHEN** the user requests tray mode
- **THEN** the application reports that the tray is unavailable instead of remaining as an invisible resident process
- **AND** a later HTTP or HTTPS invocation still routes without a tray

### Requirement: Tray menu

The English-language menu SHALL contain `Auto`, a separator, available browser configurations in YAML declaration order, a separator, `Edit configuration`, and `Close tray`. Browser selection rows SHALL reserve a checkmark column and show a checkmark only for the current selection. If no valid browsers are configured, the browser list SHALL be empty. There SHALL be no configuration-status row or unrelated controls.

#### Scenario: Inspect automatic routing

- **WHEN** the tray starts
- **THEN** `Auto` is selected
- **AND** available browser names appear as selectable rows

### Requirement: Session-only override

Selecting a browser SHALL force that browser for incoming URLs while the tray remains open. Selecting `Auto` SHALL restore configured rules and default selection. The selection SHALL exist only in memory and SHALL not rewrite the configuration. Closing or restarting the tray SHALL clear it. Removing the selected browser in a configuration reload SHALL return the selection to `Auto`.

#### Scenario: Clear an override by closing the tray

- **GIVEN** a browser is temporarily selected
- **WHEN** `Close tray` is chosen
- **THEN** the resident tray exits
- **AND** a later URL invocation uses automatic routing

### Requirement: Edit configuration

`Edit configuration` SHALL open the active `config.yaml` with the operating system's default application for YAML files, passing the path as data rather than shell code.

#### Scenario: Edit the current file

- **WHEN** `Edit configuration` is selected
- **THEN** the active configuration opens in its associated editor: executable-adjacent on Windows, per-user XDG on Linux

### Requirement: Tray autostart

The optional boolean `autostart` SHALL control whether the tray opens at user sign-in. It SHALL not control whether on-demand routing works and SHALL not start a tray merely because a URL was received. Enabling or disabling it SHALL update only this application's user-level startup registration through a platform adapter.

#### Scenario: Disable tray autostart

- **GIVEN** `autostart` is false
- **WHEN** a URL invocation is received without a running tray
- **THEN** the URL still opens
- **AND** no tray is started by that invocation
