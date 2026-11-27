## MODIFIED Requirements

### Requirement: Configuration format

The application SHALL read one YAML mapping from `config.yaml` independently of the working directory. On Windows the file SHALL remain next to the executable. On Linux it SHALL be under `$XDG_CONFIG_HOME/dynamicbrowser/`, defaulting to `$HOME/.config/dynamicbrowser/` when the XDG setting is unset or invalid. The local schema SHALL be adjacent to the configuration. Both platforms SHALL support optional top-level `default`, `autostart`, `browsers` and `rules` fields. `default` SHALL be a browser configuration name. Each named browser SHALL define `exe` and MAY define `args` as an array of strings. Executable paths containing directory components SHALL resolve relative to the application executable's directory when not absolute; bare names MAY also resolve through `PATH`. Browser arguments and destination URLs SHALL be passed directly to the executable without shell interpretation. A missing configuration SHALL be created without inventing executable paths.

#### Scenario: Launch a browser profile

- **GIVEN** a named browser defines `exe` and `args: ['--profile-directory=ExampleProfile']`
- **WHEN** that browser is selected for a URL
- **THEN** it receives the configured arguments followed by the URL as a separate argument

#### Scenario: Use per-user configuration with a read-only Linux installation

- **GIVEN** the Linux executable is in a read-only installation directory
- **WHEN** the application starts with a writable XDG configuration directory and no configuration yet
- **THEN** it creates `config.yaml` and `config.schema.json` in the user configuration directory
- **AND** it does not write next to the executable

## ADDED Requirements

### Requirement: Unavailable source metadata on Linux

Linux routing SHALL use the direct launcher executable name for `source.process` when available and not a recognized intermediary. Known desktop brokers, openers, and shells SHALL NOT stand in for the originating application. A source condition SHALL NOT match an unknown value, even with a wildcard. The same source matching policy SHALL apply with and without a resident tray. Linux SHALL support the optional boolean configuration keys `window-as-process` and `active-window-fallback`, each defaulting to `true`. Windows SHALL ignore both options and preserve its native process and window-title matching.

When `window-as-process` is enabled, ordinary `source.window` conditions SHALL match the known launcher process name on Linux, not a window title. When `active-window-fallback` is enabled and the launcher is unknown, the router SHALL try the active window's class on Hyprland only if no ordinary rule matches and no tray override applies. The class SHALL be used only to match `source.window`, not to supply `source.process`. If the class cannot be obtained or no rule matches it, the configured default SHALL apply. Disabling either option SHALL NOT disable the other.

#### Scenario: Open a desktop link without source metadata

- **GIVEN** a rule requires a non-empty source process and a later rule matches the URL domain
- **WHEN** Linux desktop activation supplies only the URL
- **THEN** the source-specific rule does not match
- **AND** the domain rule can route the link

#### Scenario: Ignore a known intermediary

- **GIVEN** a source rule uses the `*` process pattern
- **WHEN** Linux can identify only an opener or desktop broker
- **THEN** the source rule does not match, whether routing is on demand or handed to the tray

#### Scenario: Use the active class only as a last resort

- **GIVEN** a `source.window` rule matches the active Hyprland window class
- **WHEN** the launcher is unknown and no ordinary rule matches
- **THEN** the window rule can select the browser unless `active-window-fallback` is `false`
- **AND** an ordinary URL rule takes precedence whenever it matches

### Requirement: Opt-in routing diagnostics

When `DYNAMICBROWSER_TRACE_ROUTING=1` is set on the process performing routing, it SHALL write the known process name or `unknown`, any active window class consulted, the decision (rule number, default, override, or rejection), and the selected browser label to standard error. The trace SHALL be off by default and SHALL NOT include URLs, window titles, or browser arguments. For a request handed to a resident tray, tracing is controlled by the tray's environment.

#### Scenario: Diagnose a forwarded link

- **GIVEN** routing diagnostics are enabled on a resident tray
- **WHEN** a request is forwarded to it and a rule selects a browser
- **THEN** the tray reports the process, rule number, and browser label without logging the URL
