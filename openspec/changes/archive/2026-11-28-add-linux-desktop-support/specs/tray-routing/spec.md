## MODIFIED Requirements

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

### Requirement: Edit configuration

`Edit configuration` SHALL open the active `config.yaml` with the operating system's default application for YAML files, passing the path as data rather than shell code.

#### Scenario: Edit the current file

- **WHEN** `Edit configuration` is selected
- **THEN** the active configuration opens in its associated editor: executable-adjacent on Windows, per-user XDG on Linux
